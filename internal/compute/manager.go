package compute

import (
	"context"
	"fmt"
	"io"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/threatflux/libgo/pkg/logger"
)

// ComputeManager implements the unified compute management interface.
type ComputeManager struct {
	// Map fields (8 bytes)
	backends map[ComputeBackend]BackendService
	// Pointer fields (8 bytes)
	resourceTracker *ResourceTracker
	quotaManager    *QuotaManager
	eventBus        *EventBus
	logger          logger.Logger
	// Struct fields
	config ManagerConfig
	mu     sync.RWMutex
}

// ManagerConfig holds configuration for the compute manager.
type ManagerConfig struct {
	// Duration fields (8 bytes)
	HealthCheckInterval time.Duration
	MetricsInterval     time.Duration
	// Struct fields
	ResourceLimits ComputeResources
	// Enum fields
	DefaultBackend ComputeBackend
	// Bool fields (1 byte)
	AllowMixedWorkloads bool
	EnableQuotas        bool
}

// NewComputeManager creates a new unified compute manager.
func NewComputeManager(config ManagerConfig, logger logger.Logger) Manager {
	manager := &ComputeManager{
		backends:        make(map[ComputeBackend]BackendService),
		config:          config,
		logger:          logger,
		resourceTracker: NewResourceTracker(),
		quotaManager:    NewQuotaManager(),
		eventBus:        NewEventBus(),
	}

	return manager
}

// RegisterBackend registers a compute backend.
func (m *ComputeManager) RegisterBackend(backend ComputeBackend, service BackendService) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.backends[backend]; exists {
		return fmt.Errorf("backend %s already registered", backend)
	}

	m.backends[backend] = service
	m.logger.Info("Registered compute backend", logger.String("backend", string(backend)))

	return nil
}

// CreateInstance creates a new compute instance.
func (m *ComputeManager) CreateInstance(ctx context.Context, req ComputeInstanceRequest) (*ComputeInstance, error) {
	// Determine backend
	backend := req.Backend
	if backend == "" {
		backend = m.config.DefaultBackend
	}

	// Validate backend exists
	backendService, err := m.getBackend(backend)
	if err != nil {
		return nil, err
	}

	// Check quotas
	if m.config.EnableQuotas {
		if quotaErr := m.quotaManager.CheckQuota(ctx, req); quotaErr != nil {
			return nil, fmt.Errorf("quota exceeded: %w", quotaErr)
		}
	}

	// Validate configuration
	if configErr := backendService.ValidateConfig(ctx, req.Config); configErr != nil {
		return nil, fmt.Errorf("invalid configuration: %w", configErr)
	}

	// Generate UUID for instance
	req.UUID = uuid.New().String()
	if req.Name == "" {
		req.Name = fmt.Sprintf("%s-%s", strings.ToLower(string(req.Type)), req.UUID[:8])
	}

	// Create the instance
	instance, err := backendService.Create(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to create instance on backend %s: %w", backend, err)
	}

	// Update resource tracking
	m.resourceTracker.AddInstance(instance)

	// Emit event
	m.eventBus.Emit(InstanceEvent{
		ID:         uuid.New().String(),
		InstanceID: instance.ID,
		Type:       "lifecycle",
		Action:     "create",
		Status:     "success",
		Timestamp:  time.Now(),
	})

	m.logger.Info("Created compute instance",
		logger.String("id", instance.ID),
		logger.String("name", instance.Name),
		logger.String("type", string(instance.Type)),
		logger.String("backend", string(instance.Backend)))

	return instance, nil
}

// GetInstance retrieves a compute instance by ID.
func (m *ComputeManager) GetInstance(ctx context.Context, id string) (*ComputeInstance, error) {
	// Try each backend to find the instance
	m.mu.RLock()
	defer m.mu.RUnlock()

	for backend, service := range m.backends {
		instance, err := service.Get(ctx, id)
		if err == nil {
			return instance, nil
		}

		// Log non-critical errors but continue searching
		m.logger.Debug("Instance not found in backend",
			logger.String("id", id),
			logger.String("backend", string(backend)),
			logger.Error(err))
	}

	return nil, fmt.Errorf("instance %s not found in any backend", id)
}

// GetInstanceByName retrieves a compute instance by name.
func (m *ComputeManager) GetInstanceByName(ctx context.Context, name string) (*ComputeInstance, error) {
	instances, err := m.ListAllInstances(ctx, ComputeInstanceListOptions{
		Limit: 100, // Reasonable limit for name search
	})
	if err != nil {
		return nil, err
	}

	for _, instance := range instances {
		if instance.Name == name {
			return instance, nil
		}
	}

	return nil, fmt.Errorf("instance with name %s not found", name)
}

