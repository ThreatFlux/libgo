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
// Field alignment optimized: structs first, time, maps/slices, pointers, strings, enums, primitives.
type ComputeInstance struct {
	// Struct fields (largest first)
	Config      ComputeInstanceConfig `json:"config"`           // ~200+ bytes
	Resources   ComputeResources      `json:"resources"`        // ~160+ bytes
	Limits      ComputeResources      `json:"limits,omitempty"` // ~160+ bytes
	RuntimeInfo RuntimeInfo           `json:"runtime_info"`     // ~150+ bytes
	// Time fields (24 bytes each)
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	// Map and slice fields (24 bytes each)
	Labels      map[string]string      `json:"labels,omitempty"`
	Annotations map[string]string      `json:"annotations,omitempty"`
	BackendData map[string]interface{} `json:"backend_data,omitempty"`
	Storage     []StorageAttachment    `json:"storage"`
	Networks    []NetworkAttachment    `json:"networks"`
	// Pointer fields (8 bytes each)
	StartedAt  *time.Time `json:"started_at,omitempty"`
	FinishedAt *time.Time `json:"finished_at,omitempty"`
	// String fields (16 bytes each) - group together
	Name        string `json:"name"`
	Status      string `json:"status"`
	UUID        string `json:"uuid,omitempty"`
	ID          string `json:"id"`
	HealthState string `json:"health_state,omitempty"`
	// Uint fields (8 bytes on 64-bit)
	UserID uint `json:"user_id"`
	// Enum fields (4 bytes each) - group together
	Backend ComputeBackend       `json:"backend"`
	Type    ComputeInstanceType  `json:"type"`
	State   ComputeInstanceState `json:"state"`
}

// ComputeInstanceConfig holds instance configuration.
type ComputeInstanceConfig struct {
	CloudInit       *CloudInitConfig  `json:"cloud_init,omitempty"`
	SecurityContext *SecurityContext  `json:"security_context,omitempty"`
	Environment     map[string]string `json:"environment,omitempty"`
	HealthCheck     *HealthCheck      `json:"health_check,omitempty"`
	WorkingDir      string            `json:"working_dir,omitempty"`
	User            string            `json:"user,omitempty"`
	Firmware        string            `json:"firmware,omitempty"`
	Image           string            `json:"image"`
	Args            []string          `json:"args,omitempty"`
	Command         []string          `json:"command,omitempty"`
	RestartPolicy   RestartPolicy     `json:"restart_policy"`
	Capabilities    []string          `json:"capabilities,omitempty"`
	Privileged      bool              `json:"privileged,omitempty"`
	SecureBoot      bool              `json:"secure_boot,omitempty"`
	TPMEnabled      bool              `json:"tpm_enabled,omitempty"`
}

