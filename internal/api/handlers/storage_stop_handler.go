package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/threatflux/libgo/internal/libvirt/storage"
	"github.com/threatflux/libgo/pkg/logger"
)

// StorageStopHandler handles stopping storage pools
type StorageStopHandler struct {
	poolManager storage.PoolManager
	logger      logger.Logger
}

// NewStorageStopHandler creates a new storage stop handler
func NewStorageStopHandler(poolManager storage.PoolManager, logger logger.Logger) *StorageStopHandler {
	return &StorageStopHandler{
		poolManager: poolManager,
		logger:      logger,
	}
}

// Handle handles the storage stop request
func (h *StorageStopHandler) Handle(c *gin.Context) {
	ctx := c.Request.Context()

	// Get pool name from URL parameter
	poolName := c.Param("name")
	if poolName == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Pool name is required",
		})
		return
	}

	// Stop storage pool
	err := h.poolManager.Stop(ctx, poolName)
	if err != nil {
		if err == storage.ErrPoolNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Storage pool not found",
			})
			return
		}
		h.logger.Error("Failed to stop storage pool",
			logger.String("name", poolName),
			logger.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to stop storage pool",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Storage pool stopped successfully",
	})
}
