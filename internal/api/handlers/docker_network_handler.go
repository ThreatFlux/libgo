package handlers

import (
	"net/http"

	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/network"
	"github.com/gin-gonic/gin"
	dockernetwork "github.com/threatflux/libgo/internal/docker/network"
	"github.com/threatflux/libgo/pkg/logger"
)

// DockerNetworkHandler handles Docker network API requests
type DockerNetworkHandler struct {
	service dockernetwork.Service
	logger  logger.Logger
}

// NewDockerNetworkHandler creates a new Docker network handler
func NewDockerNetworkHandler(service dockernetwork.Service, logger logger.Logger) *DockerNetworkHandler {
	return &DockerNetworkHandler{
		service: service,
		logger:  logger,
	}
}

// ListNetworks handles GET /docker/networks
func (h *DockerNetworkHandler) ListNetworks(c *gin.Context) {
	ctx := c.Request.Context()

	networks, err := h.service.List(ctx, network.ListOptions{})
	if err != nil {
		h.logger.Error("Failed to list networks", logger.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list networks"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"networks": networks,
		"count":    len(networks),
	})
}

// CreateNetwork handles POST /docker/networks
func (h *DockerNetworkHandler) CreateNetwork(c *gin.Context) {
	ctx := c.Request.Context()

	var req struct {
		Options    map[string]string `json:"options"`
		Labels     map[string]string `json:"labels"`
		IPAM       *network.IPAM     `json:"ipam"`
		Name       string            `json:"name" binding:"required"`
		Driver     string            `json:"driver"`
		Internal   bool              `json:"internal"`
		Attachable bool              `json:"attachable"`
		EnableIPv6 bool              `json:"enable_ipv6"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	createOptions := network.CreateOptions{
		Driver:     req.Driver,
		Options:    req.Options,
		IPAM:       req.IPAM,
		Internal:   req.Internal,
		Attachable: req.Attachable,
		EnableIPv6: &req.EnableIPv6,
		Labels:     req.Labels,
	}

	response, err := h.service.Create(ctx, req.Name, createOptions)
	if err != nil {
		h.logger.Error("Failed to create network",
			logger.String("name", req.Name),
			logger.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create network"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message":    "Network created",
		"network_id": response.ID,
		"warnings":   response.Warning,
	})
}

// InspectNetwork handles GET /docker/networks/:id
func (h *DockerNetworkHandler) InspectNetwork(c *gin.Context) {
	ctx := c.Request.Context()
	networkID := c.Param("id")

	networkInfo, err := h.service.Inspect(ctx, networkID, network.InspectOptions{})
	if err != nil {
		h.logger.Error("Failed to inspect network",
			logger.String("network_id", networkID),
			logger.Error(err))
		c.JSON(http.StatusNotFound, gin.H{"error": "Network not found"})
		return
	}

	c.JSON(http.StatusOK, networkInfo)
}

// RemoveNetwork handles DELETE /docker/networks/:id
func (h *DockerNetworkHandler) RemoveNetwork(c *gin.Context) {
	ctx := c.Request.Context()
	networkID := c.Param("id")

	if err := h.service.Remove(ctx, networkID); err != nil {
		h.logger.Error("Failed to remove network",
			logger.String("network_id", networkID),
			logger.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to remove network"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":    "Network removed",
		"network_id": networkID,
	})
}

// ConnectContainer handles POST /docker/networks/:id/connect
func (h *DockerNetworkHandler) ConnectContainer(c *gin.Context) {
	ctx := c.Request.Context()
	networkID := c.Param("id")

	var req struct {
		EndpointSettings *network.EndpointSettings `json:"endpoint_config"`
		Container        string                    `json:"container" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.service.Connect(ctx, networkID, req.Container, req.EndpointSettings); err != nil {
		h.logger.Error("Failed to connect container to network",
			logger.String("network_id", networkID),
			logger.String("container", req.Container),
			logger.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to connect container"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":      "Container connected to network",
		"network_id":   networkID,
		"container_id": req.Container,
	})
}

// DisconnectContainer handles POST /docker/networks/:id/disconnect
func (h *DockerNetworkHandler) DisconnectContainer(c *gin.Context) {
	ctx := c.Request.Context()
	networkID := c.Param("id")

	var req struct {
		Container string `json:"container" binding:"required"`
		Force     bool   `json:"force"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.service.Disconnect(ctx, networkID, req.Container, req.Force); err != nil {
		h.logger.Error("Failed to disconnect container from network",
			logger.String("network_id", networkID),
			logger.String("container", req.Container),
			logger.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to disconnect container"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":      "Container disconnected from network",
		"network_id":   networkID,
		"container_id": req.Container,
	})
}

// PruneNetworks handles POST /docker/networks/prune
func (h *DockerNetworkHandler) PruneNetworks(c *gin.Context) {
	ctx := c.Request.Context()

	report, err := h.service.Prune(ctx, filters.Args{})
	if err != nil {
		h.logger.Error("Failed to prune networks", logger.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to prune networks"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":          "Networks pruned",
		"networks_deleted": report.NetworksDeleted,
	})
}
