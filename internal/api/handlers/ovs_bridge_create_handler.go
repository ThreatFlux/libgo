package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/threatflux/libgo/internal/ovs"
	"github.com/threatflux/libgo/pkg/logger"
)

// OVSBridgeCreateHandler handles creating OVS bridges.
type OVSBridgeCreateHandler struct {
	ovsManager ovs.Manager
	logger     logger.Logger
}

// NewOVSBridgeCreateHandler creates a new OVSBridgeCreateHandler.
func NewOVSBridgeCreateHandler(ovsManager ovs.Manager, logger logger.Logger) *OVSBridgeCreateHandler {
	return &OVSBridgeCreateHandler{
		ovsManager: ovsManager,
		logger:     logger,
	}
}

// Handle implements Handler interface.
func (h *OVSBridgeCreateHandler) Handle(c *gin.Context) {
	ctx := c.Request.Context()

	var params ovs.CreateBridgeParams
	if err := c.ShouldBindJSON(&params); err != nil {
		h.logger.Debug("Invalid request body", logger.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request body",
		})
		return
	}

	// Create the bridge
	if err := h.ovsManager.CreateBridge(ctx, params.Name); err != nil {
		h.logger.Error("Failed to create OVS bridge",
			logger.String("name", params.Name),
			logger.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	// Set controller if specified
	if params.Controller != "" {
		if err := h.ovsManager.SetController(ctx, params.Name, params.Controller); err != nil {
			h.logger.Warn("Failed to set controller on bridge",
				logger.String("bridge", params.Name),
				logger.String("controller", params.Controller),
				logger.Error(err))
		}
	}

	// Get bridge info
	bridge, err := h.ovsManager.GetBridge(ctx, params.Name)
	if err != nil {
		h.logger.Error("Failed to get bridge info after creation",
			logger.String("name", params.Name),
			logger.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Bridge created but failed to get info",
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"bridge": bridge,
	})
}
