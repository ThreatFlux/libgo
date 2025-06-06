package handlers

import (
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/docker/docker/api/types/container"
	"github.com/gin-gonic/gin"
	containerSvc "github.com/threatflux/libgo/internal/docker/container"
	"github.com/threatflux/libgo/pkg/logger"
)

// DockerContainerHandler handles Docker container operations.
type DockerContainerHandler struct {
	containerService containerSvc.Service
	logger           logger.Logger
}

// NewDockerContainerHandler creates a new Docker container handler.
func NewDockerContainerHandler(containerService containerSvc.Service, logger logger.Logger) *DockerContainerHandler {
	return &DockerContainerHandler{
		containerService: containerService,
		logger:           logger,
	}
}

// CreateContainerRequest represents the request to create a container.
type CreateContainerRequest struct {
	// Struct fields (need to be first for alignment).
	HostConfig container.HostConfig `json:"host_config,omitempty"`
	// Slice fields (24 bytes each).
	Cmd        []string `json:"cmd,omitempty"`
	Entrypoint []string `json:"entrypoint,omitempty"`
	Env        []string `json:"env,omitempty"`
	// Map fields (24 bytes each).
	Labels  map[string]string   `json:"labels,omitempty"`
	Volumes map[string]struct{} `json:"volumes,omitempty"`
	// String fields (16 bytes each).
	Name       string `json:"name" binding:"required"`
	Image      string `json:"image" binding:"required"`
	WorkingDir string `json:"working_dir,omitempty"`
	User       string `json:"user,omitempty"`
}

// ContainerResponse represents a container in responses.
type ContainerResponse struct {
	// Slice fields (24 bytes).
	Ports []PortMapping `json:"ports,omitempty"`
	// Map fields (24 bytes).
	Labels map[string]string `json:"labels,omitempty"`
	// Int64 fields (8 bytes).
	Created int64 `json:"created"`
	// String fields (16 bytes each).
	ID     string `json:"id"`
	Name   string `json:"name"`
	Image  string `json:"image"`
	State  string `json:"state"`
	Status string `json:"status"`
}

// PortMapping represents a port mapping.
type PortMapping struct {
	Type    string `json:"type"`
	IP      string `json:"ip,omitempty"`
	Private uint16 `json:"private"`
	Public  uint16 `json:"public"`
}

// CreateContainer creates a new Docker container..
func (h *DockerContainerHandler) CreateContainer(c *gin.Context) {
	contextLogger := h.logger.WithFields(
		logger.String("handler", "docker_container_create"),
		logger.String("method", c.Request.Method),
		logger.String("path", c.Request.URL.Path),
	)

	var req CreateContainerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		contextLogger.Warn("Invalid container creation request", logger.Error(err))
		HandleError(c, ErrInvalidInput)
		return
	}

	// Build container config.
	config := &container.Config{
		Image:      req.Image,
		Cmd:        req.Cmd,
		Entrypoint: req.Entrypoint,
		Env:        req.Env,
		Labels:     req.Labels,
		Volumes:    req.Volumes,
		WorkingDir: req.WorkingDir,
		User:       req.User,
	}

	// Create container.
	id, err := h.containerService.Create(c.Request.Context(), config, &req.HostConfig, req.Name)
	if err != nil {
		contextLogger.Error("Failed to create container", logger.Error(err))
		HandleError(c, fmt.Errorf("failed to create container: %w", err))
		return
	}

	contextLogger.Info("Container created successfully",
		logger.String("container_id", id),
		logger.String("name", req.Name))

	c.JSON(http.StatusCreated, gin.H{
		"id":   id,
		"name": req.Name,
	})
}

