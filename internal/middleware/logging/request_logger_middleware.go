package logging

import (
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/wroersma/libgo/pkg/logger"
)

// RequestLoggerMiddleware returns a standard HTTP middleware function for logging HTTP requests
func RequestLoggerMiddleware(log logger.Logger) func(http.ResponseWriter, *http.Request, func(http.ResponseWriter, *http.Request)) {
	return func(w http.ResponseWriter, r *http.Request, next func(http.ResponseWriter, *http.Request)) {
		// Start timer
		start := time.Now()

		// Generate a request ID if none exists
		requestID := r.Header.Get("X-Request-ID")
		if requestID == "" {
			requestID = uuid.New().String()
			w.Header().Set("X-Request-ID", requestID)
		}

		// Create a response writer wrapper to capture status code
		ww := &responseWriterWrapper{
			ResponseWriter: w,
			statusCode:     http.StatusOK, // Default to 200 OK
		}

		// Process request with wrapped response writer
		next(ww, r)

		// Log request completion
		duration := time.Since(start)
		log.Info("HTTP request processed",
			logger.String("method", r.Method),
			logger.String("path", r.URL.Path),
			logger.String("requestId", requestID),
			logger.String("remoteAddr", r.RemoteAddr),
			logger.String("userAgent", r.UserAgent()),
			logger.Int("status", ww.statusCode),
			logger.Float64("durationMs", float64(duration.Milliseconds())),
		)
	}
}

// responseWriterWrapper wraps http.ResponseWriter to capture the status code
type responseWriterWrapper struct {
	http.ResponseWriter
	statusCode int
}

// WriteHeader captures the status code before calling the underlying WriteHeader
func (w *responseWriterWrapper) WriteHeader(statusCode int) {
	w.statusCode = statusCode
	w.ResponseWriter.WriteHeader(statusCode)
}
