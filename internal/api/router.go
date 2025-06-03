package api

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/threatflux/libgo/internal/middleware/auth"
	"github.com/threatflux/libgo/internal/middleware/logging"
	"github.com/threatflux/libgo/internal/middleware/recovery"
	"github.com/threatflux/libgo/pkg/logger"
)

// RouterConfig holds the configuration for the router.
type RouterConfig struct {
	// LoggingConfig is the configuration for request logging
	LoggingConfig logging.Config

	// RecoveryConfig is the configuration for panic recovery
	RecoveryConfig recovery.Config

	// BasePath is the base path for all API routes (e.g., "/api/v1")
	BasePath string

	// EnableCORS determines if CORS support is enabled
	EnableCORS bool
}

// DefaultRouterConfig returns the default router configuration.
func DefaultRouterConfig() RouterConfig {
	return RouterConfig{
		BasePath:   "/api/v1",
		EnableCORS: true,
		LoggingConfig: logging.Config{
			SkipPaths:          []string{"/health", "/metrics"},
			MaxBodyLogSize:     4096,
			IncludeRequestBody: true,
		},
		RecoveryConfig: recovery.Config{
			DisableStackTrace: false,
		},
	}
}

// vmManagerWebSocketAdapter is defined in router_adapter.go

// SetupRouter configures the API router with standard middleware and routes.
func SetupRouter(
	engine *gin.Engine,
	log logger.Logger,
	config RouterConfig,
	authMiddleware *auth.JWTMiddleware,
	roleMiddleware *auth.RoleMiddleware,
	vmManager interface{}, // VM manager interface for WebSocket monitoring
) *gin.Engine {
	// Apply middleware to all routes
	engine.Use(recovery.Handler(log, config.RecoveryConfig))
	engine.Use(logging.RequestLogger(log, config.LoggingConfig))

	// CORS support if enabled
	if config.EnableCORS {
		engine.Use(corsMiddleware())
	}

	// Health check endpoint (not behind auth)
	engine.GET("/health", healthCheckHandler)

	// Setup API routes under base path
	api := engine.Group(config.BasePath)

	// Public routes (no auth required)
	setupPublicRoutes(api)

	// Protected routes (auth required)
	protected := api.Group("")
	protected.Use(authMiddleware.Authenticate())
	setupProtectedRoutes(protected, authMiddleware, roleMiddleware)

	// Setup Admin routes
	admin := protected.Group("/admin")
	admin.Use(roleMiddleware.RequireRole("admin"))
	setupAdminRoutes(admin)

	// Setup WebSocket routes in api.ConfigureRoutes

	// No route handler
	engine.NoRoute(noRouteHandler)

	return engine
}

// setupPublicRoutes configures routes that don't require authentication.
func setupPublicRoutes(router *gin.RouterGroup) {
	// Authentication endpoints
	auth := router.Group("/auth")
	{
		// These will be connected to actual handlers in separate files
		auth.POST("/login", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"status": "Not implemented"})
		})

		auth.POST("/refresh", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"status": "Not implemented"})
		})
	}
}

// setupProtectedRoutes configures routes that require authentication.
func setupProtectedRoutes(
	router *gin.RouterGroup,
	authMiddleware *auth.JWTMiddleware,
	roleMiddleware *auth.RoleMiddleware,
) {
	// VM management endpoints
	vms := router.Group("/vms")
	{
		// List VMs (all users with read permission)
		vms.GET("", roleMiddleware.RequirePermission("read"), func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"status": "Not implemented"})
		})

		// Get VM details (all users with read permission)
		vms.GET("/:name", roleMiddleware.RequirePermission("read"), func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"status": "Not implemented"})
		})

		// Create VM (users with create permission)
		vms.POST("", roleMiddleware.RequirePermission("create"), func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"status": "Not implemented"})
		})

		// Delete VM (users with delete permission)
		vms.DELETE("/:name", roleMiddleware.RequirePermission("delete"), func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"status": "Not implemented"})
		})

		// Start VM (users with start permission)
		vms.PUT("/:name/start", roleMiddleware.RequirePermission("start"), func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"status": "Not implemented"})
		})

		// Stop VM (users with stop permission)
		vms.PUT("/:name/stop", roleMiddleware.RequirePermission("stop"), func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"status": "Not implemented"})
		})

		// Export VM (users with export permission)
		vms.POST("/:name/export", roleMiddleware.RequirePermission("export"), func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"status": "Not implemented"})
		})
	}

	// User profile endpoints
	profile := router.Group("/profile")
	{
		// Get current user profile
		profile.GET("", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"status": "Not implemented"})
		})

		// Update current user profile
		profile.PUT("", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"status": "Not implemented"})
		})

		// Change password
		profile.PUT("/password", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"status": "Not implemented"})
		})
	}
}

// setupAdminRoutes configures routes that require admin role.
func setupAdminRoutes(router *gin.RouterGroup) {
	// User management (admin only)
	users := router.Group("/users")
	{
		// List all users
		users.GET("", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"status": "Not implemented"})
		})

		// Get user details
		users.GET("/:id", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"status": "Not implemented"})
		})

		// Create user
		users.POST("", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"status": "Not implemented"})
		})

		// Update user
		users.PUT("/:id", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"status": "Not implemented"})
		})

		// Delete user
		users.DELETE("/:id", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"status": "Not implemented"})
		})
	}

	// System management (admin only)
	system := router.Group("/system")
	{
		// Get system stats
		system.GET("/stats", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"status": "Not implemented"})
		})

		// Check libvirt connection
		system.GET("/libvirt", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"status": "Not implemented"})
		})
	}
}

// healthCheckHandler handles health check requests.
func healthCheckHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status": "up",
		"time":   time.Now().Format(time.RFC3339),
	})
}

// noRouteHandler handles requests to non-existent routes.
func noRouteHandler(c *gin.Context) {
	c.JSON(http.StatusNotFound, gin.H{
		"status":  http.StatusNotFound,
		"code":    "NOT_FOUND",
		"message": "The requested resource was not found",
	})
}

// corsMiddleware adds CORS headers to responses.
func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}
