package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/threatflux/libgo/internal/libvirt/storage"
	"github.com/threatflux/libgo/pkg/logger"
)

// StorageVolumeCreateHandler handles creating storage volumes
type StorageVolumeCreateHandler struct {
	volumeManager storage.VolumeManager
	logger        logger.Logger
}

// NewStorageVolumeCreateHandler creates a new storage volume create handler
func NewStorageVolumeCreateHandler(volumeManager storage.VolumeManager, logger logger.Logger) *StorageVolumeCreateHandler {
	return &StorageVolumeCreateHandler{
		volumeManager: volumeManager,
		logger:        logger,
	}
}

// Handle handles the storage volume create request
func (h *StorageVolumeCreateHandler) Handle(c *gin.Context) {
	ctx := c.Request.Context()

	// Get pool name from URL parameter
	poolName := c.Param("name")
	if poolName == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Pool name is required",
		})
		return
	}

	// Parse request body
	var params storage.CreateVolumeParams
	if err := c.ShouldBindJSON(&params); err != nil {
		h.logger.Error("Failed to parse create volume request", logger.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request body",
		})
		return
	}

	// Set default format if not provided
	if params.Format == "" {
		params.Format = "qcow2"
	}

	// Create storage volume
	err := h.volumeManager.Create(ctx, poolName, params.Name, params.CapacityBytes, params.Format)
	if err != nil {
		if err == storage.ErrVolumeExists {
			c.JSON(http.StatusConflict, gin.H{
				"error": "Storage volume already exists",
			})
			return
		}
		h.logger.Error("Failed to create storage volume",
			logger.String("pool", poolName),
			logger.String("volume", params.Name),
			logger.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to create storage volume",
		})
		return
	}

	// Get volume info to return
	volumeInfo, err := h.volumeManager.GetInfo(ctx, poolName, params.Name)
	if err != nil {
		h.logger.Warn("Created volume but failed to get info",
			logger.String("pool", poolName),
			logger.String("volume", params.Name),
			logger.Error(err))
		c.JSON(http.StatusCreated, gin.H{
			"message": "Volume created successfully",
			"name":    params.Name,
		})
		return
	}

	c.JSON(http.StatusCreated, volumeInfo)
}
