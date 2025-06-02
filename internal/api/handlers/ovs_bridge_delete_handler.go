package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/threatflux/libgo/internal/ovs"
	"github.com/threatflux/libgo/pkg/logger"
)

// OVSBridgeDeleteHandler handles deleting OVS bridges
type OVSBridgeDeleteHandler struct {
	ovsManager ovs.Manager
	logger     logger.Logger
}

// NewOVSBridgeDeleteHandler creates a new OVSBridgeDeleteHandler
func NewOVSBridgeDeleteHandler(ovsManager ovs.Manager, logger logger.Logger) *OVSBridgeDeleteHandler {
	return &OVSBridgeDeleteHandler{
		ovsManager: ovsManager,
		logger:     logger,
	}
}

// Handle implements Handler interface
func (h *OVSBridgeDeleteHandler) Handle(c *gin.Context) {
	ctx := c.Request.Context()
	name := c.Param("bridge")

	if name == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Bridge name is required",
		})
		return
	}

	// Delete the bridge
	if err := h.ovsManager.DeleteBridge(ctx, name); err != nil {
		h.logger.Error("Failed to delete OVS bridge",
			logger.String("name", name),
			logger.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Bridge deleted successfully",
	})
}
