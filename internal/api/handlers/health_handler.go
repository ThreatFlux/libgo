package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/wroersma/libgo/internal/health"
	"github.com/wroersma/libgo/pkg/logger"
)

// HealthHandler handles health check endpoints
type HealthHandler struct {
	checker health.Checker
	logger  logger.Logger
}

// NewHealthHandler creates a new health handler
func NewHealthHandler(checker *health.Checker, logger logger.Logger) *HealthHandler {
	return &HealthHandler{
		checker: *checker,
		logger:  logger,
	}
}

// GetHealth handles GET /health
func (h *HealthHandler) GetHealth(c *gin.Context) {
	// Run health checks
	result := h.checker.RunChecks()
	
	// Determine response status code
	statusCode := http.StatusOK
	if result.Status == health.StatusDown {
		statusCode = http.StatusServiceUnavailable
	}
	
	// Return health check result
	c.JSON(statusCode, result)
}

// GetHealthLiveness handles GET /health/liveness
func (h *HealthHandler) GetHealthLiveness(c *gin.Context) {
	// Simple liveness check - if the service can respond, it's alive
	c.JSON(http.StatusOK, gin.H{
		"status": "UP",
	})
}

// GetHealthReadiness handles GET /health/readiness
func (h *HealthHandler) GetHealthReadiness(c *gin.Context) {
	// Run full health checks for readiness
	result := h.checker.RunChecks()
	
	// Determine response status code
	statusCode := http.StatusOK
	if result.Status == health.StatusDown {
		statusCode = http.StatusServiceUnavailable
	}
	
	// Return health check result
	c.JSON(statusCode, result)
}

// RegisterHandler registers all health routes
func (h *HealthHandler) RegisterHandler(router gin.IRouter) {
	router.GET("/health", h.GetHealth)
	router.GET("/health/liveness", h.GetHealthLiveness)
	router.GET("/health/readiness", h.GetHealthReadiness)
}
