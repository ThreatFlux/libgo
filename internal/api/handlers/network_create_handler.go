package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/threatflux/libgo/internal/libvirt/network"
	"github.com/threatflux/libgo/pkg/logger"
)

// NetworkCreateHandler handles network creation
type NetworkCreateHandler struct {
	networkManager network.Manager
	logger         logger.Logger
}

// NewNetworkCreateHandler creates a new NetworkCreateHandler
func NewNetworkCreateHandler(networkManager network.Manager, logger logger.Logger) *NetworkCreateHandler {
	return &NetworkCreateHandler{
		networkManager: networkManager,
		logger:         logger,
	}
}

// Handle implements Handler interface
func (h *NetworkCreateHandler) Handle(c *gin.Context) {
	ctx := c.Request.Context()

	// Parse request body
	var params network.CreateNetworkParams
	if err := c.ShouldBindJSON(&params); err != nil {
		h.logger.Debug("Invalid request body", logger.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request body",
		})
		return
	}

	// Validate required fields
	if params.Name == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Network name is required",
		})
		return
	}

	// Set defaults if not provided
	if params.BridgeName == "" {
		params.BridgeName = "virbr-" + params.Name
	}

	if params.Forward == nil {
		params.Forward = &network.NetworkForward{
			Mode: "nat",
		}
	}

	if params.IP == nil {
		// Default to 192.168.100.0/24 with DHCP
		params.IP = &network.NetworkIP{
			Address: "192.168.100.1",
			Netmask: "255.255.255.0",
			DHCP: &network.NetworkDHCPInfo{
				Enabled: true,
			},
		}
	}

	// Create the network
	networkInfo, err := h.networkManager.Create(ctx, &params)
	if err != nil {
		h.logger.Error("Failed to create network",
			logger.String("name", params.Name),
			logger.Error(err))

		// Check for specific errors
		if err.Error() == "network already exists: "+params.Name {
			c.JSON(http.StatusConflict, gin.H{
				"error": "Network already exists",
			})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to create network",
		})
		return
	}

	h.logger.Info("Network created successfully",
		logger.String("name", networkInfo.Name),
		logger.String("uuid", networkInfo.UUID))

	// Return the created network info
	c.JSON(http.StatusCreated, networkInfo)
}
