package main

import (
	"context"
	"flag"
	"fmt"
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
	"github.com/threatflux/libgo/internal/config"
	"github.com/threatflux/libgo/internal/database"
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
	"github.com/threatflux/libgo/internal/vm"
	"github.com/threatflux/libgo/internal/vm/cloudinit"
	"github.com/threatflux/libgo/internal/vm/template"
	"github.com/threatflux/libgo/pkg/logger"
	"github.com/threatflux/libgo/pkg/utils/xml"
)

// Build information
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
		logger.String("version", version),
		logger.String("commit", commit),
		logger.String("buildDate", buildDate))

	// Initialize libvirt connection manager
	connManager, err := initLibvirt(cfg.Libvirt, log)
	if err != nil {
		log.Fatal("Failed to initialize libvirt connection", logger.Error(err))
	}
	defer connManager.Close()

	// Create context for dependency setup
	ctx := context.Background()

	// Initialize components
	components, err := initComponents(ctx, cfg, connManager, log)
	if err != nil {
		log.Fatal("Failed to initialize components", logger.Error(err))
	}

	// Initialize default users if configured
	log.Info("Default users configuration",
		logger.Int("count", len(cfg.Auth.DefaultUsers)),
		logger.String("config_path", *configPath))

	if len(cfg.Auth.DefaultUsers) > 0 {
		log.Info("Initializing default users from config",
			logger.Int("count", len(cfg.Auth.DefaultUsers)))

		for i, u := range cfg.Auth.DefaultUsers {
			log.Info("Default user config",
				logger.Int("index", i),
				logger.String("username", u.Username))
		}

		err = initDefaultUsers(ctx, components.UserService, cfg.Auth.DefaultUsers, log)
		if err != nil {
			log.Fatal("Failed to initialize default users", logger.Error(err))
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
		logger.String("host", cfg.Server.Host),
		logger.Int("port", cfg.Server.Port))

	if err := server.Start(); err != nil {
		log.Fatal("Failed to start server", logger.Error(err))
	}

	// Wait for shutdown signal
	<-stopCh
	log.Info("Shutting down gracefully")
}

// initConfig initializes configuration
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

// initLogger initializes logger
func initLogger(config config.LoggingConfig) (logger.Logger, error) {
	// Create zap logger based on configuration
	log, err := logger.NewZapLogger(config)
	if err != nil {
		return nil, fmt.Errorf("creating logger: %w", err)
	}

	return log, nil
}

// initLibvirt initializes libvirt connections
func initLibvirt(config config.LibvirtConfig, logger logger.Logger) (connection.Manager, error) {
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
	connManager.Release(conn)

	return connManager, nil
}

// ComponentDependencies holds all the component dependencies
type ComponentDependencies struct {
	// Connection managers
	ConnManager    connection.Manager
	DomainManager  domain.Manager
	StorageManager storage.VolumeManager
	PoolManager    storage.PoolManager
	NetworkManager network.Manager

	// VM related
	TemplateManager  template.Manager
	CloudInitManager cloudinit.Manager
	VMManager        vm.Manager

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

// initComponents initializes all application components
func initComponents(ctx context.Context, cfg *config.Config, connManager connection.Manager, log logger.Logger) (*ComponentDependencies, error) {
	components := &ComponentDependencies{
		ConnManager: connManager,
	}

	// Initialize XML builder for domain
	domainXMLLoader, err := xml.NewTemplateLoader(filepath.Join(cfg.TemplatesPath, "domain"))
	if err != nil {
		return nil, fmt.Errorf("creating domain template loader: %w", err)
	}
	domainXMLBuilder := domain.NewTemplateXMLBuilder(domainXMLLoader, log)

	// Initialize domain manager
	components.DomainManager = domain.NewDomainManager(connManager, domainXMLBuilder, log)

	// Initialize storage components
	storageXMLLoader, err := xml.NewTemplateLoader(filepath.Join(cfg.TemplatesPath, "storage"))
	if err != nil {
		return nil, fmt.Errorf("creating storage template loader: %w", err)
	}
	storageXMLBuilder := storage.NewTemplateXMLBuilder(storageXMLLoader, log)

	components.PoolManager = storage.NewLibvirtPoolManager(connManager, storageXMLBuilder, log)
	components.StorageManager = storage.NewLibvirtVolumeManager(connManager, components.PoolManager, storageXMLBuilder, log)

	// Initialize network components
	networkXMLLoader, err := xml.NewTemplateLoader(filepath.Join(cfg.TemplatesPath, "network"))
	if err != nil {
		return nil, fmt.Errorf("creating network template loader: %w", err)
	}
	networkXMLBuilder := network.TemplateXMLBuilderWithLoader(networkXMLLoader, log)

	components.NetworkManager = network.NewLibvirtNetworkManager(connManager, networkXMLBuilder, log)

	// Initialize cloud-init components
	cloudInitTemplateLoader, err := xml.NewTemplateLoader(filepath.Join(cfg.TemplatesPath, "cloudinit"))
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
		CloudInitDir:    "/home/vtriple/libgo-temp/cloudinit",
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

	// Initialize metrics
	metricsDeps := map[string]interface{}{
		"vm_manager":     components.VMManager,
		"export_manager": components.ExportManager,
	}
	components.MetricsCollector, err = metrics.NewCollector("prometheus", ctx, metricsDeps, log)
	if err != nil {
		return nil, fmt.Errorf("creating metrics collector: %w", err)
	}

	return components, nil
}

// initHealthChecker initializes the health checker
func initHealthChecker(components *ComponentDependencies, version, buildDate string, log logger.Logger) *health.Checker {
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

// setupRoutes configures API routes
func setupRoutes(server *api.Server, components *ComponentDependencies, healthChecker *health.Checker, cfg *config.Config, log logger.Logger) {
	// Create API handlers
	vmHandler := handlers.NewVMHandler(components.VMManager, log)
	exportHandler := handlers.NewExportHandler(components.VMManager, components.ExportManager, log)
	authHandler := handlers.NewAuthHandler(components.UserService, components.JWTGenerator, log, cfg.Auth.TokenExpiration)
	healthHandler := handlers.NewHealthHandler(healthChecker, log)
	metricsHandler := handlers.NewMetricsHandler(components.MetricsCollector, log)

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
	)
}

// initDefaultUsers initializes default users from configuration
func initDefaultUsers(ctx context.Context, userService user.Service, defaultUsers []config.DefaultUser, log logger.Logger) error {
	log.Info("Initializing default users", logger.Int("count", len(defaultUsers)))

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

// setupSignalHandler sets up signal handling for graceful shutdown
func setupSignalHandler(server *api.Server, log logger.Logger) chan os.Signal {
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
			log.Error("Error during server shutdown", logger.Error(err))
		}

		// Signal that shutdown is complete
		close(stopCh)
	}()

	return stopCh
}
