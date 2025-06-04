package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"math"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/threatflux/libgo/internal/api"
	"github.com/threatflux/libgo/internal/api/handlers"
	"github.com/threatflux/libgo/internal/auth/jwt"
	"github.com/threatflux/libgo/internal/auth/user"
	"github.com/threatflux/libgo/internal/compute"
	"github.com/threatflux/libgo/internal/config"
	"github.com/threatflux/libgo/internal/database"
	"github.com/threatflux/libgo/internal/docker"
	"github.com/threatflux/libgo/internal/docker/container"
	"github.com/threatflux/libgo/internal/export"
	"github.com/threatflux/libgo/internal/export/formats/ova"
	"github.com/threatflux/libgo/internal/health"
	"github.com/threatflux/libgo/internal/libvirt/connection"
	"github.com/threatflux/libgo/internal/libvirt/domain"
	"github.com/threatflux/libgo/internal/libvirt/network"
	"github.com/threatflux/libgo/internal/libvirt/storage"
	"github.com/threatflux/libgo/internal/metrics"
	"github.com/threatflux/libgo/internal/middleware"
	"github.com/threatflux/libgo/internal/middleware/auth"
	"github.com/threatflux/libgo/internal/middleware/logging"
	"github.com/threatflux/libgo/internal/middleware/recovery"
	vmmodels "github.com/threatflux/libgo/internal/models/vm"
	"github.com/threatflux/libgo/internal/ovs"
	"github.com/threatflux/libgo/internal/vm"
	"github.com/threatflux/libgo/internal/vm/cloudinit"
	"github.com/threatflux/libgo/internal/vm/template"
	loggerPkg "github.com/threatflux/libgo/pkg/logger"
	"github.com/threatflux/libgo/pkg/utils/exec"
	"github.com/threatflux/libgo/pkg/utils/xmlutils"
)

// Build information.
var (
	version   string = "dev"
	commit    string = "none"
	buildDate string = "unknown"
)

func main() {
	// Parse command-line flags
	configPath := flag.String("config", "configs/config.yaml", "Path to configuration file")
	showVersion := flag.Bool("version", false, "Show version information")
	flag.Parse()

	// Show version and exit if requested
	if *showVersion {
		fmt.Printf("LibGo KVM API %s (commit %s) built on %s\n", version, commit, buildDate)
		os.Exit(0)
	}

	// Initialize configuration
	cfg, err := initConfig(*configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load configuration: %v\n", err)
		os.Exit(1)
	}

	// Initialize logger
	log, err := initLogger(cfg.Logging)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to initialize logger: %v\n", err)
		os.Exit(1)
	}

	// Log startup information
	log.Info("Starting LibGo KVM API",
		loggerPkg.String("version", version),
		loggerPkg.String("commit", commit),
		loggerPkg.String("buildDate", buildDate))

	// Initialize libvirt connection manager
	connManager, err := initLibvirt(cfg.Libvirt, log)
	if err != nil {
		log.Fatal("Failed to initialize libvirt connection", loggerPkg.Error(err))
	}
	defer connManager.Close()

	// Create context for dependency setup
	ctx := context.Background()

	// Initialize components
	components, err := initComponents(ctx, cfg, connManager, log)
	if err != nil {
		log.Error("Failed to initialize components", loggerPkg.Error(err))
		return
	}

	// Ensure storage pool exists
	if poolErr := ensureStoragePool(ctx, components.PoolManager, cfg, log); poolErr != nil {
		log.Error("Failed to ensure storage pool", loggerPkg.Error(poolErr))
		return
	}

	// Initialize default users if configured
	log.Info("Default users configuration",
		loggerPkg.Int("count", len(cfg.Auth.DefaultUsers)),
		loggerPkg.String("config_path", *configPath))

	if len(cfg.Auth.DefaultUsers) > 0 {
		log.Info("Initializing default users from config",
			loggerPkg.Int("count", len(cfg.Auth.DefaultUsers)))

		for i, u := range cfg.Auth.DefaultUsers {
			log.Info("Default user config",
				loggerPkg.Int("index", i),
				loggerPkg.String("username", u.Username))
		}

		err = initDefaultUsers(ctx, components.UserService, cfg.Auth.DefaultUsers, log)
		if err != nil {
			log.Error("Failed to initialize default users", loggerPkg.Error(err))
			return
		}
	} else {
		log.Warn("No default users configured")
	}

	// Initialize health checker
	healthChecker := initHealthChecker(components, version, buildDate, log)

	// Initialize API server
	server := api.NewServer(cfg.Server, log)

	// Setup API routes
	setupRoutes(server, components, healthChecker, cfg, log)

	// Setup signal handler for graceful shutdown
	stopCh := setupSignalHandler(server, log)

	// Start the server
	log.Info("Starting HTTP server",
		loggerPkg.String("host", cfg.Server.Host),
		loggerPkg.Int("port", cfg.Server.Port))

	if err := server.Start(); err != nil {
		log.Fatal("Failed to start server", loggerPkg.Error(err))
	}

	// Wait for shutdown signal
	<-stopCh
	log.Info("Shutting down gracefully")
}