// ListInstances lists instances from a specific backend.
func (m *ComputeManager) ListInstances(ctx context.Context, opts ComputeInstanceListOptions) ([]*ComputeInstance, error) {
	backend := opts.Backend
	if backend == "" {
		backend = m.config.DefaultBackend
	}

	backendService, err := m.getBackend(backend)
	if err != nil {
		return nil, err
	}

	return backendService.List(ctx, opts)
}

// ListAllInstances lists instances from all backends.
func (m *ComputeManager) ListAllInstances(ctx context.Context, opts ComputeInstanceListOptions) ([]*ComputeInstance, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var allInstances []*ComputeInstance
	var errors []error

	for backend, service := range m.backends {
		backendOpts := opts
		backendOpts.Backend = backend

		instances, err := service.List(ctx, backendOpts)
		if err != nil {
			errors = append(errors, fmt.Errorf("backend %s: %w", backend, err))
			continue
		}

		allInstances = append(allInstances, instances...)
	}

	// Log errors but don't fail completely if some backends work
	for _, err := range errors {
		m.logger.Warn("Failed to list instances from backend", logger.Error(err))
	}

	// Sort instances by creation time (newest first)
	sort.Slice(allInstances, func(i, j int) bool {
		return allInstances[i].CreatedAt.After(allInstances[j].CreatedAt)
	})

	// Apply global limits
	if opts.Limit > 0 && len(allInstances) > opts.Limit {
		allInstances = allInstances[:opts.Limit]
	}

	return allInstances, nil
}

// UpdateInstance updates a compute instance.
func (m *ComputeManager) UpdateInstance(ctx context.Context, id string, update ComputeInstanceUpdate) (*ComputeInstance, error) {
	// Find the instance first to determine its backend
	instance, err := m.GetInstance(ctx, id)
	if err != nil {
		return nil, err
	}

	backendService, err := m.getBackend(instance.Backend)
	if err != nil {
		return nil, err
	}

	// Update the instance
	updatedInstance, err := backendService.Update(ctx, id, update)
	if err != nil {
		return nil, fmt.Errorf("failed to update instance: %w", err)
	}

	// Update resource tracking
	m.resourceTracker.UpdateInstance(updatedInstance)

	// Emit event
	m.eventBus.Emit(InstanceEvent{
		ID:         uuid.New().String(),
		InstanceID: instance.ID,
		Type:       "lifecycle",
		Action:     "update",
		Status:     "success",
		Timestamp:  time.Now(),
	})

	return updatedInstance, nil
}

// DeleteInstance deletes a compute instance.
func (m *ComputeManager) DeleteInstance(ctx context.Context, id string, force bool) error {
	// Find the instance first to determine its backend
	instance, err := m.GetInstance(ctx, id)
	if err != nil {
		return err
	}

	backendService, err := m.getBackend(instance.Backend)
	if err != nil {
		return err
	}

	// Delete the instance
	if err := backendService.Delete(ctx, id, force); err != nil {
		return fmt.Errorf("failed to delete instance: %w", err)
	}

	// Update resource tracking
	m.resourceTracker.RemoveInstance(id)

	// Emit event
	m.eventBus.Emit(InstanceEvent{
		ID:         uuid.New().String(),
		InstanceID: instance.ID,
		Type:       "lifecycle",
		Action:     "delete",
		Status:     "success",
		Timestamp:  time.Now(),
	})

	m.logger.Info("Deleted compute instance",
		logger.String("id", instance.ID),
		logger.String("name", instance.Name),
		logger.String("backend", string(instance.Backend)))

	return nil
}

// Lifecycle operations

// StartInstance starts a compute instance.
func (m *ComputeManager) StartInstance(ctx context.Context, id string) error {
	instance, err := m.GetInstance(ctx, id)
	if err != nil {
		return err
	}

	backendService, err := m.getBackend(instance.Backend)
	if err != nil {
		return err
	}

	if err := backendService.Start(ctx, id); err != nil {
		return fmt.Errorf("failed to start instance: %w", err)
	}

	m.eventBus.Emit(InstanceEvent{
		ID:         uuid.New().String(),
		InstanceID: instance.ID,
		Type:       "lifecycle",
		Action:     "start",
		Status:     "success",
		Timestamp:  time.Now(),
	})

	return nil
}

