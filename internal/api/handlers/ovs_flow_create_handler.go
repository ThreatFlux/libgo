package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/threatflux/libgo/internal/ovs"
	"github.com/threatflux/libgo/pkg/logger"
)

// OVSFlowCreateHandler handles creating OVS flow rules
type OVSFlowCreateHandler struct {
	ovsManager ovs.Manager
	logger     logger.Logger
}

// NewOVSFlowCreateHandler creates a new OVSFlowCreateHandler
func NewOVSFlowCreateHandler(ovsManager ovs.Manager, logger logger.Logger) *OVSFlowCreateHandler {
	return &OVSFlowCreateHandler{
		ovsManager: ovsManager,
		logger:     logger,
	}
}

// Handle implements Handler interface
func (h *OVSFlowCreateHandler) Handle(c *gin.Context) {
	ctx := c.Request.Context()

	var params ovs.CreateFlowParams
	if err := c.ShouldBindJSON(&params); err != nil {
		h.logger.Debug("Invalid request body", logger.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request body",
		})
		return
	}

	// Generate flow ID if not provided
	flowID := params.Cookie
	if flowID == "" {
		flowID = uuid.New().String()
	}

	// Build flow rule
	flow := &ovs.FlowRule{
		ID:       flowID,
		Table:    params.Table,
		Priority: params.Priority,
		Match:    params.Match,
		Actions:  params.Actions,
		Cookie:   flowID,
	}

	// Add the flow
	if err := h.ovsManager.AddFlow(ctx, params.Bridge, flow); err != nil {
		h.logger.Error("Failed to create OVS flow",
			logger.String("bridge", params.Bridge),
			logger.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"flow": flow,
	})
}
