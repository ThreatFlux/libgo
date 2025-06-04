package handlers

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/threatflux/libgo/internal/compute"
	apierrors "github.com/threatflux/libgo/internal/errors"
	"github.com/threatflux/libgo/pkg/logger"
)

// ComputeHandler handles unified compute instance requests.
type ComputeHandler struct {
	computeManager compute.Manager
	logger         logger.Logger
}

// NewComputeHandler creates a new compute handler.
func NewComputeHandler(computeManager compute.Manager, logger logger.Logger) *ComputeHandler {
	return &ComputeHandler{
		computeManager: computeManager,
		logger:         logger,
	}
}

// CreateInstance handles requests to create a new compute instance.
func (h *ComputeHandler) CreateInstance(c *gin.Context) {
	contextLogger := getContextLogger(c, h.logger)

	// Parse and validate request body
	var req compute.ComputeInstanceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		contextLogger.Warn("Invalid compute instance creation request", logger.Error(err))
		HandleError(c, ErrInvalidInput)
		return
	}

	// Set user ID from context
	if userID, exists := c.Get("user_id"); exists {
		if uid, ok := userID.(uint); ok {
			req.UserID = uid
		}
	}

	// Validate required fields
	if err := h.validateCreateRequest(req); err != nil {
		contextLogger.Warn("Invalid compute instance parameters",
			logger.String("name", req.Name),
			logger.String("type", string(req.Type)),
			logger.Error(err))
		HandleError(c, err)
		return
	}

	// Create the instance
	instance, err := h.computeManager.CreateInstance(c.Request.Context(), req)
	if err != nil {
		contextLogger.Error("Failed to create compute instance",
			logger.String("name", req.Name),
			logger.String("type", string(req.Type)),
			logger.String("backend", string(req.Backend)),
			logger.Error(err))
		HandleError(c, apierrors.Wrap(err, "create instance"))
		return
	}

	contextLogger.Info("Created compute instance",
		logger.String("id", instance.ID),
		logger.String("name", instance.Name),
		logger.String("type", string(instance.Type)),
		logger.String("backend", string(instance.Backend)))

	c.JSON(http.StatusCreated, gin.H{
		"instance": instance,
	})
}

// GetInstance handles requests to get a compute instance by ID.
func (h *ComputeHandler) GetInstance(c *gin.Context) {
	contextLogger := getContextLogger(c, h.logger)
	id := c.Param("id")

	if id == "" {
		contextLogger.Warn("Missing instance ID")
		HandleError(c, ErrInvalidInput)
		return
	}

	instance, err := h.computeManager.GetInstance(c.Request.Context(), id)
	if err != nil {
		contextLogger.Warn("Failed to get compute instance",
			logger.String("id", id),
			logger.Error(err))
		HandleError(c, apierrors.Wrap(err, "get instance"))
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"instance": instance,
	})
}

// GetInstanceByName handles requests to get a compute instance by name.
func (h *ComputeHandler) GetInstanceByName(c *gin.Context) {
	contextLogger := getContextLogger(c, h.logger)
	name := c.Param("name")

	if name == "" {
		contextLogger.Warn("Missing instance name")
		HandleError(c, ErrInvalidInput)
		return
	}

	instance, err := h.computeManager.GetInstanceByName(c.Request.Context(), name)
	if err != nil {
		contextLogger.Warn("Failed to get compute instance by name",
			logger.String("name", name),
			logger.Error(err))
		HandleError(c, apierrors.Wrap(err, "get instance by name"))
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"instance": instance,
	})
}