// initConfig initializes configuration.
func initConfig(configPath string) (*config.Config, error) {
	// Create config loader
	loader := config.NewYAMLLoader(configPath)

	// Load configuration
	cfg := &config.Config{}
	if err := loader.Load(cfg); err != nil {
		return nil, fmt.Errorf("loading config: %w", err)
	}

	// Apply environment variable overrides
	if err := loader.LoadWithOverrides(cfg); err != nil {
		return nil, fmt.Errorf("applying config overrides: %w", err)
	}

	// Validate configuration
	if err := config.Validate(cfg); err != nil {
		return nil, fmt.Errorf("validating config: %w", err)
	}

	return cfg, nil
}

// initLogger initializes logger.
func initLogger(config config.LoggingConfig) (loggerPkg.Logger, error) {
	// Create zap logger based on configuration
	log, err := loggerPkg.NewZapLogger(config)
	if err != nil {
		return nil, fmt.Errorf("creating logger: %w", err)
	}

	return log, nil
}

// initLibvirt initializes libvirt connections.
func initLibvirt(config config.LibvirtConfig, logger loggerPkg.Logger) (connection.Manager, error) {
	// Create connection manager
	connManager, err := connection.NewConnectionManager(config, logger)
	if err != nil {
		return nil, fmt.Errorf("creating libvirt connection manager: %w", err)
	}

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	conn, err := connManager.Connect(ctx)
	if err != nil {
		return nil, fmt.Errorf("connecting to libvirt: %w", err)
	}
	if releaseErr := connManager.Release(conn); releaseErr != nil {
		logger.Warn("Failed to release test connection", loggerPkg.Error(releaseErr))
	}

	return connManager, nil
}

// ComponentDependencies holds all the component dependencies.
type ComponentDependencies struct {
	// Connection managers
	ConnManager    connection.Manager
	DomainManager  domain.Manager
	StorageManager storage.VolumeManager
	PoolManager    storage.PoolManager
	NetworkManager network.Manager
	OVSManager     ovs.Manager

	// VM related
	TemplateManager  template.Manager
	CloudInitManager cloudinit.Manager
	VMManager        vm.Manager

	// Docker related
	DockerManager docker.Manager

	// Unified compute
	ComputeManager compute.Manager

	// Export
	ExportManager export.Manager

	// Authentication
	UserService  user.Service
	JWTGenerator jwt.Generator
	JWTValidator jwt.Validator

	// Middleware
	JWTMiddleware      *auth.JWTMiddleware
	RoleMiddleware     *auth.RoleMiddleware
	RecoveryMiddleware func(http.ResponseWriter, *http.Request, func(http.ResponseWriter, *http.Request))
	LoggingMiddleware  func(http.ResponseWriter, *http.Request, func(http.ResponseWriter, *http.Request))

	// Metrics
	MetricsCollector metrics.Collector
}

// initComponents initializes all application components.
func initComponents(ctx context.Context, cfg *config.Config, connManager connection.Manager, log loggerPkg.Logger) (*ComponentDependencies, error) {
	components := &ComponentDependencies{
		ConnManager: connManager,
	}

	// Initialize libvirt components
	if err := initLibvirtComponents(components, cfg, connManager, log); err != nil {
		return nil, fmt.Errorf("initializing libvirt components: %w", err)
	}

	// Initialize VM components  
	if err := initVMComponents(components, cfg, log); err != nil {
		return nil, fmt.Errorf("initializing VM components: %w", err)
	}

	// Initialize authentication components
	if err := initAuthComponents(components, cfg, log); err != nil {
		return nil, fmt.Errorf("initializing auth components: %w", err)
	}

	// Initialize Docker components if enabled
	if err := initDockerComponents(components, cfg, log); err != nil {
		return nil, fmt.Errorf("initializing Docker components: %w", err)
	}

	// Initialize unified compute manager
	if err := initComputeManager(components, cfg, log); err != nil {
		return nil, fmt.Errorf("initializing compute manager: %w", err)
	}

	// Initialize metrics
	if err := initMetricsComponents(ctx, components, log); err != nil {
		return nil, fmt.Errorf("initializing metrics: %w", err)
	}

	return components, nil
}

