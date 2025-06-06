package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/threatflux/libgo/internal/libvirt/storage"
	"github.com/threatflux/libgo/pkg/logger"
)

// StorageGetHandler handles getting storage pool details.
type StorageGetHandler struct {
	poolManager storage.PoolManager
	logger      logger.Logger
}

// NewStorageGetHandler creates a new storage get handler.
func NewStorageGetHandler(poolManager storage.PoolManager, logger logger.Logger) *StorageGetHandler {
	return &StorageGetHandler{
		poolManager: poolManager,
		logger:      logger,
	}
}

// Handle handles the storage get request.
func (h *StorageGetHandler) Handle(c *gin.Context) {
	ctx := c.Request.Context()

	// Get pool name from URL parameter
	poolName := c.Param("name")
	if poolName == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Pool name is required",
		})
		return
	}

	// Get storage pool info
	poolInfo, err := h.poolManager.GetInfo(ctx, poolName)
	if err != nil {
		if err == storage.ErrPoolNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Storage pool not found",
			})
			return
		}
		h.logger.Error("Failed to get storage pool info",
			logger.String("name", poolName),
			logger.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to get storage pool info",
		})
		return
	}

	c.JSON(http.StatusOK, poolInfo)
}
