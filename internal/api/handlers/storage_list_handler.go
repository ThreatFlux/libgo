package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/threatflux/libgo/internal/libvirt/storage"
	"github.com/threatflux/libgo/pkg/logger"
)

// StorageListHandler handles listing storage pools.
type StorageListHandler struct {
	poolManager storage.PoolManager
	logger      logger.Logger
}

// NewStorageListHandler creates a new storage list handler.
func NewStorageListHandler(poolManager storage.PoolManager, logger logger.Logger) *StorageListHandler {
	return &StorageListHandler{
		poolManager: poolManager,
		logger:      logger,
	}
}

// Handle handles the storage list request.
func (h *StorageListHandler) Handle(c *gin.Context) {
	ctx := c.Request.Context()

	// List storage pools
	pools, err := h.poolManager.List(ctx)
	if err != nil {
		h.logger.Error("Failed to list storage pools", logger.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to list storage pools",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"pools": pools,
	})
}
