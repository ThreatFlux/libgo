package recovery

import (
	"net/http"
	"runtime/debug"

	"github.com/gin-gonic/gin"
	"github.com/threatflux/libgo/pkg/logger"
)

// Config holds the configuration for the recovery middleware
type Config struct {
	// DisableStackTrace determines whether to disable stack trace output
	DisableStackTrace bool

	// DisableRecovery determines whether to disable recovery (useful for testing)
	DisableRecovery bool

	// RecoveryHandler is a custom handler function to be called during recovery
	RecoveryHandler func(*gin.Context, interface{})
}

// Handler returns a gin middleware for recovering from panics
func Handler(log logger.Logger, config Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		if config.DisableRecovery {
			c.Next()
			return
		}

		defer func() {
			if err := recover(); err != nil {
				// Get stack trace
				stack := []byte{}
				if !config.DisableStackTrace {
					stack = debug.Stack()
				}

				// Log the panic
				contextLogger, exists := c.Get("logger")
				if exists {
					contextLog, ok := contextLogger.(logger.Logger)
					if ok {
						contextLog.Error("Panic recovered",
							logger.Any("error", err),
							logger.String("stack", string(stack)))
					} else {
						log.Error("Panic recovered",
							logger.Any("error", err),
							logger.String("stack", string(stack)))
					}
				} else {
					log.Error("Panic recovered",
						logger.String("method", c.Request.Method),
						logger.String("path", c.Request.URL.Path),
						logger.Any("error", err),
						logger.String("stack", string(stack)))
				}

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
		}()

		c.Next()
	}
}
