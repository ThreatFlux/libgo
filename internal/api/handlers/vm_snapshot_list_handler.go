package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	vmmodels "github.com/threatflux/libgo/internal/models/vm"
	"github.com/threatflux/libgo/pkg/logger"
)

// ListSnapshotsResponse represents the response for listing VM snapshots.
type ListSnapshotsResponse struct {
	Snapshots []*vmmodels.Snapshot `json:"snapshots"`
	Count     int                  `json:"count"`
}

// ListSnapshots handles requests to list VM snapshots.
func (h *VMHandler) ListSnapshots(c *gin.Context) {
	// Get VM name from URL path
	vmName := c.Param("name")
	if vmName == "" {
		HandleError(c, ErrInvalidInput)
		return
	}

	// Get context logger
	contextLogger := getContextLogger(c, h.logger)
	contextLogger = contextLogger.WithFields(logger.String("vmName", vmName))

	// Parse query parameters
	opts := vmmodels.SnapshotListOptions{
		IncludeMetadata: false,
		Tree:            false,
	}

	// Check if metadata should be included
	if includeMetadata := c.Query("include_metadata"); includeMetadata == "true" {
		opts.IncludeMetadata = true
	}

	// Check if tree structure is requested
	if tree := c.Query("tree"); tree == "true" {
		opts.Tree = true
	}

	// List snapshots
	snapshots, err := h.vmManager.ListSnapshots(c.Request.Context(), vmName, opts)
	if err != nil {
		contextLogger.Error("Failed to list snapshots",
			logger.Error(err))
		HandleError(c, err)
		return
	}

	// Log success
	contextLogger.Info("Listed snapshots successfully",
		logger.Int("count", len(snapshots)))

	// Return response
	c.JSON(http.StatusOK, ListSnapshotsResponse{
		Snapshots: snapshots,
		Count:     len(snapshots),
	})
}

// GetSnapshot handles requests to get information about a specific VM snapshot
func (h *VMHandler) GetSnapshot(c *gin.Context) {
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

	// Get snapshot information
	snapshot, err := h.vmManager.GetSnapshot(c.Request.Context(), vmName, snapshotName)
	if err != nil {
		contextLogger.Error("Failed to get snapshot",
			logger.Error(err))
		HandleError(c, err)
		return
	}

	// Log success
	contextLogger.Info("Retrieved snapshot information successfully")

	// Return response
	c.JSON(http.StatusOK, gin.H{
		"snapshot": snapshot,
	})
}
