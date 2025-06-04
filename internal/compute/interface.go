package compute

import (
	"context"
	"io"
	"time"
)

// Service defines the unified compute service interface for managing both VMs and containers.
type Service interface {
	// Instance management
	CreateInstance(ctx context.Context, req ComputeInstanceRequest) (*ComputeInstance, error)
	GetInstance(ctx context.Context, id string) (*ComputeInstance, error)
	GetInstanceByName(ctx context.Context, name string) (*ComputeInstance, error)
	ListInstances(ctx context.Context, opts ComputeInstanceListOptions) ([]*ComputeInstance, error)
	UpdateInstance(ctx context.Context, id string, update ComputeInstanceUpdate) (*ComputeInstance, error)
	DeleteInstance(ctx context.Context, id string, force bool) error

	// Instance lifecycle
	StartInstance(ctx context.Context, id string) error
	StopInstance(ctx context.Context, id string, force bool) error
	RestartInstance(ctx context.Context, id string, force bool) error
	PauseInstance(ctx context.Context, id string) error
	UnpauseInstance(ctx context.Context, id string) error

	// Instance operations
	AttachConsole(ctx context.Context, id string, opts ConsoleOptions) (io.ReadWriteCloser, error)
	ExecuteCommand(ctx context.Context, id string, cmd ExecRequest) (*ExecResult, error)
	GetLogs(ctx context.Context, id string, opts LogOptions) (io.ReadCloser, error)

	// Snapshots (primarily for VMs, but can support container commits)
	CreateSnapshot(ctx context.Context, id, name, description string) (*Snapshot, error)
	ListSnapshots(ctx context.Context, id string) ([]*Snapshot, error)
	RestoreSnapshot(ctx context.Context, id, snapshotID string) error
	DeleteSnapshot(ctx context.Context, id, snapshotID string) error

	// Resource management
	GetResourceUsage(ctx context.Context, id string) (*ResourceUsage, error)
	GetResourceUsageHistory(ctx context.Context, id string, opts ResourceHistoryOptions) ([]*ResourceUsage, error)
	UpdateResourceLimits(ctx context.Context, id string, resources ComputeResources) error

	// Migration and export (primarily for VMs)
	MigrateInstance(ctx context.Context, id, targetHost string, opts MigrationOptions) error
	ExportInstance(ctx context.Context, id string, opts ExportOptions) (*ExportJob, error)
	ImportInstance(ctx context.Context, source string, opts ImportOptions) (*ComputeInstance, error)

	// Network management
	AttachNetwork(ctx context.Context, id string, network NetworkAttachment) error
	DetachNetwork(ctx context.Context, id, networkName string) error
	ListNetworkAttachments(ctx context.Context, id string) ([]*NetworkAttachment, error)

	// Storage management
	AttachStorage(ctx context.Context, id string, storage StorageAttachment) error
	DetachStorage(ctx context.Context, id, storageName string) error
	ListStorageAttachments(ctx context.Context, id string) ([]*StorageAttachment, error)

	// Monitoring and metrics
	StreamMetrics(ctx context.Context, id string, opts MetricsOptions) (<-chan ResourceUsage, error)
	GetInstanceEvents(ctx context.Context, id string, opts EventOptions) ([]*InstanceEvent, error)
	StreamInstanceEvents(ctx context.Context, id string, opts EventOptions) (<-chan InstanceEvent, error)

	// Bulk operations
	BulkAction(ctx context.Context, action string, ids []string, opts BulkActionOptions) ([]*BulkActionResult, error)

	// Backend-specific operations
	GetBackendInfo(ctx context.Context, backend ComputeBackend) (*BackendInfo, error)
	ValidateInstanceConfig(ctx context.Context, config ComputeInstanceConfig, backend ComputeBackend) error
}

