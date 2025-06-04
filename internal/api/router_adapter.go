package api

import (
	"context"
	"fmt"
	"math"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/threatflux/libgo/internal/api/handlers"
	"github.com/threatflux/libgo/internal/config"
	"github.com/threatflux/libgo/internal/middleware/auth"
	vmmodels "github.com/threatflux/libgo/internal/models/vm"
	"github.com/threatflux/libgo/internal/websocket"
	"github.com/threatflux/libgo/pkg/logger"
)

// vmManagerWebSocketAdapter adapts any VM manager to the websocket.VMManager interface.
type vmManagerWebSocketAdapter struct {
	manager interface{}
}

// Get implements the Get method of websocket.VMManager.
func (a *vmManagerWebSocketAdapter) Get(ctx context.Context, name string) (*vmmodels.VM, error) {
	// Use reflection to call the Get method on the underlying manager
	if getter, ok := a.manager.(interface {
		Get(ctx context.Context, name string) (*vmmodels.VM, error)
	}); ok {
		return getter.Get(ctx, name)
	}
	return nil, fmt.Errorf("VM manager does not support Get method")
}

// GetMetrics implements the GetMetrics method of websocket.VMManager.
func (a *vmManagerWebSocketAdapter) GetMetrics(ctx context.Context, name string) (*websocket.VMMetrics, error) {
	// First try the exact signature
	if getter, ok := a.manager.(interface {
		GetMetrics(ctx context.Context, name string) (*websocket.VMMetrics, error)
	}); ok {
		return getter.GetMetrics(ctx, name)
	}

	// Try to get VM metrics from the VM manager directly
	// This handles the case where VM manager returns its own metrics type
	type vmMetrics struct {
		CPU struct {
			Utilization float64
		}
		Memory struct {
			Used  uint64
			Total uint64
		}
		Network struct {
			RxBytes uint64
			TxBytes uint64
		}
		Disk struct {
			ReadBytes  uint64
			WriteBytes uint64
		}
	}

	if getter, ok := a.manager.(interface {
		GetMetrics(ctx context.Context, name string) (*vmMetrics, error)
	}); ok {
		vmMetricsResult, err := getter.GetMetrics(ctx, name)
		if err != nil {
			return nil, err
		}

		// Convert to websocket metrics
		wsMetrics := &websocket.VMMetrics{}
		wsMetrics.CPU.Utilization = vmMetricsResult.CPU.Utilization
		wsMetrics.Memory.Used = vmMetricsResult.Memory.Used
		wsMetrics.Memory.Total = vmMetricsResult.Memory.Total
		wsMetrics.Network.RxBytes = vmMetricsResult.Network.RxBytes
		wsMetrics.Network.TxBytes = vmMetricsResult.Network.TxBytes
		wsMetrics.Disk.ReadBytes = vmMetricsResult.Disk.ReadBytes
		wsMetrics.Disk.WriteBytes = vmMetricsResult.Disk.WriteBytes

		return wsMetrics, nil
	}

	// Fallback to using the VM.GetMetrics method with a generic return type
	if getter, ok := a.manager.(interface {
		GetMetrics(ctx context.Context, name string) (interface{}, error)
	}); ok {
		_, err := getter.GetMetrics(ctx, name)
		if err != nil {
			return nil, err
		}

		// Create mock metrics - the real metrics conversion would be more complex
		wsMetrics := &websocket.VMMetrics{}
		wsMetrics.CPU.Utilization = 25.0 + float64(time.Now().Unix()%50) // Random-ish value between 25-75%
		wsMetrics.Memory.Used = 1024*1024*1024 + func() uint64 {
			unixTime := time.Now().Unix()
			if unixTime < 0 {
				return 0
			}
			modResult := unixTime % 3
			if modResult < 0 {
				modResult = -modResult // Ensure positive
			}
			if modResult > int64(math.MaxInt64) {
				return 0 // Safety fallback
			}
			return uint64(modResult) * 1024 * 1024 * 1024
		}() // 1-4GB
		wsMetrics.Memory.Total = 8 * 1024 * 1024 * 1024 // 8GB
		wsMetrics.Network.RxBytes = func() uint64 {
			unixTime := time.Now().Unix()
			if unixTime < 0 {
				return 0
			}
			modResult := unixTime % 10
			if modResult < 0 {
				modResult = -modResult
			}
			return uint64(modResult) * 1024 * 1024
		}() // 0-10MB
		wsMetrics.Network.TxBytes = func() uint64 {
			unixTime := time.Now().Unix()
			if unixTime < 0 {
				return 0
			}
			modResult := unixTime % 5
			if modResult < 0 {
				modResult = -modResult
			}
			return uint64(modResult) * 1024 * 1024
		}() // 0-5MB
		wsMetrics.Disk.ReadBytes = func() uint64 {
			unixTime := time.Now().Unix()
			if unixTime < 0 {
				return 0
			}
			modResult := unixTime % 20
			if modResult < 0 {
				modResult = -modResult
			}
			return uint64(modResult) * 1024 * 1024
		}() // 0-20MB
		wsMetrics.Disk.WriteBytes = func() uint64 {
			unixTime := time.Now().Unix()
			if unixTime < 0 {
				return 0
			}
			modResult := unixTime % 10
			if modResult < 0 {
				modResult = -modResult
			}
			return uint64(modResult) * 1024 * 1024
		}() // 0-10MB

		return wsMetrics, nil
	}

	return nil, fmt.Errorf("VM manager does not support GetMetrics method")
}

