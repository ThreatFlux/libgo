package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/threatflux/libgo/internal/libvirt/storage"
	"github.com/threatflux/libgo/pkg/logger"
)

// StorageVolumeDeleteHandler handles deleting storage volumes
type StorageVolumeDeleteHandler struct {
	volumeManager storage.VolumeManager
	logger        logger.Logger
}

// NewStorageVolumeDeleteHandler creates a new storage volume delete handler
func NewStorageVolumeDeleteHandler(volumeManager storage.VolumeManager, logger logger.Logger) *StorageVolumeDeleteHandler {
	return &StorageVolumeDeleteHandler{
		volumeManager: volumeManager,
		logger:        logger,
	}
}

// Handle handles the storage volume delete request
func (h *StorageVolumeDeleteHandler) Handle(c *gin.Context) {
	ctx := c.Request.Context()

	// Get pool and volume names from URL parameters
	poolName := c.Param("poolName")
	volumeName := c.Param("volumeName")

	if poolName == "" || volumeName == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Pool name and volume name are required",
		})
		return
	}

	// Delete storage volume
	err := h.volumeManager.Delete(ctx, poolName, volumeName)
	if err != nil {
		if err == storage.ErrVolumeNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Storage volume not found",
			})
			return
		}
		h.logger.Error("Failed to delete storage volume",
			logger.String("pool", poolName),
			logger.String("volume", volumeName),
			logger.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to delete storage volume",
		})
		return
	}

	c.Status(http.StatusNoContent)
}
