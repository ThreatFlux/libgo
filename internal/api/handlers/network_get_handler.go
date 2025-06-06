package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/threatflux/libgo/internal/libvirt/network"
	"github.com/threatflux/libgo/pkg/logger"
)

// NetworkGetHandler handles getting network details.
type NetworkGetHandler struct {
	networkManager network.Manager
	logger         logger.Logger
}

// NewNetworkGetHandler creates a new NetworkGetHandler.
func NewNetworkGetHandler(networkManager network.Manager, logger logger.Logger) *NetworkGetHandler {
	return &NetworkGetHandler{
		networkManager: networkManager,
		logger:         logger,
	}
}

// Handle implements Handler interface.
func (h *NetworkGetHandler) Handle(c *gin.Context) {
	ctx := c.Request.Context()

	// Get network name from URL parameter.
	name := c.Param("name")
	if name == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Network name is required",
		})
		return
	}

	// Get network information.
	networkInfo, err := h.networkManager.GetInfo(ctx, name)
	if err != nil {
		h.logger.Debug("Failed to get network info",
			logger.String("name", name),
			logger.Error(err))

		// Check if network not found.
		if err.Error() == "looking up network: "+err.Error() {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Network not found",
			})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to get network information",
		})
		return
	}

	// Return the network info.
	c.JSON(http.StatusOK, networkInfo)
}