// CloudInitConfig holds cloud-init configuration for VMs.
type CloudInitConfig struct {
	UserData    string           `json:"user_data,omitempty"`
	MetaData    string           `json:"meta_data,omitempty"`
	NetworkData string           `json:"network_data,omitempty"`
	VendorData  string           `json:"vendor_data,omitempty"`
	Hostname    string           `json:"hostname,omitempty"`
	KeyPairs    []string         `json:"key_pairs,omitempty"`
	Users       []CloudInitUser  `json:"users,omitempty"`
	Packages    []string         `json:"packages,omitempty"`
	RunCommands []string         `json:"run_commands,omitempty"`
	WriteFiles  []CloudInitFile  `json:"write_files,omitempty"`
	MountPoints []CloudInitMount `json:"mount_points,omitempty"`
	Enabled     bool             `json:"enabled"`
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
// Field alignment optimized: slices first, then structs by size.
type ComputeResources struct {
	// Slice fields (24 bytes)
	GPU []GPUResources `json:"gpu,omitempty"`
	// Struct fields (ordered by size - largest first)
	Memory  MemoryResources  `json:"memory"`            // ~40 bytes
	CPU     CPUResources     `json:"cpu"`               // ~80 bytes with embedded struct
	Storage StorageResources `json:"storage,omitempty"` // ~24 bytes
	Network NetworkResources `json:"network,omitempty"` // ~16 bytes
}

// CPUResources represents CPU allocation.
// Field alignment optimized: structs first, then strings, then float64, then int64.
type CPUResources struct {
	// Struct fields (largest first)
	Topology CPUTopology `json:"topology,omitempty"` // ~24 bytes (3 ints)
	// String fields (16 bytes each)
	SetCPUs string `json:"set_cpus,omitempty"`
	SetMems string `json:"set_mems,omitempty"`
	// Float64 fields (8 bytes)
	Cores float64 `json:"cores"`
	// Int64 fields (8 bytes each) - group together
	Shares int64 `json:"shares,omitempty"`
	Quota  int64 `json:"quota,omitempty"`
	Period int64 `json:"period,omitempty"`
}

// CPUTopology defines CPU topology for VMs.
type CPUTopology struct {
	Sockets int `json:"sockets"`
	Cores   int `json:"cores"`
	Threads int `json:"threads"`
}

// MemoryResources represents memory allocation.
// Field alignment optimized: int64s first, then pointers.
type MemoryResources struct {
	// Int64 fields (8 bytes each) - group together
	Limit       int64 `json:"limit"`
	Request     int64 `json:"request,omitempty"`
	Swap        int64 `json:"swap,omitempty"`
	Reservation int64 `json:"reservation,omitempty"`
	// Pointer fields (8 bytes)
	Swappiness *int `json:"swappiness,omitempty"`
}

// StorageResources represents storage allocation.
type StorageResources struct {
	IOPS       *IOPSLimits `json:"iops,omitempty"`
	TotalSpace int64       `json:"total_space,omitempty"`
	UsedSpace  int64       `json:"used_space,omitempty"`
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
// Field alignment optimized: maps/slices first, then pointers, then strings.
type NetworkAttachment struct {
	// Map and slice fields (24 bytes each)
	Options map[string]string `json:"options,omitempty"`
	DNS     []string          `json:"dns,omitempty"`
	Routes  []Route           `json:"routes,omitempty"`
	// Pointer fields (8 bytes)
	Firewall *FirewallConfig `json:"firewall,omitempty"`
	// String fields (16 bytes each) - group together
	Name        string `json:"name"`
	Network     string `json:"network"`
	Interface   string `json:"interface"`
	IPAddress   string `json:"ip_address,omitempty"`
	NetworkID   string `json:"network_id,omitempty"`
	Driver      string `json:"driver,omitempty"`
	MacAddress  string `json:"mac_address,omitempty"`
	IPv6Address string `json:"ipv6_address,omitempty"`
	Gateway     string `json:"gateway,omitempty"`
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
	Rules   []FirewallRule `json:"rules,omitempty"`
	Enabled bool           `json:"enabled"`
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
	Encryption *EncryptionConfig `json:"encryption,omitempty"`
	Mode       string            `json:"mode,omitempty"`
	IOMode     string            `json:"io_mode,omitempty"`
	VolumeID   string            `json:"volume_id,omitempty"`
	Driver     string            `json:"driver,omitempty"`
	Type       string            `json:"type"`
	Name       string            `json:"name"`
	Source     string            `json:"source"`
	Target     string            `json:"target"`
	Format     string            `json:"format,omitempty"`
	Cache      string            `json:"cache,omitempty"`
	Options    []string          `json:"options,omitempty"`
	Size       int64             `json:"size,omitempty"`
	Backup     bool              `json:"backup,omitempty"`
	Snapshot   bool              `json:"snapshot,omitempty"`
}

// EncryptionConfig represents storage encryption configuration.
type EncryptionConfig struct {
	Algorithm string `json:"algorithm,omitempty"`
	KeyID     string `json:"key_id,omitempty"`
	Enabled   bool   `json:"enabled"`
}

// RuntimeInfo holds runtime information about a compute instance.
type RuntimeInfo struct {
	Networks      map[string]NetworkRuntimeInfo `json:"networks,omitempty"`
	Storage       map[string]StorageRuntimeInfo `json:"storage,omitempty"`
	HostInfo      HostInfo                      `json:"host_info"`
	ResourceUsage ResourceUsage                 `json:"resource_usage"`
	Performance   PerformanceMetrics            `json:"performance,omitempty"`
	ProcessID     int                           `json:"process_id,omitempty"`
	ExitCode      int                           `json:"exit_code,omitempty"`
}

// ResourceUsage represents current resource usage.
type ResourceUsage struct {
	Timestamp time.Time    `json:"timestamp"`
	GPU       []GPUUsage   `json:"gpu,omitempty"`
	Memory    MemoryUsage  `json:"memory"`
	Network   NetworkUsage `json:"network"`
	Storage   StorageUsage `json:"storage"`
	CPU       CPUUsage     `json:"cpu"`
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
	State         string `json:"state"`
	MTU           int    `json:"mtu"`
	Speed         int64  `json:"speed,omitempty"`
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
	Hypervisor    string `json:"hypervisor"`
	CPUCount      int    `json:"cpu_count"`
	TotalMemory   int64  `json:"total_memory"`
}

// ComputeInstanceRequest represents a request to create a compute instance.
type ComputeInstanceRequest struct {
	Limits      *ComputeResources     `json:"limits,omitempty"`
	Labels      map[string]string     `json:"labels,omitempty"`
	Annotations map[string]string     `json:"annotations,omitempty"`
	Backend     ComputeBackend        `json:"backend,omitempty"`
	UUID        string                `json:"uuid,omitempty"`
	Type        ComputeInstanceType   `json:"type" validate:"required"`
	Name        string                `json:"name" validate:"required"`
	Storage     []StorageAttachment   `json:"storage,omitempty"`
	Networks    []NetworkAttachment   `json:"networks,omitempty"`
	Config      ComputeInstanceConfig `json:"config" validate:"required"`
	Resources   ComputeResources      `json:"resources" validate:"required"`
	UserID      uint                  `json:"user_id,omitempty"`
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
	Labels    map[string]string    `json:"labels,omitempty"`
	Backend   ComputeBackend       `json:"backend,omitempty"`
	Type      ComputeInstanceType  `json:"type,omitempty"`
	State     ComputeInstanceState `json:"state,omitempty"`
	SortBy    string               `json:"sort_by,omitempty"`
	SortOrder string               `json:"sort_order,omitempty"`
	UserID    uint                 `json:"user_id,omitempty"`
	Limit     int                  `json:"limit,omitempty"`
	Offset    int                  `json:"offset,omitempty"`
}
