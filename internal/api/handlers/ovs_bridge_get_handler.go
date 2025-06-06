package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/threatflux/libgo/internal/ovs"
	"github.com/threatflux/libgo/pkg/logger"
)

// OVSBridgeGetHandler handles getting a specific OVS bridge.
type OVSBridgeGetHandler struct {
	ovsManager ovs.Manager
	logger     logger.Logger
}

// NewOVSBridgeGetHandler creates a new OVSBridgeGetHandler.
func NewOVSBridgeGetHandler(ovsManager ovs.Manager, logger logger.Logger) *OVSBridgeGetHandler {
	return &OVSBridgeGetHandler{
		ovsManager: ovsManager,
		logger:     logger,
	}
}

// Handle implements Handler interface.
func (h *OVSBridgeGetHandler) Handle(c *gin.Context) {
	ctx := c.Request.Context()
	name := c.Param("bridge")

	if name == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Bridge name is required",
		})
		return
	}

	// Get bridge info
	bridge, err := h.ovsManager.GetBridge(ctx, name)
	if err != nil {
		h.logger.Error("Failed to get OVS bridge",
			logger.String("name", name),
			logger.Error(err))
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Bridge not found",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"bridge": bridge,
	})
}
