package compute

import (
	"time"
)

// ComputeBackend represents the underlying virtualization technology.
type ComputeBackend string

const (
	BackendKVM    ComputeBackend = "kvm"
	BackendDocker ComputeBackend = "docker"
)

// ComputeInstanceType defines the type of compute instance.
type ComputeInstanceType string

const (
	InstanceTypeVM        ComputeInstanceType = "vm"
	InstanceTypeContainer ComputeInstanceType = "container"
	InstanceTypeService   ComputeInstanceType = "service" // Docker compose service
)

// ComputeInstanceState represents the state of a compute instance.
type ComputeInstanceState string

const (
	StateCreated    ComputeInstanceState = "created"
	StateStarting   ComputeInstanceState = "starting"
	StateRunning    ComputeInstanceState = "running"
	StateStopping   ComputeInstanceState = "stopping"
	StateStopped    ComputeInstanceState = "stopped"
	StatePaused     ComputeInstanceState = "paused"
	StateError      ComputeInstanceState = "error"
	StateMigrating  ComputeInstanceState = "migrating"
	StateRestarting ComputeInstanceState = "restarting"
	StateUnknown    ComputeInstanceState = "unknown"
)

// ComputeInstance represents a unified compute resource (VM or Container).
type ComputeInstance struct {
	// Identity
	ID     string `json:"id"`
	Name   string `json:"name"`
	UUID   string `json:"uuid,omitempty"`
	UserID uint   `json:"user_id"`

	// Type and Backend
	Type    ComputeInstanceType `json:"type"`
	Backend ComputeBackend      `json:"backend"`

	// State
	State       ComputeInstanceState `json:"state"`
	Status      string               `json:"status"` // Detailed status message
	HealthState string               `json:"health_state,omitempty"`

	// Configuration
	Config ComputeInstanceConfig `json:"config"`

	// Resources
	Resources ComputeResources `json:"resources"`
	Limits    ComputeResources `json:"limits,omitempty"`

	// Networking
	Networks []NetworkAttachment `json:"networks"`

	// Storage
	Storage []StorageAttachment `json:"storage"`

	// Metadata
	Labels      map[string]string `json:"labels,omitempty"`
	Annotations map[string]string `json:"annotations,omitempty"`

	// Runtime Information
	RuntimeInfo RuntimeInfo `json:"runtime_info"`

	// Timestamps
	CreatedAt  time.Time  `json:"created_at"`
	UpdatedAt  time.Time  `json:"updated_at"`
	StartedAt  *time.Time `json:"started_at,omitempty"`
	FinishedAt *time.Time `json:"finished_at,omitempty"`

	// Backend-specific data (opaque)
	BackendData map[string]interface{} `json:"backend_data,omitempty"`
}

// ComputeInstanceConfig holds instance configuration.
type ComputeInstanceConfig struct {
	// Common configuration
	Image       string            `json:"image"`
	Command     []string          `json:"command,omitempty"`
	Args        []string          `json:"args,omitempty"`
	Environment map[string]string `json:"environment,omitempty"`
	WorkingDir  string            `json:"working_dir,omitempty"`
	User        string            `json:"user,omitempty"`

	// VM-specific configuration
	Firmware   string           `json:"firmware,omitempty"`
	SecureBoot bool             `json:"secure_boot,omitempty"`
	TPMEnabled bool             `json:"tpm_enabled,omitempty"`
	CloudInit  *CloudInitConfig `json:"cloud_init,omitempty"`

	// Container-specific configuration
	Privileged      bool             `json:"privileged,omitempty"`
	Capabilities    []string         `json:"capabilities,omitempty"`
	SecurityContext *SecurityContext `json:"security_context,omitempty"`

	// Restart policy
	RestartPolicy RestartPolicy `json:"restart_policy"`

	// Health check
	HealthCheck *HealthCheck `json:"health_check,omitempty"`
}

// CloudInitConfig holds cloud-init configuration for VMs.
type CloudInitConfig struct {
	Enabled     bool             `json:"enabled"`
	UserData    string           `json:"user_data,omitempty"`
	MetaData    string           `json:"meta_data,omitempty"`
	NetworkData string           `json:"network_data,omitempty"`
	VendorData  string           `json:"vendor_data,omitempty"`
	KeyPairs    []string         `json:"key_pairs,omitempty"`
	Hostname    string           `json:"hostname,omitempty"`
	Users       []CloudInitUser  `json:"users,omitempty"`
	Packages    []string         `json:"packages,omitempty"`
	RunCommands []string         `json:"run_commands,omitempty"`
	WriteFiles  []CloudInitFile  `json:"write_files,omitempty"`
	MountPoints []CloudInitMount `json:"mount_points,omitempty"`
}

