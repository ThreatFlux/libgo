package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/threatflux/libgo/internal/ovs"
	"github.com/threatflux/libgo/pkg/logger"
)

// OVSPortDeleteHandler handles deleting OVS ports
type OVSPortDeleteHandler struct {
	ovsManager ovs.Manager
	logger     logger.Logger
}

// NewOVSPortDeleteHandler creates a new OVSPortDeleteHandler
func NewOVSPortDeleteHandler(ovsManager ovs.Manager, logger logger.Logger) *OVSPortDeleteHandler {
	return &OVSPortDeleteHandler{
		ovsManager: ovsManager,
		logger:     logger,
	}
}

// Handle implements Handler interface
func (h *OVSPortDeleteHandler) Handle(c *gin.Context) {
	ctx := c.Request.Context()
	bridge := c.Param("bridge")
	port := c.Param("port")

	if bridge == "" || port == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Bridge and port names are required",
		})
		return
	}

	// Delete the port
	if err := h.ovsManager.DeletePort(ctx, bridge, port); err != nil {
		h.logger.Error("Failed to delete OVS port",
			logger.String("bridge", bridge),
			logger.String("port", port),
			logger.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Port deleted successfully",
	})
}
