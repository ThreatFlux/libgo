package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/threatflux/libgo/pkg/logger"
)

// DeleteSnapshotResponse represents the response for a VM snapshot deletion request
type DeleteSnapshotResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

// DeleteSnapshot handles requests to delete a VM snapshot
func (h *VMHandler) DeleteSnapshot(c *gin.Context) {
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

	// Delete the snapshot
	if err := h.vmManager.DeleteSnapshot(c.Request.Context(), vmName, snapshotName); err != nil {
		contextLogger.Error("Failed to delete snapshot",
			logger.Error(err))
		HandleError(c, err)
		return
	}

	// Log success
	contextLogger.Info("Snapshot deleted successfully")

	// Return response
	c.JSON(http.StatusOK, DeleteSnapshotResponse{
		Success: true,
		Message: "Snapshot deleted successfully",
	})
}