// Manager provides higher-level orchestration across multiple backends.
type Manager interface {
	Service

	// Multi-backend operations
	ListAllInstances(ctx context.Context, opts ComputeInstanceListOptions) ([]*ComputeInstance, error)
	GetClusterStatus(ctx context.Context) (*ClusterStatus, error)
	GetResourceQuotas(ctx context.Context, userID uint) (*ResourceQuotas, error)
	SetResourceQuotas(ctx context.Context, userID uint, quotas ResourceQuotas) error

	// Template management
	CreateTemplate(ctx context.Context, instanceID string, template InstanceTemplate) (*InstanceTemplate, error)
	GetTemplate(ctx context.Context, templateID string) (*InstanceTemplate, error)
	ListTemplates(ctx context.Context, opts TemplateListOptions) ([]*InstanceTemplate, error)
	DeleteTemplate(ctx context.Context, templateID string) error
	CloneFromTemplate(ctx context.Context, templateID string, req CloneRequest) (*ComputeInstance, error)

	// Compose/orchestration support
	DeployCompose(ctx context.Context, composeData []byte, opts ComposeDeployOptions) (*ComposeDeployment, error)
	GetComposeDeployment(ctx context.Context, deploymentID string) (*ComposeDeployment, error)
	ListComposeDeployments(ctx context.Context, opts ComposeListOptions) ([]*ComposeDeployment, error)
	UpdateComposeDeployment(ctx context.Context, deploymentID string, composeData []byte) (*ComposeDeployment, error)
	DeleteComposeDeployment(ctx context.Context, deploymentID string, force bool) error

	// Health and maintenance
	HealthCheck(ctx context.Context) (*HealthStatus, error)
	PerformMaintenance(ctx context.Context, opts MaintenanceOptions) error
}

// BackendService defines the interface that each backend (KVM, Docker) must implement.
type BackendService interface {
	// Basic CRUD operations
	Create(ctx context.Context, req ComputeInstanceRequest) (*ComputeInstance, error)
	Get(ctx context.Context, id string) (*ComputeInstance, error)
	List(ctx context.Context, opts ComputeInstanceListOptions) ([]*ComputeInstance, error)
	Update(ctx context.Context, id string, update ComputeInstanceUpdate) (*ComputeInstance, error)
	Delete(ctx context.Context, id string, force bool) error

	// Lifecycle operations
	Start(ctx context.Context, id string) error
	Stop(ctx context.Context, id string, force bool) error
	Restart(ctx context.Context, id string, force bool) error
	Pause(ctx context.Context, id string) error
	Unpause(ctx context.Context, id string) error

	// Resource operations
	GetResourceUsage(ctx context.Context, id string) (*ResourceUsage, error)
	UpdateResourceLimits(ctx context.Context, id string, resources ComputeResources) error

	// Backend-specific information
	GetBackendInfo(ctx context.Context) (*BackendInfo, error)
	ValidateConfig(ctx context.Context, config ComputeInstanceConfig) error

	// Type identification
	GetBackendType() ComputeBackend
	GetSupportedInstanceTypes() []ComputeInstanceType
}

// Supporting types for the service interface

// ConsoleOptions represents options for console attachment.
type ConsoleOptions struct {
	Type   string `json:"type"` // vnc, spice, serial, web
	Width  int    `json:"width,omitempty"`
	Height int    `json:"height,omitempty"`
	Secure bool   `json:"secure,omitempty"`
}

// ExecRequest represents a command execution request.
type ExecRequest struct {
	// Slice and map fields (8 bytes)
	Command []string          `json:"command"`
	Env     map[string]string `json:"env,omitempty"`
	// String fields (8 bytes)
	WorkDir string `json:"work_dir,omitempty"`
	User    string `json:"user,omitempty"`
	// Int fields (4 bytes)
	Timeout int `json:"timeout,omitempty"` // seconds
	// Bool fields (1 byte)
	TTY    bool `json:"tty,omitempty"`
	Stdin  bool `json:"stdin,omitempty"`
	Stdout bool `json:"stdout,omitempty"`
	Stderr bool `json:"stderr,omitempty"`
}

// ExecResult represents the result of command execution.
type ExecResult struct {
	// String fields (8 bytes)
	Stdout string `json:"stdout,omitempty"`
	Stderr string `json:"stderr,omitempty"`
	Error  string `json:"error,omitempty"`
	// Int64 fields (8 bytes)
	Duration int64 `json:"duration"` // milliseconds
	// Int fields (4 bytes)
	ExitCode int `json:"exit_code"`
	// Bool fields (1 byte)
	Timeout bool `json:"timeout"`
}

// LogOptions represents options for log retrieval.
type LogOptions struct {
	// Pointer fields (8 bytes)
	Since *TimeStamp `json:"since,omitempty"`
	Until *TimeStamp `json:"until,omitempty"`
	// Int fields (4 bytes)
	Tail int `json:"tail,omitempty"`
	// Bool fields (1 byte)
	Follow     bool `json:"follow,omitempty"`
	Timestamps bool `json:"timestamps,omitempty"`
	Details    bool `json:"details,omitempty"`
}

// TimeStamp represents a timestamp that can be parsed from various formats.
type TimeStamp struct {
	time.Time
}

