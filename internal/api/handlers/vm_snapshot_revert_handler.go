package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/threatflux/libgo/pkg/logger"
)

// RevertSnapshotResponse represents the response for a VM snapshot revert request
type RevertSnapshotResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

// RevertSnapshot handles requests to revert a VM to a snapshot
func (h *VMHandler) RevertSnapshot(c *gin.Context) {
	// Get VM name and snapshot name from URL path
	vmName := c.Param("name")
	snapshotName := c.Param("snapshot")
	
	if vmName == "" || snapshotName == "" {
		HandleError(c, ErrInvalidInput)
		return
	}

	// Get context logger
	contextLogger := getContextLogger(c, h.logger)
	contextLogger = contextLogger.WithFields(
		logger.String("vmName", vmName),
		logger.String("snapshotName", snapshotName))

	// Revert to the snapshot
	if err := h.vmManager.RevertSnapshot(c.Request.Context(), vmName, snapshotName); err != nil {
		contextLogger.Error("Failed to revert to snapshot",
			logger.Error(err))
		HandleError(c, err)
		return
	}

	// Log success
	contextLogger.Info("VM reverted to snapshot successfully")

	// Return response
	c.JSON(http.StatusOK, RevertSnapshotResponse{
		Success: true,
		Message: "VM reverted to snapshot successfully",
	})
}