// StopInstance stops a compute instance.
func (m *ComputeManager) StopInstance(ctx context.Context, id string, force bool) error {
	instance, err := m.GetInstance(ctx, id)
	if err != nil {
		return err
	}

	backendService, err := m.getBackend(instance.Backend)
	if err != nil {
		return err
	}

	if err := backendService.Stop(ctx, id, force); err != nil {
		return fmt.Errorf("failed to stop instance: %w", err)
	}

	m.eventBus.Emit(InstanceEvent{
		ID:         uuid.New().String(),
		InstanceID: instance.ID,
		Type:       "lifecycle",
		Action:     "stop",
		Status:     "success",
		Timestamp:  time.Now(),
	})

	return nil
}

// RestartInstance restarts a compute instance.
func (m *ComputeManager) RestartInstance(ctx context.Context, id string, force bool) error {
	instance, err := m.GetInstance(ctx, id)
	if err != nil {
		return err
	}

	backendService, err := m.getBackend(instance.Backend)
	if err != nil {
		return err
	}

	if err := backendService.Restart(ctx, id, force); err != nil {
		return fmt.Errorf("failed to restart instance: %w", err)
	}

	m.eventBus.Emit(InstanceEvent{
		ID:         uuid.New().String(),
		InstanceID: instance.ID,
		Type:       "lifecycle",
		Action:     "restart",
		Status:     "success",
		Timestamp:  time.Now(),
	})

	return nil
}

// PauseInstance pauses a compute instance.
func (m *ComputeManager) PauseInstance(ctx context.Context, id string) error {
	instance, err := m.GetInstance(ctx, id)
	if err != nil {
		return err
	}

	backendService, err := m.getBackend(instance.Backend)
	if err != nil {
		return err
	}

	if err := backendService.Pause(ctx, id); err != nil {
		return fmt.Errorf("failed to pause instance: %w", err)
	}

	m.eventBus.Emit(InstanceEvent{
		ID:         uuid.New().String(),
		InstanceID: instance.ID,
		Type:       "lifecycle",
		Action:     "pause",
		Status:     "success",
		Timestamp:  time.Now(),
	})

	return nil
}

// UnpauseInstance unpauses a compute instance.
func (m *ComputeManager) UnpauseInstance(ctx context.Context, id string) error {
	instance, err := m.GetInstance(ctx, id)
	if err != nil {
		return err
	}

	backendService, err := m.getBackend(instance.Backend)
	if err != nil {
		return err
	}

	if err := backendService.Unpause(ctx, id); err != nil {
		return fmt.Errorf("failed to unpause instance: %w", err)
	}

	m.eventBus.Emit(InstanceEvent{
		ID:         uuid.New().String(),
		InstanceID: instance.ID,
		Type:       "lifecycle",
		Action:     "unpause",
		Status:     "success",
		Timestamp:  time.Now(),
	})

	return nil
}

// Resource management

// GetResourceUsage gets current resource usage for an instance.
func (m *ComputeManager) GetResourceUsage(ctx context.Context, id string) (*ResourceUsage, error) {
	instance, err := m.GetInstance(ctx, id)
	if err != nil {
		return nil, err
	}

	backendService, err := m.getBackend(instance.Backend)
	if err != nil {
		return nil, err
	}

	return backendService.GetResourceUsage(ctx, id)
}

// GetResourceUsageHistory gets historical resource usage (placeholder).
func (m *ComputeManager) GetResourceUsageHistory(ctx context.Context, id string, opts ResourceHistoryOptions) ([]*ResourceUsage, error) {
	// This would be implemented with a metrics storage backend
	return nil, fmt.Errorf("resource usage history not implemented yet")
}

// UpdateResourceLimits updates resource limits for an instance.
func (m *ComputeManager) UpdateResourceLimits(ctx context.Context, id string, resources ComputeResources) error {
	instance, err := m.GetInstance(ctx, id)
	if err != nil {
		return err
	}

	backendService, err := m.getBackend(instance.Backend)
	if err != nil {
		return err
	}

	return backendService.UpdateResourceLimits(ctx, id, resources)
}

// GetClusterStatus returns the overall cluster status.
func (m *ComputeManager) GetClusterStatus(ctx context.Context) (*ClusterStatus, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	status := &ClusterStatus{
		Backends:    make(map[ComputeBackend]*BackendInfo),
		LastUpdated: time.Now(),
	}

	// Get status from each backend
	for backend, service := range m.backends {
		backendInfo, err := service.GetBackendInfo(ctx)
		if err != nil {
			m.logger.Warn("Failed to get backend info",
				logger.String("backend", string(backend)),
				logger.Error(err))
			continue
		}
		status.Backends[backend] = backendInfo
	}

	// Get overall instance counts
	instances, err := m.ListAllInstances(ctx, ComputeInstanceListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get instance counts: %w", err)
	}

	status.TotalInstances = len(instances)
	for _, instance := range instances {
		switch instance.State {
		case StateRunning:
			status.RunningInstances++
		case StateStopped:
			status.StoppedInstances++
		case StateError:
			status.ErrorInstances++
		}
	}

	return status, nil
}