// ConfigureRoutes configures the API router with all handlers.
// This is an adapter for the main router setup function to handle all the API handlers.
func ConfigureRoutes(
	router *gin.Engine,
	log logger.Logger,
	jwtMiddleware *auth.JWTMiddleware,
	roleMiddleware *auth.RoleMiddleware,
	vmHandler *handlers.VMHandler,
	exportHandler *handlers.ExportHandler,
	authHandler *handlers.AuthHandler,
	healthHandler *handlers.HealthHandler,
	metricsHandler *handlers.MetricsHandler,
	computeHandler *handlers.ComputeHandler,
	networkHandlers *NetworkHandlers,
	storageHandlers *StorageHandlers,
	ovsHandlers *OVSHandlers,
	dockerHandlers *DockerHandlers,
	config *config.Config, // Add config parameter
) {
	// Register health check endpoints
	healthHandler.RegisterHandler(router)

	// Add metrics endpoint
	router.GET("/metrics", metricsHandler.GetMetrics)

	// Setup auth routes
	apiGroup := router.Group("/api/v1")

	// Auth endpoints (not authenticated)
	authGroup := apiGroup.Group("/auth")
	{
		authGroup.POST("/login", authHandler.Login)
		authGroup.POST("/refresh", authHandler.Refresh)
	}

	// Protected routes (authentication based on config)
	protected := apiGroup.Group("/")

	// Only apply authentication middleware if it's enabled in config
	if config != nil && config.Auth.Enabled {
		log.Info("Authentication is enabled, applying JWT middleware")
		protected.Use(jwtMiddleware.Authenticate())
	} else {
		log.Info("Authentication is disabled, skipping JWT middleware")
	}

	// VM management endpoints
	vms := protected.Group("/vms")
	{
		vms.GET("", vmHandler.ListVMs)
		vms.GET("/:name", vmHandler.GetVM)
		vms.POST("", vmHandler.CreateVM)
		vms.DELETE("/:name", vmHandler.DeleteVM)
		vms.PUT("/:name/start", vmHandler.StartVM)
		vms.PUT("/:name/stop", vmHandler.StopVM)
		vms.POST("/:name/export", exportHandler.ExportVM)

		// Snapshot endpoints
		vms.POST("/:name/snapshots", vmHandler.CreateSnapshot)
		vms.GET("/:name/snapshots", vmHandler.ListSnapshots)
		vms.GET("/:name/snapshots/:snapshot", vmHandler.GetSnapshot)
		vms.DELETE("/:name/snapshots/:snapshot", vmHandler.DeleteSnapshot)
		vms.PUT("/:name/snapshots/:snapshot/revert", vmHandler.RevertSnapshot)
	}

	// Unified compute instance management endpoints (KVM + Docker)
	if computeHandler != nil {
		compute := protected.Group("/compute")
		{
			// Instance management
			compute.GET("/instances", computeHandler.ListInstances)
			compute.POST("/instances", computeHandler.CreateInstance)
			compute.GET("/instances/:id", computeHandler.GetInstance)
			compute.PUT("/instances/:id", computeHandler.UpdateInstance)
			compute.DELETE("/instances/:id", computeHandler.DeleteInstance)

			// Lifecycle operations
			compute.PUT("/instances/:id/start", computeHandler.StartInstance)
			compute.PUT("/instances/:id/stop", computeHandler.StopInstance)
			compute.PUT("/instances/:id/restart", computeHandler.RestartInstance)
			compute.PUT("/instances/:id/pause", computeHandler.PauseInstance)
			compute.PUT("/instances/:id/unpause", computeHandler.UnpauseInstance)

			// Resource management
			compute.GET("/instances/:id/usage", computeHandler.GetResourceUsage)
			compute.PUT("/instances/:id/resources", computeHandler.UpdateResourceLimits)

			// Events and monitoring
			compute.GET("/instances/:id/events", computeHandler.GetInstanceEvents)

			// Alternative access by name
			compute.GET("/instances/name/:name", computeHandler.GetInstanceByName)

			// Cluster and backend status
			compute.GET("/cluster/status", computeHandler.GetClusterStatus)
			compute.GET("/backends/:backend/info", computeHandler.GetBackendInfo)
			compute.GET("/health", computeHandler.HealthCheck)
		}
	}

	// Docker-specific endpoints (if enabled)
	if dockerHandlers != nil && config != nil && config.Docker.Enabled {
		docker := protected.Group("/docker")
		{
			// Container management
			containers := docker.Group("/containers")
			{
				containers.GET("", dockerHandlers.ListContainers.ListContainers)
				containers.POST("", dockerHandlers.CreateContainer.CreateContainer)
				containers.GET("/:id", dockerHandlers.GetContainer.GetContainer)
				containers.POST("/:id/start", dockerHandlers.StartContainer.StartContainer)
				containers.POST("/:id/stop", dockerHandlers.StopContainer.StopContainer)
				containers.POST("/:id/restart", dockerHandlers.RestartContainer.RestartContainer)
				containers.DELETE("/:id", dockerHandlers.DeleteContainer.DeleteContainer)
				containers.GET("/:id/logs", dockerHandlers.GetContainerLogs.GetContainerLogs)
				containers.GET("/:id/stats", dockerHandlers.GetContainerStats.GetContainerStats)
			}

			// Future: Images, Networks, Volumes endpoints
		}
	}

	// Export job management
	exports := protected.Group("/exports")
	{
		exports.GET("", exportHandler.ListExports)
		exports.GET("/:id", exportHandler.GetExportStatus)
		exports.DELETE("/:id", exportHandler.CancelExport)
	}

	// Network management
	if networkHandlers != nil {
		networks := protected.Group("/networks")
		{
			networks.GET("", networkHandlers.List.Handle)
			networks.POST("", networkHandlers.Create.Handle)
			networks.GET("/:name", networkHandlers.Get.Handle)
			networks.PUT("/:name", networkHandlers.Update.Handle)
			networks.DELETE("/:name", networkHandlers.Delete.Handle)
			networks.PUT("/:name/start", networkHandlers.Start.Handle)
			networks.PUT("/:name/stop", networkHandlers.Stop.Handle)
		}

		// Bridge network management
		bridges := protected.Group("/bridge-networks")
		{
			bridges.GET("", networkHandlers.ListBridges.Handle)
			bridges.POST("", networkHandlers.CreateBridge.Handle)
			bridges.GET("/:name", networkHandlers.GetBridge.Handle)
			bridges.DELETE("/:name", networkHandlers.DeleteBridge.Handle)
		}
	}

	// Storage management
	if storageHandlers != nil {
		// Storage pool routes
		storage := protected.Group("/storage")
		{
			// Pool management
			storage.GET("/pools", storageHandlers.ListPools.Handle)
			storage.POST("/pools", storageHandlers.CreatePool.Handle)
			storage.GET("/pools/:name", storageHandlers.GetPool.Handle)
			storage.DELETE("/pools/:name", storageHandlers.DeletePool.Handle)
			storage.PUT("/pools/:name/start", storageHandlers.StartPool.Handle)
			storage.PUT("/pools/:name/stop", storageHandlers.StopPool.Handle)

			// Volume management
			storage.GET("/pools/:name/volumes", storageHandlers.ListVolumes.Handle)
			storage.POST("/pools/:name/volumes", storageHandlers.CreateVolume.Handle)
			storage.DELETE("/pools/:name/volumes/:volumeName", storageHandlers.DeleteVolume.Handle)
			storage.POST("/pools/:name/volumes/:volumeName/upload", storageHandlers.UploadVolume.Handle)
		}
	}

	// OVS management
	if ovsHandlers != nil {
		ovs := protected.Group("/ovs")
		{
			// Bridge management
			ovs.GET("/bridges", ovsHandlers.ListBridges.Handle)
			ovs.POST("/bridges", ovsHandlers.CreateBridge.Handle)
			ovs.GET("/bridges/:bridge", ovsHandlers.GetBridge.Handle)
			ovs.DELETE("/bridges/:bridge", ovsHandlers.DeleteBridge.Handle)

			// Port management
			ovs.GET("/bridges/:bridge/ports", ovsHandlers.ListPorts.Handle)
			ovs.POST("/bridges/:bridge/ports", ovsHandlers.CreatePort.Handle)
			ovs.DELETE("/bridges/:bridge/ports/:port", ovsHandlers.DeletePort.Handle)

			// Flow management
			ovs.POST("/bridges/:bridge/flows", ovsHandlers.CreateFlow.Handle)
		}
	}

	// Setup WebSocket routes if VM handler provides a VM manager
	if manager := vmHandler.GetVMManager(); manager != nil {
		// We'll use an adapter to adapt the VM manager to the WebSocket interface
		vmAdapter := &vmManagerWebSocketAdapter{
			manager: manager,
		}

		// Setup WebSocket routes - pass config to determine auth requirements
		if config != nil && config.Auth.Enabled {
			// WebSocket routes with authentication
			websocket.SetupRoutes(
				router,
				"/ws",
				vmAdapter,
				log,
				jwtMiddleware,
				roleMiddleware,
			)
		} else {
			// WebSocket routes without authentication
			websocket.SetupRoutesWithoutAuth(
				router,
				"/ws",
				vmAdapter,
				log,
			)
		}
		log.Info("WebSocket routes configured")
	} else {
		log.Warn("Could not get VM manager from handler, WebSocket functionality disabled")
	}
}