// ListInstances handles requests to list compute instances.
func (h *ComputeHandler) ListInstances(c *gin.Context) {
	contextLogger := getContextLogger(c, h.logger)

	// Parse query parameters
	opts := h.parseListOptions(c)

	// Set user ID filter from context if not admin
	if userID, exists := c.Get("user_id"); exists {
		if uid, ok := userID.(uint); ok {
			// Check if user is admin
			if roles, exists := c.Get("user_roles"); exists {
				if rolesSlice, ok := roles.([]string); ok {
					isAdmin := false
					for _, role := range rolesSlice {
						if role == "admin" {
							isAdmin = true
							break
						}
					}
					// If not admin, filter by user ID
					if !isAdmin {
						opts.UserID = uid
					}
				}
			}
		}
	}

	var instances []*compute.ComputeInstance
	var err error

	// List from all backends or specific backend
	if opts.Backend == "" {
		instances, err = h.computeManager.ListAllInstances(c.Request.Context(), opts)
	} else {
		instances, err = h.computeManager.ListInstances(c.Request.Context(), opts)
	}

	if err != nil {
		contextLogger.Error("Failed to list compute instances",
			logger.String("backend", string(opts.Backend)),
			logger.Error(err))
		HandleError(c, apierrors.Wrap(err, "list instances"))
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"instances": instances,
		"count":     len(instances),
	})
}

// UpdateInstance handles requests to update a compute instance.
func (h *ComputeHandler) UpdateInstance(c *gin.Context) {
	contextLogger := getContextLogger(c, h.logger)
	id := c.Param("id")

	if id == "" {
		contextLogger.Warn("Missing instance ID")
		HandleError(c, ErrInvalidInput)
		return
	}

	// Parse update request
	var update compute.ComputeInstanceUpdate
	if err := c.ShouldBindJSON(&update); err != nil {
		contextLogger.Warn("Invalid update request", logger.Error(err))
		HandleError(c, ErrInvalidInput)
		return
	}

	instance, err := h.computeManager.UpdateInstance(c.Request.Context(), id, update)
	if err != nil {
		contextLogger.Error("Failed to update compute instance",
			logger.String("id", id),
			logger.Error(err))
		HandleError(c, apierrors.Wrap(err, "update instance"))
		return
	}

	contextLogger.Info("Updated compute instance",
		logger.String("id", id),
		logger.String("name", instance.Name))

	c.JSON(http.StatusOK, gin.H{
		"instance": instance,
	})
}

// DeleteInstance handles requests to delete a compute instance.
func (h *ComputeHandler) DeleteInstance(c *gin.Context) {
	contextLogger := getContextLogger(c, h.logger)
	id := c.Param("id")

	if id == "" {
		contextLogger.Warn("Missing instance ID")
		HandleError(c, ErrInvalidInput)
		return
	}

	// Parse force parameter
	force := c.Query("force") == "true"

	err := h.computeManager.DeleteInstance(c.Request.Context(), id, force)
	if err != nil {
		contextLogger.Error("Failed to delete compute instance",
			logger.String("id", id),
			logger.Bool("force", force),
			logger.Error(err))
		HandleError(c, apierrors.Wrap(err, "delete instance"))
		return
	}

	contextLogger.Info("Deleted compute instance",
		logger.String("id", id),
		logger.Bool("force", force))

	c.JSON(http.StatusOK, gin.H{
		"message": "Instance deleted successfully",
	})
}

// Lifecycle operations.

// StartInstance handles requests to start a compute instance.
func (h *ComputeHandler) StartInstance(c *gin.Context) {
	h.performLifecycleAction(c, "start", func(id string) error {
		return h.computeManager.StartInstance(c.Request.Context(), id)
	})
}

// StopInstance handles requests to stop a compute instance.
func (h *ComputeHandler) StopInstance(c *gin.Context) {
	force := c.Query("force") == "true"
	h.performLifecycleAction(c, "stop", func(id string) error {
		return h.computeManager.StopInstance(c.Request.Context(), id, force)
	})
}

// RestartInstance handles requests to restart a compute instance.
func (h *ComputeHandler) RestartInstance(c *gin.Context) {
	force := c.Query("force") == "true"
	h.performLifecycleAction(c, "restart", func(id string) error {
		return h.computeManager.RestartInstance(c.Request.Context(), id, force)
	})
}

