package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/threatflux/libgo/internal/libvirt/network"
	"github.com/threatflux/libgo/pkg/logger"
)

// BridgeNetworkDeleteHandler handles deleting bridge networks.
type BridgeNetworkDeleteHandler struct {
	networkManager network.Manager
	logger         logger.Logger
}

// NewBridgeNetworkDeleteHandler creates a new BridgeNetworkDeleteHandler.
func NewBridgeNetworkDeleteHandler(networkManager network.Manager, logger logger.Logger) *BridgeNetworkDeleteHandler {
	return &BridgeNetworkDeleteHandler{
		networkManager: networkManager,
		logger:         logger,
	}
}

// Handle implements Handler interface.
func (h *BridgeNetworkDeleteHandler) Handle(c *gin.Context) {
	ctx := c.Request.Context()
	name := c.Param("name")

	if name == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Network name is required",
		})
		return
	}

	// Get network details first to verify it's a bridge network.
	details, err := h.networkManager.GetInfo(ctx, name)
	if err != nil {
		h.logger.Error("Failed to get network details",
			logger.String("name", name),
			logger.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	// Check if it's a bridge network.
	if details.Forward.Mode != "bridge" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Network is not a bridge network",
		})
		return
	}

	// Stop the network first if it's active.
	if details.Active {
		if err := h.networkManager.Stop(ctx, name); err != nil {
			h.logger.Error("Failed to stop network before deletion",
				logger.String("name", name),
				logger.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})
			return
		}
	}

	// Delete the network.
	if err := h.networkManager.Delete(ctx, name); err != nil {
		h.logger.Error("Failed to delete bridge network",
			logger.String("name", name),
			logger.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	h.logger.Info("Bridge network deleted", logger.String("name", name))

	c.JSON(http.StatusOK, gin.H{
		"message": "Bridge network deleted successfully",
	})
}
