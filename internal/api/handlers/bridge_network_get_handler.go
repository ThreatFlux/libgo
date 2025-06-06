package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/threatflux/libgo/internal/libvirt/network"
	"github.com/threatflux/libgo/pkg/logger"
)

// BridgeNetworkGetHandler handles getting bridge network details.
type BridgeNetworkGetHandler struct {
	networkManager network.Manager
	logger         logger.Logger
}

// NewBridgeNetworkGetHandler creates a new BridgeNetworkGetHandler.
func NewBridgeNetworkGetHandler(networkManager network.Manager, logger logger.Logger) *BridgeNetworkGetHandler {
	return &BridgeNetworkGetHandler{
		networkManager: networkManager,
		logger:         logger,
	}
}

// Handle implements Handler interface.
func (h *BridgeNetworkGetHandler) Handle(c *gin.Context) {
	ctx := c.Request.Context()
	name := c.Param("name")

	if name == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Network name is required",
		})
		return
	}

	// Get network details.
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
	if details.Forward.Mode != bridgeNetworkMode {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Network is not a bridge network",
		})
		return
	}

	response := BridgeNetworkInfo{
		Name:        details.Name,
		BridgeName:  details.BridgeName,
		Active:      details.Active,
		AutoStart:   details.Autostart,
		ForwardMode: details.Forward.Mode,
	}

	h.logger.Debug("Retrieved bridge network details", logger.String("name", name))

	c.JSON(http.StatusOK, response)
}
