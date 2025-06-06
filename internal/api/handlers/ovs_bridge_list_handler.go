package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/threatflux/libgo/internal/ovs"
	"github.com/threatflux/libgo/pkg/logger"
)

// OVSBridgeListHandler handles listing OVS bridges.
type OVSBridgeListHandler struct {
	ovsManager ovs.Manager
	logger     logger.Logger
}

// NewOVSBridgeListHandler creates a new OVSBridgeListHandler.
func NewOVSBridgeListHandler(ovsManager ovs.Manager, logger logger.Logger) *OVSBridgeListHandler {
	return &OVSBridgeListHandler{
		ovsManager: ovsManager,
		logger:     logger,
	}
}

// Handle implements Handler interface.
func (h *OVSBridgeListHandler) Handle(c *gin.Context) {
	ctx := c.Request.Context()

	// List all bridges
	bridges, err := h.ovsManager.ListBridges(ctx)
	if err != nil {
		h.logger.Error("Failed to list OVS bridges", logger.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to list bridges",
		})
		return
	}

	// Ensure bridges is not nil
	if bridges == nil {
		bridges = []ovs.BridgeInfo{}
	}

	c.JSON(http.StatusOK, gin.H{
		"bridges": bridges,
		"count":   len(bridges),
	})
}