// GetResourceQuotas gets resource quotas for a user.
func (m *ComputeManager) GetResourceQuotas(ctx context.Context, userID uint) (*ResourceQuotas, error) {
	return m.quotaManager.GetQuotas(userID), nil
}

// SetResourceQuotas sets resource quotas for a user.
func (m *ComputeManager) SetResourceQuotas(ctx context.Context, userID uint, quotas ResourceQuotas) error {
	return m.quotaManager.SetQuotas(userID, quotas)
}

// GetBackendInfo gets information about a specific backend.
func (m *ComputeManager) GetBackendInfo(ctx context.Context, backend ComputeBackend) (*BackendInfo, error) {
	backendService, err := m.getBackend(backend)
	if err != nil {
		return nil, err
	}

	return backendService.GetBackendInfo(ctx)
}

// ValidateInstanceConfig validates an instance configuration.
func (m *ComputeManager) ValidateInstanceConfig(ctx context.Context, config ComputeInstanceConfig, backend ComputeBackend) error {
	backendService, err := m.getBackend(backend)
	if err != nil {
		return err
	}

	return backendService.ValidateConfig(ctx, config)
}

// Stub implementations for unimplemented methods

// AttachConsole attaches to an instance console.
func (m *ComputeManager) AttachConsole(ctx context.Context, id string, opts ConsoleOptions) (io.ReadWriteCloser, error) {
	return nil, fmt.Errorf("console attachment not implemented yet")
}

// ExecuteCommand executes a command in an instance.
func (m *ComputeManager) ExecuteCommand(ctx context.Context, id string, cmd ExecRequest) (*ExecResult, error) {
	return nil, fmt.Errorf("command execution not implemented yet")
}

// GetLogs gets logs from an instance.
func (m *ComputeManager) GetLogs(ctx context.Context, id string, opts LogOptions) (io.ReadCloser, error) {
	return nil, fmt.Errorf("log retrieval not implemented yet")
}

// Snapshot operations (stubs).
func (m *ComputeManager) CreateSnapshot(ctx context.Context, id, name, description string) (*Snapshot, error) {
	return nil, fmt.Errorf("snapshots not implemented yet")
}

func (m *ComputeManager) ListSnapshots(ctx context.Context, id string) ([]*Snapshot, error) {
	return nil, fmt.Errorf("snapshots not implemented yet")
}

func (m *ComputeManager) RestoreSnapshot(ctx context.Context, id, snapshotID string) error {
	return fmt.Errorf("snapshots not implemented yet")
}

func (m *ComputeManager) DeleteSnapshot(ctx context.Context, id, snapshotID string) error {
	return fmt.Errorf("snapshots not implemented yet")
}

// Migration and export (stubs).
func (m *ComputeManager) MigrateInstance(ctx context.Context, id, targetHost string, opts MigrationOptions) error {
	return fmt.Errorf("migration not implemented yet")
}

func (m *ComputeManager) ExportInstance(ctx context.Context, id string, opts ExportOptions) (*ExportJob, error) {
	return nil, fmt.Errorf("export not implemented yet")
}

func (m *ComputeManager) ImportInstance(ctx context.Context, source string, opts ImportOptions) (*ComputeInstance, error) {
	return nil, fmt.Errorf("import not implemented yet")
}

// Network and storage attachment (stubs).
func (m *ComputeManager) AttachNetwork(ctx context.Context, id string, network NetworkAttachment) error {
	return fmt.Errorf("network attachment not implemented yet")
}

func (m *ComputeManager) DetachNetwork(ctx context.Context, id, networkName string) error {
	return fmt.Errorf("network detachment not implemented yet")
}

func (m *ComputeManager) ListNetworkAttachments(ctx context.Context, id string) ([]*NetworkAttachment, error) {
	return nil, fmt.Errorf("network listing not implemented yet")
}

func (m *ComputeManager) AttachStorage(ctx context.Context, id string, storage StorageAttachment) error {
	return fmt.Errorf("storage attachment not implemented yet")
}

func (m *ComputeManager) DetachStorage(ctx context.Context, id, storageName string) error {
	return fmt.Errorf("storage detachment not implemented yet")
}

