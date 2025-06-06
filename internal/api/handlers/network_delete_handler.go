package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/threatflux/libgo/internal/libvirt/network"
	"github.com/threatflux/libgo/pkg/logger"
)

// NetworkDeleteHandler handles network deletion.
type NetworkDeleteHandler struct {
	networkManager network.Manager
	logger         logger.Logger
}

// NewNetworkDeleteHandler creates a new NetworkDeleteHandler.
func NewNetworkDeleteHandler(networkManager network.Manager, logger logger.Logger) *NetworkDeleteHandler {
	return &NetworkDeleteHandler{
		networkManager: networkManager,
		logger:         logger,
	}
}

// Handle implements Handler interface.
func (h *NetworkDeleteHandler) Handle(c *gin.Context) {
	ctx := c.Request.Context()

	// Get network name from URL parameter.
	name := c.Param("name")
	if name == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Network name is required",
		})
		return
	}

	// Delete the network.
	err := h.networkManager.Delete(ctx, name)
	if err != nil {
		h.logger.Error("Failed to delete network",
			logger.String("name", name),
			logger.Error(err))

		// Check for specific errors.
		if err.Error() == "looking up network: "+err.Error() {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Network not found",
			})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to delete network",
		})
		return
	}

	h.logger.Info("Network deleted successfully",
		logger.String("name", name))

	// Return success.
	c.JSON(http.StatusOK, gin.H{
		"message": "Network deleted successfully",
	})
}