// PauseInstance handles requests to pause a compute instance.
func (h *ComputeHandler) PauseInstance(c *gin.Context) {
	h.performLifecycleAction(c, "pause", func(id string) error {
		return h.computeManager.PauseInstance(c.Request.Context(), id)
	})
}

// UnpauseInstance handles requests to unpause a compute instance.
func (h *ComputeHandler) UnpauseInstance(c *gin.Context) {
	h.performLifecycleAction(c, "unpause", func(id string) error {
		return h.computeManager.UnpauseInstance(c.Request.Context(), id)
	})
}

// Resource operations.

// GetResourceUsage handles requests to get instance resource usage.
func (h *ComputeHandler) GetResourceUsage(c *gin.Context) {
	contextLogger := getContextLogger(c, h.logger)
	id := c.Param("id")

	if id == "" {
		contextLogger.Warn("Missing instance ID")
		HandleError(c, ErrInvalidInput)
		return
	}

	usage, err := h.computeManager.GetResourceUsage(c.Request.Context(), id)
	if err != nil {
		contextLogger.Error("Failed to get resource usage",
			logger.String("id", id),
			logger.Error(err))
		HandleError(c, apierrors.Wrap(err, "get resource usage"))
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"usage": usage,
	})
}

// UpdateResourceLimits handles requests to update instance resource limits.
func (h *ComputeHandler) UpdateResourceLimits(c *gin.Context) {
	contextLogger := getContextLogger(c, h.logger)
	id := c.Param("id")

	if id == "" {
		contextLogger.Warn("Missing instance ID")
		HandleError(c, ErrInvalidInput)
		return
	}

	// Parse resource limits
	var resources compute.ComputeResources
	if err := c.ShouldBindJSON(&resources); err != nil {
		contextLogger.Warn("Invalid resource limits", logger.Error(err))
		HandleError(c, ErrInvalidInput)
		return
	}

	err := h.computeManager.UpdateResourceLimits(c.Request.Context(), id, resources)
	if err != nil {
		contextLogger.Error("Failed to update resource limits",
			logger.String("id", id),
			logger.Error(err))
		HandleError(c, apierrors.Wrap(err, "update resource limits"))
		return
	}

	contextLogger.Info("Updated resource limits",
		logger.String("id", id))

	c.JSON(http.StatusOK, gin.H{
		"message": "Resource limits updated successfully",
	})
}

// Cluster and backend operations.

// GetClusterStatus handles requests to get overall cluster status.
func (h *ComputeHandler) GetClusterStatus(c *gin.Context) {
	contextLogger := getContextLogger(c, h.logger)

	status, err := h.computeManager.GetClusterStatus(c.Request.Context())
	if err != nil {
		contextLogger.Error("Failed to get cluster status", logger.Error(err))
		HandleError(c, apierrors.Wrap(err, "get cluster status"))
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": status,
	})
}

// GetBackendInfo handles requests to get backend information.
func (h *ComputeHandler) GetBackendInfo(c *gin.Context) {
	contextLogger := getContextLogger(c, h.logger)
	backendStr := c.Param("backend")

	if backendStr == "" {
		contextLogger.Warn("Missing backend parameter")
		HandleError(c, ErrInvalidInput)
		return
	}

	backend := compute.ComputeBackend(backendStr)
	info, err := h.computeManager.GetBackendInfo(c.Request.Context(), backend)
	if err != nil {
		contextLogger.Error("Failed to get backend info",
			logger.String("backend", string(backend)),
			logger.Error(err))
		HandleError(c, apierrors.Wrap(err, "get backend info"))
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"info": info,
	})
}

