package handlers

import (
	"net/http"

	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/volume"
	"github.com/gin-gonic/gin"
	dockervolume "github.com/threatflux/libgo/internal/docker/volume"
	"github.com/threatflux/libgo/pkg/logger"
)

// DockerVolumeHandler handles Docker volume API requests.
type DockerVolumeHandler struct {
	service dockervolume.Service
	logger  logger.Logger
}

// NewDockerVolumeHandler creates a new Docker volume handler.
func NewDockerVolumeHandler(service dockervolume.Service, logger logger.Logger) *DockerVolumeHandler {
	return &DockerVolumeHandler{
		service: service,
		logger:  logger,
	}
}

// ListVolumes handles GET /docker/volumes.
func (h *DockerVolumeHandler) ListVolumes(c *gin.Context) {
	ctx := c.Request.Context()

	response, err := h.service.List(ctx, volume.ListOptions{})
	if err != nil {
		h.logger.Error("Failed to list volumes", logger.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list volumes"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"volumes": response.Volumes,
		"count":   len(response.Volumes),
	})
}

// CreateVolume handles POST /docker/volumes.
func (h *DockerVolumeHandler) CreateVolume(c *gin.Context) {
	ctx := c.Request.Context()

	var req struct {
		DriverOpts map[string]string `json:"driver_opts"`
		Labels     map[string]string `json:"labels"`
		Name       string            `json:"name"`
		Driver     string            `json:"driver"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	createOptions := volume.CreateOptions{
		Name:       req.Name,
		Driver:     req.Driver,
		DriverOpts: req.DriverOpts,
		Labels:     req.Labels,
	}

	vol, err := h.service.Create(ctx, createOptions)
	if err != nil {
		h.logger.Error("Failed to create volume",
			logger.String("name", req.Name),
			logger.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create volume"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Volume created",
		"volume":  vol,
	})
}

// InspectVolume handles GET /docker/volumes/:name.
func (h *DockerVolumeHandler) InspectVolume(c *gin.Context) {
	ctx := c.Request.Context()
	volumeName := c.Param("name")

	vol, err := h.service.Inspect(ctx, volumeName)
	if err != nil {
		h.logger.Error("Failed to inspect volume",
			logger.String("volume_name", volumeName),
			logger.Error(err))
		c.JSON(http.StatusNotFound, gin.H{"error": "Volume not found"})
		return
	}

	c.JSON(http.StatusOK, vol)
}

// RemoveVolume handles DELETE /docker/volumes/:name.
func (h *DockerVolumeHandler) RemoveVolume(c *gin.Context) {
	ctx := c.Request.Context()
	volumeName := c.Param("name")
	force := c.Query("force") == "true"

	if err := h.service.Remove(ctx, volumeName, force); err != nil {
		h.logger.Error("Failed to remove volume",
			logger.String("volume_name", volumeName),
			logger.Bool("force", force),
			logger.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to remove volume"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":     "Volume removed",
		"volume_name": volumeName,
	})
}

// PruneVolumes handles POST /docker/volumes/prune.
func (h *DockerVolumeHandler) PruneVolumes(c *gin.Context) {
	ctx := c.Request.Context()

	report, err := h.service.Prune(ctx, filters.Args{})
	if err != nil {
		h.logger.Error("Failed to prune volumes", logger.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to prune volumes"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":         "Volumes pruned",
		"volumes_deleted": report.VolumesDeleted,
		"space_reclaimed": report.SpaceReclaimed,
	})
}
