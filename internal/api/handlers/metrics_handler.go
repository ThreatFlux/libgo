package handlers

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/wroersma/libgo/internal/metrics"
	"github.com/wroersma/libgo/pkg/logger"
)

// MetricsHandler handles metrics endpoints
type MetricsHandler struct {
	collector metrics.Collector
	logger    logger.Logger
}

// NewMetricsHandler creates a new metrics handler
func NewMetricsHandler(collector metrics.Collector, logger logger.Logger) *MetricsHandler {
	return &MetricsHandler{
		collector: collector,
		logger:    logger,
	}
}

// GetMetrics handles GET /metrics
func (h *MetricsHandler) GetMetrics(c *gin.Context) {
	h.logger.Debug("Serving metrics request")

	// Use Prometheus HTTP handler
	promHandler := promhttp.Handler()
	promHandler.ServeHTTP(c.Writer, c.Request)
}

// RegisterHandler registers all metrics routes
func (h *MetricsHandler) RegisterHandler(router gin.IRouter) {
	router.GET("/metrics", h.GetMetrics)
}

// CollectRequestMetrics middleware collects metrics for API requests
func (h *MetricsHandler) CollectRequestMetrics() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Track request start time
		startTime := time.Now()

		// Process request
		c.Next()

		// Calculate duration
		duration := time.Since(startTime)

		// Record metrics
		h.collector.RecordRequest(
			c.Request.Method,
			c.Request.URL.Path,
			c.Writer.Status(),
			duration,
		)
	}
}
