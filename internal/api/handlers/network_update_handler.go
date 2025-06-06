package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/threatflux/libgo/internal/libvirt/network"
	"github.com/threatflux/libgo/pkg/logger"
)

// NetworkUpdateHandler handles network updates.
type NetworkUpdateHandler struct {
	networkManager network.Manager
	logger         logger.Logger
}

// NewNetworkUpdateHandler creates a new NetworkUpdateHandler.
func NewNetworkUpdateHandler(networkManager network.Manager, logger logger.Logger) *NetworkUpdateHandler {
	return &NetworkUpdateHandler{
		networkManager: networkManager,
		logger:         logger,
	}
}

// Handle implements Handler interface.
func (h *NetworkUpdateHandler) Handle(c *gin.Context) {
	ctx := c.Request.Context()

	// Get network name from URL parameter.
	name := c.Param("name")
	if name == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Network name is required",
		})
		return
	}

	// Parse request body.
	var params network.UpdateNetworkParams
	if err := c.ShouldBindJSON(&params); err != nil {
		h.logger.Debug("Invalid request body", logger.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request body",
		})
		return
	}

	// Update the network.
	networkInfo, err := h.networkManager.Update(ctx, name, &params)
	if err != nil {
		h.logger.Error("Failed to update network",
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
			"error": "Failed to update network",
		})
		return
	}

	h.logger.Info("Network updated successfully",
		logger.String("name", networkInfo.Name),
		logger.String("uuid", networkInfo.UUID))

	// Return the updated network info.
	c.JSON(http.StatusOK, networkInfo)
}
