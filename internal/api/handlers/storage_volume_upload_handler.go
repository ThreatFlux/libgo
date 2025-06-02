package handlers

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/threatflux/libgo/internal/libvirt/storage"
	"github.com/threatflux/libgo/pkg/logger"
)

// StorageVolumeUploadHandler handles uploading data to storage volumes
type StorageVolumeUploadHandler struct {
	volumeManager storage.VolumeManager
	logger        logger.Logger
}

// NewStorageVolumeUploadHandler creates a new storage volume upload handler
func NewStorageVolumeUploadHandler(volumeManager storage.VolumeManager, logger logger.Logger) *StorageVolumeUploadHandler {
	return &StorageVolumeUploadHandler{
		volumeManager: volumeManager,
		logger:        logger,
	}
}

// Handle handles the storage volume upload request
func (h *StorageVolumeUploadHandler) Handle(c *gin.Context) {
	ctx := c.Request.Context()

	// Get pool and volume names from URL parameters
	poolName := c.Param("name")
	volumeName := c.Param("volumeName")

	if poolName == "" || volumeName == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Pool name and volume name are required",
		})
		return
	}

	// Check if this is a multipart upload
	if strings.HasPrefix(c.GetHeader("Content-Type"), "multipart/form-data") {
		// Handle multipart file upload
		file, header, err := c.Request.FormFile("file")
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Failed to get uploaded file",
			})
			return
		}
		defer file.Close()

		// Upload data to volume
		err = h.volumeManager.Upload(ctx, poolName, volumeName, file)
		if err != nil {
			if err == storage.ErrVolumeNotFound {
				c.JSON(http.StatusNotFound, gin.H{
					"error": "Storage volume not found",
				})
				return
			}
			h.logger.Error("Failed to upload to storage volume",
				logger.String("pool", poolName),
				logger.String("volume", volumeName),
				logger.String("filename", header.Filename),
				logger.Int64("size", header.Size),
				logger.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to upload to storage volume",
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message":  "File uploaded successfully",
			"pool":     poolName,
			"volume":   volumeName,
			"filename": header.Filename,
			"bytes":    header.Size,
		})
	} else {
		// Handle raw data upload
		contentLength := c.Request.ContentLength
		if contentLength <= 0 {
			// Try to get from header
			if clHeader := c.GetHeader("Content-Length"); clHeader != "" {
				if cl, err := strconv.ParseInt(clHeader, 10, 64); err == nil {
					contentLength = cl
				}
			}
		}

		// Upload data to volume
		err := h.volumeManager.Upload(ctx, poolName, volumeName, c.Request.Body)
		if err != nil {
			if err == storage.ErrVolumeNotFound {
				c.JSON(http.StatusNotFound, gin.H{
					"error": "Storage volume not found",
				})
				return
			}
			h.logger.Error("Failed to upload to storage volume",
				logger.String("pool", poolName),
				logger.String("volume", volumeName),
				logger.Int64("contentLength", contentLength),
				logger.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to upload to storage volume",
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message": "Data uploaded successfully",
			"pool":    poolName,
			"volume":  volumeName,
			"bytes":   contentLength,
		})
	}
}