// CloudInitUser represents a user in cloud-init.
type CloudInitUser struct {
	Name              string   `json:"name"`
	Shell             string   `json:"shell,omitempty"`
	Groups            []string `json:"groups,omitempty"`
	Sudo              string   `json:"sudo,omitempty"`
	SSHAuthorizedKeys []string `json:"ssh_authorized_keys,omitempty"`
	LockPassword      bool     `json:"lock_password,omitempty"`
}

// CloudInitFile represents a file to write via cloud-init.
type CloudInitFile struct {
	Path        string `json:"path"`
	Content     string `json:"content"`
	Permissions string `json:"permissions,omitempty"`
	Owner       string `json:"owner,omitempty"`
	Encoding    string `json:"encoding,omitempty"`
}

// CloudInitMount represents a mount point in cloud-init.
type CloudInitMount struct {
	Source     string   `json:"source"`
	Target     string   `json:"target"`
	Filesystem string   `json:"filesystem,omitempty"`
	Options    []string `json:"options,omitempty"`
}

// SecurityContext holds security-related configuration.
type SecurityContext struct {
	RunAsUser       *int64           `json:"run_as_user,omitempty"`
	RunAsGroup      *int64           `json:"run_as_group,omitempty"`
	RunAsNonRoot    *bool            `json:"run_as_non_root,omitempty"`
	ReadOnlyRootFS  *bool            `json:"read_only_root_fs,omitempty"`
	AllowPrivileged *bool            `json:"allow_privileged,omitempty"`
	Capabilities    *Capabilities    `json:"capabilities,omitempty"`
	SELinuxOptions  *SELinuxOptions  `json:"selinux_options,omitempty"`
	SeccompProfile  *SeccompProfile  `json:"seccomp_profile,omitempty"`
	AppArmorProfile *AppArmorProfile `json:"apparmor_profile,omitempty"`
}

// Capabilities holds Linux capability configuration.
type Capabilities struct {
	Add  []string `json:"add,omitempty"`
	Drop []string `json:"drop,omitempty"`
}

// SELinuxOptions holds SELinux configuration.
type SELinuxOptions struct {
	User  string `json:"user,omitempty"`
	Role  string `json:"role,omitempty"`
	Type  string `json:"type,omitempty"`
	Level string `json:"level,omitempty"`
}

// SeccompProfile holds seccomp configuration.
type SeccompProfile struct {
	Type             string `json:"type"`
	LocalhostProfile string `json:"localhost_profile,omitempty"`
}

// AppArmorProfile holds AppArmor configuration.
type AppArmorProfile struct {
	Type             string `json:"type"`
	LocalhostProfile string `json:"localhost_profile,omitempty"`
}

// RestartPolicy defines restart behavior.
type RestartPolicy struct {
	Policy            string `json:"policy"` // always, on-failure, unless-stopped, no
	MaximumRetryCount int    `json:"maximum_retry_count,omitempty"`
}

// HealthCheck defines health check configuration.
type HealthCheck struct {
	Test        []string      `json:"test"`
	Interval    time.Duration `json:"interval"`
	Timeout     time.Duration `json:"timeout"`
	StartPeriod time.Duration `json:"start_period,omitempty"`
	Retries     int           `json:"retries"`
}

// ComputeResources represents resource allocation.
type ComputeResources struct {
	CPU     CPUResources     `json:"cpu"`
	Memory  MemoryResources  `json:"memory"`
	Storage StorageResources `json:"storage,omitempty"`
	Network NetworkResources `json:"network,omitempty"`
	GPU     []GPUResources   `json:"gpu,omitempty"`
}

// CPUResources represents CPU allocation.
type CPUResources struct {
	Cores    float64     `json:"cores"`              // Number of CPU cores (can be fractional)
	Shares   int64       `json:"shares,omitempty"`   // CPU shares (relative weight)
	Quota    int64       `json:"quota,omitempty"`    // CPU quota in microseconds
	Period   int64       `json:"period,omitempty"`   // CPU period in microseconds
	SetCPUs  string      `json:"set_cpus,omitempty"` // CPU affinity (e.g., "0-3,8-11")
	SetMems  string      `json:"set_mems,omitempty"` // Memory node affinity
	Topology CPUTopology `json:"topology,omitempty"` // CPU topology for VMs
}