// initLibvirtComponents initializes libvirt-related components.
func initLibvirtComponents(components *ComponentDependencies, cfg *config.Config, connManager connection.Manager, log loggerPkg.Logger) error {
	// Initialize XML builder for domain
	domainXMLLoader, err := xmlutils.NewTemplateLoader(filepath.Join(cfg.TemplatesPath, "domain"))
	if err != nil {
		return fmt.Errorf("creating domain template loader: %w", err)
	}
	domainXMLBuilder := domain.NewTemplateXMLBuilder(domainXMLLoader, log)

	// Initialize domain manager
	components.DomainManager = domain.NewDomainManager(connManager, domainXMLBuilder, log)

	// Initialize storage components
	storageXMLLoader, err := xmlutils.NewTemplateLoader(filepath.Join(cfg.TemplatesPath, "storage"))
	if err != nil {
		return nil, fmt.Errorf("creating storage template loader: %w", err)
	}
	storageXMLBuilder := storage.NewTemplateXMLBuilder(storageXMLLoader, log)

	components.PoolManager = storage.NewLibvirtPoolManager(connManager, storageXMLBuilder, log)
	components.StorageManager = storage.NewLibvirtVolumeManager(connManager, components.PoolManager, storageXMLBuilder, log)

	// Initialize network components
	networkXMLLoader, err := xmlutils.NewTemplateLoader(filepath.Join(cfg.TemplatesPath, "network"))
	if err != nil {
		return nil, fmt.Errorf("creating network template loader: %w", err)
	}
	networkXMLBuilder := network.TemplateXMLBuilderWithLoader(networkXMLLoader, log)

	components.NetworkManager = network.NewLibvirtNetworkManager(connManager, networkXMLBuilder, log)

	// Initialize OVS manager with sudo wrapper
	commandExecutor := &exec.DefaultCommandExecutor{}
	sudoExecutor := ovs.NewSudoExecutor(commandExecutor)
	components.OVSManager = ovs.NewOVSManager(sudoExecutor, log)

	// Initialize cloud-init components
	cloudInitTemplateLoader, err := xmlutils.NewTemplateLoader(filepath.Join(cfg.TemplatesPath, "cloudinit"))
	if err != nil {
		return nil, fmt.Errorf("creating cloud-init template loader: %w", err)
	}
	components.CloudInitManager, err = cloudinit.NewCloudInitGenerator(cloudInitTemplateLoader.GetTemplatePath(), log)
	if err != nil {
		return nil, fmt.Errorf("creating cloud-init generator: %w", err)
	}

	// Initialize VM template manager
	components.TemplateManager, err = template.NewTemplateManager(cfg.TemplatesPath, log)
	if err != nil {
		return nil, fmt.Errorf("creating template manager: %w", err)
	}

	// Initialize export manager with appropriate converters
	_, err = ova.NewOVFTemplateGenerator(log) // Initialize but we'll use it later in a separate PR
	if err != nil {
		return nil, fmt.Errorf("creating OVF template generator: %w", err)
	}

	components.ExportManager, err = export.NewExportManager(
		components.StorageManager,
		components.DomainManager,
		cfg.Export.OutputDir,
		log,
	)
	if err != nil {
		return nil, fmt.Errorf("creating export manager: %w", err)
	}

	// Initialize VM manager
	vmConfig := vm.Config{
		StoragePoolName: cfg.Libvirt.PoolName,
		NetworkName:     cfg.Libvirt.NetworkName,
		WorkDir:         filepath.Join(cfg.Export.TempDir, "vms"),
		CloudInitDir:    filepath.Join(cfg.Export.TempDir, "cloudinit"),
	}

	components.VMManager = vm.NewVMManager(
		components.DomainManager,
		components.StorageManager,
		components.NetworkManager,
		components.TemplateManager,
		components.CloudInitManager,
		vmConfig,
		log,
	)

	// Initialize database connection
	db, err := database.NewConnection(cfg.Database)
	if err != nil {
		return nil, fmt.Errorf("initializing database connection: %w", err)
	}

	// Initialize authentication components
	components.UserService, err = user.NewGormUserService(db, log)
	if err != nil {
		return nil, fmt.Errorf("initializing user service: %w", err)
	}
	components.JWTGenerator = jwt.NewJWTGenerator(cfg.Auth)
	components.JWTValidator = jwt.NewJWTValidator(cfg.Auth)

	// Initialize middleware
	components.JWTMiddleware = auth.NewJWTMiddleware(components.JWTValidator, components.UserService, log)
	components.RoleMiddleware = auth.NewRoleMiddleware(components.UserService, log)
	components.RecoveryMiddleware = recovery.RecoveryMiddleware(log)
	components.LoggingMiddleware = logging.RequestLoggerMiddleware(log)

	// Initialize Docker manager if enabled
	if cfg.Docker.Enabled {
		log.Info("Initializing Docker client manager")

		dockerOpts := []docker.ClientOption{
			docker.WithHost(cfg.Docker.Host),
			docker.WithAPIVersion(cfg.Docker.APIVersion),
			docker.WithTLSVerify(cfg.Docker.TLSVerify),
			docker.WithRequestTimeout(cfg.Docker.RequestTimeout),
			docker.WithRetry(cfg.Docker.MaxRetries, cfg.Docker.RetryDelay),
			docker.WithLogger(log),
		}

		if cfg.Docker.TLSVerify {
			dockerOpts = append(dockerOpts, docker.WithTLSConfig(
				cfg.Docker.TLSCertPath,
				cfg.Docker.TLSKeyPath,
				cfg.Docker.TLSCAPath,
			))
		}

		components.DockerManager, err = docker.NewManager(dockerOpts...)
		if err != nil {
			return nil, fmt.Errorf("creating Docker manager: %w", err)
		}

		log.Info("Docker client manager initialized successfully")
	} else {
		log.Info("Docker support disabled in configuration")
	}

	// Initialize unified compute manager
	log.Info("Initializing unified compute manager")

	computeConfig := compute.ManagerConfig{
		DefaultBackend:      compute.ComputeBackend(cfg.Compute.DefaultBackend),
		AllowMixedWorkloads: cfg.Compute.AllowMixedDeployments,
		ResourceLimits:      convertConfigResourceLimits(cfg.Compute.ResourceLimits),
		HealthCheckInterval: cfg.Compute.HealthCheckInterval,
		MetricsInterval:     cfg.Compute.MetricsCollectionInterval,
		EnableQuotas:        true, // Enable quotas by default
	}

	components.ComputeManager = compute.NewComputeManager(computeConfig, log)

	// Register KVM backend through VM manager wrapper
	kvmBackend := NewKVMBackendAdapter(components.VMManager, log)
	if concreteManager, ok := components.ComputeManager.(*compute.ComputeManager); ok {
		if kvmErr := concreteManager.RegisterBackend(compute.BackendKVM, kvmBackend); kvmErr != nil {
			return nil, fmt.Errorf("registering KVM backend: %w", kvmErr)
		}
	} else {
		return nil, fmt.Errorf("compute manager is not a concrete ComputeManager")
	}

	// Register Docker backend if enabled
	if cfg.Docker.Enabled && components.DockerManager != nil {
		dockerBackend := docker.NewBackendService(components.DockerManager, log)
		if concreteManager, ok := components.ComputeManager.(*compute.ComputeManager); ok {
			if dockerErr := concreteManager.RegisterBackend(compute.BackendDocker, dockerBackend); dockerErr != nil {
				return nil, fmt.Errorf("registering Docker backend: %w", dockerErr)
			}
			log.Info("Docker backend registered with compute manager")
		}
	}

	log.Info("Unified compute manager initialized successfully")

	// Initialize metrics
	metricsDeps := map[string]interface{}{
		"vm_manager":      components.VMManager,
		"export_manager":  components.ExportManager,
		"compute_manager": components.ComputeManager,
	}
	if components.DockerManager != nil {
		metricsDeps["docker_manager"] = components.DockerManager
	}

	components.MetricsCollector, err = metrics.NewCollector("prometheus", ctx, metricsDeps, log)
	if err != nil {
		return nil, fmt.Errorf("creating metrics collector: %w", err)
	}

	return components, nil
}

