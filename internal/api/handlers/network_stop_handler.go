package handlers

import (
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
	networkOperationHandler(c, h.networkManager, h.logger, h.networkManager.Stop, "stop")
}
