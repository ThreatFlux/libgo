package recovery

import (
	"net/http"
	"runtime/debug"

	"github.com/gin-gonic/gin"
	"github.com/threatflux/libgo/pkg/logger"
)

// Config holds the configuration for the recovery middleware.
type Config struct {
	// RecoveryHandler is a custom handler function to be called during recovery
	RecoveryHandler func(*gin.Context, interface{})

	// DisableStackTrace determines whether to disable stack trace output
	DisableStackTrace bool

	// DisableRecovery determines whether to disable recovery (useful for testing)
	DisableRecovery bool
}

// Handler returns a gin middleware for recovering from panics.
func Handler(log logger.Logger, config Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		if config.DisableRecovery {
			c.Next()
			return
		}

		defer func() {
			if err := recover(); err != nil {
				handlePanic(c, log, config, err)
			}
		}()

		c.Next()
	}
}

// handlePanic handles the panic recovery logic.
func handlePanic(c *gin.Context, log logger.Logger, config Config, err interface{}) {
	// Get stack trace
	stack := getStackTrace(config.DisableStackTrace)

	// Log the panic
	logPanic(c, log, err, stack)

	// Use custom recovery handler if provided
	if config.RecoveryHandler != nil {
		config.RecoveryHandler(c, err)
		return
	}

	// Default error response
	c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
		"status":  http.StatusInternalServerError,
		"code":    "INTERNAL_SERVER_ERROR",
		"message": "Internal server error",
	})
}

// getStackTrace returns the stack trace if enabled.
func getStackTrace(disabled bool) string {
	if disabled {
		return ""
	}
	return string(debug.Stack())
}

// logPanic logs the panic with appropriate context.
func logPanic(c *gin.Context, fallbackLog logger.Logger, err interface{}, stack string) {
	contextLogger, exists := c.Get("logger")
	if exists {
		if contextLog, ok := contextLogger.(logger.Logger); ok {
			contextLog.Error("Panic recovered",
				logger.Any("error", err),
				logger.String("stack", stack))
			return
		}
	}

	// Fallback to default logger with request context
	fallbackLog.Error("Panic recovered",
		logger.String("method", c.Request.Method),
		logger.String("path", c.Request.URL.Path),
		logger.Any("error", err),
		logger.String("stack", stack))
}
