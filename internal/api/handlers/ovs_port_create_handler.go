package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/threatflux/libgo/internal/ovs"
	"github.com/threatflux/libgo/pkg/logger"
)

// OVSPortCreateHandler handles creating OVS ports.
type OVSPortCreateHandler struct {
	ovsManager ovs.Manager
	logger     logger.Logger
}

// NewOVSPortCreateHandler creates a new OVSPortCreateHandler.
func NewOVSPortCreateHandler(ovsManager ovs.Manager, logger logger.Logger) *OVSPortCreateHandler {
	return &OVSPortCreateHandler{
		ovsManager: ovsManager,
		logger:     logger,
	}
}

// Handle implements Handler interface.
func (h *OVSPortCreateHandler) Handle(c *gin.Context) {
	ctx := c.Request.Context()

	var params ovs.CreatePortParams
	if err := c.ShouldBindJSON(&params); err != nil {
		h.logger.Debug("Invalid request body", logger.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request body",
		})
		return
	}

	// Build port options
	options := &ovs.PortOptions{
		Type:        params.Type,
		Tag:         params.Tag,
		Trunks:      params.Trunks,
		PeerPort:    params.PeerPort,
		RemoteIP:    params.RemoteIP,
		TunnelType:  params.TunnelType,
		ExternalIDs: params.ExternalIDs,
		OtherConfig: params.OtherConfig,
	}

	// Add the port
	if err := h.ovsManager.AddPort(ctx, params.Bridge, params.Name, options); err != nil {
		h.logger.Error("Failed to create OVS port",
			logger.String("name", params.Name),
			logger.String("bridge", params.Bridge),
			logger.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	// Get port info
	port, err := h.ovsManager.GetPort(ctx, params.Bridge, params.Name)
	if err != nil {
		h.logger.Error("Failed to get port info after creation",
			logger.String("name", params.Name),
			logger.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Port created but failed to get info",
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"port": port,
	})
}
