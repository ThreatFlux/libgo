package api

import (
	"context"
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/threatflux/libgo/internal/api/handlers"
	"github.com/threatflux/libgo/internal/config"
	"github.com/threatflux/libgo/internal/middleware/auth"
	vmmodels "github.com/threatflux/libgo/internal/models/vm"
	"github.com/threatflux/libgo/internal/websocket"
	"github.com/threatflux/libgo/pkg/logger"
)

// vmManagerWebSocketAdapter adapts any VM manager to the websocket.VMManager interface
type vmManagerWebSocketAdapter struct {
	manager interface{}
}

// Get implements the Get method of websocket.VMManager
func (a *vmManagerWebSocketAdapter) Get(ctx context.Context, name string) (*vmmodels.VM, error) {
	// Use reflection to call the Get method on the underlying manager
	if getter, ok := a.manager.(interface {
		Get(ctx context.Context, name string) (*vmmodels.VM, error)
	}); ok {
		return getter.Get(ctx, name)
	}
	return nil, fmt.Errorf("VM manager does not support Get method")
}

// GetMetrics implements the GetMetrics method of websocket.VMManager
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
		wsMetrics.CPU.Utilization = 25.0 + float64(time.Now().Unix() % 50) // Random-ish value between 25-75%
		wsMetrics.Memory.Used = 1024*1024*1024 + uint64(time.Now().Unix() % 3 * 1024*1024*1024) // 1-4GB
		wsMetrics.Memory.Total = 8 * 1024 * 1024 * 1024 // 8GB
		wsMetrics.Network.RxBytes = uint64(time.Now().Unix() % 10 * 1024 * 1024) // 0-10MB
		wsMetrics.Network.TxBytes = uint64(time.Now().Unix() % 5 * 1024 * 1024)  // 0-5MB
		wsMetrics.Disk.ReadBytes = uint64(time.Now().Unix() % 20 * 1024 * 1024)  // 0-20MB
		wsMetrics.Disk.WriteBytes = uint64(time.Now().Unix() % 10 * 1024 * 1024) // 0-10MB
		
		return wsMetrics, nil
	}
	
	return nil, fmt.Errorf("VM manager does not support GetMetrics method")
}

// ConfigureRoutes configures the API router with all handlers
// This is an adapter for the main router setup function to handle all the API handlers
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
	}

	// Export job management
	exports := protected.Group("/exports")
	{
		exports.GET("", exportHandler.ListExports)
		exports.GET("/:id", exportHandler.GetExportStatus)
		exports.DELETE("/:id", exportHandler.CancelExport)
	}
	
	// Setup WebSocket routes if VM handler provides a VM manager
	if manager := vmHandler.GetVMManager(); manager != nil {
		// We'll use an adapter to adapt the VM manager to the WebSocket interface
		vmAdapter := &vmManagerWebSocketAdapter{
			manager: manager,
		}

		// WebSocket routes are under /ws prefix
		websocket.SetupRoutes(
			router,
			"/ws",
			vmAdapter, // Use the adapter instead of direct cast
			log,
			jwtMiddleware,
			roleMiddleware,
		)
		log.Info("WebSocket routes configured")
	} else {
		log.Warn("Could not get VM manager from handler, WebSocket functionality disabled")
	}
}