func (m *ComputeManager) ListStorageAttachments(ctx context.Context, id string) ([]*StorageAttachment, error) {
	return nil, fmt.Errorf("storage listing not implemented yet")
}

// Monitoring (stubs).
func (m *ComputeManager) StreamMetrics(ctx context.Context, id string, opts MetricsOptions) (<-chan ResourceUsage, error) {
	return nil, fmt.Errorf("metrics streaming not implemented yet")
}

func (m *ComputeManager) GetInstanceEvents(ctx context.Context, id string, opts EventOptions) ([]*InstanceEvent, error) {
	return m.eventBus.GetEvents(id, opts), nil
}

func (m *ComputeManager) StreamInstanceEvents(ctx context.Context, id string, opts EventOptions) (<-chan InstanceEvent, error) {
	return m.eventBus.StreamEvents(id, opts), nil
}

// Bulk operations (stubs).
func (m *ComputeManager) BulkAction(ctx context.Context, action string, ids []string, opts BulkActionOptions) ([]*BulkActionResult, error) {
	return nil, fmt.Errorf("bulk operations not implemented yet")
}

// Template management (stubs).
func (m *ComputeManager) CreateTemplate(ctx context.Context, instanceID string, template InstanceTemplate) (*InstanceTemplate, error) {
	return nil, fmt.Errorf("templates not implemented yet")
}

func (m *ComputeManager) GetTemplate(ctx context.Context, templateID string) (*InstanceTemplate, error) {
	return nil, fmt.Errorf("templates not implemented yet")
}

func (m *ComputeManager) ListTemplates(ctx context.Context, opts TemplateListOptions) ([]*InstanceTemplate, error) {
	return nil, fmt.Errorf("templates not implemented yet")
}

func (m *ComputeManager) DeleteTemplate(ctx context.Context, templateID string) error {
	return fmt.Errorf("templates not implemented yet")
}

func (m *ComputeManager) CloneFromTemplate(ctx context.Context, templateID string, req CloneRequest) (*ComputeInstance, error) {
	return nil, fmt.Errorf("templates not implemented yet")
}

// Compose operations (stubs).
func (m *ComputeManager) DeployCompose(ctx context.Context, composeData []byte, opts ComposeDeployOptions) (*ComposeDeployment, error) {
	return nil, fmt.Errorf("compose not implemented yet")
}

func (m *ComputeManager) GetComposeDeployment(ctx context.Context, deploymentID string) (*ComposeDeployment, error) {
	return nil, fmt.Errorf("compose not implemented yet")
}

func (m *ComputeManager) ListComposeDeployments(ctx context.Context, opts ComposeListOptions) ([]*ComposeDeployment, error) {
	return nil, fmt.Errorf("compose not implemented yet")
}

func (m *ComputeManager) UpdateComposeDeployment(ctx context.Context, deploymentID string, composeData []byte) (*ComposeDeployment, error) {
	return nil, fmt.Errorf("compose not implemented yet")
}

func (m *ComputeManager) DeleteComposeDeployment(ctx context.Context, deploymentID string, force bool) error {
	return fmt.Errorf("compose not implemented yet")
}

// Health and maintenance.
func (m *ComputeManager) HealthCheck(ctx context.Context) (*HealthStatus, error) {
	// Check all backends
	m.mu.RLock()
	defer m.mu.RUnlock()

	allHealthy := true
	var messages []string

	for backend, service := range m.backends {
		info, err := service.GetBackendInfo(ctx)
		if err != nil {
			allHealthy = false
			messages = append(messages, fmt.Sprintf("%s: %v", backend, err))
			continue
		}

		if info.HealthCheck != nil && info.HealthCheck.Status != "healthy" {
			allHealthy = false
			messages = append(messages, fmt.Sprintf("%s: %s", backend, info.HealthCheck.Message))
		}
	}

	status := "healthy"
	if !allHealthy {
		status = "unhealthy"
	}

	return &HealthStatus{
		Status:     status,
		Message:    strings.Join(messages, "; "),
		LastCheck:  time.Now(),
		CheckCount: 1,
	}, nil
}

func (m *ComputeManager) PerformMaintenance(ctx context.Context, opts MaintenanceOptions) error {
	return fmt.Errorf("maintenance operations not implemented yet")
}

// Helper methods

func (m *ComputeManager) getBackend(backend ComputeBackend) (BackendService, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	service, exists := m.backends[backend]
	if !exists {
		return nil, fmt.Errorf("backend %s not found", backend)
	}

	return service, nil
}
