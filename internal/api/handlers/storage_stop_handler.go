package handlers

import (
	"github.com/gin-gonic/gin"
	"github.com/threatflux/libgo/internal/libvirt/storage"
	"github.com/threatflux/libgo/pkg/logger"
)

// StorageStopHandler handles stopping storage pools.
type StorageStopHandler struct {
	poolManager storage.PoolManager
	logger      logger.Logger
}

// NewStorageStopHandler creates a new storage stop handler.
func NewStorageStopHandler(poolManager storage.PoolManager, logger logger.Logger) *StorageStopHandler {
	return &StorageStopHandler{
		poolManager: poolManager,
		logger:      logger,
	}
}

// Handle handles the storage stop request.
func (h *StorageStopHandler) Handle(c *gin.Context) {
	storageOperationHandler(c, h.poolManager, h.logger, h.poolManager.Stop, "stop")
}