// ListContainers lists Docker containers..
func (h *DockerContainerHandler) ListContainers(c *gin.Context) {
	contextLogger := h.logger.WithFields(
		logger.String("handler", "docker_container_list"),
		logger.String("method", c.Request.Method),
		logger.String("path", c.Request.URL.Path),
	)

	// Parse query parameters.
	all := c.DefaultQuery("all", "false") == trueString
	limit, err := strconv.Atoi(c.DefaultQuery("limit", "0"))
	if err != nil {
		limit = 0 // Default to no limit if parsing fails
	}

	options := container.ListOptions{
		All:   all,
		Limit: limit,
	}

	// Add filters if provided.
	if labelFilter := c.Query("label"); labelFilter != "" {
		options.Filters.Add("label", labelFilter)
	}
	if statusFilter := c.Query("status"); statusFilter != "" {
		options.Filters.Add("status", statusFilter)
	}

	containers, err := h.containerService.List(c.Request.Context(), options)
	if err != nil {
		contextLogger.Error("Failed to list containers", logger.Error(err))
		HandleError(c, fmt.Errorf("failed to list containers: %w", err))
		return
	}

	// Convert to response format.
	response := make([]ContainerResponse, len(containers))
	for i, cnt := range containers {
		response[i] = ContainerResponse{
			ID:      cnt.ID,
			Name:    cnt.Names[0], // Docker prefixes with "/"
			Image:   cnt.Image,
			State:   cnt.State,
			Status:  cnt.Status,
			Created: cnt.Created,
			Labels:  cnt.Labels,
		}

		// Convert ports.
		for _, port := range cnt.Ports {
			response[i].Ports = append(response[i].Ports, PortMapping{
				Private: port.PrivatePort,
				Public:  port.PublicPort,
				Type:    port.Type,
				IP:      port.IP,
			})
		}
	}

	c.JSON(http.StatusOK, response)
}

// GetContainer gets a specific container by ID.
func (h *DockerContainerHandler) GetContainer(c *gin.Context) {
	contextLogger := h.logger.WithFields(
		logger.String("handler", "docker_container_get"),
		logger.String("method", c.Request.Method),
		logger.String("path", c.Request.URL.Path),
	)

	containerID := c.Param("id")
	if containerID == "" {
		HandleError(c, ErrInvalidInput)
		return
	}

	containerJSON, err := h.containerService.Inspect(c.Request.Context(), containerID)
	if err != nil {
		contextLogger.Error("Failed to inspect container",
			logger.String("container_id", containerID),
			logger.Error(err))
		HandleError(c, fmt.Errorf("failed to inspect container: %w", err))
		return
	}

	c.JSON(http.StatusOK, containerJSON)
}

// StartContainer starts a container..
func (h *DockerContainerHandler) StartContainer(c *gin.Context) {
	contextLogger := h.logger.WithFields(
		logger.String("handler", "docker_container_start"),
		logger.String("method", c.Request.Method),
		logger.String("path", c.Request.URL.Path),
	)

	containerID := c.Param("id")
	if containerID == "" {
		HandleError(c, ErrInvalidInput)
		return
	}

	if err := h.containerService.Start(c.Request.Context(), containerID); err != nil {
		contextLogger.Error("Failed to start container",
			logger.String("container_id", containerID),
			logger.Error(err))
		HandleError(c, fmt.Errorf("failed to start container: %w", err))
		return
	}

	c.JSON(http.StatusNoContent, nil)
}

// StopContainer stops a container..
func (h *DockerContainerHandler) StopContainer(c *gin.Context) {
	contextLogger := h.logger.WithFields(
		logger.String("handler", "docker_container_stop"),
		logger.String("method", c.Request.Method),
		logger.String("path", c.Request.URL.Path),
	)

	containerID := c.Param("id")
	if containerID == "" {
		HandleError(c, ErrInvalidInput)
		return
	}

	// Parse timeout from query.
	var timeout *int
	if t := c.Query("timeout"); t != "" {
		if timeoutVal, err := strconv.Atoi(t); err == nil {
			timeout = &timeoutVal
		}
	}

	if err := h.containerService.Stop(c.Request.Context(), containerID, timeout); err != nil {
		contextLogger.Error("Failed to stop container",
			logger.String("container_id", containerID),
			logger.Error(err))
		HandleError(c, fmt.Errorf("failed to stop container: %w", err))
		return
	}

	c.JSON(http.StatusNoContent, nil)
}

// RestartContainer restarts a container..
func (h *DockerContainerHandler) RestartContainer(c *gin.Context) {
	contextLogger := h.logger.WithFields(
		logger.String("handler", "docker_container_restart"),
		logger.String("method", c.Request.Method),
		logger.String("path", c.Request.URL.Path),
	)

	containerID := c.Param("id")
	if containerID == "" {
		HandleError(c, ErrInvalidInput)
		return
	}

	// Parse timeout from query.
	var timeout *int
	if t := c.Query("timeout"); t != "" {
		if timeoutVal, err := strconv.Atoi(t); err == nil {
			timeout = &timeoutVal
		}
	}

	if err := h.containerService.Restart(c.Request.Context(), containerID, timeout); err != nil {
		contextLogger.Error("Failed to restart container",
			logger.String("container_id", containerID),
			logger.Error(err))
		HandleError(c, fmt.Errorf("failed to restart container: %w", err))
		return
	}

	c.JSON(http.StatusNoContent, nil)
}