// Snapshot represents a compute instance snapshot.
type Snapshot struct {
	// String fields (8 bytes)
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	InstanceID  string `json:"instance_id"`
	State       string `json:"state"`
	Parent      string `json:"parent,omitempty"`
	// Slice and map fields (8 bytes)
	Children []string          `json:"children,omitempty"`
	Metadata map[string]string `json:"metadata,omitempty"`
	// Int64 fields (8 bytes)
	Size int64 `json:"size,omitempty"`
	// Time fields (8 bytes)
	CreatedAt time.Time `json:"created_at"`
}

// ResourceHistoryOptions represents options for resource usage history.
type ResourceHistoryOptions struct {
	// Pointer fields (8 bytes)
	Start *TimeStamp `json:"start,omitempty"`
	End   *TimeStamp `json:"end,omitempty"`
	// String fields (8 bytes)
	Interval string `json:"interval,omitempty"` // 1m, 5m, 1h, etc.
	// Slice fields (8 bytes)
	Metrics []string `json:"metrics,omitempty"` // cpu, memory, network, storage
}

// MigrationOptions represents options for instance migration.
type MigrationOptions struct {
	// Slice and map fields (8 bytes)
	Flags      []string          `json:"flags,omitempty"`
	Parameters map[string]string `json:"parameters,omitempty"`
	// Int64 fields (8 bytes)
	Bandwidth int64 `json:"bandwidth,omitempty"` // MB/s
	// Int fields (4 bytes)
	Timeout int `json:"timeout,omitempty"` // seconds
	// Bool fields (1 byte)
	Live       bool `json:"live,omitempty"`
	Offline    bool `json:"offline,omitempty"`
	Persistent bool `json:"persistent,omitempty"`
	Undefine   bool `json:"undefine,omitempty"`
	Compressed bool `json:"compressed,omitempty"`
}

// ExportOptions represents options for instance export.
type ExportOptions struct {
	// String fields (8 bytes)
	Format      string `json:"format"`      // ova, qcow2, vmdk, tar
	Destination string `json:"destination"` // local path or remote URL
	// Map fields (8 bytes)
	Parameters map[string]string `json:"parameters,omitempty"`
	// Bool fields (1 byte)
	Compress  bool `json:"compress,omitempty"`
	Metadata  bool `json:"metadata,omitempty"`
	Snapshots bool `json:"snapshots,omitempty"`
}

// ImportOptions represents options for instance import.
type ImportOptions struct {
	// String fields (8 bytes)
	Name string `json:"name"`
	// Slice and map fields (8 bytes)
	Networks   []NetworkAttachment `json:"networks,omitempty"`
	Storage    []StorageAttachment `json:"storage,omitempty"`
	Parameters map[string]string   `json:"parameters,omitempty"`
	// Pointer fields (8 bytes)
	Resources *ComputeResources `json:"resources,omitempty"`
	// Enum fields (typically 4 bytes)
	Backend ComputeBackend `json:"backend,omitempty"`
	// Bool fields (1 byte)
	AutoStart bool `json:"auto_start,omitempty"`
}

// ExportJob represents an export operation.
type ExportJob struct {
	// String fields (8 bytes)
	ID          string `json:"id"`
	InstanceID  string `json:"instance_id"`
	Format      string `json:"format"`
	Destination string `json:"destination"`
	State       string `json:"state"`
	Error       string `json:"error,omitempty"`
	Checksum    string `json:"checksum,omitempty"`
	// Pointer fields (8 bytes)
	CompletedAt *time.Time `json:"completed_at,omitempty"`
	// Float64 fields (8 bytes)
	Progress float64 `json:"progress"`
	// Int64 fields (8 bytes)
	Size int64 `json:"size,omitempty"`
	// Time fields (8 bytes)
	StartedAt time.Time `json:"started_at"`
}

// MetricsOptions represents options for metrics streaming.
type MetricsOptions struct {
	// Slice fields (8 bytes)
	Metrics []string `json:"metrics,omitempty"`
	// Duration fields (8 bytes)
	Interval time.Duration `json:"interval,omitempty"`
	// Int fields (4 bytes)
	Buffer int `json:"buffer,omitempty"`
}

// EventOptions represents options for event retrieval.
type EventOptions struct {
	// Slice fields (8 bytes)
	Types []string `json:"types,omitempty"`
	// Pointer fields (8 bytes)
	Since *TimeStamp `json:"since,omitempty"`
	Until *TimeStamp `json:"until,omitempty"`
	// Int fields (4 bytes)
	Limit int `json:"limit,omitempty"`
	// Bool fields (1 byte)
	Follow bool `json:"follow,omitempty"`
}

