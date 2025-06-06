package handlers

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/threatflux/libgo/internal/libvirt/network"
	"github.com/threatflux/libgo/pkg/logger"
)

// networkOperationHandler is a generic handler for network start/stop operations.
func networkOperationHandler(c *gin.Context, networkManager network.Manager, log logger.Logger,
	operation func(context.Context, string) error, operationName string) {
	ctx := c.Request.Context()

	// Get network name from URL parameter.
	name := c.Param("name")
	if name == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Network name is required",
		})
		return
	}

	// Execute the operation.
	err := operation(ctx, name)
	if err != nil {
		log.Error("Failed to "+operationName+" network",
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
			"error": "Failed to " + operationName + " network",
		})
		return
	}

	log.Info("Network "+operationName+"ped successfully",
		logger.String("name", name))

	// Get updated network info.
	networkInfo, err := networkManager.GetInfo(ctx, name)
	if err != nil {
		log.Warn("Failed to get network info after "+operationName,
			logger.String("name", name),
			logger.Error(err))

		// Still return success since the operation succeeded.
		c.JSON(http.StatusOK, gin.H{
			"message": "Network " + operationName + "ped successfully",
		})
		return
	}

	// Return updated network info.
	c.JSON(http.StatusOK, networkInfo)
}

// NetworkStartHandler handles starting a network.
type NetworkStartHandler struct {
	networkManager network.Manager
	logger         logger.Logger
}

// NewNetworkStartHandler creates a new NetworkStartHandler.
func NewNetworkStartHandler(networkManager network.Manager, logger logger.Logger) *NetworkStartHandler {
	return &NetworkStartHandler{
		networkManager: networkManager,
		logger:         logger,
	}
}

// Handle implements Handler interface.
func (h *NetworkStartHandler) Handle(c *gin.Context) {
	networkOperationHandler(c, h.networkManager, h.logger, h.networkManager.Start, "start")
}
