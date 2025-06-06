package handlers

import (
	"net/http"

	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/image"
	"github.com/gin-gonic/gin"
	dockerimage "github.com/threatflux/libgo/internal/docker/image"
	"github.com/threatflux/libgo/pkg/logger"
)

const trueString = "true"

// DockerImageHandler handles Docker image API requests.
type DockerImageHandler struct {
	service dockerimage.Service
	logger  logger.Logger
}

// NewDockerImageHandler creates a new Docker image handler.
func NewDockerImageHandler(service dockerimage.Service, logger logger.Logger) *DockerImageHandler {
	return &DockerImageHandler{
		service: service,
		logger:  logger,
	}
}

// ListImages handles GET /docker/images.
func (h *DockerImageHandler) ListImages(c *gin.Context) {
	ctx := c.Request.Context()

	images, err := h.service.List(ctx, image.ListOptions{})
	if err != nil {
		h.logger.Error("Failed to list images", logger.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list images"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"images": images,
		"count":  len(images),
	})
}

// PullImage handles POST /docker/images/pull.
func (h *DockerImageHandler) PullImage(c *gin.Context) {
	ctx := c.Request.Context()

	var req struct {
		Image string `json:"image" binding:"required"`
		Tag   string `json:"tag"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	imageName := req.Image
	if req.Tag != "" {
		imageName += ":" + req.Tag
	}

	reader, err := h.service.Pull(ctx, imageName, image.PullOptions{})
	if err != nil {
		h.logger.Error("Failed to pull image",
			logger.String("image", imageName),
			logger.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to pull image"})
		return
	}
	defer reader.Close()

	c.JSON(http.StatusOK, gin.H{
		"message": "Image pull started",
		"image":   imageName,
	})
}

// RemoveImage handles DELETE /docker/images/:id.
func (h *DockerImageHandler) RemoveImage(c *gin.Context) {
	ctx := c.Request.Context()
	imageID := c.Param("id")

	removeOptions := image.RemoveOptions{
		Force:         c.Query("force") == trueString,
		PruneChildren: c.Query("prune") == trueString,
	}

	deleteResponses, err := h.service.Remove(ctx, imageID, removeOptions)
	if err != nil {
		h.logger.Error("Failed to remove image",
			logger.String("image_id", imageID),
			logger.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to remove image"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":        "Image removed",
		"deleted_images": deleteResponses,
		"deleted_count":  len(deleteResponses),
	})
}

// InspectImage handles GET /docker/images/:id.
func (h *DockerImageHandler) InspectImage(c *gin.Context) {
	ctx := c.Request.Context()
	imageID := c.Param("id")

	imageInfo, err := h.service.Inspect(ctx, imageID)
	if err != nil {
		h.logger.Error("Failed to inspect image",
			logger.String("image_id", imageID),
			logger.Error(err))
		c.JSON(http.StatusNotFound, gin.H{"error": "Image not found"})
		return
	}

	c.JSON(http.StatusOK, imageInfo)
}

// PruneImages handles POST /docker/images/prune.
func (h *DockerImageHandler) PruneImages(c *gin.Context) {
	ctx := c.Request.Context()

	report, err := h.service.Prune(ctx, filters.Args{})
	if err != nil {
		h.logger.Error("Failed to prune images", logger.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to prune images"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":         "Images pruned",
		"deleted_images":  report.ImagesDeleted,
		"space_reclaimed": report.SpaceReclaimed,
	})
}
