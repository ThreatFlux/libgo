package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/threatflux/libgo/internal/libvirt/storage"
	"github.com/threatflux/libgo/pkg/logger"
)

// StorageCreateHandler handles creating storage pools
type StorageCreateHandler struct {
	poolManager storage.PoolManager
	logger      logger.Logger
}

// NewStorageCreateHandler creates a new storage create handler
func NewStorageCreateHandler(poolManager storage.PoolManager, logger logger.Logger) *StorageCreateHandler {
	return &StorageCreateHandler{
		poolManager: poolManager,
		logger:      logger,
	}
}

// Handle handles the storage create request
func (h *StorageCreateHandler) Handle(c *gin.Context) {
	ctx := c.Request.Context()

	// Parse request body
	var params storage.CreatePoolParams
	if err := c.ShouldBindJSON(&params); err != nil {
		h.logger.Error("Failed to parse create pool request", logger.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request body",
		})
		return
	}

	// Create storage pool
	poolInfo, err := h.poolManager.Create(ctx, &params)
	if err != nil {
		if err == storage.ErrPoolExists {
			c.JSON(http.StatusConflict, gin.H{
				"error": "Storage pool already exists",
			})
			return
		}
		h.logger.Error("Failed to create storage pool",
			logger.String("name", params.Name),
			logger.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to create storage pool",
		})
		return
	}

	c.JSON(http.StatusCreated, poolInfo)
}
