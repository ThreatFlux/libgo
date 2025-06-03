package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/threatflux/libgo/internal/libvirt/network"
	"github.com/threatflux/libgo/pkg/logger"
)

// NetworkListHandler handles listing all networks
type NetworkListHandler struct {
	networkManager network.Manager
	logger         logger.Logger
}

// NewNetworkListHandler creates a new NetworkListHandler
func NewNetworkListHandler(networkManager network.Manager, logger logger.Logger) *NetworkListHandler {
	return &NetworkListHandler{
		networkManager: networkManager,
		logger:         logger,
	}
}

// Handle implements Handler interface
func (h *NetworkListHandler) Handle(c *gin.Context) {
	ctx := c.Request.Context()

	// List all networks
	networks, err := h.networkManager.List(ctx)
	if err != nil {
		h.logger.Error("Failed to list networks", logger.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to list networks",
		})
		return
	}

	// Ensure networks is not nil
	if networks == nil {
		networks = []*network.NetworkInfo{}
	}

	// Return the list
	c.JSON(http.StatusOK, gin.H{
		"networks": networks,
		"count":    len(networks),
	})
}
