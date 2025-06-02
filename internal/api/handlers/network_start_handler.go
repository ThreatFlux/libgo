package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/threatflux/libgo/internal/libvirt/network"
	"github.com/threatflux/libgo/pkg/logger"
)

// NetworkStartHandler handles starting a network
type NetworkStartHandler struct {
	networkManager network.Manager
	logger         logger.Logger
}

// NewNetworkStartHandler creates a new NetworkStartHandler
func NewNetworkStartHandler(networkManager network.Manager, logger logger.Logger) *NetworkStartHandler {
	return &NetworkStartHandler{
		networkManager: networkManager,
		logger:         logger,
	}
}

// Handle implements Handler interface
func (h *NetworkStartHandler) Handle(c *gin.Context) {
	ctx := c.Request.Context()

	// Get network name from URL parameter
	name := c.Param("name")
	if name == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Network name is required",
		})
		return
	}

	// Start the network
	err := h.networkManager.Start(ctx, name)
	if err != nil {
		h.logger.Error("Failed to start network",
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
			"error": "Failed to start network",
		})
		return
	}

	h.logger.Info("Network started successfully",
		logger.String("name", name))

	// Get updated network info
	networkInfo, err := h.networkManager.GetInfo(ctx, name)
	if err != nil {
		h.logger.Warn("Failed to get network info after start",
			logger.String("name", name),
			logger.Error(err))

		// Still return success since the start operation succeeded
		c.JSON(http.StatusOK, gin.H{
			"message": "Network started successfully",
		})
		return
	}

	// Return updated network info
	c.JSON(http.StatusOK, networkInfo)
}
