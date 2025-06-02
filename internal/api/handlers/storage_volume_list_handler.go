package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/threatflux/libgo/internal/libvirt/storage"
	"github.com/threatflux/libgo/pkg/logger"
)

// StorageVolumeListHandler handles listing storage volumes
type StorageVolumeListHandler struct {
	volumeManager storage.VolumeManager
	logger        logger.Logger
}

// NewStorageVolumeListHandler creates a new storage volume list handler
func NewStorageVolumeListHandler(volumeManager storage.VolumeManager, logger logger.Logger) *StorageVolumeListHandler {
	return &StorageVolumeListHandler{
		volumeManager: volumeManager,
		logger:        logger,
	}
}

// Handle handles the storage volume list request
func (h *StorageVolumeListHandler) Handle(c *gin.Context) {
	ctx := c.Request.Context()

	// Get pool name from URL parameter
	poolName := c.Param("poolName")
	if poolName == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Pool name is required",
		})
		return
	}

	// List volumes in the pool
	volumes, err := h.volumeManager.List(ctx, poolName)
	if err != nil {
		h.logger.Error("Failed to list storage volumes",
			logger.String("pool", poolName),
			logger.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to list storage volumes",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"volumes": volumes,
	})
}
