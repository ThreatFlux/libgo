package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/threatflux/libgo/internal/libvirt/storage"
	"github.com/threatflux/libgo/pkg/logger"
)

// StorageStartHandler handles starting storage pools
type StorageStartHandler struct {
	poolManager storage.PoolManager
	logger      logger.Logger
}

// NewStorageStartHandler creates a new storage start handler
func NewStorageStartHandler(poolManager storage.PoolManager, logger logger.Logger) *StorageStartHandler {
	return &StorageStartHandler{
		poolManager: poolManager,
		logger:      logger,
	}
}

// Handle handles the storage start request
func (h *StorageStartHandler) Handle(c *gin.Context) {
	ctx := c.Request.Context()

	// Get pool name from URL parameter
	poolName := c.Param("name")
	if poolName == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Pool name is required",
		})
		return
	}

	// Start storage pool
	err := h.poolManager.Start(ctx, poolName)
	if err != nil {
		if err == storage.ErrPoolNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Storage pool not found",
			})
			return
		}
		h.logger.Error("Failed to start storage pool",
			logger.String("name", poolName),
			logger.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to start storage pool",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Storage pool started successfully",
	})
}
