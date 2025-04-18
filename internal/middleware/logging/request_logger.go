package logging

import (
	"bytes"
	"io"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/wroersma/libgo/pkg/logger"
)

// Config holds the configuration for the request logger
type Config struct {
	// SkipPaths are paths that will not be logged
	SkipPaths []string

	// MaxBodyLogSize is the maximum body size to log (in bytes), 0 means no logging
	MaxBodyLogSize int

	// IncludeRequestBody determines whether to log request bodies
	IncludeRequestBody bool

	// IncludeResponseBody determines whether to log response bodies
	IncludeResponseBody bool
}

// RequestLogger returns a gin middleware for logging HTTP requests
func RequestLogger(log logger.Logger, config Config) gin.HandlerFunc {
	// Create a skip paths map for faster lookup
	skipPaths := make(map[string]bool)
	for _, path := range config.SkipPaths {
		skipPaths[path] = true
	}

	return func(c *gin.Context) {
		// Skip logging for specified paths
		if skipPaths[c.Request.URL.Path] {
			c.Next()
			return
		}

		// Start timer
		start := time.Now()

		// Generate a request ID if none exists
		requestID := c.GetHeader("X-Request-ID")
		if requestID == "" {
			requestID = uuid.New().String()
			c.Header("X-Request-ID", requestID)
		}

		// Store the logger with request ID in context
		contextLogger := log.WithFields(
			logger.String("requestId", requestID),
			logger.String("method", c.Request.Method),
			logger.String("path", c.Request.URL.Path),
			logger.String("ip", c.ClientIP()),
		)

		c.Set("logger", contextLogger)

		// Log request body if enabled
		var requestBody []byte
		if config.IncludeRequestBody && config.MaxBodyLogSize > 0 && c.Request.Body != nil {
			requestBody, _ = io.ReadAll(c.Request.Body)
			// Limit the size if needed
			if len(requestBody) > config.MaxBodyLogSize {
				requestBody = requestBody[:config.MaxBodyLogSize]
			}
			// Replace the body for downstream handlers
			c.Request.Body = io.NopCloser(bytes.NewBuffer(requestBody))
		}

		// Create a custom response writer to capture the response
		var responseBodyBuffer *bytes.Buffer
		var responseBody []byte
		if config.IncludeResponseBody && config.MaxBodyLogSize > 0 {
			responseBodyBuffer = &bytes.Buffer{}
			blw := &bodyLogWriter{
				ResponseWriter: c.Writer,
				bodyBuffer:     responseBodyBuffer,
				maxSize:        config.MaxBodyLogSize,
			}
			c.Writer = blw
		}

		// Process request
		c.Next()

		// Get response status and size
		status := c.Writer.Status()
		size := c.Writer.Size()

		// Get response body if enabled
		if config.IncludeResponseBody && config.MaxBodyLogSize > 0 && responseBodyBuffer != nil {
			responseBody = responseBodyBuffer.Bytes()
		}

		// Log the request
		endTime := time.Now()
		latency := endTime.Sub(start)

		var logEntry logger.Logger = contextLogger.WithFields(
			logger.Int("status", status),
			logger.Int("size", size),
			logger.Float64("latency", latency.Seconds()),
		)

		// Add user ID if available
		if userID, exists := c.Get("userId"); exists {
			logEntry = logEntry.WithFields(logger.String("userId", userID.(string)))
		}

		// Add request body if available
		if len(requestBody) > 0 {
			logEntry = logEntry.WithFields(logger.String("requestBody", string(requestBody)))
		}

		// Add response body if available
		if len(responseBody) > 0 {
			logEntry = logEntry.WithFields(logger.String("responseBody", string(responseBody)))
		}

		// Log with appropriate level based on status code
		if status >= 500 {
			logEntry.Error("Server error")
		} else if status >= 400 {
			logEntry.Warn("Client error")
		} else {
			logEntry.Info("Request handled")
		}
	}
}

// bodyLogWriter is a custom response writer that captures the response body
type bodyLogWriter struct {
	gin.ResponseWriter
	bodyBuffer *bytes.Buffer
	maxSize    int
}

// Write captures the response body and writes it to the original response writer
func (w *bodyLogWriter) Write(b []byte) (int, error) {
	// If we haven't reached the max size, add to buffer
	if w.bodyBuffer.Len() < w.maxSize {
		// Calculate how much more we can add to buffer
		remainingSpace := w.maxSize - w.bodyBuffer.Len()
		if remainingSpace > 0 {
			if remainingSpace >= len(b) {
				w.bodyBuffer.Write(b)
			} else {
				w.bodyBuffer.Write(b[:remainingSpace])
			}
		}
	}
	return w.ResponseWriter.Write(b)
}

// WriteString is here to implement the gin.ResponseWriter interface fully
func (w *bodyLogWriter) WriteString(s string) (int, error) {
	return w.Write([]byte(s))
}
