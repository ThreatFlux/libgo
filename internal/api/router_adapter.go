package api

import (
	"github.com/gin-gonic/gin"
	"github.com/threatflux/libgo/internal/api/handlers"
	"github.com/threatflux/libgo/internal/middleware/auth"
	"github.com/threatflux/libgo/pkg/logger"
)

// ConfigureRoutes configures the API router with all handlers
// This is an adapter for the main router setup function to handle all the API handlers
func ConfigureRoutes(
	router *gin.Engine,
	log logger.Logger,
	jwtMiddleware *auth.JWTMiddleware,
	vmHandler *handlers.VMHandler,
	exportHandler *handlers.ExportHandler,
	authHandler *handlers.AuthHandler,
	healthHandler *handlers.HealthHandler,
	metricsHandler *handlers.MetricsHandler,
) {
	// Add health check endpoint
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "healthy"})
	})

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

	// Protected routes (require authentication)
	protected := apiGroup.Group("/")
	protected.Use(jwtMiddleware.Authenticate())

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
}