// initHealthChecker initializes the health checker.
func initHealthChecker(components *ComponentDependencies, version, buildDate string, log loggerPkg.Logger) *health.Checker {
	// Create health checker
	checker := health.NewChecker(version, buildDate)

	// Add libvirt connection health check
	checker.AddCheck(health.NewLibvirtConnectionCheck(components.ConnManager, log))

	// Add storage pool health check
	checker.AddCheck(health.NewStoragePoolCheck(components.PoolManager, "default", log))

	// Add network health check
	checker.AddCheck(health.NewNetworkCheck(components.NetworkManager, "default", log))

	return checker
}

// setupRoutes configures API routes.
func setupRoutes(server *api.Server, components *ComponentDependencies, healthChecker *health.Checker, cfg *config.Config, log loggerPkg.Logger) {
	// Create API handlers
	vmHandler := handlers.NewVMHandler(components.VMManager, log)
	exportHandler := handlers.NewExportHandler(components.VMManager, components.ExportManager, log)
	authHandler := handlers.NewAuthHandler(components.UserService, components.JWTGenerator, log, cfg.Auth.TokenExpiration)
	healthHandler := handlers.NewHealthHandler(healthChecker, log)
	metricsHandler := handlers.NewMetricsHandler(components.MetricsCollector, log)

	// Create unified compute handler
	computeHandler := handlers.NewComputeHandler(components.ComputeManager, log)

	// Create network handlers
	networkHandlers := &api.NetworkHandlers{
		List:   handlers.NewNetworkListHandler(components.NetworkManager, log),
		Create: handlers.NewNetworkCreateHandler(components.NetworkManager, log),
		Get:    handlers.NewNetworkGetHandler(components.NetworkManager, log),
		Update: handlers.NewNetworkUpdateHandler(components.NetworkManager, log),
		Delete: handlers.NewNetworkDeleteHandler(components.NetworkManager, log),
		Start:  handlers.NewNetworkStartHandler(components.NetworkManager, log),
		Stop:   handlers.NewNetworkStopHandler(components.NetworkManager, log),

		// Bridge network handlers
		ListBridges:  handlers.NewBridgeNetworkListHandler(components.NetworkManager, log),
		CreateBridge: handlers.NewBridgeNetworkCreateHandler(components.NetworkManager, log),
		GetBridge:    handlers.NewBridgeNetworkGetHandler(components.NetworkManager, log),
		DeleteBridge: handlers.NewBridgeNetworkDeleteHandler(components.NetworkManager, log),
	}

	// Create storage handlers
	storageHandlers := &api.StorageHandlers{
		ListPools:    handlers.NewStorageListHandler(components.PoolManager, log),
		CreatePool:   handlers.NewStorageCreateHandler(components.PoolManager, log),
		GetPool:      handlers.NewStorageGetHandler(components.PoolManager, log),
		DeletePool:   handlers.NewStorageDeleteHandler(components.PoolManager, log),
		StartPool:    handlers.NewStorageStartHandler(components.PoolManager, log),
		StopPool:     handlers.NewStorageStopHandler(components.PoolManager, log),
		ListVolumes:  handlers.NewStorageVolumeListHandler(components.StorageManager, log),
		CreateVolume: handlers.NewStorageVolumeCreateHandler(components.StorageManager, log),
		DeleteVolume: handlers.NewStorageVolumeDeleteHandler(components.StorageManager, log),
		UploadVolume: handlers.NewStorageVolumeUploadHandler(components.StorageManager, log),
	}

	// Create OVS handlers
	ovsHandlers := &api.OVSHandlers{
		CreateBridge: handlers.NewOVSBridgeCreateHandler(components.OVSManager, log),
		ListBridges:  handlers.NewOVSBridgeListHandler(components.OVSManager, log),
		GetBridge:    handlers.NewOVSBridgeGetHandler(components.OVSManager, log),
		DeleteBridge: handlers.NewOVSBridgeDeleteHandler(components.OVSManager, log),
		CreatePort:   handlers.NewOVSPortCreateHandler(components.OVSManager, log),
		ListPorts:    handlers.NewOVSPortListHandler(components.OVSManager, log),
		DeletePort:   handlers.NewOVSPortDeleteHandler(components.OVSManager, log),
		CreateFlow:   handlers.NewOVSFlowCreateHandler(components.OVSManager, log),
	}

	// Create Docker handlers if Docker is enabled
	var dockerHandlers *api.DockerHandlers
	if cfg.Docker.Enabled && components.DockerManager != nil {
		log.Info("Creating Docker API handlers")

		// Create container service
		containerService := container.NewService(components.DockerManager, log)

		// Create container handler
		containerHandler := handlers.NewDockerContainerHandler(containerService, log)

		// Build Docker handlers structure
		dockerHandlers = &api.DockerHandlers{
			CreateContainer:   containerHandler,
			ListContainers:    containerHandler,
			GetContainer:      containerHandler,
			StartContainer:    containerHandler,
			StopContainer:     containerHandler,
			RestartContainer:  containerHandler,
			DeleteContainer:   containerHandler,
			GetContainerLogs:  containerHandler,
			GetContainerStats: containerHandler,
		}

		log.Info("Docker API handlers created successfully")
	} else {
		log.Info("Docker handlers disabled or Docker manager not available")
	}

	// Setup router
	router := server.Router()

	// Add middleware
	router.Use(middleware.RecoveryToGin(components.RecoveryMiddleware))
	router.Use(middleware.LoggingToGin(components.LoggingMiddleware))
	router.Use(metricsHandler.CollectRequestMetrics())

	// Setup API routes
	api.ConfigureRoutes(
		router,
		log,
		components.JWTMiddleware,
		components.RoleMiddleware,
		vmHandler,
		exportHandler,
		authHandler,
		healthHandler,
		metricsHandler,
		computeHandler,
		networkHandlers,
		storageHandlers,
		ovsHandlers,
		dockerHandlers,
		cfg, // Pass the configuration
	)
}

