package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/threatflux/libgo/internal/libvirt/storage"
	"github.com/threatflux/libgo/pkg/logger"
)

// StorageDeleteHandler handles deleting storage pools
type StorageDeleteHandler struct {
	poolManager storage.PoolManager
	logger      logger.Logger
}

// NewStorageDeleteHandler creates a new storage delete handler
func NewStorageDeleteHandler(poolManager storage.PoolManager, logger logger.Logger) *StorageDeleteHandler {
	return &StorageDeleteHandler{
		poolManager: poolManager,
		logger:      logger,
	}
}

// Handle handles the storage delete request
func (h *StorageDeleteHandler) Handle(c *gin.Context) {
	ctx := c.Request.Context()

	// Get pool name from URL parameter
	poolName := c.Param("name")
	if poolName == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Pool name is required",
		})
		return
	}

	// Delete storage pool
	err := h.poolManager.Delete(ctx, poolName)
	if err != nil {
		if err == storage.ErrPoolNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Storage pool not found",
			})
			return
		}
		h.logger.Error("Failed to delete storage pool",
			logger.String("name", poolName),
			logger.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to delete storage pool",
		})
		return
	}

	c.Status(http.StatusNoContent)
}