// GetInstanceEvents handles requests to get instance events.
func (h *ComputeHandler) GetInstanceEvents(c *gin.Context) {
	contextLogger := getContextLogger(c, h.logger)
	id := c.Param("id")

	if id == "" {
		contextLogger.Warn("Missing instance ID")
		HandleError(c, ErrInvalidInput)
		return
	}

	// Parse event options
	opts := h.parseEventOptions(c)

	events, err := h.computeManager.GetInstanceEvents(c.Request.Context(), id, opts)
	if err != nil {
		contextLogger.Error("Failed to get instance events",
			logger.String("id", id),
			logger.Error(err))
		HandleError(c, apierrors.Wrap(err, "get instance events"))
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"events": events,
		"count":  len(events),
	})
}

// HealthCheck handles requests to check compute manager health.
func (h *ComputeHandler) HealthCheck(c *gin.Context) {
	contextLogger := getContextLogger(c, h.logger)

	health, err := h.computeManager.HealthCheck(c.Request.Context())
	if err != nil {
		contextLogger.Error("Health check failed", logger.Error(err))
		HandleError(c, apierrors.Wrap(err, "health check"))
		return
	}

	status := http.StatusOK
	if health.Status != "healthy" {
		status = http.StatusServiceUnavailable
	}

	c.JSON(status, gin.H{
		"health": health,
	})
}

// Helper methods.

func (h *ComputeHandler) validateCreateRequest(req compute.ComputeInstanceRequest) error {
	if req.Name == "" {
		return apierrors.ErrInvalidParameter
	}

	if req.Type == "" {
		return apierrors.ErrInvalidParameter
	}

	if req.Config.Image == "" {
		return apierrors.ErrInvalidParameter
	}

	return nil
}

func (h *ComputeHandler) parseListOptions(c *gin.Context) compute.ComputeInstanceListOptions {
	opts := compute.ComputeInstanceListOptions{}

	if backend := c.Query("backend"); backend != "" {
		opts.Backend = compute.ComputeBackend(backend)
	}

	if state := c.Query("state"); state != "" {
		opts.State = compute.ComputeInstanceState(state)
	}

	if limitStr := c.Query("limit"); limitStr != "" {
		if limit, err := strconv.Atoi(limitStr); err == nil && limit > 0 {
			opts.Limit = limit
		}
	}

	if offsetStr := c.Query("offset"); offsetStr != "" {
		if offset, err := strconv.Atoi(offsetStr); err == nil && offset >= 0 {
			opts.Offset = offset
		}
	}

	if labels := c.Query("labels"); labels != "" {
		opts.Labels = make(map[string]string)
		for _, pair := range strings.Split(labels, ",") {
			kv := strings.SplitN(pair, "=", 2)
			if len(kv) == 2 {
				opts.Labels[kv[0]] = kv[1]
			}
		}
	}

	return opts
}

func (h *ComputeHandler) parseEventOptions(c *gin.Context) compute.EventOptions {
	opts := compute.EventOptions{}

	if limitStr := c.Query("limit"); limitStr != "" {
		if limit, err := strconv.Atoi(limitStr); err == nil && limit > 0 {
			opts.Limit = limit
		}
	}

	if follow := c.Query("follow"); follow == "true" {
		opts.Follow = true
	}

	if types := c.Query("types"); types != "" {
		opts.Types = strings.Split(types, ",")
	}

	return opts
}

func (h *ComputeHandler) performLifecycleAction(c *gin.Context, action string, actionFunc func(string) error) {
	contextLogger := getContextLogger(c, h.logger)
	id := c.Param("id")

	if id == "" {
		contextLogger.Warn("Missing instance ID")
		HandleError(c, ErrInvalidInput)
		return
	}

	err := actionFunc(id)
	if err != nil {
		contextLogger.Error("Failed to perform lifecycle action",
			logger.String("id", id),
			logger.String("action", action),
			logger.Error(err))
		HandleError(c, apierrors.Wrap(err, action+" instance"))
		return
	}

	contextLogger.Info("Performed lifecycle action",
		logger.String("id", id),
		logger.String("action", action))

	c.JSON(http.StatusOK, gin.H{
		"message": action + " operation completed successfully",
	})
}
