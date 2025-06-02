package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/threatflux/libgo/internal/libvirt/network"
	"github.com/threatflux/libgo/pkg/logger"
)

// BridgeNetworkCreateHandler handles creating libvirt bridge networks
type BridgeNetworkCreateHandler struct {
	networkManager network.Manager
	logger         logger.Logger
}

// NewBridgeNetworkCreateHandler creates a new BridgeNetworkCreateHandler
func NewBridgeNetworkCreateHandler(networkManager network.Manager, logger logger.Logger) *BridgeNetworkCreateHandler {
	return &BridgeNetworkCreateHandler{
		networkManager: networkManager,
		logger:         logger,
	}
}

// CreateBridgeNetworkParams represents the request parameters for creating a bridge network
type CreateBridgeNetworkParams struct {
	Name       string `json:"name" binding:"required"`
	BridgeName string `json:"bridge_name" binding:"required"`
	AutoStart  bool   `json:"auto_start,omitempty"`
}

// Handle implements Handler interface
func (h *BridgeNetworkCreateHandler) Handle(c *gin.Context) {
	ctx := c.Request.Context()
	var params CreateBridgeNetworkParams

	if err := c.ShouldBindJSON(&params); err != nil {
		h.logger.Debug("Invalid request body", logger.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	// Create bridge network parameters
	networkParams := &network.CreateNetworkParams{
		Name:       params.Name,
		BridgeName: params.BridgeName,
		Forward: &network.NetworkForward{
			Mode: "bridge",
		},
		Autostart: params.AutoStart,
	}

	// Create the network
	if _, err := h.networkManager.Create(ctx, networkParams); err != nil {
		h.logger.Error("Failed to create bridge network",
			logger.String("name", params.Name),
			logger.String("bridge", params.BridgeName),
			logger.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	h.logger.Info("Bridge network created",
		logger.String("name", params.Name),
		logger.String("bridge", params.BridgeName))

	c.JSON(http.StatusCreated, gin.H{
		"message": "Bridge network created successfully",
		"name":    params.Name,
	})
}