// initDefaultUsers initializes default users from configuration.
func initDefaultUsers(ctx context.Context, userService user.Service, defaultUsers []config.DefaultUser, log loggerPkg.Logger) error {
	log.Info("Initializing default users", loggerPkg.Int("count", len(defaultUsers)))

	// Convert DefaultUser config structs to the format expected by InitializeDefaultUsers
	userConfigs := make([]user.DefaultUserConfig, len(defaultUsers))

	for i, u := range defaultUsers {
		userConfigs[i] = user.DefaultUserConfig{
			Username: u.Username,
			Password: u.Password,
			Email:    u.Email,
			Roles:    u.Roles,
		}
	}

	// Initialize default users
	return userService.InitializeDefaultUsers(ctx, userConfigs)
}

// setupSignalHandler sets up signal handling for graceful shutdown.
func setupSignalHandler(server *api.Server, log loggerPkg.Logger) chan os.Signal {
	stopCh := make(chan os.Signal, 1)
	signal.Notify(stopCh, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)

	go func() {
		<-stopCh
		log.Info("Received shutdown signal")

		// Create shutdown context with timeout
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		// Shutdown the server
		if err := server.Stop(ctx); err != nil {
			log.Error("Error during server shutdown", loggerPkg.Error(err))
		}

		// Signal that shutdown is complete
		close(stopCh)
	}()

	return stopCh
}

