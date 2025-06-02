package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/threatflux/libgo/internal/libvirt/network"
	"github.com/threatflux/libgo/pkg/logger"
)

// NetworkStopHandler handles stopping a network
type NetworkStopHandler struct {
	networkManager network.Manager
	logger         logger.Logger
}

// NewNetworkStopHandler creates a new NetworkStopHandler
func NewNetworkStopHandler(networkManager network.Manager, logger logger.Logger) *NetworkStopHandler {
	return &NetworkStopHandler{
		networkManager: networkManager,
		logger:         logger,
	}
}

// Handle implements Handler interface
func (h *NetworkStopHandler) Handle(c *gin.Context) {
	ctx := c.Request.Context()

	// Get network name from URL parameter
	name := c.Param("name")
	if name == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Network name is required",
		})
		return
	}

	// Stop the network
	err := h.networkManager.Stop(ctx, name)
	if err != nil {
		h.logger.Error("Failed to stop network",
			logger.String("name", name),
			logger.Error(err))

		// Check for specific errors
		if err.Error() == "looking up network: "+err.Error() {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Network not found",
			})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to stop network",
		})
		return
	}

	h.logger.Info("Network stopped successfully",
		logger.String("name", name))

	// Get updated network info
	networkInfo, err := h.networkManager.GetInfo(ctx, name)
	if err != nil {
		h.logger.Warn("Failed to get network info after stop",
			logger.String("name", name),
			logger.Error(err))

		// Still return success since the stop operation succeeded
		c.JSON(http.StatusOK, gin.H{
			"message": "Network stopped successfully",
		})
		return
	}

	// Return updated network info
	c.JSON(http.StatusOK, networkInfo)
}