// CPUTopology defines CPU topology for VMs.
type CPUTopology struct {
	Sockets int `json:"sockets"`
	Cores   int `json:"cores"`
	Threads int `json:"threads"`
}

// MemoryResources represents memory allocation.
type MemoryResources struct {
	Limit       int64 `json:"limit"`                 // Memory limit in bytes
	Request     int64 `json:"request,omitempty"`     // Memory request in bytes
	Swap        int64 `json:"swap,omitempty"`        // Swap limit in bytes
	Reservation int64 `json:"reservation,omitempty"` // Memory reservation in bytes
	Swappiness  *int  `json:"swappiness,omitempty"`  // Swappiness (0-100)
}

// StorageResources represents storage allocation.
type StorageResources struct {
	TotalSpace int64       `json:"total_space,omitempty"` // Total storage space in bytes
	UsedSpace  int64       `json:"used_space,omitempty"`  // Used storage space in bytes
	IOPS       *IOPSLimits `json:"iops,omitempty"`        // IOPS limits
}

// IOPSLimits represents IOPS limitations.
type IOPSLimits struct {
	ReadIOPS  int64 `json:"read_iops,omitempty"`
	WriteIOPS int64 `json:"write_iops,omitempty"`
	ReadBPS   int64 `json:"read_bps,omitempty"`
	WriteBPS  int64 `json:"write_bps,omitempty"`
}

// NetworkResources represents network resource allocation.
type NetworkResources struct {
	BandwidthLimit int64 `json:"bandwidth_limit,omitempty"` // Network bandwidth limit in bps
	PacketLimit    int64 `json:"packet_limit,omitempty"`    // Packet rate limit in pps
}

// GPUResources represents GPU allocation.
type GPUResources struct {
	DeviceID    string  `json:"device_id"`
	Model       string  `json:"model,omitempty"`
	Memory      int64   `json:"memory,omitempty"`      // GPU memory in bytes
	Utilization float64 `json:"utilization,omitempty"` // GPU utilization percentage
}

// NetworkAttachment represents a network interface attachment.
type NetworkAttachment struct {
	Name        string            `json:"name"`
	Interface   string            `json:"interface"`
	Network     string            `json:"network"`
	NetworkID   string            `json:"network_id,omitempty"`
	Driver      string            `json:"driver,omitempty"`
	MacAddress  string            `json:"mac_address,omitempty"`
	IPAddress   string            `json:"ip_address,omitempty"`
	IPv6Address string            `json:"ipv6_address,omitempty"`
	Gateway     string            `json:"gateway,omitempty"`
	DNS         []string          `json:"dns,omitempty"`
	Routes      []Route           `json:"routes,omitempty"`
	Options     map[string]string `json:"options,omitempty"`
	Firewall    *FirewallConfig   `json:"firewall,omitempty"`
}

// Route represents a network route.
type Route struct {
	Destination string `json:"destination"`
	Gateway     string `json:"gateway"`
	Interface   string `json:"interface,omitempty"`
	Metric      int    `json:"metric,omitempty"`
}

// FirewallConfig represents firewall configuration for a network interface.
type FirewallConfig struct {
	Enabled bool           `json:"enabled"`
	Rules   []FirewallRule `json:"rules,omitempty"`
}

// FirewallRule represents a firewall rule.
type FirewallRule struct {
	Action    string `json:"action"`    // allow, deny, reject
	Direction string `json:"direction"` // in, out
	Protocol  string `json:"protocol,omitempty"`
	Port      string `json:"port,omitempty"`
	Source    string `json:"source,omitempty"`
	Target    string `json:"target,omitempty"`
}

// StorageAttachment represents a storage volume attachment.
type StorageAttachment struct {
	Name       string            `json:"name"`
	Source     string            `json:"source"`
	Target     string            `json:"target"`
	VolumeID   string            `json:"volume_id,omitempty"`
	Driver     string            `json:"driver,omitempty"`
	Type       string            `json:"type"`           // bind, volume, tmpfs
	Mode       string            `json:"mode,omitempty"` // ro, rw
	Options    []string          `json:"options,omitempty"`
	Size       int64             `json:"size,omitempty"`
	Format     string            `json:"format,omitempty"`  // qcow2, raw, etc.
	Cache      string            `json:"cache,omitempty"`   // none, writethrough, writeback
	IOMode     string            `json:"io_mode,omitempty"` // native, threads
	Backup     bool              `json:"backup,omitempty"`
	Snapshot   bool              `json:"snapshot,omitempty"`
	Encryption *EncryptionConfig `json:"encryption,omitempty"`
}