// InstanceEvent represents an instance event.
type InstanceEvent struct {
	// String fields (8 bytes)
	ID         string `json:"id"`
	InstanceID string `json:"instance_id"`
	Type       string `json:"type"`
	Action     string `json:"action"`
	Status     string `json:"status"`
	Message    string `json:"message,omitempty"`
	User       string `json:"user,omitempty"`
	// Map fields (8 bytes)
	Details map[string]interface{} `json:"details,omitempty"`
	// Time fields (8 bytes)
	Timestamp time.Time `json:"timestamp"`
}

// BulkActionOptions represents options for bulk operations.
type BulkActionOptions struct {
	// Map fields (8 bytes)
	Parameters map[string]string `json:"parameters,omitempty"`
	// Int fields (4 bytes)
	BatchSize int `json:"batch_size,omitempty"`
	Timeout   int `json:"timeout,omitempty"` // seconds
	// Bool fields (1 byte)
	Force    bool `json:"force,omitempty"`
	Parallel bool `json:"parallel,omitempty"`
}

// BulkActionResult represents the result of a bulk action on an instance.
type BulkActionResult struct {
	// String fields (8 bytes)
	InstanceID string `json:"instance_id"`
	Action     string `json:"action"`
	Error      string `json:"error,omitempty"`
	// Pointer fields (8 bytes)
	CompletedAt *time.Time `json:"completed_at,omitempty"`
	// Time fields (8 bytes)
	StartedAt time.Time `json:"started_at"`
	// Bool fields (1 byte)
	Success bool `json:"success"`
}

// BackendInfo represents information about a compute backend.
type BackendInfo struct {
	// String fields (8 bytes)
	Version    string `json:"version"`
	APIVersion string `json:"api_version"`
	Status     string `json:"status"`
	// Slice and map fields (8 bytes)
	Capabilities   []string               `json:"capabilities"`
	SupportedTypes []ComputeInstanceType  `json:"supported_types"`
	Configuration  map[string]interface{} `json:"configuration,omitempty"`
	// Pointer fields (8 bytes)
	HealthCheck *HealthStatus `json:"health_check,omitempty"`
	// Struct fields
	ResourceLimits     ComputeResources `json:"resource_limits"`
	AvailableResources ComputeResources `json:"available_resources"`
	// Enum fields
	Type ComputeBackend `json:"type"`
}

// ClusterStatus represents the overall status of the compute cluster.
type ClusterStatus struct {
	// Map fields (8 bytes)
	Backends map[ComputeBackend]*BackendInfo `json:"backends"`
	// Pointer fields (8 bytes)
	Health *HealthStatus `json:"health"`
	// Duration fields (8 bytes)
	Uptime time.Duration `json:"uptime"`
	// Time fields (8 bytes)
	LastUpdated time.Time `json:"last_updated"`
	// Struct fields
	ResourceUsage  ComputeResources `json:"resource_usage"`
	ResourceLimits ComputeResources `json:"resource_limits"`
	// Int fields (4 bytes)
	TotalInstances   int `json:"total_instances"`
	RunningInstances int `json:"running_instances"`
	StoppedInstances int `json:"stopped_instances"`
	ErrorInstances   int `json:"error_instances"`
}

// ResourceQuotas represents resource quotas for a user.
type ResourceQuotas struct {
	// Slice fields (8 bytes)
	AllowedBackends []ComputeBackend      `json:"allowed_backends"`
	AllowedTypes    []ComputeInstanceType `json:"allowed_types"`
	// Float64 fields (8 bytes)
	MaxCPUCores float64 `json:"max_cpu_cores"`
	// Time fields (8 bytes)
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	// Int fields (4 bytes)
	MaxInstances int `json:"max_instances"`
	MaxMemoryGB  int `json:"max_memory_gb"`
	MaxStorageGB int `json:"max_storage_gb"`
	MaxNetworks  int `json:"max_networks"`
	// Uint fields (4 bytes)
	UserID uint `json:"user_id"`
}

// InstanceTemplate represents a template for creating instances.
type InstanceTemplate struct {
	// String fields (8 bytes)
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	// Slice and map fields (8 bytes)
	Networks []NetworkAttachment `json:"networks"`
	Storage  []StorageAttachment `json:"storage"`
	Labels   map[string]string   `json:"labels,omitempty"`
	// Time fields (8 bytes)
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	// Struct fields
	Config    ComputeInstanceConfig `json:"config"`
	Resources ComputeResources      `json:"resources"`
	// Enum fields
	Type    ComputeInstanceType `json:"type"`
	Backend ComputeBackend      `json:"backend"`
	// Uint fields (4 bytes)
	UserID uint `json:"user_id"`
	// Bool fields (1 byte)
	Public bool `json:"public"`
}