// ensureStoragePool ensures the required storage pool exists and is active.
func ensureStoragePool(ctx context.Context, poolManager storage.PoolManager, cfg *config.Config, log loggerPkg.Logger) error {
	poolName, poolPath := resolvePoolConfiguration(cfg, log)

	if err := createPoolDirectory(poolPath); err != nil {
		return err
	}

	if err := poolManager.EnsureExists(ctx, poolName, poolPath); err != nil {
		return fmt.Errorf("failed to ensure storage pool %s exists: %w", poolName, err)
	}

	return copyTemplateImages(cfg.Storage.Templates, poolPath, log)
}

// resolvePoolConfiguration resolves pool name and path from configuration.
func resolvePoolConfiguration(cfg *config.Config, log loggerPkg.Logger) (string, string) {
	poolName := cfg.Storage.DefaultPool
	poolPath := cfg.Storage.PoolPath

	if poolName == "" {
		poolName = cfg.Libvirt.PoolName
		log.Warn("Storage default pool not specified, using libvirt pool name",
			loggerPkg.String("pool", poolName))
	}

	if poolPath == "" {
		poolPath = "/tmp/libgo-storage"
		log.Warn("Storage pool path not specified, using default path",
			loggerPkg.String("path", poolPath))
	}

	return poolName, poolPath
}

// createPoolDirectory creates the storage pool directory.
func createPoolDirectory(poolPath string) error {
	if err := os.MkdirAll(poolPath, 0777); err != nil {
		return fmt.Errorf("failed to create storage path %s: %w", poolPath, err)
	}
	return nil
}

// copyTemplateImages copies template images to the storage pool.
func copyTemplateImages(templates map[string]string, poolPath string, log loggerPkg.Logger) error {
	for templateName, imagePath := range templates {
		if err := copyTemplateImage(templateName, imagePath, poolPath, log); err != nil {
			// Log error but continue with other templates
			log.Warn("Failed to copy template image",
				loggerPkg.String("template", templateName),
				loggerPkg.Error(err))
		}
	}
	return nil
}

// copyTemplateImage copies a single template image to the pool.
func copyTemplateImage(templateName, imagePath, poolPath string, log loggerPkg.Logger) error {
	if _, err := os.Stat(imagePath); os.IsNotExist(err) {
		log.Warn("Template image not found, skipping copy",
			loggerPkg.String("template", templateName),
			loggerPkg.String("path", imagePath))
		return nil
	}

	destPath := filepath.Join(poolPath, filepath.Base(imagePath))
	if _, err := os.Stat(destPath); !os.IsNotExist(err) {
		return nil // Destination already exists
	}

	log.Info("Copying template image to storage pool",
		loggerPkg.String("template", templateName),
		loggerPkg.String("source", imagePath),
		loggerPkg.String("destination", destPath))

	return performImageCopy(imagePath, destPath, templateName, log)
}