// DeleteContainer deletes a container..
func (h *DockerContainerHandler) DeleteContainer(c *gin.Context) {
	contextLogger := h.logger.WithFields(
		logger.String("handler", "docker_container_delete"),
		logger.String("method", c.Request.Method),
		logger.String("path", c.Request.URL.Path),
	)

	containerID := c.Param("id")
	if containerID == "" {
		HandleError(c, ErrInvalidInput)
		return
	}

	force := c.Query("force") == trueString

	if err := h.containerService.Remove(c.Request.Context(), containerID, force); err != nil {
		contextLogger.Error("Failed to delete container",
			logger.String("container_id", containerID),
			logger.Bool("force", force),
			logger.Error(err))
		HandleError(c, fmt.Errorf("failed to delete container: %w", err))
		return
	}

	c.JSON(http.StatusNoContent, nil)
}

// GetContainerLogs gets container logs..
func (h *DockerContainerHandler) GetContainerLogs(c *gin.Context) {
	contextLogger := h.logger.WithFields(
		logger.String("handler", "docker_container_logs"),
		logger.String("method", c.Request.Method),
		logger.String("path", c.Request.URL.Path),
	)

	containerID := c.Param("id")
	if containerID == "" {
		HandleError(c, ErrInvalidInput)
		return
	}

	// Parse options.
	options := container.LogsOptions{
		ShowStdout: c.DefaultQuery("stdout", "true") == trueString,
		ShowStderr: c.DefaultQuery("stderr", "true") == trueString,
		Follow:     c.Query("follow") == trueString,
		Timestamps: c.Query("timestamps") == trueString,
	}

	// Parse tail option.
	if tail := c.Query("tail"); tail != "" {
		options.Tail = tail
	}

	// Parse since option.
	if since := c.Query("since"); since != "" {
		options.Since = since
	}

	logs, err := h.containerService.Logs(c.Request.Context(), containerID, options)
	if err != nil {
		contextLogger.Error("Failed to get container logs",
			logger.String("container_id", containerID),
			logger.Error(err))
		HandleError(c, fmt.Errorf("failed to get container logs: %w", err))
		return
	}
	defer logs.Close()

	// Stream logs to response.
	c.Header("Content-Type", "text/plain; charset=utf-8")
	c.Header("Transfer-Encoding", "chunked")

	if options.Follow {
		c.Header("X-Content-Type-Options", "nosniff")
	}

	c.Status(http.StatusOK)

	// Copy logs to response.
	if _, err := io.Copy(c.Writer, logs); err != nil {
		contextLogger.Error("Failed to stream logs",
			logger.String("container_id", containerID),
			logger.Error(err))
	}
}

// GetContainerStats gets container statistics..
func (h *DockerContainerHandler) GetContainerStats(c *gin.Context) {
	contextLogger := h.logger.WithFields(
		logger.String("handler", "docker_container_stats"),
		logger.String("method", c.Request.Method),
		logger.String("path", c.Request.URL.Path),
	)

	containerID := c.Param("id")
	if containerID == "" {
		HandleError(c, ErrInvalidInput)
		return
	}

	stream := c.Query("stream") == trueString

	stats, err := h.containerService.Stats(c.Request.Context(), containerID, stream)
	if err != nil {
		contextLogger.Error("Failed to get container stats",
			logger.String("container_id", containerID),
			logger.Error(err))
		HandleError(c, fmt.Errorf("failed to get container stats: %w", err))
		return
	}
	defer stats.Body.Close()

	// Set appropriate headers.
	c.Header("Content-Type", "application/json")
	if stream {
		c.Header("Transfer-Encoding", "chunked")
		c.Header("X-Content-Type-Options", "nosniff")
	}

	c.Status(http.StatusOK)

	// Copy stats to response.
	if _, err := io.Copy(c.Writer, stats.Body); err != nil {
		contextLogger.Error("Failed to stream stats",
			logger.String("container_id", containerID),
			logger.Error(err))
	}
}
