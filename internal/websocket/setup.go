package websocket

import (
	"github.com/gin-gonic/gin"
	"github.com/threatflux/libgo/internal/middleware/auth"
	"github.com/threatflux/libgo/pkg/logger"
)

// SetupRoutes configures the WebSocket routes with authentication
func SetupRoutes(
	router *gin.Engine,
	basePath string,
	vmManager VMManager,
	logger logger.Logger,
	authMiddleware *auth.JWTMiddleware,
	roleMiddleware *auth.RoleMiddleware,
) *Handler {
	// Create WebSocket handler
	handler := NewHandler(logger)

	// Create VM monitor
	monitor := NewVMMonitor(handler, vmManager, logger)
	monitor.Start()

	// Setup WebSocket routes
	ws := router.Group(basePath)
	{
		// VM monitoring endpoint
		ws.GET("/vms/:name", authMiddleware.Authenticate(), roleMiddleware.RequirePermission("read"), func(c *gin.Context) {
			vmName := c.Param("name")
			monitor.RegisterVM(vmName)
			handler.HandleVM(c)
		})

		// VM console endpoint
		ws.GET("/vms/:name/console", authMiddleware.Authenticate(), roleMiddleware.RequirePermission("console"), func(c *gin.Context) {
			vmName := c.Param("name")
			monitor.RegisterVM(vmName)
			handler.HandleVMConsole(c)
		})
	}

	return handler
}

// SetupRoutesWithoutAuth configures the WebSocket routes without authentication
func SetupRoutesWithoutAuth(
	router *gin.Engine,
	basePath string,
	vmManager VMManager,
	logger logger.Logger,
) *Handler {
	// Create WebSocket handler
	handler := NewHandler(logger)

	// Create VM monitor
	monitor := NewVMMonitor(handler, vmManager, logger)
	monitor.Start()

	// Setup WebSocket routes without authentication
	ws := router.Group(basePath)
	{
		// VM monitoring endpoint
		ws.GET("/vms/:name", func(c *gin.Context) {
			vmName := c.Param("name")
			monitor.RegisterVM(vmName)
			// Set a default userID when auth is disabled
			c.Set("userID", "anonymous")
			handler.HandleVM(c)
		})

		// VM console endpoint
		ws.GET("/vms/:name/console", func(c *gin.Context) {
			vmName := c.Param("name")
			monitor.RegisterVM(vmName)
			// Set a default userID when auth is disabled
			c.Set("userID", "anonymous")
			handler.HandleVMConsole(c)
		})
	}

	return handler
}
