package handlers

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/threatflux/libgo/internal/libvirt/storage"
	"github.com/threatflux/libgo/pkg/logger"
)

// storageOperationHandler is a generic handler for storage pool start/stop operations
func storageOperationHandler(c *gin.Context, poolManager storage.PoolManager, log logger.Logger,
	operation func(context.Context, string) error, operationName string) {
	ctx := c.Request.Context()

	// Get pool name from URL parameter
	poolName := c.Param("name")
	if poolName == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Pool name is required",
		})
		return
	}

	// Execute the operation
	err := operation(ctx, poolName)
	if err != nil {
		if err == storage.ErrPoolNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Storage pool not found",
			})
			return
		}
		log.Error("Failed to "+operationName+" storage pool",
			logger.String("name", poolName),
			logger.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to " + operationName + " storage pool",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Storage pool " + operationName + "ped successfully",
	})
}

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
	storageOperationHandler(c, h.poolManager, h.logger, h.poolManager.Start, "start")
}
