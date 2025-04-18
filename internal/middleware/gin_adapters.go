package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/wroersma/libgo/internal/middleware/logging"
	"github.com/wroersma/libgo/internal/middleware/recovery"
	"github.com/wroersma/libgo/pkg/logger"
)

// RecoveryToGin adapts a standard HTTP recovery middleware to Gin middleware
func RecoveryToGin(middleware func(http.ResponseWriter, *http.Request, func(http.ResponseWriter, *http.Request))) gin.HandlerFunc {
	return func(c *gin.Context) {
		middleware(c.Writer, c.Request, func(w http.ResponseWriter, r *http.Request) {
			c.Next()
		})
	}
}

// LoggingToGin adapts a standard HTTP logging middleware to Gin middleware
func LoggingToGin(middleware func(http.ResponseWriter, *http.Request, func(http.ResponseWriter, *http.Request))) gin.HandlerFunc {
	return func(c *gin.Context) {
		middleware(c.Writer, c.Request, func(w http.ResponseWriter, r *http.Request) {
			c.Next()
		})
	}
}

// RecoveryGinMiddleware creates a Gin recovery middleware
func RecoveryGinMiddleware(log logger.Logger) gin.HandlerFunc {
	return RecoveryToGin(recovery.RecoveryMiddleware(log))
}

// LoggingGinMiddleware creates a Gin logging middleware
func LoggingGinMiddleware(log logger.Logger) gin.HandlerFunc {
	return LoggingToGin(logging.RequestLoggerMiddleware(log))
}