// performImageCopy performs the actual file copy operation.
func performImageCopy(imagePath, destPath, templateName string, log loggerPkg.Logger) error {
	src, err := os.Open(imagePath)
	if err != nil {
		return fmt.Errorf("failed to open source file: %w", err)
	}
	defer src.Close()

	dst, err := os.Create(destPath)
	if err != nil {
		return fmt.Errorf("failed to create destination file: %w", err)
	}
	defer dst.Close()

	if err := dst.Chmod(0666); err != nil {
		log.Warn("Failed to set permissions on destination file",
			loggerPkg.String("path", destPath),
			loggerPkg.Error(err))
	}

	if _, err := io.Copy(dst, src); err != nil {
		return fmt.Errorf("failed to copy file: %w", err)
	}

	log.Info("Template image copied successfully",
		loggerPkg.String("template", templateName),
		loggerPkg.String("destination", destPath))

	return nil
}

// convertConfigResourceLimits converts config resource limits to compute resource limits.
func convertConfigResourceLimits(limits config.ResourceLimits) compute.ComputeResources {
	return compute.ComputeResources{
		CPU: compute.CPUResources{
			Cores: float64(limits.MaxCPUCores),
		},
		Memory: compute.MemoryResources{
			Limit: int64(limits.MaxMemoryGB) * 1024 * 1024 * 1024, // Convert GB to bytes
		},
		Storage: compute.StorageResources{
			TotalSpace: int64(limits.MaxStorageGB) * 1024 * 1024 * 1024, // Convert GB to bytes
		},
		Network: compute.NetworkResources{
			BandwidthLimit: int64(limits.MaxNetworkMbps) * 1024 * 1024, // Convert Mbps to bps
		},
	}
}

// NewKVMBackendAdapter creates an adapter that wraps the VM manager to implement the BackendService interface.
func NewKVMBackendAdapter(vmManager vm.Manager, logger loggerPkg.Logger) compute.BackendService {
	return &kvmBackendAdapter{
		vmManager: vmManager,
		logger:    logger,
	}
}

// kvmBackendAdapter adapts the VM manager to the compute backend interface.
type kvmBackendAdapter struct {
	vmManager vm.Manager
	logger    loggerPkg.Logger
}

// Create creates a new KVM instance.
func (a *kvmBackendAdapter) Create(ctx context.Context, req compute.ComputeInstanceRequest) (*compute.ComputeInstance, error) {
	// Convert compute request to VM request
	vmReq := a.convertToVMRequest(req)

	// Create VM
	vmInstance, err := a.vmManager.Create(ctx, vmReq)
	if err != nil {
		return nil, fmt.Errorf("failed to create VM: %w", err)
	}

	// Convert VM to compute instance
	return a.convertFromVM(vmInstance), nil
}

// Get retrieves a KVM instance by ID.
func (a *kvmBackendAdapter) Get(ctx context.Context, id string) (*compute.ComputeInstance, error) {
	vmInstance, err := a.vmManager.Get(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get VM: %w", err)
	}

	return a.convertFromVM(vmInstance), nil
}

// List retrieves KVM instances.
func (a *kvmBackendAdapter) List(ctx context.Context, opts compute.ComputeInstanceListOptions) ([]*compute.ComputeInstance, error) {
	// VM manager List method doesn't take options - just list all VMs
	vmInstances, err := a.vmManager.List(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list VMs: %w", err)
	}

	instances := make([]*compute.ComputeInstance, len(vmInstances))
	for i, vm := range vmInstances {
		instances[i] = a.convertFromVM(vm)
	}

	return instances, nil
}

// Update updates a KVM instance.
func (a *kvmBackendAdapter) Update(ctx context.Context, id string, update compute.ComputeInstanceUpdate) (*compute.ComputeInstance, error) {
	// For now, just return the current instance since VM updates are complex
	return a.Get(ctx, id)
}

// Delete removes a KVM instance.
func (a *kvmBackendAdapter) Delete(ctx context.Context, id string, force bool) error {
	return a.vmManager.Delete(ctx, id)
}

// Start starts a KVM instance.
func (a *kvmBackendAdapter) Start(ctx context.Context, id string) error {
	return a.vmManager.Start(ctx, id)
}

// Stop stops a KVM instance.
func (a *kvmBackendAdapter) Stop(ctx context.Context, id string, force bool) error {
	return a.vmManager.Stop(ctx, id)
}

// Restart restarts a KVM instance.
func (a *kvmBackendAdapter) Restart(ctx context.Context, id string, force bool) error {
	return a.vmManager.Restart(ctx, id)
}

// Pause pauses a KVM instance.
func (a *kvmBackendAdapter) Pause(ctx context.Context, id string) error {
	return fmt.Errorf("pause not implemented for KVM backend")
}

// Unpause unpauses a KVM instance.
func (a *kvmBackendAdapter) Unpause(ctx context.Context, id string) error {
	return fmt.Errorf("unpause not implemented for KVM backend")
}

