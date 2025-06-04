package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/threatflux/libgo/internal/libvirt/network"
	"github.com/threatflux/libgo/pkg/logger"
)

// BridgeNetworkListHandler handles listing libvirt bridge networks
type BridgeNetworkListHandler struct {
	networkManager network.Manager
	logger         logger.Logger
}

// NewBridgeNetworkListHandler creates a new BridgeNetworkListHandler
func NewBridgeNetworkListHandler(networkManager network.Manager, logger logger.Logger) *BridgeNetworkListHandler {
	return &BridgeNetworkListHandler{
		networkManager: networkManager,
		logger:         logger,
	}
}

// BridgeNetworkInfo represents bridge network information
type BridgeNetworkInfo struct {
	Name        string `json:"name"`
	BridgeName  string `json:"bridge_name"`
	ForwardMode string `json:"forward_mode"`
	Active      bool   `json:"active"`
	AutoStart   bool   `json:"auto_start"`
}

// Handle implements Handler interface
func (h *BridgeNetworkListHandler) Handle(c *gin.Context) {
	ctx := c.Request.Context()

	// Get all networks
	networks, err := h.networkManager.List(ctx)
	if err != nil {
		h.logger.Error("Failed to list networks", logger.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	// Filter for bridge networks only
	var bridgeNetworks []BridgeNetworkInfo
	for _, net := range networks {
		if net.Forward.Mode == "bridge" {
			bridgeNetworks = append(bridgeNetworks, BridgeNetworkInfo{
				Name:        net.Name,
				BridgeName:  net.BridgeName,
				Active:      net.Active,
				AutoStart:   net.Autostart,
				ForwardMode: net.Forward.Mode,
			})
		}
	}

	h.logger.Debug("Listed bridge networks", logger.Int("count", len(bridgeNetworks)))

	c.JSON(http.StatusOK, gin.H{
		"networks": bridgeNetworks,
		"count":    len(bridgeNetworks),
	})
}