// EncryptionConfig represents storage encryption configuration.
type EncryptionConfig struct {
	Enabled   bool   `json:"enabled"`
	Algorithm string `json:"algorithm,omitempty"`
	KeyID     string `json:"key_id,omitempty"`
}

// RuntimeInfo holds runtime information about a compute instance.
type RuntimeInfo struct {
	// Process information
	ProcessID int `json:"process_id,omitempty"`
	ExitCode  int `json:"exit_code,omitempty"`

	// Resource usage
	ResourceUsage ResourceUsage `json:"resource_usage"`

	// Network information
	Networks map[string]NetworkRuntimeInfo `json:"networks,omitempty"`

	// Storage information
	Storage map[string]StorageRuntimeInfo `json:"storage,omitempty"`

	// Performance metrics
	Performance PerformanceMetrics `json:"performance,omitempty"`

	// Host information
	HostInfo HostInfo `json:"host_info"`
}

// ResourceUsage represents current resource usage.
type ResourceUsage struct {
	CPU       CPUUsage     `json:"cpu"`
	Memory    MemoryUsage  `json:"memory"`
	Network   NetworkUsage `json:"network"`
	Storage   StorageUsage `json:"storage"`
	GPU       []GPUUsage   `json:"gpu,omitempty"`
	Timestamp time.Time    `json:"timestamp"`
}

// CPUUsage represents CPU usage statistics.
type CPUUsage struct {
	Usage         float64 `json:"usage"`          // CPU usage percentage
	UsageNanos    int64   `json:"usage_nanos"`    // CPU usage in nanoseconds
	SystemUsage   int64   `json:"system_usage"`   // System CPU usage
	OnlineCPUs    int     `json:"online_cpus"`    // Number of online CPUs
	ThrottledTime int64   `json:"throttled_time"` // Time throttled in nanoseconds
}

// MemoryUsage represents memory usage statistics.
type MemoryUsage struct {
	Usage        int64   `json:"usage"`         // Memory usage in bytes
	MaxUsage     int64   `json:"max_usage"`     // Maximum memory usage in bytes
	Limit        int64   `json:"limit"`         // Memory limit in bytes
	Available    int64   `json:"available"`     // Available memory in bytes
	UsagePercent float64 `json:"usage_percent"` // Memory usage percentage
	Cache        int64   `json:"cache"`         // Cache memory in bytes
	RSS          int64   `json:"rss"`           // Resident set size in bytes
	Swap         int64   `json:"swap"`          // Swap usage in bytes
}

// NetworkUsage represents network usage statistics.
type NetworkUsage struct {
	RxBytes   int64 `json:"rx_bytes"`   // Bytes received
	TxBytes   int64 `json:"tx_bytes"`   // Bytes transmitted
	RxPackets int64 `json:"rx_packets"` // Packets received
	TxPackets int64 `json:"tx_packets"` // Packets transmitted
	RxErrors  int64 `json:"rx_errors"`  // Receive errors
	TxErrors  int64 `json:"tx_errors"`  // Transmit errors
	RxDropped int64 `json:"rx_dropped"` // Dropped received packets
	TxDropped int64 `json:"tx_dropped"` // Dropped transmitted packets
}

// StorageUsage represents storage usage statistics.
type StorageUsage struct {
	ReadBytes  int64 `json:"read_bytes"`  // Bytes read
	WriteBytes int64 `json:"write_bytes"` // Bytes written
	ReadOps    int64 `json:"read_ops"`    // Read operations
	WriteOps   int64 `json:"write_ops"`   // Write operations
	ReadTime   int64 `json:"read_time"`   // Time spent reading (ms)
	WriteTime  int64 `json:"write_time"`  // Time spent writing (ms)
}

// GPUUsage represents GPU usage statistics.
type GPUUsage struct {
	DeviceID         string  `json:"device_id"`
	Utilization      float64 `json:"utilization"`        // GPU utilization percentage
	MemoryUsed       int64   `json:"memory_used"`        // GPU memory used in bytes
	MemoryTotal      int64   `json:"memory_total"`       // Total GPU memory in bytes
	Temperature      float64 `json:"temperature"`        // GPU temperature in Celsius
	PowerUsage       float64 `json:"power_usage"`        // Power usage in watts
	ClockSpeed       int64   `json:"clock_speed"`        // Clock speed in MHz
	MemoryClockSpeed int64   `json:"memory_clock_speed"` // Memory clock speed in MHz
}