// GetResourceUsage gets current resource usage for a KVM instance.
func (a *kvmBackendAdapter) GetResourceUsage(ctx context.Context, id string) (*compute.ResourceUsage, error) {
	// This would integrate with the VM manager's resource monitoring
	return nil, fmt.Errorf("resource usage not implemented for KVM backend")
}

// UpdateResourceLimits updates resource limits for a KVM instance.
func (a *kvmBackendAdapter) UpdateResourceLimits(ctx context.Context, id string, resources compute.ComputeResources) error {
	// This would involve updating VM configuration
	return fmt.Errorf("resource limit updates not implemented for KVM backend")
}

// GetBackendInfo returns information about the KVM backend.
func (a *kvmBackendAdapter) GetBackendInfo(ctx context.Context) (*compute.BackendInfo, error) {
	return &compute.BackendInfo{
		Type:           compute.BackendKVM,
		Version:        "libvirt",
		APIVersion:     "1.0",
		Status:         "running",
		Capabilities:   []string{"vms", "snapshots", "migration", "export"},
		SupportedTypes: []compute.ComputeInstanceType{compute.InstanceTypeVM},
		HealthCheck: &compute.HealthStatus{
			Status:    "healthy",
			LastCheck: time.Now(),
		},
	}, nil
}

// ValidateConfig validates a compute instance configuration for KVM.
func (a *kvmBackendAdapter) ValidateConfig(ctx context.Context, config compute.ComputeInstanceConfig) error {
	if config.Image == "" {
		return fmt.Errorf("image is required for KVM VMs")
	}
	return nil
}

// GetBackendType returns the backend type.
func (a *kvmBackendAdapter) GetBackendType() compute.ComputeBackend {
	return compute.BackendKVM
}

// GetSupportedInstanceTypes returns supported instance types.
func (a *kvmBackendAdapter) GetSupportedInstanceTypes() []compute.ComputeInstanceType {
	return []compute.ComputeInstanceType{compute.InstanceTypeVM}
}

// Helper conversion methods.
func (a *kvmBackendAdapter) convertToVMRequest(req compute.ComputeInstanceRequest) vmmodels.VMParams {
	return vmmodels.VMParams{
		Name:     req.Name,
		Template: req.Config.Image, // Use image as template name
		CPU: vmmodels.CPUParams{
			Count: int(req.Resources.CPU.Cores),
		},
		Memory: vmmodels.MemoryParams{
			SizeBytes: func() uint64 {
				if req.Resources.Memory.Limit < 0 {
					return 0
				}
				return uint64(req.Resources.Memory.Limit)
			}(),
		},
		Disk: vmmodels.DiskParams{
			SizeBytes: func() uint64 {
				if req.Resources.Storage.TotalSpace < 0 {
					return 0
				}
				return uint64(req.Resources.Storage.TotalSpace)
			}(),
			Format: vmmodels.DiskFormatQCOW2,
		},
		Network: vmmodels.NetParams{
			Type:   vmmodels.NetworkTypeBridge,
			Source: "default", // Use default network for now
		},
	}
}

func (a *kvmBackendAdapter) convertFromVM(vmInstance *vmmodels.VM) *compute.ComputeInstance {
	state := a.convertVMState(vmInstance.Status)

	return &compute.ComputeInstance{
		ID:      vmInstance.UUID, // Use UUID as ID
		Name:    vmInstance.Name,
		UUID:    vmInstance.UUID,
		Type:    compute.InstanceTypeVM,
		Backend: compute.BackendKVM,
		State:   state,
		Status:  string(vmInstance.Status),
		Config:  compute.ComputeInstanceConfig{
			// Image/template info would need to be tracked separately
		},
		Resources: compute.ComputeResources{
			CPU: compute.CPUResources{
				Cores: float64(vmInstance.CPU.Count),
			},
			Memory: compute.MemoryResources{
				Limit: func() int64 {
				if vmInstance.Memory.SizeBytes > uint64(math.MaxInt64) {
					return math.MaxInt64
				}
				return int64(vmInstance.Memory.SizeBytes)
			}(),
			},
		},
		CreatedAt: vmInstance.CreatedAt,
		BackendData: map[string]interface{}{
			"kvm_uuid": vmInstance.UUID,
			"kvm_name": vmInstance.Name,
		},
	}
}

func (a *kvmBackendAdapter) convertVMState(vmState vmmodels.VMStatus) compute.ComputeInstanceState {
	switch vmState {
	case vmmodels.VMStatusRunning:
		return compute.StateRunning
	case vmmodels.VMStatusStopped, vmmodels.VMStatusShutdown:
		return compute.StateStopped
	case vmmodels.VMStatusPaused:
		return compute.StatePaused
	case vmmodels.VMStatusCrashed:
		return compute.StateError
	default:
		return compute.StateUnknown
	}
}