// TemplateListOptions represents options for listing templates.
type TemplateListOptions struct {
	// String fields (8 bytes)
	SortBy    string `json:"sort_by,omitempty"`
	SortOrder string `json:"sort_order,omitempty"`
	// Map fields (8 bytes)
	Labels map[string]string `json:"labels,omitempty"`
	// Pointer fields (8 bytes)
	UserID *uint `json:"user_id,omitempty"`
	Public *bool `json:"public,omitempty"`
	// Enum fields
	Type    ComputeInstanceType `json:"type,omitempty"`
	Backend ComputeBackend      `json:"backend,omitempty"`
	// Int fields (4 bytes)
	Limit  int `json:"limit,omitempty"`
	Offset int `json:"offset,omitempty"`
}

// CloneRequest represents a request to clone from a template.
type CloneRequest struct {
	// String fields (8 bytes)
	Name string `json:"name"`
	// Slice and map fields (8 bytes)
	Networks    []NetworkAttachment `json:"networks,omitempty"`
	Storage     []StorageAttachment `json:"storage,omitempty"`
	Labels      map[string]string   `json:"labels,omitempty"`
	Annotations map[string]string   `json:"annotations,omitempty"`
	// Pointer fields (8 bytes)
	Resources *ComputeResources      `json:"resources,omitempty"`
	Overrides *ComputeInstanceConfig `json:"overrides,omitempty"`
	// Bool fields (1 byte)
	AutoStart bool `json:"auto_start,omitempty"`
}

// ComposeDeployment represents a Docker Compose deployment.
type ComposeDeployment struct {
	ID          string                      `json:"id"`
	Name        string                      `json:"name"`
	ProjectName string                      `json:"project_name"`
	Content     string                      `json:"content,omitempty"`
	State       string                      `json:"state"`
	Services    map[string]*ComputeInstance `json:"services"`
	Networks    []string                    `json:"networks"`
	Volumes     []string                    `json:"volumes"`
	Environment map[string]string           `json:"environment,omitempty"`
	Labels      map[string]string           `json:"labels,omitempty"`
	UserID      uint                        `json:"user_id"`
	CreatedAt   time.Time                   `json:"created_at"`
	UpdatedAt   time.Time                   `json:"updated_at"`
	DeployedAt  *time.Time                  `json:"deployed_at,omitempty"`
}

// ComposeDeployOptions represents options for Compose deployment.
type ComposeDeployOptions struct {
	ProjectName string            `json:"project_name,omitempty"`
	Environment map[string]string `json:"environment,omitempty"`
	Labels      map[string]string `json:"labels,omitempty"`
	AutoStart   bool              `json:"auto_start,omitempty"`
	Force       bool              `json:"force,omitempty"`
}

// ComposeListOptions represents options for listing Compose deployments.
type ComposeListOptions struct {
	UserID    *uint             `json:"user_id,omitempty"`
	State     string            `json:"state,omitempty"`
	Labels    map[string]string `json:"labels,omitempty"`
	Limit     int               `json:"limit,omitempty"`
	Offset    int               `json:"offset,omitempty"`
	SortBy    string            `json:"sort_by,omitempty"`
	SortOrder string            `json:"sort_order,omitempty"`
}

// HealthStatus represents health check results.
type HealthStatus struct {
	Status       string            `json:"status"` // healthy, unhealthy, unknown
	Message      string            `json:"message,omitempty"`
	Details      map[string]string `json:"details,omitempty"`
	LastCheck    time.Time         `json:"last_check"`
	CheckCount   int               `json:"check_count"`
	FailureCount int               `json:"failure_count"`
}

// MaintenanceOptions represents options for maintenance operations.
type MaintenanceOptions struct {
	Type       string            `json:"type"` // cleanup, backup, update
	Force      bool              `json:"force,omitempty"`
	DryRun     bool              `json:"dry_run,omitempty"`
	Timeout    int               `json:"timeout,omitempty"`  // seconds
	Schedule   string            `json:"schedule,omitempty"` // cron expression
	Notify     bool              `json:"notify,omitempty"`
	Parameters map[string]string `json:"parameters,omitempty"`
}
