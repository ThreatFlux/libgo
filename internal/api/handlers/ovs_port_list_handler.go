package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/threatflux/libgo/internal/ovs"
	"github.com/threatflux/libgo/pkg/logger"
)

// OVSPortListHandler handles listing OVS ports
type OVSPortListHandler struct {
	ovsManager ovs.Manager
	logger     logger.Logger
}

// NewOVSPortListHandler creates a new OVSPortListHandler
func NewOVSPortListHandler(ovsManager ovs.Manager, logger logger.Logger) *OVSPortListHandler {
	return &OVSPortListHandler{
		ovsManager: ovsManager,
		logger:     logger,
	}
}

// Handle implements Handler interface
func (h *OVSPortListHandler) Handle(c *gin.Context) {
	ctx := c.Request.Context()
	bridge := c.Param("bridge")

	if bridge == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Bridge name is required",
		})
		return
	}

	// List all ports on the bridge
	ports, err := h.ovsManager.ListPorts(ctx, bridge)
	if err != nil {
		h.logger.Error("Failed to list OVS ports",
			logger.String("bridge", bridge),
			logger.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to list ports",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"ports": ports,
		"count": len(ports),
	})
}
