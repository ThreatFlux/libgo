package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	vmmodels "github.com/threatflux/libgo/internal/models/vm"
	"github.com/threatflux/libgo/pkg/logger"
)

// CreateSnapshotResponse represents the response for a VM snapshot creation request
type CreateSnapshotResponse struct {
	Snapshot *vmmodels.Snapshot `json:"snapshot"`
}

// CreateSnapshot handles requests to create a new VM snapshot
func (h *VMHandler) CreateSnapshot(c *gin.Context) {
	// Get VM name from URL path
	vmName := c.Param("name")
	if vmName == "" {
		HandleError(c, ErrInvalidInput)
		return
	}

	// Get context logger
	contextLogger := getContextLogger(c, h.logger)
	contextLogger = contextLogger.WithFields(logger.String("vmName", vmName))

	// Parse and validate request body
	var params vmmodels.SnapshotParams
	if err := c.ShouldBindJSON(&params); err != nil {
		contextLogger.Warn("Invalid snapshot creation request",
			logger.Error(err))
		HandleError(c, ErrInvalidInput)
		return
	}

	// Validate parameters
	if params.Name == "" {
		contextLogger.Warn("Invalid snapshot parameters - missing name")
		HandleError(c, ErrInvalidInput)
		return
	}

	// Create the snapshot
	snapshot, err := h.vmManager.CreateSnapshot(c.Request.Context(), vmName, params)
	if err != nil {
		contextLogger.Error("Failed to create snapshot",
			logger.String("snapshotName", params.Name),
			logger.Error(err))
		HandleError(c, err)
		return
	}

	// Log success
	contextLogger.Info("Snapshot created successfully",
		logger.String("snapshotName", snapshot.Name))

	// Return response
	c.JSON(http.StatusCreated, CreateSnapshotResponse{
		Snapshot: snapshot,
	})
}