// NetworkRuntimeInfo represents runtime network interface information.
type NetworkRuntimeInfo struct {
	InterfaceName string `json:"interface_name"`
	IPAddress     string `json:"ip_address"`
	IPv6Address   string `json:"ipv6_address,omitempty"`
	MacAddress    string `json:"mac_address"`
	MTU           int    `json:"mtu"`
	State         string `json:"state"`           // up, down, unknown
	Speed         int64  `json:"speed,omitempty"` // Link speed in bps
}

// StorageRuntimeInfo represents runtime storage information.
type StorageRuntimeInfo struct {
	MountPoint   string  `json:"mount_point"`
	Filesystem   string  `json:"filesystem"`
	Size         int64   `json:"size"`
	Used         int64   `json:"used"`
	Available    int64   `json:"available"`
	UsagePercent float64 `json:"usage_percent"`
}

// PerformanceMetrics represents performance-related metrics.
type PerformanceMetrics struct {
	Uptime          time.Duration `json:"uptime"`
	LoadAverage     LoadAverage   `json:"load_average"`
	ProcessCount    int           `json:"process_count"`
	ThreadCount     int           `json:"thread_count"`
	FileDescriptors int           `json:"file_descriptors"`
	ContextSwitches int64         `json:"context_switches"`
	Interrupts      int64         `json:"interrupts"`
}

// LoadAverage represents system load averages.
type LoadAverage struct {
	Load1  float64 `json:"load1"`  // 1-minute load average
	Load5  float64 `json:"load5"`  // 5-minute load average
	Load15 float64 `json:"load15"` // 15-minute load average
}

// HostInfo represents information about the host system.
type HostInfo struct {
	NodeName      string `json:"node_name"`
	KernelVersion string `json:"kernel_version"`
	OSType        string `json:"os_type"`
	Architecture  string `json:"architecture"`
	CPUCount      int    `json:"cpu_count"`
	TotalMemory   int64  `json:"total_memory"`
	Hypervisor    string `json:"hypervisor"` // kvm, docker
}

// ComputeInstanceRequest represents a request to create a compute instance.
type ComputeInstanceRequest struct {
	Name        string                `json:"name" validate:"required"`
	UUID        string                `json:"uuid,omitempty"`
	UserID      uint                  `json:"user_id,omitempty"`
	Type        ComputeInstanceType   `json:"type" validate:"required"`
	Backend     ComputeBackend        `json:"backend,omitempty"`
	Config      ComputeInstanceConfig `json:"config" validate:"required"`
	Resources   ComputeResources      `json:"resources" validate:"required"`
	Limits      *ComputeResources     `json:"limits,omitempty"`
	Networks    []NetworkAttachment   `json:"networks,omitempty"`
	Storage     []StorageAttachment   `json:"storage,omitempty"`
	Labels      map[string]string     `json:"labels,omitempty"`
	Annotations map[string]string     `json:"annotations,omitempty"`
	AutoStart   bool                  `json:"auto_start,omitempty"`
}

// ComputeInstanceUpdate represents an update to a compute instance.
type ComputeInstanceUpdate struct {
	Resources   *ComputeResources      `json:"resources,omitempty"`
	Limits      *ComputeResources      `json:"limits,omitempty"`
	Labels      map[string]string      `json:"labels,omitempty"`
	Annotations map[string]string      `json:"annotations,omitempty"`
	Config      *ComputeInstanceConfig `json:"config,omitempty"`
}

// ComputeInstanceListOptions represents options for listing compute instances.
type ComputeInstanceListOptions struct {
	Backend   ComputeBackend       `json:"backend,omitempty"`
	Type      ComputeInstanceType  `json:"type,omitempty"`
	State     ComputeInstanceState `json:"state,omitempty"`
	Labels    map[string]string    `json:"labels,omitempty"`
	UserID    uint                 `json:"user_id,omitempty"`
	Limit     int                  `json:"limit,omitempty"`
	Offset    int                  `json:"offset,omitempty"`
	SortBy    string               `json:"sort_by,omitempty"`
	SortOrder string               `json:"sort_order,omitempty"`
}
