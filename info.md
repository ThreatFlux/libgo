# Source Code Files & Function Details

## Table of Contents
- [Configuration System](#configuration-system)
- [Logging System](#logging-system)
- [Utilities](#utilities)
- [Libvirt Integration](#libvirt-integration)
- [VM Management](#vm-management)
- [Authentication System](#authentication-system)
- [API Layer](#api-layer)
- [Export Functionality](#export-functionality)
- [Metrics and Monitoring](#metrics-and-monitoring)
- [Application Entry Point](#application-entry-point)

## Configuration System

### `internal/config/types.go`

**Purpose**: Define configuration data structures

**Types**:
```go
// Config holds all application configuration
type Config struct {
    Server   ServerConfig   `yaml:"server" json:"server"`
    Libvirt  LibvirtConfig  `yaml:"libvirt" json:"libvirt"`
    Auth     AuthConfig     `yaml:"auth" json:"auth"`
    Logging  LoggingConfig  `yaml:"logging" json:"logging"`
    Storage  StorageConfig  `yaml:"storage" json:"storage"`
    Export   ExportConfig   `yaml:"export" json:"export"`
    Features FeaturesConfig `yaml:"features" json:"features"`
}

// ServerConfig holds HTTP server configuration
type ServerConfig struct {
    Host           string        `yaml:"host" json:"host"`
    Port           int           `yaml:"port" json:"port"`
    ReadTimeout    time.Duration `yaml:"readTimeout" json:"readTimeout"`
    WriteTimeout   time.Duration `yaml:"writeTimeout" json:"writeTimeout"`
    MaxHeaderBytes int           `yaml:"maxHeaderBytes" json:"maxHeaderBytes"`
    TLS            TLSConfig     `yaml:"tls" json:"tls"`
}

// LibvirtConfig holds libvirt connection settings
type LibvirtConfig struct {
    URI               string        `yaml:"uri" json:"uri"`
    ConnectionTimeout time.Duration `yaml:"connectionTimeout" json:"connectionTimeout"`
    MaxConnections    int           `yaml:"maxConnections" json:"maxConnections"`
    PoolName          string        `yaml:"poolName" json:"poolName"`
    NetworkName       string        `yaml:"networkName" json:"networkName"`
}

// More config types for other components...
```

### `internal/config/loader_interface.go`

**Purpose**: Define the configuration loading interface

**Interfaces**:
```go
// Loader is the interface for loading configuration
type Loader interface {
    // Load loads configuration from a source into the provided config struct
    Load(cfg *types.Config) error
    
    // LoadFromFile loads configuration from a specific file
    LoadFromFile(filePath string, cfg *types.Config) error
    
    // LoadWithOverrides loads configuration with environment variable overrides
    LoadWithOverrides(cfg *types.Config) error
}
```

### `internal/config/yaml_loader.go`

**Purpose**: Implement YAML configuration loader

**Types**:
```go
// YAMLLoader implements Loader for YAML files
type YAMLLoader struct {
    // Default config file path
    DefaultPath string
}
```

**Functions**:
```go
// NewYAMLLoader creates a new YAML config loader
func NewYAMLLoader(defaultPath string) *YAMLLoader
    Returns: A new YAML loader instance

// Load implements Loader.Load for YAML files
func (l *YAMLLoader) Load(cfg *types.Config) error
    Params: cfg - Pointer to Config structure to populate
    Returns: Error if loading fails

// LoadFromFile implements Loader.LoadFromFile for YAML files
func (l *YAMLLoader) LoadFromFile(filePath string, cfg *types.Config) error
    Params: 
        filePath - Path to YAML config file
        cfg - Pointer to Config structure to populate
    Returns: Error if loading fails

// LoadWithOverrides implements Loader.LoadWithOverrides
func (l *YAMLLoader) LoadWithOverrides(cfg *types.Config) error
    Params: cfg - Pointer to Config structure to populate with overrides
    Returns: Error if loading or overriding fails
```

### `internal/config/validator.go`

**Purpose**: Validate configuration values

**Functions**:
```go
// Validate checks if the configuration is valid
func Validate(cfg *types.Config) error
    Params: cfg - Configuration to validate
    Returns: Error if validation fails

// ValidateServer validates server configuration
func ValidateServer(server types.ServerConfig) error
    Params: server - Server configuration to validate
    Returns: Error if validation fails

// ValidateLibvirt validates libvirt configuration
func ValidateLibvirt(libvirt types.LibvirtConfig) error
    Params: libvirt - Libvirt configuration to validate
    Returns: Error if validation fails

// Additional validation functions for other config sections...
```

## Logging System

### `pkg/logger/interface.go`

**Purpose**: Define logging interface

**Interfaces**:
```go
// Logger defines the interface for logging
type Logger interface {
    // Debug logs a message at debug level
    Debug(msg string, fields ...Field)
    
    // Info logs a message at info level
    Info(msg string, fields ...Field)
    
    // Warn logs a message at warning level
    Warn(msg string, fields ...Field)
    
    // Error logs a message at error level
    Error(msg string, fields ...Field)
    
    // Fatal logs a message at fatal level then calls os.Exit(1)
    Fatal(msg string, fields ...Field)
    
    // WithFields returns a new Logger with the given fields added
    WithFields(fields ...Field) Logger
    
    // WithError returns a new Logger with the given error attached
    WithError(err error) Logger
    
    // Sync flushes any buffered log entries
    Sync() error
}

// Field represents a structured log field
type Field struct {
    Key   string
    Value interface{}
}
```

**Functions**:
```go
// String creates a string Field
func String(key, value string) Field
    Params: key, value - Key-value pair for the field
    Returns: A Field with string value

// Int creates an integer Field
func Int(key string, value int) Field
    Params: key, value - Key-value pair for the field
    Returns: A Field with int value

// Error creates an error Field
func Error(err error) Field
    Params: err - Error to log
    Returns: A Field with error value

// Additional field creator functions...
```

### `pkg/logger/zap_logger.go`

**Purpose**: Implement logging using Zap

**Types**:
```go
// ZapLogger implements Logger using zap
type ZapLogger struct {
    logger *zap.Logger
}
```

**Functions**:
```go
// NewZapLogger creates a new ZapLogger
func NewZapLogger(config types.LoggingConfig) (*ZapLogger, error)
    Params: config - Logging configuration
    Returns: 
        - A configured ZapLogger instance
        - Error if initialization fails

// Debug implements Logger.Debug
func (l *ZapLogger) Debug(msg string, fields ...Field)
    Params: 
        msg - Message to log
        fields - Optional structured fields

// Info implements Logger.Info
func (l *ZapLogger) Info(msg string, fields ...Field)
    Params: 
        msg - Message to log
        fields - Optional structured fields

// Error implements Logger.Error
func (l *ZapLogger) Error(msg string, fields ...Field)
    Params: 
        msg - Message to log
        fields - Optional structured fields

// WithFields implements Logger.WithFields
func (l *ZapLogger) WithFields(fields ...Field) Logger
    Params: fields - Fields to add to the logger
    Returns: A new logger with the fields added

// Sync implements Logger.Sync
func (l *ZapLogger) Sync() error
    Returns: Error if sync fails

// Additional methods implementing the Logger interface...
```

## Utilities

### `pkg/utils/exec/command.go`

**Purpose**: Safe command execution utilities

**Types**:
```go
// CommandOptions holds options for command execution
type CommandOptions struct {
    Timeout       time.Duration
    Directory     string
    Environment   []string
    StdinData     []byte
    CombinedOutput bool
}
```

**Functions**:
```go
// ExecuteCommand executes a system command with the given options
func ExecuteCommand(ctx context.Context, name string, args []string, opts CommandOptions) ([]byte, error)
    Params:
        ctx - Context for timeout/cancellation
        name - Command name to execute
        args - Command arguments
        opts - Execution options
    Returns:
        - Command output as bytes
        - Error if execution fails

// ExecuteCommandWithInput executes a command with input data
func ExecuteCommandWithInput(ctx context.Context, name string, args []string, input []byte, opts CommandOptions) ([]byte, error)
    Params:
        ctx - Context for timeout/cancellation
        name - Command name to execute
        args - Command arguments
        input - Data to send to stdin
        opts - Execution options
    Returns:
        - Command output as bytes
        - Error if execution fails
```

### `pkg/utils/xml/template.go`

**Purpose**: XML template utilities

**Types**:
```go
// TemplateLoader handles loading and rendering XML templates
type TemplateLoader struct {
    TemplateDir string
    templates   map[string]*template.Template
}
```

**Functions**:
```go
// NewTemplateLoader creates a template loader
func NewTemplateLoader(templateDir string) (*TemplateLoader, error)
    Params: templateDir - Directory containing templates
    Returns:
        - New template loader instance
        - Error if initialization fails

// LoadTemplate loads a template from the template directory
func (l *TemplateLoader) LoadTemplate(templateName string) (*template.Template, error)
    Params: templateName - Name of template to load
    Returns:
        - Loaded template
        - Error if loading fails

// RenderTemplate renders a template with the given data
func (l *TemplateLoader) RenderTemplate(templateName string, data interface{}) (string, error)
    Params:
        templateName - Name of template to render
        data - Data to use in template rendering
    Returns:
        - Rendered template as string
        - Error if rendering fails
```

### `pkg/utils/xml/parser.go`

**Purpose**: XML parsing utilities

**Functions**:
```go
// ParseXML parses XML data into a structured object
func ParseXML(data []byte, v interface{}) error
    Params:
        data - XML data to parse
        v - Target object to populate
    Returns: Error if parsing fails

// ParseXMLFile parses an XML file into a structured object
func ParseXMLFile(filePath string, v interface{}) error
    Params:
        filePath - XML file to parse
        v - Target object to populate
    Returns: Error if parsing fails

// GetElementByXPath retrieves an XML element using XPath
func GetElementByXPath(doc *etree.Document, xpath string) (*etree.Element, error)
    Params:
        doc - XML document to search
        xpath - XPath query string
    Returns:
        - Found element or nil
        - Error if search fails
```

## Libvirt Integration

### `internal/libvirt/connection/interface.go`

**Purpose**: Define libvirt connection interface

**Interfaces**:
```go
// Manager defines the interface for managing libvirt connections
type Manager interface {
    // Connect establishes a connection to libvirt
    Connect(ctx context.Context) (Connection, error)
    
    // Release returns a connection to the pool
    Release(conn Connection) error
    
    // Close closes all connections in the pool
    Close() error
}

// Connection defines interface for a libvirt connection
type Connection interface {
    // GetLibvirtConnection returns the underlying libvirt connection
    GetLibvirtConnection() *libvirt.Connect
    
    // Close closes the connection
    Close() error
    
    // IsActive checks if connection is active
    IsActive() bool
}
```

### `internal/libvirt/connection/manager.go`

**Purpose**: Implement connection management for libvirt

**Types**:
```go
// ConnectionManager implements Manager for libvirt connections
type ConnectionManager struct {
    uri            string
    connPool       chan *libvirtConnection
    maxConnections int
    timeout        time.Duration
    logger         logger.Logger
}

// libvirtConnection implements Connection interface
type libvirtConnection struct {
    conn   *libvirt.Connect
    active bool
    manager *ConnectionManager
}
```

**Functions**:
```go
// NewConnectionManager creates a new ConnectionManager
func NewConnectionManager(cfg types.LibvirtConfig, logger logger.Logger) (*ConnectionManager, error)
    Params:
        cfg - Libvirt configuration
        logger - Logger instance
    Returns:
        - New connection manager
        - Error if creation fails

// Connect implements Manager.Connect
func (m *ConnectionManager) Connect(ctx context.Context) (Connection, error)
    Params: ctx - Context for timeout/cancellation
    Returns:
        - Libvirt connection
        - Error if connection fails

// Release implements Manager.Release
func (m *ConnectionManager) Release(conn Connection) error
    Params: conn - Connection to release
    Returns: Error if release fails

// Close implements Manager.Close
func (m *ConnectionManager) Close() error
    Returns: Error if closing fails

// GetLibvirtConnection implements Connection.GetLibvirtConnection
func (c *libvirtConnection) GetLibvirtConnection() *libvirt.Connect
    Returns: Underlying libvirt.Connect object

// IsActive implements Connection.IsActive
func (c *libvirtConnection) IsActive() bool
    Returns: Whether connection is active
```

### `internal/libvirt/domain/interface.go`

**Purpose**: Define domain (VM) management interface

**Interfaces**:
```go
// Manager defines the interface for managing libvirt domains
type Manager interface {
    // Create creates a new domain (VM)
    Create(ctx context.Context, params models.VMParams) (*models.VM, error)
    
    // Get retrieves information about a domain
    Get(ctx context.Context, name string) (*models.VM, error)
    
    // List lists all domains
    List(ctx context.Context) ([]*models.VM, error)
    
    // Start starts a domain
    Start(ctx context.Context, name string) error
    
    // Stop stops a domain (graceful shutdown)
    Stop(ctx context.Context, name string) error
    
    // ForceStop forces a domain to stop
    ForceStop(ctx context.Context, name string) error
    
    // Delete deletes a domain
    Delete(ctx context.Context, name string) error
    
    // GetXML gets the XML configuration of a domain
    GetXML(ctx context.Context, name string) (string, error)
}
```

### `internal/libvirt/domain/manager.go`

**Purpose**: Implement domain management operations

**Types**:
```go
// DomainManager implements Manager for libvirt domains
type DomainManager struct {
    connManager connection.Manager
    xmlBuilder  XMLBuilder
    logger      logger.Logger
}
```

**Functions**:
```go
// NewDomainManager creates a new DomainManager
func NewDomainManager(connManager connection.Manager, xmlBuilder XMLBuilder, logger logger.Logger) *DomainManager
    Params:
        connManager - Libvirt connection manager
        xmlBuilder - XML builder for domain definitions
        logger - Logger instance
    Returns: New domain manager

// Create implements Manager.Create
func (m *DomainManager) Create(ctx context.Context, params models.VMParams) (*models.VM, error)
    Params:
        ctx - Context for timeout/cancellation
        params - VM creation parameters
    Returns:
        - Created VM info
        - Error if creation fails

// Get implements Manager.Get
func (m *DomainManager) Get(ctx context.Context, name string) (*models.VM, error)
    Params:
        ctx - Context for timeout/cancellation
        name - VM name
    Returns:
        - VM information
        - Error if retrieval fails

// List implements Manager.List
func (m *DomainManager) List(ctx context.Context) ([]*models.VM, error)
    Params: ctx - Context for timeout/cancellation
    Returns:
        - List of VMs
        - Error if listing fails

// Start implements Manager.Start
func (m *DomainManager) Start(ctx context.Context, name string) error
    Params:
        ctx - Context for timeout/cancellation
        name - VM name
    Returns: Error if starting fails

// Stop implements Manager.Stop
func (m *DomainManager) Stop(ctx context.Context, name string) error
    Params:
        ctx - Context for timeout/cancellation
        name - VM name
    Returns: Error if stopping fails

// domainToVM converts libvirt domain to VM model
func (m *DomainManager) domainToVM(domain *libvirt.Domain) (*models.VM, error)
    Params: domain - Libvirt domain object
    Returns:
        - VM model object
        - Error if conversion fails
```

### `internal/libvirt/domain/xml_builder.go`

**Purpose**: Generate XML for domain definitions

**Interfaces**:
```go
// XMLBuilder defines interface for building domain XML
type XMLBuilder interface {
    // BuildDomainXML builds XML for domain creation
    BuildDomainXML(params models.VMParams) (string, error)
}
```

**Types**:
```go
// TemplateXMLBuilder implements XMLBuilder using templates
type TemplateXMLBuilder struct {
    templateLoader *utils.TemplateLoader
}
```

**Functions**:
```go
// NewTemplateXMLBuilder creates a new TemplateXMLBuilder
func NewTemplateXMLBuilder(templateLoader *utils.TemplateLoader) *TemplateXMLBuilder
    Params: templateLoader - Template loading utility
    Returns: New XML builder for domains

// BuildDomainXML implements XMLBuilder.BuildDomainXML
func (b *TemplateXMLBuilder) BuildDomainXML(params models.VMParams) (string, error)
    Params: params - VM parameters
    Returns:
        - Domain XML string
        - Error if building fails
```

### `internal/libvirt/storage/interface.go`

**Purpose**: Define storage management interface

**Interfaces**:
```go
// PoolManager defines interface for managing storage pools
type PoolManager interface {
    // EnsureExists ensures that a storage pool exists
    EnsureExists(ctx context.Context, name string, path string) error
    
    // Delete deletes a storage pool
    Delete(ctx context.Context, name string) error
    
    // Get gets a storage pool
    Get(ctx context.Context, name string) (*libvirt.StoragePool, error)
}

// VolumeManager defines interface for managing storage volumes
type VolumeManager interface {
    // Create creates a new storage volume
    Create(ctx context.Context, poolName string, volName string, capacityBytes uint64, format string) error
    
    // CreateFromImage creates a volume from an existing image
    CreateFromImage(ctx context.Context, poolName string, volName string, imagePath string, format string) error
    
    // Delete deletes a storage volume
    Delete(ctx context.Context, poolName string, volName string) error
    
    // Resize resizes a storage volume
    Resize(ctx context.Context, poolName string, volName string, capacityBytes uint64) error
    
    // GetPath gets the path of a storage volume
    GetPath(ctx context.Context, poolName string, volName string) (string, error)
    
    // Clone clones a storage volume
    Clone(ctx context.Context, poolName string, sourceVolName string, destVolName string) error
}
```

### `internal/libvirt/storage/pool_manager.go`

**Purpose**: Implement storage pool management

**Types**:
```go
// LibvirtPoolManager implements PoolManager for libvirt
type LibvirtPoolManager struct {
    connManager connection.Manager
    xmlBuilder  XMLBuilder
    logger      logger.Logger
}
```

**Functions**:
```go
// NewLibvirtPoolManager creates a new LibvirtPoolManager
func NewLibvirtPoolManager(connManager connection.Manager, xmlBuilder XMLBuilder, logger logger.Logger) *LibvirtPoolManager
    Params:
        connManager - Libvirt connection manager
        xmlBuilder - XML builder for storage pool definitions
        logger - Logger instance
    Returns: New storage pool manager

// EnsureExists implements PoolManager.EnsureExists
func (m *LibvirtPoolManager) EnsureExists(ctx context.Context, name string, path string) error
    Params:
        ctx - Context for timeout/cancellation
        name - Pool name
        path - Pool storage path
    Returns: Error if operation fails

// Delete implements PoolManager.Delete
func (m *LibvirtPoolManager) Delete(ctx context.Context, name string) error
    Params:
        ctx - Context for timeout/cancellation
        name - Pool name
    Returns: Error if deletion fails

// Get implements PoolManager.Get
func (m *LibvirtPoolManager) Get(ctx context.Context, name string) (*libvirt.StoragePool, error)
    Params:
        ctx - Context for timeout/cancellation
        name - Pool name
    Returns:
        - Storage pool object
        - Error if retrieval fails
```

### `internal/libvirt/storage/volume_manager.go`

**Purpose**: Implement storage volume management

**Types**:
```go
// LibvirtVolumeManager implements VolumeManager for libvirt
type LibvirtVolumeManager struct {
    connManager connection.Manager
    poolManager PoolManager
    xmlBuilder  XMLBuilder
    logger      logger.Logger
}
```

**Functions**:
```go
// NewLibvirtVolumeManager creates a new LibvirtVolumeManager
func NewLibvirtVolumeManager(connManager connection.Manager, poolManager PoolManager, xmlBuilder XMLBuilder, logger logger.Logger) *LibvirtVolumeManager
    Params:
        connManager - Libvirt connection manager
        poolManager - Storage pool manager
        xmlBuilder - XML builder for volume definitions
        logger - Logger instance
    Returns: New volume manager

// Create implements VolumeManager.Create
func (m *LibvirtVolumeManager) Create(ctx context.Context, poolName string, volName string, capacityBytes uint64, format string) error
    Params:
        ctx - Context for timeout/cancellation
        poolName - Storage pool name
        volName - Volume name
        capacityBytes - Volume size in bytes
        format - Volume format (qcow2, raw, etc.)
    Returns: Error if creation fails

// CreateFromImage implements VolumeManager.CreateFromImage
func (m *LibvirtVolumeManager) CreateFromImage(ctx context.Context, poolName string, volName string, imagePath string, format string) error
    Params:
        ctx - Context for timeout/cancellation
        poolName - Storage pool name
        volName - Volume name
        imagePath - Path to source image
        format - Volume format
    Returns: Error if creation fails

// Delete implements VolumeManager.Delete
func (m *LibvirtVolumeManager) Delete(ctx context.Context, poolName string, volName string) error
    Params:
        ctx - Context for timeout/cancellation
        poolName - Storage pool name
        volName - Volume name
    Returns: Error if deletion fails

// Resize implements VolumeManager.Resize
func (m *LibvirtVolumeManager) Resize(ctx context.Context, poolName string, volName string, capacityBytes uint64) error
    Params:
        ctx - Context for timeout/cancellation
        poolName - Storage pool name
        volName - Volume name
        capacityBytes - New size in bytes
    Returns: Error if resize fails
```

### `internal/libvirt/network/interface.go`

**Purpose**: Define network management interface

**Interfaces**:
```go
// Manager defines interface for managing libvirt networks
type Manager interface {
    // EnsureExists ensures a network exists
    EnsureExists(ctx context.Context, name string, bridgeName string, cidr string, dhcp bool) error
    
    // Delete deletes a network
    Delete(ctx context.Context, name string) error
    
    // Get gets a network
    Get(ctx context.Context, name string) (*libvirt.Network, error)
    
    // GetDHCPLeases gets the DHCP leases for a network
    GetDHCPLeases(ctx context.Context, name string) ([]libvirt.NetworkDHCPLease, error)
    
    // FindIPByMAC finds the IP address of a MAC address in the network
    FindIPByMAC(ctx context.Context, networkName string, mac string) (string, error)
}
```

### `internal/libvirt/network/manager.go`

**Purpose**: Implement network management

**Types**:
```go
// LibvirtNetworkManager implements Manager for libvirt networks
type LibvirtNetworkManager struct {
    connManager connection.Manager
    xmlBuilder  XMLBuilder
    logger      logger.Logger
}
```

**Functions**:
```go
// NewLibvirtNetworkManager creates a new LibvirtNetworkManager
func NewLibvirtNetworkManager(connManager connection.Manager, xmlBuilder XMLBuilder, logger logger.Logger) *LibvirtNetworkManager
    Params:
        connManager - Libvirt connection manager
        xmlBuilder - XML builder for network definitions
        logger - Logger instance
    Returns: New network manager

// EnsureExists implements Manager.EnsureExists
func (m *LibvirtNetworkManager) EnsureExists(ctx context.Context, name string, bridgeName string, cidr string, dhcp bool) error
    Params:
        ctx - Context for timeout/cancellation
        name - Network name
        bridgeName - Bridge device name
        cidr - Network CIDR (e.g., 192.168.122.0/24)
        dhcp - Whether to enable DHCP
    Returns: Error if operation fails

// Delete implements Manager.Delete
func (m *LibvirtNetworkManager) Delete(ctx context.Context, name string) error
    Params:
        ctx - Context for timeout/cancellation
        name - Network name
    Returns: Error if deletion fails

// Get implements Manager.Get
func (m *LibvirtNetworkManager) Get(ctx context.Context, name string) (*libvirt.Network, error)
    Params:
        ctx - Context for timeout/cancellation
        name - Network name
    Returns:
        - Network object
        - Error if retrieval fails

// FindIPByMAC implements Manager.FindIPByMAC
func (m *LibvirtNetworkManager) FindIPByMAC(ctx context.Context, networkName string, mac string) (string, error)
    Params:
        ctx - Context for timeout/cancellation
        networkName - Network name
        mac - MAC address to look up
    Returns:
        - IP address if found
        - Error if lookup fails
```

## VM Management

### `internal/models/vm/vm.go`

**Purpose**: Define VM data structure

**Types**:
```go
// VM represents a virtual machine
type VM struct {
    Name        string     `json:"name"`
    UUID        string     `json:"uuid"`
    Status      VMStatus   `json:"status"`
    CPU         CPUInfo    `json:"cpu"`
    Memory      MemoryInfo `json:"memory"`
    Disks       []DiskInfo `json:"disks"`
    Networks    []NetInfo  `json:"networks"`
    CreatedAt   time.Time  `json:"createdAt"`
    Description string     `json:"description,omitempty"`
}

// VMStatus represents the status of a VM
type VMStatus string

// Status constants
const (
    VMStatusRunning  VMStatus = "running"
    VMStatusStopped  VMStatus = "stopped"
    VMStatusPaused   VMStatus = "paused"
    VMStatusShutdown VMStatus = "shutdown"
    VMStatusCrashed  VMStatus = "crashed"
    VMStatusUnknown  VMStatus = "unknown"
)
```

### `internal/models/vm/params.go`

**Purpose**: Define VM creation parameters

**Types**:
```go
// VMParams contains parameters for VM creation
type VMParams struct {
    Name        string       `json:"name" validate:"required,hostname_rfc1123"`
    Description string       `json:"description,omitempty"`
    CPU         CPUParams    `json:"cpu" validate:"required"`
    Memory      MemoryParams `json:"memory" validate:"required"`
    Disk        DiskParams   `json:"disk" validate:"required"`
    Network     NetParams    `json:"network"`
    CloudInit   CloudInitConfig `json:"cloudInit,omitempty"`
}

// CPUParams contains CPU parameters
type CPUParams struct {
    Count  int    `json:"count" validate:"required,min=1,max=128"`
    Model  string `json:"model,omitempty"`
    Socket int    `json:"socket,omitempty" validate:"omitempty,min=1"`
    Cores  int    `json:"cores,omitempty" validate:"omitempty,min=1"`
    Threads int   `json:"threads,omitempty" validate:"omitempty,min=1"`
}

// MemoryParams contains memory parameters
type MemoryParams struct {
    SizeBytes uint64 `json:"sizeBytes" validate:"required,min=134217728"` // Minimum 128MB
}

// DiskParams contains disk parameters
type DiskParams struct {
    SizeBytes    uint64 `json:"sizeBytes" validate:"required,min=1073741824"` // Minimum 1GB
    Format       string `json:"format" validate:"required,oneof=qcow2 raw"` 
    SourceImage  string `json:"sourceImage,omitempty"`
    StoragePool  string `json:"storagePool,omitempty"`
}

// NetParams contains network parameters
type NetParams struct {
    Type         string `json:"type" validate:"required,oneof=bridge network direct"`
    Source       string `json:"source" validate:"required"`
    Model        string `json:"model,omitempty" validate:"omitempty,oneof=virtio e1000 rtl8139"`
    MacAddress   string `json:"macAddress,omitempty" validate:"omitempty,mac"`
}

// CloudInitConfig contains cloud-init configuration
type CloudInitConfig struct {
    UserData     string `json:"userData,omitempty"`
    MetaData     string `json:"metaData,omitempty"`
    NetworkConfig string `json:"networkConfig,omitempty"`
}
```

### `internal/vm/interface.go`

**Purpose**: Define VM manager interface

**Interfaces**:
```go
// Manager defines the interface for VM management
type Manager interface {
    // Create creates a new VM
    Create(ctx context.Context, params models.VMParams) (*models.VM, error)
    
    // Get gets a VM by name
    Get(ctx context.Context, name string) (*models.VM, error)
    
    // List lists all VMs
    List(ctx context.Context) ([]*models.VM, error)
    
    // Delete deletes a VM
    Delete(ctx context.Context, name string) error
    
    // Start starts a VM
    Start(ctx context.Context, name string) error
    
    // Stop stops a VM
    Stop(ctx context.Context, name string) error
    
    // Restart restarts a VM
    Restart(ctx context.Context, name string) error
    
    // Export exports a VM
    Export(ctx context.Context, name string, exportParams export.Params) (*export.Job, error)
    
    // GetExportJob gets an export job
    GetExportJob(ctx context.Context, jobID string) (*export.Job, error)
}
```

### `internal/vm/manager.go`

**Purpose**: Implement VM management

**Types**:
```go
// VMManager implements Manager interface
type VMManager struct {
    domainManager  domain.Manager
    storageManager storage.VolumeManager
    networkManager network.Manager
    templateManager template.Manager
    cloudInitManager cloudinit.Manager
    exportManager  export.Manager
    config         types.Config
    logger         logger.Logger
}
```

**Functions**:
```go
// NewVMManager creates a new VMManager
func NewVMManager(
    domainManager domain.Manager,
    storageManager storage.VolumeManager,
    networkManager network.Manager,
    templateManager template.Manager,
    cloudInitManager cloudinit.Manager,
    exportManager export.Manager,
    config types.Config,
    logger logger.Logger,
) *VMManager
    Params:
        domainManager - Domain manager
        storageManager - Storage manager
        networkManager - Network manager
        templateManager - Template manager
        cloudInitManager - Cloud-init manager
        exportManager - Export manager
        config - Global configuration
        logger - Logger instance
    Returns: New VM manager

// Create implements Manager.Create
func (m *VMManager) Create(ctx context.Context, params models.VMParams) (*models.VM, error)
    Params:
        ctx - Context for timeout/cancellation
        params - VM creation parameters
    Returns:
        - Created VM information
        - Error if creation fails

// Get implements Manager.Get
func (m *VMManager) Get(ctx context.Context, name string) (*models.VM, error)
    Params:
        ctx - Context for timeout/cancellation
        name - VM name
    Returns:
        - VM information
        - Error if retrieval fails

// List implements Manager.List
func (m *VMManager) List(ctx context.Context) ([]*models.VM, error)
    Params: ctx - Context for timeout/cancellation
    Returns:
        - List of VMs
        - Error if listing fails

// Delete implements Manager.Delete
func (m *VMManager) Delete(ctx context.Context, name string) error
    Params:
        ctx - Context for timeout/cancellation
        name - VM name
    Returns: Error if deletion fails

// Start implements Manager.Start
func (m *VMManager) Start(ctx context.Context, name string) error
    Params:
        ctx - Context for timeout/cancellation
        name - VM name
    Returns: Error if starting fails

// Export implements Manager.Export
func (m *VMManager) Export(ctx context.Context, name string, exportParams export.Params) (*export.Job, error)
    Params:
        ctx - Context for timeout/cancellation
        name - VM name
        exportParams - Export parameters
    Returns:
        - Export job
        - Error if export initiation fails
```

### `internal/vm/template/interface.go`

**Purpose**: Define VM template interface

**Interfaces**:
```go
// Manager defines interface for VM templates
type Manager interface {
    // GetTemplate gets a VM template by name
    GetTemplate(name string) (*models.VMParams, error)
    
    // ListTemplates lists all available templates
    ListTemplates() ([]string, error)
    
    // ApplyTemplate applies a template to VM parameters
    ApplyTemplate(templateName string, params *models.VMParams) error
}
```

### `internal/vm/template/manager.go`

**Purpose**: Implement VM template management

**Types**:
```go
// TemplateManager implements Manager for VM templates
type TemplateManager struct {
    templates  map[string]models.VMParams
    logger     logger.Logger
}
```

**Functions**:
```go
// NewTemplateManager creates a new TemplateManager
func NewTemplateManager(templatePath string, logger logger.Logger) (*TemplateManager, error)
    Params:
        templatePath - Path to template definitions
        logger - Logger instance
    Returns:
        - New template manager
        - Error if initialization fails

// GetTemplate implements Manager.GetTemplate
func (m *TemplateManager) GetTemplate(name string) (*models.VMParams, error)
    Params: name - Template name
    Returns:
        - VM parameters based on template
        - Error if template not found

// ListTemplates implements Manager.ListTemplates
func (m *TemplateManager) ListTemplates() ([]string, error)
    Returns:
        - List of template names
        - Error if listing fails

// ApplyTemplate implements Manager.ApplyTemplate
func (m *TemplateManager) ApplyTemplate(templateName string, params *models.VMParams) error
    Params:
        templateName - Template to apply
        params - VM parameters to modify
    Returns: Error if template application fails
```

### `internal/vm/cloudinit/interface.go`

**Purpose**: Define cloud-init interface

**Interfaces**:
```go
// Manager defines interface for cloud-init
type Manager interface {
    // GenerateISO generates a cloud-init ISO
    GenerateISO(ctx context.Context, config models.CloudInitConfig, targetPath string) error
    
    // GenerateUserData generates user-data from template
    GenerateUserData(params models.VMParams) (string, error)
    
    // GenerateMetaData generates meta-data from template
    GenerateMetaData(params models.VMParams) (string, error)
    
    // GenerateNetworkConfig generates network configuration
    GenerateNetworkConfig(params models.VMParams) (string, error)
}
```

### `internal/vm/cloudinit/generator.go`

**Purpose**: Implement cloud-init data generation

**Types**:
```go
// CloudInitGenerator implements Manager for cloud-init
type CloudInitGenerator struct {
    templateLoader *utils.TemplateLoader
    logger         logger.Logger
}
```

**Functions**:
```go
// NewCloudInitGenerator creates a new CloudInitGenerator
func NewCloudInitGenerator(templateLoader *utils.TemplateLoader, logger logger.Logger) *CloudInitGenerator
    Params:
        templateLoader - Template loader for cloud-init templates
        logger - Logger instance
    Returns: New cloud-init generator

// GenerateUserData implements Manager.GenerateUserData
func (g *CloudInitGenerator) GenerateUserData(params models.VMParams) (string, error)
    Params: params - VM parameters
    Returns:
        - Generated user-data content
        - Error if generation fails

// GenerateMetaData implements Manager.GenerateMetaData
func (g *CloudInitGenerator) GenerateMetaData(params models.VMParams) (string, error)
    Params: params - VM parameters
    Returns:
        - Generated meta-data content
        - Error if generation fails

// GenerateNetworkConfig implements Manager.GenerateNetworkConfig
func (g *CloudInitGenerator) GenerateNetworkConfig(params models.VMParams) (string, error)
    Params: params - VM parameters
    Returns:
        - Generated network configuration
        - Error if generation fails
```

### `internal/vm/cloudinit/iso_builder.go`

**Purpose**: Implement cloud-init ISO creation

**Functions**:
```go
// GenerateISO implements Manager.GenerateISO
func (g *CloudInitGenerator) GenerateISO(ctx context.Context, config models.CloudInitConfig, targetPath string) error
    Params:
        ctx - Context for timeout/cancellation
        config - Cloud-init configuration
        targetPath - Path for the generated ISO
    Returns: Error if ISO generation fails

// buildISOFiles creates the files for cloud-init ISO
func (g *CloudInitGenerator) buildISOFiles(config models.CloudInitConfig, tempDir string) error
    Params:
        config - Cloud-init configuration
        tempDir - Temporary directory for files
    Returns: Error if file creation fails

// createISO creates ISO image from files
func (g *CloudInitGenerator) createISO(ctx context.Context, sourceDir string, targetPath string) error
    Params:
        ctx - Context for timeout/cancellation
        sourceDir - Source directory with cloud-init files
        targetPath - Target ISO path
    Returns: Error if ISO creation fails
```

## Authentication System

### `internal/models/user/user.go`

**Purpose**: Define user model

**Types**:
```go
// User represents a user in the system
type User struct {
    ID        string    `json:"id"`
    Username  string    `json:"username"`
    Password  string    `json:"-"` // Hashed password, not exposed in JSON
    Email     string    `json:"email"`
    Roles     []string  `json:"roles"`
    Active    bool      `json:"active"`
    CreatedAt time.Time `json:"createdAt"`
    UpdatedAt time.Time `json:"updatedAt"`
}
```

### `internal/models/user/role.go`

**Purpose**: Define user roles

**Constants**:
```go
// User roles
const (
    RoleAdmin     = "admin"
    RoleOperator  = "operator"
    RoleViewer    = "viewer"
)

// Permissions
const (
    PermCreate    = "create"
    PermRead      = "read"
    PermUpdate    = "update"
    PermDelete    = "delete"
    PermStart     = "start"
    PermStop      = "stop"
    PermExport    = "export"
)

// RolePermissions maps roles to their permissions
var RolePermissions = map[string][]string{
    RoleAdmin: {
        PermCreate, PermRead, PermUpdate, PermDelete,
        PermStart, PermStop, PermExport,
    },
    RoleOperator: {
        PermRead, PermUpdate, PermStart, PermStop, PermExport,
    },
    RoleViewer: {
        PermRead,
    },
}
```

### `internal/auth/jwt/claims.go`

**Purpose**: Define JWT claims structure

**Types**:
```go
// Claims represents custom JWT claims
type Claims struct {
    jwt.RegisteredClaims
    UserID   string   `json:"userId"`
    Username string   `json:"username"`
    Roles    []string `json:"roles"`
}
```

### `internal/auth/jwt/generator.go`

**Purpose**: Generate JWT tokens

**Interfaces**:
```go
// Generator defines interface for JWT token generation
type Generator interface {
    // Generate generates a JWT token for a user
    Generate(user *models.User) (string, error)
    
    // GenerateWithExpiration generates a JWT token with specific expiration
    GenerateWithExpiration(user *models.User, expiration time.Duration) (string, error)
    
    // Parse parses and validates a JWT token
    Parse(tokenString string) (*Claims, error)
}
```

**Types**:
```go
// JWTGenerator implements Generator
type JWTGenerator struct {
    secretKey []byte
    algorithm jwt.SigningMethod
    issuer    string
    expiresIn time.Duration
}
```

**Functions**:
```go
// NewJWTGenerator creates a new JWTGenerator
func NewJWTGenerator(config types.AuthConfig) *JWTGenerator
    Params: config - Auth configuration
    Returns: New JWT generator

// Generate implements Generator.Generate
func (g *JWTGenerator) Generate(user *models.User) (string, error)
    Params: user - User to generate token for
    Returns:
        - JWT token string
        - Error if generation fails

// GenerateWithExpiration implements Generator.GenerateWithExpiration
func (g *JWTGenerator) GenerateWithExpiration(user *models.User, expiration time.Duration) (string, error)
    Params:
        user - User to generate token for
        expiration - Token expiration duration
    Returns:
        - JWT token string
        - Error if generation fails

// Parse implements Generator.Parse
func (g *JWTGenerator) Parse(tokenString string) (*Claims, error)
    Params: tokenString - JWT token to parse
    Returns:
        - Parsed JWT claims
        - Error if parsing or validation fails
```

### `internal/auth/jwt/validator.go`

**Purpose**: Validate JWT tokens

**Interfaces**:
```go
// Validator defines interface for JWT token validation
type Validator interface {
    // Validate validates a JWT token
    Validate(tokenString string) (*Claims, error)
    
    // ValidateWithClaims validates a token and populates the claims
    ValidateWithClaims(tokenString string, claims jwt.Claims) error
}
```

**Types**:
```go
// JWTValidator implements Validator
type JWTValidator struct {
    secretKey []byte
    algorithm jwt.SigningMethod
    issuer    string
}
```

**Functions**:
```go
// NewJWTValidator creates a new JWTValidator
func NewJWTValidator(config types.AuthConfig) *JWTValidator
    Params: config - Auth configuration
    Returns: New JWT validator

// Validate implements Validator.Validate
func (v *JWTValidator) Validate(tokenString string) (*Claims, error)
    Params: tokenString - JWT token to validate
    Returns:
        - Validated JWT claims
        - Error if validation fails

// ValidateWithClaims implements Validator.ValidateWithClaims
func (v *JWTValidator) ValidateWithClaims(tokenString string, claims jwt.Claims) error
    Params:
        tokenString - JWT token to validate
        claims - Claims structure to populate
    Returns: Error if validation fails
```

### `internal/auth/user/service_interface.go`

**Purpose**: Define user service interface

**Interfaces**:
```go
// Service defines interface for user management
type Service interface {
    // Authenticate authenticates a user
    Authenticate(ctx context.Context, username, password string) (*models.User, error)
    
    // GetByID gets a user by ID
    GetByID(ctx context.Context, id string) (*models.User, error)
    
    // HasPermission checks if a user has a permission
    HasPermission(ctx context.Context, userID string, permission string) (bool, error)
}
```

### `internal/auth/user/service.go`

**Purpose**: Implement user service

**Types**:
```go
// UserService implements Service
type UserService struct {
    users  map[string]*models.User
    logger logger.Logger
}
```

**Functions**:
```go
// NewUserService creates a new UserService
func NewUserService(logger logger.Logger) *UserService
    Params: logger - Logger instance
    Returns: New user service

// Authenticate implements Service.Authenticate
func (s *UserService) Authenticate(ctx context.Context, username, password string) (*models.User, error)
    Params:
        ctx - Context for timeout/cancellation
        username - Username to authenticate
        password - Password to verify
    Returns:
        - User if authentication succeeds
        - Error if authentication fails

// GetByID implements Service.GetByID
func (s *UserService) GetByID(ctx context.Context, id string) (*models.User, error)
    Params:
        ctx - Context for timeout/cancellation
        id - User ID to retrieve
    Returns:
        - User if found
        - Error if user not found

// HasPermission implements Service.HasPermission
func (s *UserService) HasPermission(ctx context.Context, userID string, permission string) (bool, error)
    Params:
        ctx - Context for timeout/cancellation
        userID - User ID to check
        permission - Permission to check
    Returns:
        - Whether user has permission
        - Error if check fails
```

### `internal/auth/user/password.go`

**Purpose**: Password hashing and verification

**Functions**:
```go
// HashPassword creates a password hash
func HashPassword(password string) (string, error)
    Params: password - Password to hash
    Returns:
        - Password hash
        - Error if hashing fails

// VerifyPassword verifies a password against a hash
func VerifyPassword(password, hash string) bool
    Params:
        password - Password to verify
        hash - Hash to verify against
    Returns: Whether password matches hash
```

## API Layer

### `internal/api/server.go`

**Purpose**: HTTP server setup

**Types**:
```go
// Server represents the HTTP server
type Server struct {
    router       *gin.Engine
    httpServer   *http.Server
    config       types.ServerConfig
    logger       logger.Logger
}
```

**Functions**:
```go
// NewServer creates a new API server
func NewServer(config types.ServerConfig, logger logger.Logger) *Server
    Params:
        config - Server configuration
        logger - Logger instance
    Returns: New server instance

// Start starts the HTTP server
func (s *Server) Start() error
    Returns: Error if server start fails

// Stop stops the HTTP server gracefully
func (s *Server) Stop(ctx context.Context) error
    Params: ctx - Context for timeout/cancellation
    Returns: Error if server stop fails

// Router returns the Gin router
func (s *Server) Router() *gin.Engine
    Returns: Gin router instance
```

### `internal/api/router.go`

**Purpose**: API route configuration

**Functions**:
```go
// SetupRouter configures the API router
func SetupRouter(
    engine *gin.Engine,
    logger logger.Logger,
    authMiddleware *middleware.JWTMiddleware,
    vmHandler *handlers.VMHandler,
    exportHandler *handlers.ExportHandler,
    authHandler *handlers.AuthHandler,
    healthHandler *handlers.HealthHandler,
    metricsHandler *handlers.MetricsHandler,
) *gin.Engine
    Params:
        engine - Gin engine
        logger - Logger instance
        authMiddleware - JWT authentication middleware
        vmHandler - VM management handler
        exportHandler - VM export handler
        authHandler - Authentication handler
        healthHandler - Health check handler
        metricsHandler - Metrics handler
    Returns: Configured Gin engine
```

### `internal/middleware/auth/jwt_middleware.go`

**Purpose**: JWT authentication middleware

**Types**:
```go
// JWTMiddleware implements JWT authentication
type JWTMiddleware struct {
    validator jwt.Validator
    userService user.Service
    logger    logger.Logger
}
```

**Functions**:
```go
// NewJWTMiddleware creates a new JWTMiddleware
func NewJWTMiddleware(validator jwt.Validator, userService user.Service, logger logger.Logger) *JWTMiddleware
    Params:
        validator - JWT validator
        userService - User service
        logger - Logger instance
    Returns: New JWT middleware

// Authenticate middleware for authentication
func (m *JWTMiddleware) Authenticate() gin.HandlerFunc
    Returns: Gin middleware function for authentication

// Authorize middleware for authorization
func (m *JWTMiddleware) Authorize(permission string) gin.HandlerFunc
    Params: permission - Required permission
    Returns: Gin middleware function for authorization
```

### `internal/middleware/auth/role_middleware.go`

**Purpose**: Role-based access control

**Types**:
```go
// RoleMiddleware implements role-based access control
type RoleMiddleware struct {
    userService user.Service
    logger    logger.Logger
}
```

**Functions**:
```go
// NewRoleMiddleware creates a new RoleMiddleware
func NewRoleMiddleware(userService user.Service, logger logger.Logger) *RoleMiddleware
    Params:
        userService - User service
        logger - Logger instance
    Returns: New role middleware

// RequireRole middleware to require a specific role
func (m *RoleMiddleware) RequireRole(role string) gin.HandlerFunc
    Params: role - Required role
    Returns: Gin middleware function for role checking

// RequirePermission middleware to require a specific permission
func (m *RoleMiddleware) RequirePermission(permission string) gin.HandlerFunc
    Params: permission - Required permission
    Returns: Gin middleware function for permission checking
```

### `internal/api/handlers/vm_list_handler.go`

**Purpose**: Handle VM listing

**Functions**:
```go
// ListVMs handles GET /vms
func (h *VMHandler) ListVMs(c *gin.Context)
    Params: c - Gin context
    Action: Returns list of VMs
```

### `internal/api/handlers/vm_create_handler.go`

**Purpose**: Handle VM creation

**Functions**:
```go
// CreateVM handles POST /vms
func (h *VMHandler) CreateVM(c *gin.Context)
    Params: c - Gin context
    Action: Creates a new VM

// validateCreateParams validates VM creation parameters
func (h *VMHandler) validateCreateParams(params models.VMParams) error
    Params: params - VM creation parameters
    Returns: Validation error if any
```

### `internal/api/handlers/vm_get_handler.go`

**Purpose**: Handle VM details retrieval

**Functions**:
```go
// GetVM handles GET /vms/:name
func (h *VMHandler) GetVM(c *gin.Context)
    Params: c - Gin context
    Action: Returns VM details
```

### `internal/api/handlers/vm_delete_handler.go`

**Purpose**: Handle VM deletion

**Functions**:
```go
// DeleteVM handles DELETE /vms/:name
func (h *VMHandler) DeleteVM(c *gin.Context)
    Params: c - Gin context
    Action: Deletes a VM
```

### `internal/api/handlers/vm_start_handler.go`

**Purpose**: Handle VM starting

**Functions**:
```go
// StartVM handles PUT /vms/:name/start
func (h *VMHandler) StartVM(c *gin.Context)
    Params: c - Gin context
    Action: Starts a VM
```

### `internal/api/handlers/vm_stop_handler.go`

**Purpose**: Handle VM stopping

**Functions**:
```go
// StopVM handles PUT /vms/:name/stop
func (h *VMHandler) StopVM(c *gin.Context)
    Params: c - Gin context
    Action: Stops a VM
```

### `internal/api/handlers/vm_export_handler.go`

**Purpose**: Handle VM export

**Functions**:
```go
// ExportVM handles POST /vms/:name/export
func (h *ExportHandler) ExportVM(c *gin.Context)
    Params: c - Gin context
    Action: Exports a VM

// GetExportStatus handles GET /exports/:id
func (h *ExportHandler) GetExportStatus(c *gin.Context)
    Params: c - Gin context
    Action: Returns export job status

// validateExportParams validates export parameters
func (h *ExportHandler) validateExportParams(params export.Params) error
    Params: params - Export parameters
    Returns: Validation error if any
```

### `internal/api/handlers/auth_login_handler.go`

**Purpose**: Handle authentication

**Types**:
```go
// LoginRequest represents login request
type LoginRequest struct {
    Username string `json:"username" binding:"required"`
    Password string `json:"password" binding:"required"`
}

// LoginResponse represents login response
type LoginResponse struct {
    Token   string       `json:"token"`
    Expires time.Time    `json:"expires"`
    User    *models.User `json:"user"`
}
```

**Functions**:
```go
// Login handles POST /login
func (h *AuthHandler) Login(c *gin.Context)
    Params: c - Gin context
    Action: Authenticates user and returns JWT token
```

## Export Functionality

### `internal/export/interface.go`

**Purpose**: Define export interface

**Types**:
```go
// Params represents export parameters
type Params struct {
    Format   string            `json:"format" binding:"required,oneof=qcow2 vmdk vdi ova raw"`
    Options  map[string]string `json:"options,omitempty"`
    FileName string            `json:"fileName,omitempty"`
}

// Status represents export job status
type Status string

// Job status constants
const (
    StatusPending   Status = "pending"
    StatusRunning   Status = "running"
    StatusCompleted Status = "completed"
    StatusFailed    Status = "failed"
    StatusCancelled Status = "cancelled"
)

// Job represents an export job
type Job struct {
    ID         string    `json:"id"`
    VMName     string    `json:"vmName"`
    Format     string    `json:"format"`
    Status     Status    `json:"status"`
    Progress   int       `json:"progress"`
    StartTime  time.Time `json:"startTime"`
    EndTime    time.Time `json:"endTime,omitempty"`
    Error      string    `json:"error,omitempty"`
    OutputPath string    `json:"outputPath,omitempty"`
    Options    map[string]string `json:"options,omitempty"`
}
```

**Interfaces**:
```go
// Manager defines interface for export management
type Manager interface {
    // CreateExportJob creates a new export job
    CreateExportJob(ctx context.Context, vmName string, params Params) (*Job, error)
    
    // GetJob gets an export job by ID
    GetJob(ctx context.Context, jobID string) (*Job, error)
    
    // CancelJob cancels an export job
    CancelJob(ctx context.Context, jobID string) error
    
    // ListJobs lists all export jobs
    ListJobs(ctx context.Context) ([]*Job, error)
}
```

### `internal/export/manager.go`

**Purpose**: Implement export management

**Types**:
```go
// ExportManager implements Manager
type ExportManager struct {
    jobs          map[string]*Job
    formatManagers map[string]formats.Converter
    storageManager storage.VolumeManager
    domainManager  domain.Manager
    logger         logger.Logger
    config         types.ExportConfig
    mu            sync.RWMutex
}
```

**Functions**:
```go
// NewExportManager creates a new ExportManager
func NewExportManager(
    storageManager storage.VolumeManager,
    domainManager domain.Manager,
    config types.ExportConfig,
    logger logger.Logger,
) (*ExportManager, error)
    Params:
        storageManager - Storage volume manager
        domainManager - Domain manager
        config - Export configuration
        logger - Logger instance
    Returns:
        - New export manager
        - Error if initialization fails

// CreateExportJob implements Manager.CreateExportJob
func (m *ExportManager) CreateExportJob(ctx context.Context, vmName string, params Params) (*Job, error)
    Params:
        ctx - Context for timeout/cancellation
        vmName - VM name to export
        params - Export parameters
    Returns:
        - Export job
        - Error if job creation fails

// GetJob implements Manager.GetJob
func (m *ExportManager) GetJob(ctx context.Context, jobID string) (*Job, error)
    Params:
        ctx - Context for timeout/cancellation
        jobID - Job ID to retrieve
    Returns:
        - Export job
        - Error if job not found

// processExportJob processes an export job
func (m *ExportManager) processExportJob(job *Job)
    Params: job - Job to process
    Action: Performs export in background

// updateJobStatus updates job status
func (m *ExportManager) updateJobStatus(jobID string, status Status, progress int, err error)
    Params:
        jobID - Job ID to update
        status - New status
        progress - Progress percentage
        err - Error if any
    Action: Updates job status
```

### `internal/export/formats/interface.go`

**Purpose**: Define format converter interface

**Interfaces**:
```go
// Converter defines interface for format converters
type Converter interface {
    // Convert converts a VM disk to the target format
    Convert(ctx context.Context, sourcePath string, destPath string, options map[string]string) error
    
    // GetFormatName returns the format name
    GetFormatName() string
    
    // ValidateOptions validates conversion options
    ValidateOptions(options map[string]string) error
}
```

### `internal/export/formats/qcow2/converter.go`

**Purpose**: Implement QCOW2 conversion

**Types**:
```go
// QCOW2Converter implements Converter for QCOW2 format
type QCOW2Converter struct {
    logger logger.Logger
}
```

**Functions**:
```go
// NewQCOW2Converter creates a new QCOW2Converter
func NewQCOW2Converter(logger logger.Logger) *QCOW2Converter
    Params: logger - Logger instance
    Returns: New QCOW2 converter

// Convert implements Converter.Convert
func (c *QCOW2Converter) Convert(ctx context.Context, sourcePath string, destPath string, options map[string]string) error
    Params:
        ctx - Context for timeout/cancellation
        sourcePath - Source disk path
        destPath - Destination path
        options - Conversion options
    Returns: Error if conversion fails

// GetFormatName implements Converter.GetFormatName
func (c *QCOW2Converter) GetFormatName() string
    Returns: "qcow2"

// ValidateOptions implements Converter.ValidateOptions
func (c *QCOW2Converter) ValidateOptions(options map[string]string) error
    Params: options - Options to validate
    Returns: Error if options are invalid
```

### `internal/export/formats/vmdk/converter.go`

**Purpose**: Implement VMDK conversion

**Types**:
```go
// VMDKConverter implements Converter for VMDK format
type VMDKConverter struct {
    logger logger.Logger
}
```

**Functions**:
```go
// NewVMDKConverter creates a new VMDKConverter
func NewVMDKConverter(logger logger.Logger) *VMDKConverter
    Params: logger - Logger instance
    Returns: New VMDK converter

// Convert implements Converter.Convert
func (c *VMDKConverter) Convert(ctx context.Context, sourcePath string, destPath string, options map[string]string) error
    Params:
        ctx - Context for timeout/cancellation
        sourcePath - Source disk path
        destPath - Destination path
        options - Conversion options
    Returns: Error if conversion fails

// GetFormatName implements Converter.GetFormatName
func (c *VMDKConverter) GetFormatName() string
    Returns: "vmdk"

// ValidateOptions implements Converter.ValidateOptions
func (c *VMDKConverter) ValidateOptions(options map[string]string) error
    Params: options - Options to validate
    Returns: Error if options are invalid
```

### `internal/export/formats/ova/converter.go`

**Purpose**: Implement OVA conversion

**Types**:
```go
// OVAConverter implements Converter for OVA format
type OVAConverter struct {
    templateGenerator *OVFTemplateGenerator
    logger           logger.Logger
}
```

**Functions**:
```go
// NewOVAConverter creates a new OVAConverter
func NewOVAConverter(templateGenerator *OVFTemplateGenerator, logger logger.Logger) *OVAConverter
    Params:
        templateGenerator - OVF template generator
        logger - Logger instance
    Returns: New OVA converter

// Convert implements Converter.Convert
func (c *OVAConverter) Convert(ctx context.Context, sourcePath string, destPath string, options map[string]string) error
    Params:
        ctx - Context for timeout/cancellation
        sourcePath - Source disk path
        destPath - Destination path
        options - Conversion options
    Returns: Error if conversion fails

// GetFormatName implements Converter.GetFormatName
func (c *OVAConverter) GetFormatName() string
    Returns: "ova"

// convertToDisk converts source to VMDK format
func (c *OVAConverter) convertToDisk(ctx context.Context, sourcePath string, destPath string) error
    Params:
        ctx - Context for timeout/cancellation
        sourcePath - Source disk path
        destPath - Destination path
    Returns: Error if conversion fails

// packageOVA packages VMDK and OVF into OVA
func (c *OVAConverter) packageOVA(ctx context.Context, vmdkPath string, ovfPath string, ovaPath string) error
    Params:
        ctx - Context for timeout/cancellation
        vmdkPath - VMDK disk path
        ovfPath - OVF descriptor path
        ovaPath - Destination OVA path
    Returns: Error if packaging fails
```

### `internal/export/formats/ova/ovf_template.go`

**Purpose**: Generate OVF templates for OVA export

**Types**:
```go
// OVFTemplateGenerator generates OVF templates
type OVFTemplateGenerator struct {
    templateLoader *utils.TemplateLoader
    logger         logger.Logger
}
```

**Functions**:
```go
// NewOVFTemplateGenerator creates a new OVFTemplateGenerator
func NewOVFTemplateGenerator(templateLoader *utils.TemplateLoader, logger logger.Logger) *OVFTemplateGenerator
    Params:
        templateLoader - Template loader
        logger - Logger instance
    Returns: New OVF template generator

// GenerateOVF generates an OVF descriptor
func (g *OVFTemplateGenerator) GenerateOVF(vm *models.VM, diskPath string, diskSize uint64) (string, error)
    Params:
        vm - VM information
        diskPath - Disk path
        diskSize - Disk size in bytes
    Returns:
        - OVF descriptor
        - Error if generation fails

// writeOVFToFile writes OVF to a file
func (g *OVFTemplateGenerator) writeOVFToFile(ovfContent string, outPath string) error
    Params:
        ovfContent - OVF content
        outPath - Output file path
    Returns: Error if writing fails
```

## Metrics and Monitoring

### `internal/metrics/prometheus.go`

**Purpose**: Configure Prometheus metrics

**Types**:
```go
// PrometheusMetrics implements metrics collection
type PrometheusMetrics struct {
    requestDuration *prometheus.HistogramVec
    requests        *prometheus.CounterVec
    vmOperations    *prometheus.CounterVec
    vmCount         prometheus.GaugeFunc
    exportCount     prometheus.GaugeFunc
    libvirtErrors   *prometheus.CounterVec
}
```

**Functions**:
```go
// NewPrometheusMetrics creates a new PrometheusMetrics
func NewPrometheusMetrics(vmManager vm.Manager, exportManager export.Manager) *PrometheusMetrics
    Params:
        vmManager - VM manager for VM metrics
        exportManager - Export manager for export metrics
    Returns: New Prometheus metrics

// RecordRequest records an API request
func (m *PrometheusMetrics) RecordRequest(method, path string, status int, duration time.Duration)
    Params:
        method - HTTP method
        path - Request path
        status - HTTP status code
        duration - Request duration
    Action: Records request metrics

// RecordVMOperation records a VM operation
func (m *PrometheusMetrics) RecordVMOperation(operation string, vmName string, success bool)
    Params:
        operation - Operation name
        vmName - VM name
        success - Whether operation succeeded
    Action: Records VM operation metrics

// RecordLibvirtError records a libvirt error
func (m *PrometheusMetrics) RecordLibvirtError(operation string, errorType string)
    Params:
        operation - Operation that failed
        errorType - Error type
    Action: Records libvirt error metrics
```

### `internal/health/checker.go`

**Purpose**: Implement health checking

**Types**:
```go
// Status represents health status
type Status string

// Health status constants
const (
    StatusUp   Status = "UP"
    StatusDown Status = "DOWN"
)

// Check represents a health check
type Check struct {
    Name    string            `json:"name"`
    Status  Status            `json:"status"`
    Details map[string]string `json:"details,omitempty"`
}

// Result represents health check result
type Result struct {
    Status  Status   `json:"status"`
    Checks  []Check  `json:"checks"`
    Version string   `json:"version"`
}

// Checker performs health checks
type Checker struct {
    checks  []CheckFunction
    version string
}

// CheckFunction represents a health check function
type CheckFunction func() Check
```

**Functions**:
```go
// NewChecker creates a new health Checker
func NewChecker(version string) *Checker
    Params: version - Application version
    Returns: New health checker

// AddCheck adds a health check
func (c *Checker) AddCheck(check CheckFunction)
    Params: check - Check function to add
    Action: Adds check to the checker

// RunChecks runs all health checks
func (c *Checker) RunChecks() Result
    Returns: Health check results
```

### `internal/health/libvirt_checker.go`

**Purpose**: Implement libvirt-specific health checks

**Functions**:
```go
// NewLibvirtConnectionCheck creates a check for libvirt connection
func NewLibvirtConnectionCheck(connManager connection.Manager) CheckFunction
    Params: connManager - Libvirt connection manager
    Returns: Check function for libvirt connection

// NewStoragePoolCheck creates a check for storage pool
func NewStoragePoolCheck(poolManager storage.PoolManager, poolName string) CheckFunction
    Params:
        poolManager - Storage pool manager
        poolName - Pool to check
    Returns: Check function for storage pool

// NewNetworkCheck creates a check for network
func NewNetworkCheck(networkManager network.Manager, networkName string) CheckFunction
    Params:
        networkManager - Network manager
        networkName - Network to check
    Returns: Check function for network
```

## Application Entry Point

### `cmd/server/main.go`

**Purpose**: Application entry point

**Functions**:
```go
// main is the entry point of the application
func main()
    Action: Starts the application

// initConfig initializes configuration
func initConfig() (*types.Config, error)
    Returns:
        - Configuration
        - Error if loading fails

// initLogger initializes logger
func initLogger(config types.LoggingConfig) (logger.Logger, error)
    Params: config - Logging configuration
    Returns:
        - Logger instance
        - Error if initialization fails

// initLibvirt initializes libvirt connections
func initLibvirt(config types.LibvirtConfig, logger logger.Logger) (connection.Manager, error)
    Params:
        config - Libvirt configuration
        logger - Logger instance
    Returns:
        - Libvirt connection manager
        - Error if initialization fails

// setupSignalHandler sets up signal handling
func setupSignalHandler(server *api.Server, logger logger.Logger) chan os.Signal
    Params:
        server - API server
        logger - Logger instance
    Returns: Signal channel
```# Source Code Files & Function Details

## Table of Contents
- [Configuration System](#configuration-system)
- [Logging System](#logging-system)
- [Utilities](#utilities)
- [Libvirt Integration](#libvirt-integration)
- [VM Management](#vm-management)
- [Authentication System](#authentication-system)
- [API Layer](#api-layer)
- [Export Functionality](#export-functionality)
- [Metrics and Monitoring](#metrics-and-monitoring)
- [Application Entry Point](#application-entry-point)

## Configuration System

### `internal/config/types.go`

**Purpose**: Define configuration data structures

**Types**:
```go
// Config holds all application configuration
type Config struct {
    Server   ServerConfig   `yaml:"server" json:"server"`
    Libvirt  LibvirtConfig  `yaml:"libvirt" json:"libvirt"`
    Auth     AuthConfig     `yaml:"auth" json:"auth"`
    Logging  LoggingConfig  `yaml:"logging" json:"logging"`
    Storage  StorageConfig  `yaml:"storage" json:"storage"`
    Export   ExportConfig   `yaml:"export" json:"export"`
    Features FeaturesConfig `yaml:"features" json:"features"`
}

// ServerConfig holds HTTP server configuration
type ServerConfig struct {
    Host           string        `yaml:"host" json:"host"`
    Port           int           `yaml:"port" json:"port"`
    ReadTimeout    time.Duration `yaml:"readTimeout" json:"readTimeout"`
    WriteTimeout   time.Duration `yaml:"writeTimeout" json:"writeTimeout"`
    MaxHeaderBytes int           `yaml:"maxHeaderBytes" json:"maxHeaderBytes"`
    TLS            TLSConfig     `yaml:"tls" json:"tls"`
}

// LibvirtConfig holds libvirt connection settings
type LibvirtConfig struct {
    URI               string        `yaml:"uri" json:"uri"`
    ConnectionTimeout time.Duration `yaml:"connectionTimeout" json:"connectionTimeout"`
    MaxConnections    int           `yaml:"maxConnections" json:"maxConnections"`
    PoolName          string        `yaml:"poolName" json:"poolName"`
    NetworkName       string        `yaml:"networkName" json:"networkName"`
}

// More config types for other components...
```

### `internal/config/loader_interface.go`

**Purpose**: Define the configuration loading interface

**Interfaces**:
```go
// Loader is the interface for loading configuration
type Loader interface {
    // Load loads configuration from a source into the provided config struct
    Load(cfg *types.Config) error
    
    // LoadFromFile loads configuration from a specific file
    LoadFromFile(filePath string, cfg *types.Config) error
    
    // LoadWithOverrides loads configuration with environment variable overrides
    LoadWithOverrides(cfg *types.Config) error
}
```

### `internal/config/yaml_loader.go`

**Purpose**: Implement YAML configuration loader

**Types**:
```go
// YAMLLoader implements Loader for YAML files
type YAMLLoader struct {
    // Default config file path
    DefaultPath string
}
```

**Functions**:
```go
// NewYAMLLoader creates a new YAML config loader
func NewYAMLLoader(defaultPath string) *YAMLLoader
    Returns: A new YAML loader instance

// Load implements Loader.Load for YAML files
func (l *YAMLLoader) Load(cfg *types.Config) error
    Params: cfg - Pointer to Config structure to populate
    Returns: Error if loading fails

// LoadFromFile implements Loader.LoadFromFile for YAML files
func (l *YAMLLoader) LoadFromFile(filePath string, cfg *types.Config) error
    Params: 
        filePath - Path to YAML config file
        cfg - Pointer to Config structure to populate
    Returns: Error if loading fails

// LoadWithOverrides implements Loader.LoadWithOverrides
func (l *YAMLLoader) LoadWithOverrides(cfg *types.Config) error
    Params: cfg - Pointer to Config structure to populate with overrides
    Returns: Error if loading or overriding fails
```

### `internal/config/validator.go`

**Purpose**: Validate configuration values

**Functions**:
```go
// Validate checks if the configuration is valid
func Validate(cfg *types.Config) error
    Params: cfg - Configuration to validate
    Returns: Error if validation fails

// ValidateServer validates server configuration
func ValidateServer(server types.ServerConfig) error
    Params: server - Server configuration to validate
    Returns: Error if validation fails

// ValidateLibvirt validates libvirt configuration
func ValidateLibvirt(libvirt types.LibvirtConfig) error
    Params: libvirt - Libvirt configuration to validate
    Returns: Error if validation fails

// Additional validation functions for other config sections...
```

## Logging System

### `pkg/logger/interface.go`

**Purpose**: Define logging interface

**Interfaces**:
```go
// Logger defines the interface for logging
type Logger interface {
    // Debug logs a message at debug level
    Debug(msg string, fields ...Field)
    
    // Info logs a message at info level
    Info(msg string, fields ...Field)
    
    // Warn logs a message at warning level
    Warn(msg string, fields ...Field)
    
    // Error logs a message at error level
    Error(msg string, fields ...Field)
    
    // Fatal logs a message at fatal level then calls os.Exit(1)
    Fatal(msg string, fields ...Field)
    
    // WithFields returns a new Logger with the given fields added
    WithFields(fields ...Field) Logger
    
    // WithError returns a new Logger with the given error attached
    WithError(err error) Logger
    
    // Sync flushes any buffered log entries
    Sync() error
}

// Field represents a structured log field
type Field struct {
    Key   string
    Value interface{}
}
```

**Functions**:
```go
// String creates a string Field
func String(key, value string) Field
    Params: key, value - Key-value pair for the field
    Returns: A Field with string value

// Int creates an integer Field
func Int(key string, value int) Field
    Params: key, value - Key-value pair for the field
    Returns: A Field with int value

// Error creates an error Field
func Error(err error) Field
    Params: err - Error to log
    Returns: A Field with error value

// Additional field creator functions...
```

### `pkg/logger/zap_logger.go`

**Purpose**: Implement logging using Zap

**Types**:
```go
// ZapLogger implements Logger using zap
type ZapLogger struct {
    logger *zap.Logger
}
```

**Functions**:
```go
// NewZapLogger creates a new ZapLogger
func NewZapLogger(config types.LoggingConfig) (*ZapLogger, error)
    Params: config - Logging configuration
    Returns: 
        - A configured ZapLogger instance
        - Error if initialization fails

// Debug implements Logger.Debug
func (l *ZapLogger) Debug(msg string, fields ...Field)
    Params: 
        msg - Message to log
        fields - Optional structured fields

// Info implements Logger.Info
func (l *ZapLogger) Info(msg string, fields ...Field)
    Params: 
        msg - Message to log
        fields - Optional structured fields

// Error implements Logger.Error
func (l *ZapLogger) Error(msg string, fields ...Field)
    Params: 
        msg - Message to log
        fields - Optional structured fields

// WithFields implements Logger.WithFields
func (l *ZapLogger) WithFields(fields ...Field) Logger
    Params: fields - Fields to add to the logger
    Returns: A new logger with the fields added

// Sync implements Logger.Sync
func (l *ZapLogger) Sync() error
    Returns: Error if sync fails

// Additional methods implementing the Logger interface...
```

## Utilities

### `pkg/utils/exec/command.go`

**Purpose**: Safe command execution utilities

**Types**:
```go
// CommandOptions holds options for command execution
type CommandOptions struct {
    Timeout       time.Duration
    Directory     string
    Environment   []string
    StdinData     []byte
    CombinedOutput bool
}
```

**Functions**:
```go
// ExecuteCommand executes a system command with the given options
func ExecuteCommand(ctx context.Context, name string, args []string, opts CommandOptions) ([]byte, error)
    Params:
        ctx - Context for timeout/cancellation
        name - Command name to execute
        args - Command arguments
        opts - Execution options
    Returns:
        - Command output as bytes
        - Error if execution fails

// ExecuteCommandWithInput executes a command with input data
func ExecuteCommandWithInput(ctx context.Context, name string, args []string, input []byte, opts CommandOptions) ([]byte, error)
    Params:
        ctx - Context for timeout/cancellation
        name - Command name to execute
        args - Command arguments
        input - Data to send to stdin
        opts - Execution options
    Returns:
        - Command output as bytes
        - Error if execution fails
```

### `pkg/utils/xml/template.go`

**Purpose**: XML template utilities

**Types**:
```go
// TemplateLoader handles loading and rendering XML templates
type TemplateLoader struct {
    TemplateDir string
    templates   map[string]*template.Template
}
```

**Functions**:
```go
// NewTemplateLoader creates a template loader
func NewTemplateLoader(templateDir string) (*TemplateLoader, error)
    Params: templateDir - Directory containing templates
    Returns:
        - New template loader instance
        - Error if initialization fails

// LoadTemplate loads a template from the template directory
func (l *TemplateLoader) LoadTemplate(templateName string) (*template.Template, error)
    Params: templateName - Name of template to load
    Returns:
        - Loaded template
        - Error if loading fails

// RenderTemplate renders a template with the given data
func (l *TemplateLoader) RenderTemplate(templateName string, data interface{}) (string, error)
    Params:
        templateName - Name of template to render
        data - Data to use in template rendering
    Returns:
        - Rendered template as string
        - Error if rendering fails
```

### `pkg/utils/xml/parser.go`

**Purpose**: XML parsing utilities

**Functions**:
```go
// ParseXML parses XML data into a structured object
func ParseXML(data []byte, v interface{}) error
    Params:
        data - XML data to parse
        v - Target object to populate
    Returns: Error if parsing fails

// ParseXMLFile parses an XML file into a structured object
func ParseXMLFile(filePath string, v interface{}) error
    Params:
        filePath - XML file to parse
        v - Target object to populate
    Returns: Error if parsing fails

// GetElementByXPath retrieves an XML element using XPath
func GetElementByXPath(doc *etree.Document, xpath string) (*etree.Element, error)
    Params:
        doc - XML document to search
        xpath - XPath query string
    Returns:
        - Found element or nil
        - Error if search fails
```

## Libvirt Integration

### `internal/libvirt/connection/interface.go`

**Purpose**: Define libvirt connection interface

**Interfaces**:
```go
// Manager defines the interface for managing libvirt connections
type Manager interface {
    // Connect establishes a connection to libvirt
    Connect(ctx context.Context) (Connection, error)
    
    // Release returns a connection to the pool
    Release(conn Connection) error
    
    // Close closes all connections in the pool
    Close() error
}

// Connection defines interface for a libvirt connection
type Connection interface {
    // GetLibvirtConnection returns the underlying libvirt connection
    GetLibvirtConnection() *libvirt.Connect
    
    // Close closes the connection
    Close() error
    
    // IsActive checks if connection is active
    IsActive() bool
}
```

### `internal/libvirt/connection/manager.go`

**Purpose**: Implement connection management for libvirt

**Types**:
```go
// ConnectionManager implements Manager for libvirt connections
type ConnectionManager struct {
    uri            string
    connPool       chan *libvirtConnection
    maxConnections int
    timeout        time.Duration
    logger         logger.Logger
}

// libvirtConnection implements Connection interface
type libvirtConnection struct {
    conn   *libvirt.Connect
    active bool
    manager *ConnectionManager
}
```

**Functions**:
```go
// NewConnectionManager creates a new ConnectionManager
func NewConnectionManager(cfg types.LibvirtConfig, logger logger.Logger) (*ConnectionManager, error)
    Params:
        cfg - Libvirt configuration
        logger - Logger instance
    Returns:
        - New connection manager
        - Error if creation fails

// Connect implements Manager.Connect
func (m *ConnectionManager) Connect(ctx context.Context) (Connection, error)
    Params: ctx - Context for timeout/cancellation
    Returns:
        - Libvirt connection
        - Error if connection fails

// Release implements Manager.Release
func (m *ConnectionManager) Release(conn Connection) error
    Params: conn - Connection to release
    Returns: Error if release fails

// Close implements Manager.Close
func (m *ConnectionManager) Close() error
    Returns: Error if closing fails

// GetLibvirtConnection implements Connection.GetLibvirtConnection
func (c *libvirtConnection) GetLibvirtConnection() *libvirt.Connect
    Returns: Underlying libvirt.Connect object

// IsActive implements Connection.IsActive
func (c *libvirtConnection) IsActive() bool
    Returns: Whether connection is active
```

### `internal/libvirt/domain/interface.go`

**Purpose**: Define domain (VM) management interface

**Interfaces**:
```go
// Manager defines the interface for managing libvirt domains
type Manager interface {
    // Create creates a new domain (VM)
    Create(ctx context.Context, params models.VMParams) (*models.VM, error)
    
    // Get retrieves information about a domain
    Get(ctx context.Context, name string) (*models.VM, error)
    
    // List lists all domains
    List(ctx context.Context) ([]*models.VM, error)
    
    // Start starts a domain
    Start(ctx context.Context, name string) error
    
    // Stop stops a domain (graceful shutdown)
    Stop(ctx context.Context, name string) error
    
    // ForceStop forces a domain to stop
    ForceStop(ctx context.Context, name string) error
    
    // Delete deletes a domain
    Delete(ctx context.Context, name string) error
    
    // GetXML gets the XML configuration of a domain
    GetXML(ctx context.Context, name string) (string, error)
}
```

### `internal/libvirt/domain/manager.go`

**Purpose**: Implement domain management operations

**Types**:
```go
// DomainManager implements Manager for libvirt domains
type DomainManager struct {
    connManager connection.Manager
    xmlBuilder  XMLBuilder
    logger      logger.Logger
}
```

**Functions**:
```go
// NewDomainManager creates a new DomainManager
func NewDomainManager(connManager connection.Manager, xmlBuilder XMLBuilder, logger logger.Logger) *DomainManager
    Params:
        connManager - Libvirt connection manager
        xmlBuilder - XML builder for domain definitions
        logger - Logger instance
    Returns: New domain manager

// Create implements Manager.Create
func (m *DomainManager) Create(ctx context.Context, params models.VMParams) (*models.VM, error)
    Params:
        ctx - Context for timeout/cancellation
        params - VM creation parameters
    Returns:
        - Created VM info
        - Error if creation fails

// Get implements Manager.Get
func (m *DomainManager) Get(ctx context.Context, name string) (*models.VM, error)
    Params:
        ctx - Context for timeout/cancellation
        name - VM name
    Returns:
        - VM information
        - Error if retrieval fails

// List implements Manager.List
func (m *DomainManager) List(ctx context.Context) ([]*models.VM, error)
    Params: ctx - Context for timeout/cancellation
    Returns:
        - List of VMs
        - Error if listing fails

// Start implements Manager.Start
func (m *DomainManager) Start(ctx context.Context, name string) error
    Params:
        ctx - Context for timeout/cancellation
        name - VM name
    Returns: Error if starting fails

// Stop implements Manager.Stop
func (m *DomainManager) Stop(ctx context.Context, name string) error
    Params:
        ctx - Context for timeout/cancellation
        name - VM name
    Returns: Error if stopping fails

// domainToVM converts libvirt domain to VM model
func (m *DomainManager) domainToVM(domain *libvirt.Domain) (*models.VM, error)
    Params: domain - Libvirt domain object
    Returns:
        - VM model object
        - Error if conversion fails
```

### `internal/libvirt/domain/xml_builder.go`

**Purpose**: Generate XML for domain definitions

**Interfaces**:
```go
// XMLBuilder defines interface for building domain XML
type XMLBuilder interface {
    // BuildDomainXML builds XML for domain creation
    BuildDomainXML(params models.VMParams) (string, error)
}
```

**Types**:
```go
// TemplateXMLBuilder implements XMLBuilder using templates
type TemplateXMLBuilder struct {
    templateLoader *utils.TemplateLoader
}
```

**Functions**:
```go
// NewTemplateXMLBuilder creates a new TemplateXMLBuilder
func NewTemplateXMLBuilder(templateLoader *utils.TemplateLoader) *TemplateXMLBuilder
    Params: templateLoader - Template loading utility
    Returns: New XML builder for domains

// BuildDomainXML implements XMLBuilder.BuildDomainXML
func (b *TemplateXMLBuilder) BuildDomainXML(params models.VMParams) (string, error)
    Params: params - VM parameters
    Returns:
        - Domain XML string
        - Error if building fails
```

### `internal/libvirt/storage/interface.go`

**Purpose**: Define storage management interface

**Interfaces**:
```go
// PoolManager defines interface for managing storage pools
type PoolManager interface {
    // EnsureExists ensures that a storage pool exists
    EnsureExists(ctx context.Context, name string, path string) error
    
    // Delete deletes a storage pool
    Delete(ctx context.Context, name string) error
    
    // Get gets a storage pool
    Get(ctx context.Context, name string) (*libvirt.StoragePool, error)
}

// VolumeManager defines interface for managing storage volumes
type VolumeManager interface {
    // Create creates a new storage volume
    Create(ctx context.Context, poolName string, volName string, capacityBytes uint64, format string) error
    
    // CreateFromImage creates a volume from an existing image
    CreateFromImage(ctx context.Context, poolName string, volName string, imagePath string, format string) error
    
    // Delete deletes a storage volume
    Delete(ctx context.Context, poolName string, volName string) error
    
    // Resize resizes a storage volume
    Resize(ctx context.Context, poolName string, volName string, capacityBytes uint64) error
    
    // GetPath gets the path of a storage volume
    GetPath(ctx context.Context, poolName string, volName string) (string, error)
    
    // Clone clones a storage volume
    Clone(ctx context.Context, poolName string, sourceVolName string, destVolName string) error
}
```

### `internal/libvirt/storage/pool_manager.go`

**Purpose**: Implement storage pool management

**Types**:
```go
// LibvirtPoolManager implements PoolManager for libvirt
type LibvirtPoolManager struct {
    connManager connection.Manager
    xmlBuilder  XMLBuilder
    logger      logger.Logger
}
```

**Functions**:
```go
// NewLibvirtPoolManager creates a new LibvirtPoolManager
func NewLibvirtPoolManager(connManager connection.Manager, xmlBuilder XMLBuilder, logger logger.Logger) *LibvirtPoolManager
    Params:
        connManager - Libvirt connection manager
        xmlBuilder - XML builder for storage pool definitions
        logger - Logger instance
    Returns: New storage pool manager

// EnsureExists implements PoolManager.EnsureExists
func (m *LibvirtPoolManager) EnsureExists(ctx context.Context, name string, path string) error
    Params:
        ctx - Context for timeout/cancellation
        name - Pool name
        path - Pool storage path
    Returns: Error if operation fails

// Delete implements PoolManager.Delete
func (m *LibvirtPoolManager) Delete(ctx context.Context, name string) error
    Params:
        ctx - Context for timeout/cancellation
        name - Pool name
    Returns: Error if deletion fails

// Get implements PoolManager.Get
func (m *LibvirtPoolManager) Get(ctx context.Context, name string) (*libvirt.StoragePool, error)
    Params:
        ctx - Context for timeout/cancellation
        name - Pool name
    Returns:
        - Storage pool object
        - Error if retrieval fails
```

### `internal/libvirt/storage/volume_manager.go`

**Purpose**: Implement storage volume management

**Types**:
```go
// LibvirtVolumeManager implements VolumeManager for libvirt
type LibvirtVolumeManager struct {
    connManager connection.Manager
    poolManager PoolManager
    xmlBuilder  XMLBuilder
    logger      logger.Logger
}
```

**Functions**:
```go
// NewLibvirtVolumeManager creates a new LibvirtVolumeManager
func NewLibvirtVolumeManager(connManager connection.Manager, poolManager PoolManager, xmlBuilder XMLBuilder, logger logger.Logger) *LibvirtVolumeManager
    Params:
        connManager - Libvirt connection manager
        poolManager - Storage pool manager
        xmlBuilder - XML builder for volume definitions
        logger - Logger instance
    Returns: New volume manager

// Create implements VolumeManager.Create
func (m *LibvirtVolumeManager) Create(ctx context.Context, poolName string, volName string, capacityBytes uint64, format string) error
    Params:
        ctx - Context for timeout/cancellation
        poolName - Storage pool name
        volName - Volume name
        capacityBytes - Volume size in bytes
        format - Volume format (qcow2, raw, etc.)
    Returns: Error if creation fails

// CreateFromImage implements VolumeManager.CreateFromImage
func (m *LibvirtVolumeManager) CreateFromImage(ctx context.Context, poolName string, volName string, imagePath string, format string) error
    Params:
        ctx - Context for timeout/cancellation
        poolName - Storage pool name
        volName - Volume name
        imagePath - Path to source image
        format - Volume format
    Returns: Error if creation fails

// Delete implements VolumeManager.Delete
func (m *LibvirtVolumeManager) Delete(ctx context.Context, poolName string, volName string) error
    Params:
        ctx - Context for timeout/cancellation
        poolName - Storage pool name
        volName - Volume name
    Returns: Error if deletion fails

// Resize implements VolumeManager.Resize
func (m *LibvirtVolumeManager) Resize(ctx context.Context, poolName string, volName string, capacityBytes uint64) error
    Params:
        ctx - Context for timeout/cancellation
        poolName - Storage pool name
        volName - Volume name
        capacityBytes - New size in bytes
    Returns: Error if resize fails
```

### `internal/libvirt/network/interface.go`

**Purpose**: Define network management interface

**Interfaces**:
```go
// Manager defines interface for managing libvirt networks
type Manager interface {
    // EnsureExists ensures a network exists
    EnsureExists(ctx context.Context, name string, bridgeName string, cidr string, dhcp bool) error
    
    // Delete deletes a network
    Delete(ctx context.Context, name string) error
    
    // Get gets a network
    Get(ctx context.Context, name string) (*libvirt.Network, error)
    
    // GetDHCPLeases gets the DHCP leases for a network
    GetDHCPLeases(ctx context.Context, name string) ([]libvirt.NetworkDHCPLease, error)
    
    // FindIPByMAC finds the IP address of a MAC address in the network
    FindIPByMAC(ctx context.Context, networkName string, mac string) (string, error)
}
```

### `internal/libvirt/network/manager.go`

**Purpose**: Implement network management

**Types**:
```go
// LibvirtNetworkManager implements Manager for libvirt networks
type LibvirtNetworkManager struct {
    connManager connection.Manager
    xmlBuilder  XMLBuilder
    logger      logger.Logger
}
```

**Functions**:
```go
// NewLibvirtNetworkManager creates a new LibvirtNetworkManager
func NewLibvirtNetworkManager(connManager connection.Manager, xmlBuilder XMLBuilder, logger logger.Logger) *LibvirtNetworkManager
    Params:
        connManager - Libvirt connection manager
        xmlBuilder - XML builder for network definitions
        logger - Logger instance
    Returns: New network manager

// EnsureExists implements Manager.EnsureExists
func (m *LibvirtNetworkManager) EnsureExists(ctx context.Context, name string, bridgeName string, cidr string, dhcp bool) error
    Params:
        ctx - Context for timeout/cancellation
        name - Network name
        bridgeName - Bridge device name
        cidr - Network CIDR (e.g., 192.168.122.0/24)
        dhcp - Whether to enable DHCP
    Returns: Error if operation fails

// Delete implements Manager.Delete
func (m *LibvirtNetworkManager) Delete(ctx context.Context, name string) error
    Params:
        ctx - Context for timeout/cancellation
        name - Network name
    Returns: Error if deletion fails

// Get implements Manager.Get
func (m *LibvirtNetworkManager) Get(ctx context.Context, name string) (*libvirt.Network, error)
    Params:
        ctx - Context for timeout/cancellation
        name - Network name
    Returns:
        - Network object
        - Error if retrieval fails

// FindIPByMAC implements Manager.FindIPByMAC
func (m *LibvirtNetworkManager) FindIPByMAC(ctx context.Context, networkName string, mac string) (string, error)
    Params:
        ctx - Context for timeout/cancellation
        networkName - Network name
        mac - MAC address to look up
    Returns:
        - IP address if found
        - Error if lookup fails
```

## VM Management

### `internal/models/vm/vm.go`

**Purpose**: Define VM data structure

**Types**:
```go
// VM represents a virtual machine
type VM struct {
    Name        string     `json:"name"`
    UUID        string     `json:"uuid"`
    Status      VMStatus   `json:"status"`
    CPU         CPUInfo    `json:"cpu"`
    Memory      MemoryInfo `json:"memory"`
    Disks       []DiskInfo `json:"disks"`
    Networks    []NetInfo  `json:"networks"`
    CreatedAt   time.Time  `json:"createdAt"`
    Description string     `json:"description,omitempty"`
}

// VMStatus represents the status of a VM
type VMStatus string

// Status constants
const (
    VMStatusRunning  VMStatus = "running"
    VMStatusStopped  VMStatus = "stopped"
    VMStatusPaused   VMStatus = "paused"
    VMStatusShutdown VMStatus = "shutdown"
    VMStatusCrashed  VMStatus = "crashed"
    VMStatusUnknown  VMStatus = "unknown"
)
```

### `internal/models/vm/params.go`

**Purpose**: Define VM creation parameters

**Types**:
```go
// VMParams contains parameters for VM creation
type VMParams struct {
    Name        string       `json:"name" validate:"required,hostname_rfc1123"`
    Description string       `json:"description,omitempty"`
    CPU         CPUParams    `json:"cpu" validate:"required"`
    Memory      MemoryParams `json:"memory" validate:"required"`
    Disk        DiskParams   `json:"disk" validate:"required"`
    Network     NetParams    `json:"network"`
    CloudInit   CloudInitConfig `json:"cloudInit,omitempty"`
}

// CPUParams contains CPU parameters
type CPUParams struct {
    Count  int    `json:"count" validate:"required,min=1,max=128"`
    Model  string `json:"model,omitempty"`
    Socket int    `json:"socket,omitempty" validate:"omitempty,min=1"`
    Cores  int    `json:"cores,omitempty" validate:"omitempty,min=1"`
    Threads int   `json:"threads,omitempty" validate:"omitempty,min=1"`
}

// MemoryParams contains memory parameters
type MemoryParams struct {
    SizeBytes uint64 `json:"sizeBytes" validate:"required,min=134217728"` // Minimum 128MB
}

// DiskParams contains disk parameters
type DiskParams struct {
    SizeBytes    uint64 `json:"sizeBytes" validate:"required,min=1073741824"` // Minimum 1GB
    Format       string `json:"format" validate:"required,oneof=qcow2 raw"` 
    SourceImage  string `json:"sourceImage,omitempty"`
    StoragePool  string `json:"storagePool,omitempty"`
}

// NetParams contains network parameters
type NetParams struct {
    Type         string `json:"type" validate:"required,oneof=bridge network direct"`
    Source       string `json:"source" validate:"required"`
    Model        string `json:"model,omitempty" validate:"omitempty,oneof=virtio e1000 rtl8139"`
    MacAddress   string `json:"macAddress,omitempty" validate:"omitempty,mac"`
}

// CloudInitConfig contains cloud-init configuration
type CloudInitConfig struct {
    UserData     string `json:"userData,omitempty"`
    MetaData     string `json:"metaData,omitempty"`
    NetworkConfig string `json:"networkConfig,omitempty"`
}
```

### `internal/vm/interface.go`

**Purpose**: Define VM manager interface

**Interfaces**:
```go
// Manager defines the interface for VM management
type Manager interface {
    // Create creates a new VM
    Create(ctx context.Context, params models.VMParams) (*models.VM, error)
    
    // Get gets a VM by name
    Get(ctx context.Context, name string) (*models.VM, error)
    
    // List lists all VMs
    List(ctx context.Context) ([]*models.VM, error)
    
    // Delete deletes a VM
    Delete(ctx context.Context, name string) error
    
    // Start starts a VM
    Start(ctx context.Context, name string) error
    
    // Stop stops a VM
    Stop(ctx context.Context, name string) error
    
    // Restart restarts a VM
    Restart(ctx context.Context, name string) error
    
    // Export exports a VM
    Export(ctx context.Context, name string, exportParams export.Params) (*export.Job, error)
    
    // GetExportJob gets an export job
    GetExportJob(ctx context.Context, jobID string) (*export.Job, error)
}
```

### `internal/vm/manager.go`

**Purpose**: Implement VM management

**Types**:
```go
// VMManager implements Manager interface
type VMManager struct {
    domainManager  domain.Manager
    storageManager storage.VolumeManager
    networkManager network.Manager
    templateManager template.Manager
    cloudInitManager cloudinit.Manager
    exportManager  export.Manager
    config         types.Config
    logger         logger.Logger
}
```

**Functions**:
```go
// NewVMManager creates a new VMManager
func NewVMManager(
    domainManager domain.Manager,
    storageManager storage.VolumeManager,
    networkManager network.Manager,
    templateManager template.Manager,
    cloudInitManager cloudinit.Manager,
    exportManager export.Manager,
    config types.Config,
    logger logger.Logger,
) *VMManager
    Params:
        domainManager - Domain manager
        storageManager - Storage manager
        networkManager - Network manager
        templateManager - Template manager
        cloudInitManager - Cloud-init manager
        exportManager - Export manager
        config - Global configuration
        logger - Logger instance
    Returns: New VM manager

// Create implements Manager.Create
func (m *VMManager) Create(ctx context.Context, params models.VMParams) (*models.VM, error)
    Params:
        ctx - Context for timeout/cancellation
        params - VM creation parameters
    Returns:
        - Created VM information
        - Error if creation fails

// Get implements Manager.Get
func (m *VMManager) Get(ctx context.Context, name string) (*models.VM, error)
    Params:
        ctx - Context for timeout/cancellation
        name - VM name
    Returns:
        - VM information
        - Error if retrieval fails

// List implements Manager.List
func (m *VMManager) List(ctx context.Context) ([]*models.VM, error)
    Params: ctx - Context for timeout/cancellation
    Returns:
        - List of VMs
        - Error if listing fails

// Delete implements Manager.Delete
func (m *VMManager) Delete(ctx context.Context, name string) error
    Params:
        ctx - Context for timeout/cancellation
        name - VM name
    Returns: Error if deletion fails

// Start implements Manager.Start
func (m *VMManager) Start(ctx context.Context, name string) error
    Params:
        ctx - Context for timeout/cancellation
        name - VM name
    Returns: Error if starting fails

// Export implements Manager.Export
func (m *VMManager) Export(ctx context.Context, name string, exportParams export.Params) (*export.Job, error)
    Params:
        ctx - Context for timeout/cancellation
        name - VM name
        exportParams - Export parameters
    Returns:
        - Export job
        - Error if export initiation fails
```

### `internal/vm/template/interface.go`

**Purpose**: Define VM template interface

**Interfaces**:
```go
// Manager defines interface for VM templates
type Manager interface {
    // GetTemplate gets a VM template by name
    GetTemplate(name string) (*models.VMParams, error)
    
    // ListTemplates lists all available templates
    ListTemplates() ([]string, error)
    
    // ApplyTemplate applies a template to VM parameters
    ApplyTemplate(templateName string, params *models.VMParams) error
}
```

### `internal/vm/template/manager.go`

**Purpose**: Implement VM template management

**Types**:
```go
// TemplateManager implements Manager for VM templates
type TemplateManager struct {
    templates  map[string]models.VMParams
    logger     logger.Logger
}
```

**Functions**:
```go
// NewTemplateManager creates a new TemplateManager
func NewTemplateManager(templatePath string, logger logger.Logger) (*TemplateManager, error)
    Params:
        templatePath - Path to template definitions
        logger - Logger instance
    Returns:
        - New template manager
        - Error if initialization fails

// GetTemplate implements Manager.GetTemplate
func (m *TemplateManager) GetTemplate(name string) (*models.VMParams, error)
    Params: name - Template name
    Returns:
        - VM parameters based on template
        - Error if template not found

// ListTemplates implements Manager.ListTemplates
func (m *TemplateManager) ListTemplates() ([]string, error)
    Returns:
        - List of template names
        - Error if listing fails

// ApplyTemplate implements Manager.ApplyTemplate
func (m *TemplateManager) ApplyTemplate(templateName string, params *models.VMParams) error
    Params:
        templateName - Template to apply
        params - VM parameters to modify
    Returns: Error if template application fails
```

### `internal/vm/cloudinit/interface.go`

**Purpose**: Define cloud-init interface

**Interfaces**:
```go
// Manager defines interface for cloud-init
type Manager interface {
    // GenerateISO generates a cloud-init ISO
    GenerateISO(ctx context.Context, config models.CloudInitConfig, targetPath string) error
    
    // GenerateUserData generates user-data from template
    GenerateUserData(params models.VMParams) (string, error)
    
    // GenerateMetaData generates meta-data from template
    GenerateMetaData(params models.VMParams) (string, error)
    
    // GenerateNetworkConfig generates network configuration
    GenerateNetworkConfig(params models.VMParams) (string, error)
}
```

### `internal/vm/cloudinit/generator.go`

**Purpose**: Implement cloud-init data generation

**Types**:
```go
// CloudInitGenerator implements Manager for cloud-init
type CloudInitGenerator struct {
    templateLoader *utils.TemplateLoader
    logger         logger.Logger
}
```

**Functions**:
```go
// NewCloudInitGenerator creates a new CloudInitGenerator
func NewCloudInitGenerator(templateLoader *utils.TemplateLoader, logger logger.Logger) *CloudInitGenerator
    Params:
        templateLoader - Template loader for cloud-init templates
        logger - Logger instance
    Returns: New cloud-init generator

// GenerateUserData implements Manager.GenerateUserData
func (g *CloudInitGenerator) GenerateUserData(params models.VMParams) (string, error)
    Params: params - VM parameters
    Returns:
        - Generated user-data content
        - Error if generation fails

// GenerateMetaData implements Manager.GenerateMetaData
func (g *CloudInitGenerator) GenerateMetaData(params models.VMParams) (string, error)
    Params: params - VM parameters
    Returns:
        - Generated meta-data content
        - Error if generation fails

// GenerateNetworkConfig implements Manager.GenerateNetworkConfig
func (g *CloudInitGenerator) GenerateNetworkConfig(params models.VMParams) (string, error)
    Params: params - VM parameters
    Returns:
        - Generated network configuration
        - Error if generation fails
```

### `internal/vm/cloudinit/iso_builder.go`

**Purpose**: Implement cloud-init ISO creation

**Functions**:
```go
// GenerateISO implements Manager.GenerateISO
func (g *CloudInitGenerator) GenerateISO(ctx context.Context, config models.CloudInitConfig, targetPath string) error
    Params:
        ctx - Context for timeout/cancellation
        config - Cloud-init configuration
        targetPath - Path for the generated ISO
    Returns: Error if ISO generation fails

// buildISOFiles creates the files for cloud-init ISO
func (g *CloudInitGenerator) buildISOFiles(config models.CloudInitConfig, tempDir string) error
    Params:
        config - Cloud-init configuration
        tempDir - Temporary directory for files
    Returns: Error if file creation fails

// createISO creates ISO image from files
func (g *CloudInitGenerator) createISO(ctx context.Context, sourceDir string, targetPath string) error
    Params:
        ctx - Context for timeout/cancellation
        sourceDir - Source directory with cloud-init files
        targetPath - Target ISO path
    Returns: Error if ISO creation fails
```

## Authentication System

### `internal/models/user/user.go`

**Purpose**: Define user model

**Types**:
```go
// User represents a user in the system
type User struct {
    ID        string    `json:"id"`
    Username  string    `json:"username"`
    Password  string    `json:"-"` // Hashed password, not exposed in JSON
    Email     string    `json:"email"`
    Roles     []string  `json:"roles"`
    Active    bool      `json:"active"`
    CreatedAt time.Time `json:"createdAt"`
    UpdatedAt time.Time `json:"updatedAt"`
}
```

### `internal/models/user/role.go`

**Purpose**: Define user roles

**Constants**:
```go
// User roles
const (
    RoleAdmin     = "admin"
    RoleOperator  = "operator"
    RoleViewer    = "viewer"
)

// Permissions
const (
    PermCreate    = "create"
    PermRead      = "read"
    PermUpdate    = "update"
    PermDelete    = "delete"
    PermStart     = "start"
    PermStop      = "stop"
    PermExport    = "export"
)

// RolePermissions maps roles to their permissions
var RolePermissions = map[string][]string{
    RoleAdmin: {
        PermCreate, PermRead, PermUpdate, PermDelete,
        PermStart, PermStop, PermExport,
    },
    RoleOperator: {
        PermRead, PermUpdate, PermStart, PermStop, PermExport,
    },
    RoleViewer: {
        PermRead,
    },
}
```

### `internal/auth/jwt/claims.go`

**Purpose**: Define JWT claims structure

**Types**:
```go
// Claims represents custom JWT claims
type Claims struct {
    jwt.RegisteredClaims
    UserID   string   `json:"userId"`
    Username string   `json:"username"`
    Roles    []string `json:"roles"`
}
```

### `internal/auth/jwt/generator.go`

**Purpose**: Generate JWT tokens

**Interfaces**:
```go
// Generator defines interface for JWT token generation
type Generator interface {
    // Generate generates a JWT token for a user
    Generate(user *models.User) (string, error)
    
    // GenerateWithExpiration generates a JWT token with specific expiration
    GenerateWithExpiration(user *models.User, expiration time.Duration) (string, error)
    
    // Parse parses and validates a JWT token
    Parse(tokenString string) (*Claims, error)
}
```

**Types**:
```go
// JWTGenerator implements Generator
type JWTGenerator struct {
    secretKey []byte
    algorithm jwt.SigningMethod
    issuer    string
    expiresIn time.Duration
}
```

**Functions**:
```go
// NewJWTGenerator creates a new JWTGenerator
func NewJWTGenerator(config types.AuthConfig) *JWTGenerator
    Params: config - Auth configuration
    Returns: New JWT generator

// Generate implements Generator.Generate
func (g *JWTGenerator) Generate(user *models.User) (string, error)
    Params: user - User to generate token for
    Returns:
        - JWT token string
        - Error if generation fails

// GenerateWithExpiration implements Generator.GenerateWithExpiration
func (g *JWTGenerator) GenerateWithExpiration(user *models.User, expiration time.Duration) (string, error)
    Params:
        user - User to generate token for
        expiration - Token expiration duration
    Returns:
        - JWT token string
        - Error if generation fails

// Parse implements Generator.Parse
func (g *JWTGenerator) Parse(tokenString string) (*Claims, error)
    Params: tokenString - JWT token to parse
    Returns:
        - Parsed JWT claims
        - Error if parsing or validation fails
```

### `internal/auth/jwt/validator.go`

**Purpose**: Validate JWT tokens

**Interfaces**:
```go
// Validator defines interface for JWT token validation
type Validator interface {
    // Validate validates a JWT token
    Validate(tokenString string) (*Claims, error)
    
    // ValidateWithClaims validates a token and populates the claims
    ValidateWithClaims(tokenString string, claims jwt.Claims) error
}
```

**Types**:
```go
// JWTValidator implements Validator
type JWTValidator struct {
    secretKey []byte
    algorithm jwt.SigningMethod
    issuer    string
}
```

**Functions**:
```go
// NewJWTValidator creates a new JWTValidator
func NewJWTValidator(config types.AuthConfig) *JWTValidator
    Params: config - Auth configuration
    Returns: New JWT validator

// Validate implements Validator.Validate
func (v *JWTValidator) Validate(tokenString string) (*Claims, error)
    Params: tokenString - JWT token to validate
    Returns:
        - Validated JWT claims
        - Error if validation fails

// ValidateWithClaims implements Validator.ValidateWithClaims
func (v *JWTValidator) ValidateWithClaims(tokenString string, claims jwt.Claims) error
    Params:
        tokenString - JWT token to validate
        claims - Claims structure to populate
    Returns: Error if validation fails
```

### `internal/auth/user/service_interface.go`

**Purpose**: Define user service interface

**Interfaces**:
```go
// Service defines interface for user management
type Service interface {
    // Authenticate authenticates a user
    Authenticate(ctx context.Context, username, password string) (*models.User, error)
    
    // GetByID gets a user by ID
    GetByID(ctx context.Context, id string) (*models.User, error)
    
    // HasPermission checks if a user has a permission
    HasPermission(ctx context.Context, userID string, permission string) (bool, error)
}
```

### `internal/auth/user/service.go`

**Purpose**: Implement user service

**Types**:
```go
// UserService implements Service
type UserService struct {
    users  map[string]*models.User
    logger logger.Logger
}
```

**Functions**:
```go
// NewUserService creates a new UserService
func NewUserService(logger logger.Logger) *UserService
    Params: logger - Logger instance
    Returns: New user service

// Authenticate implements Service.Authenticate
func (s *UserService) Authenticate(ctx context.Context, username, password string) (*models.User, error)
    Params:
        ctx - Context for timeout/cancellation
        username - Username to authenticate
        password - Password to verify
    Returns:
        - User if authentication succeeds
        - Error if authentication fails

// GetByID implements Service.GetByID
func (s *UserService) GetByID(ctx context.Context, id string) (*models.User, error)
    Params:
        ctx - Context for timeout/cancellation
        id - User ID to retrieve
    Returns:
        - User if found
        - Error if user not found

// HasPermission implements Service.HasPermission
func (s *UserService) HasPermission(ctx context.Context, userID string, permission string) (bool, error)
    Params:
        ctx - Context for timeout/cancellation
        userID - User ID to check
        permission - Permission to check
    Returns:
        - Whether user has permission
        - Error if check fails
```

### `internal/auth/user/password.go`

**Purpose**: Password hashing and verification

**Functions**:
```go
// HashPassword creates a password hash
func HashPassword(password string) (string, error)
    Params: password - Password to hash
    Returns:
        - Password hash
        - Error if hashing fails

// VerifyPassword verifies a password against a hash
func VerifyPassword(password, hash string) bool
    Params:
        password - Password to verify
        hash - Hash to verify against
    Returns: Whether password matches hash
```

## API Layer

### `internal/api/server.go`

**Purpose**: HTTP server setup

**Types**:
```go
// Server represents the HTTP server
type Server struct {
    router       *gin.Engine
    httpServer   *http.Server
    config       types.ServerConfig
    logger       logger.Logger
}
```

**Functions**:
```go
// NewServer creates a new API server
func NewServer(config types.ServerConfig, logger logger.Logger) *Server
    Params:
        config - Server configuration
        logger - Logger instance
    Returns: New server instance

// Start starts the HTTP server
func (s *Server) Start() error
    Returns: Error if server start fails

// Stop stops the HTTP server gracefully
func (s *Server) Stop(ctx context.Context) error
    Params: ctx - Context for timeout/cancellation
    Returns: Error if server stop fails

// Router returns the Gin router
func (s *Server) Router() *gin.Engine
    Returns: Gin router instance
```

### `internal/api/router.go`

**Purpose**: API route configuration

**Functions**:
```go
// SetupRouter configures the API router
func SetupRouter(
    engine *gin.Engine,
    logger logger.Logger,
    authMiddleware *middleware.JWTMiddleware,
    vmHandler *handlers.VMHandler,
    exportHandler *handlers.ExportHandler,
    authHandler *handlers.AuthHandler,
    healthHandler *handlers.HealthHandler,
    metricsHandler *handlers.MetricsHandler,
) *gin.Engine
    Params:
        engine - Gin engine
        logger - Logger instance
        authMiddleware - JWT authentication middleware
        vmHandler - VM management handler
        exportHandler - VM export handler
        authHandler - Authentication handler
        healthHandler - Health check handler
        metricsHandler - Metrics handler
    Returns: Configured Gin engine
```

### `internal/middleware/auth/jwt_middleware.go`

**Purpose**: JWT authentication middleware

**Types**:
```go
// JWTMiddleware implements JWT authentication
type JWTMiddleware struct {
    validator jwt.Validator
    userService user.Service
    logger    logger.Logger
}
```

**Functions**:
```go
// NewJWTMiddleware creates a new JWTMiddleware
func NewJWTMiddleware(validator jwt.Validator, userService user.Service, logger logger.Logger) *JWTMiddleware
    Params:
        validator - JWT validator
        userService - User service
        logger - Logger instance
    Returns: New JWT middleware

// Authenticate middleware for authentication
func (m *JWTMiddleware) Authenticate() gin.HandlerFunc
    Returns: Gin middleware function for authentication

// Authorize middleware for authorization
func (m *JWTMiddleware) Authorize(permission string) gin.HandlerFunc
    Params: permission - Required permission
    Returns: Gin middleware function for authorization
```

### `internal/middleware/auth/role_middleware.go`

**Purpose**: Role-based access control

**Types**:
```go
// RoleMiddleware implements role-based access control
type RoleMiddleware struct {
    userService user.Service
    logger    logger.Logger
}
```

**Functions**:
```go
// NewRoleMiddleware creates a new RoleMiddleware
func NewRoleMiddleware(userService user.Service, logger logger.Logger) *RoleMiddleware
    Params:
        userService - User service
        logger - Logger instance
    Returns: New role middleware

// RequireRole middleware to require a specific role
func (m *RoleMiddleware) RequireRole(role string) gin.HandlerFunc
    Params: role - Required role
    Returns: Gin middleware function for role checking

// RequirePermission middleware to require a specific permission
func (m *RoleMiddleware) RequirePermission(permission string) gin.HandlerFunc
    Params: permission - Required permission
    Returns: Gin middleware function for permission checking
```

### `internal/api/handlers/vm_list_handler.go`

**Purpose**: Handle VM listing

**Functions**:
```go
// ListVMs handles GET /vms
func (h *VMHandler) ListVMs(c *gin.Context)
    Params: c - Gin context
    Action: Returns list of VMs
```

### `internal/api/handlers/vm_create_handler.go`

**Purpose**: Handle VM creation

**Functions**:
```go
// CreateVM handles POST /vms
func (h *VMHandler) CreateVM(c *gin.Context)
    Params: c - Gin context
    Action: Creates a new VM

// validateCreateParams validates VM creation parameters
func (h *VMHandler) validateCreateParams(params models.VMParams) error
    Params: params - VM creation parameters
    Returns: Validation error if any
```

### `internal/api/handlers/vm_get_handler.go`

**Purpose**: Handle VM details retrieval

**Functions**:
```go
// GetVM handles GET /vms/:name
func (h *VMHandler) GetVM(c *gin.Context)
    Params: c - Gin context
    Action: Returns VM details
```

### `internal/api/handlers/vm_delete_handler.go`

**Purpose**: Handle VM deletion

**Functions**:
```go
// DeleteVM handles DELETE /vms/:name
func (h *VMHandler) DeleteVM(c *gin.Context)
    Params: c - Gin context
    Action: Deletes a VM
```

### `internal/api/handlers/vm_start_handler.go`

**Purpose**: Handle VM starting

**Functions**:
```go
// StartVM handles PUT /vms/:name/start
func (h *VMHandler) StartVM(c *gin.Context)
    Params: c - Gin context
    Action: Starts a VM
```

### `internal/api/handlers/vm_stop_handler.go`

**Purpose**: Handle VM stopping

**Functions**:
```go
// StopVM handles PUT /vms/:name/stop
func (h *VMHandler) StopVM(c *gin.Context)
    Params: c - Gin context
    Action: Stops a VM
```

### `internal/api/handlers/vm_export_handler.go`

**Purpose**: Handle VM export

**Functions**:
```go
// ExportVM handles POST /vms/:name/export
func (h *ExportHandler) ExportVM(c *gin.Context)
    Params: c - Gin context
    Action: Exports a VM

// GetExportStatus handles GET /exports/:id
func (h *ExportHandler) GetExportStatus(c *gin.Context)
    Params: c - Gin context
    Action: Returns export job status

// validateExportParams validates export parameters
func (h *ExportHandler) validateExportParams(params export.Params) error
    Params: params - Export parameters
    Returns: Validation error if any
```

### `internal/api/handlers/auth_login_handler.go`

**Purpose**: Handle authentication

**Types**:
```go
// LoginRequest represents login request
type LoginRequest struct {
    Username string `json:"username" binding:"required"`
    Password string `json:"password" binding:"required"`
}

// LoginResponse represents login response
type LoginResponse struct {
    Token   string       `json:"token"`
    Expires time.Time    `json:"expires"`
    User    *models.User `json:"user"`
}
```

**Functions**:
```go
// Login handles POST /login
func (h *AuthHandler) Login(c *gin.Context)
    Params: c - Gin context
    Action: Authenticates user and returns JWT token
```

## Export Functionality

### `internal/export/interface.go`

**Purpose**: Define export interface

**Types**:
```go
// Params represents export parameters
type Params struct {
    Format   string            `json:"format" binding:"required,oneof=qcow2 vmdk vdi ova raw"`
    Options  map[string]string `json:"options,omitempty"`
    FileName string            `json:"fileName,omitempty"`
}

// Status represents export job status
type Status string

// Job status constants
const (
    StatusPending   Status = "pending"
    StatusRunning   Status = "running"
    StatusCompleted Status = "completed"
    StatusFailed    Status = "failed"
    StatusCancelled Status = "cancelled"
)

// Job represents an export job
type Job struct {
    ID         string    `json:"id"`
    VMName     string    `json:"vmName"`
    Format     string    `json:"format"`
    Status     Status    `json:"status"`
    Progress   int       `json:"progress"`
    StartTime  time.Time `json:"startTime"`
    EndTime    time.Time `json:"endTime,omitempty"`
    Error      string    `json:"error,omitempty"`
    OutputPath string    `json:"outputPath,omitempty"`
    Options    map[string]string `json:"options,omitempty"`
}
```

**Interfaces**:
```go
// Manager defines interface for export management
type Manager interface {
    // CreateExportJob creates a new export job
    CreateExportJob(ctx context.Context, vmName string, params Params) (*Job, error)
    
    // GetJob gets an export job by ID
    GetJob(ctx context.Context, jobID string) (*Job, error)
    
    // CancelJob cancels an export job
    CancelJob(ctx context.Context, jobID string) error
    
    // ListJobs lists all export jobs
    ListJobs(ctx context.Context) ([]*Job, error)
}
```

### `internal/export/manager.go`

**Purpose**: Implement export management

**Types**:
```go
// ExportManager implements Manager
type ExportManager struct {
    jobs          map[string]*Job
    formatManagers map[string]formats.Converter
    storageManager storage.VolumeManager
    domainManager  domain.Manager
    logger         logger.Logger
    config         types.ExportConfig
    mu            sync.RWMutex
}
```

**Functions**:
```go
// NewExportManager creates a new ExportManager
func NewExportManager(
    storageManager storage.VolumeManager,
    domainManager domain.Manager,
    config types.ExportConfig,
    logger logger.Logger,
) (*ExportManager, error)
    Params:
        storageManager - Storage volume manager
        domainManager - Domain manager
        config - Export configuration
        logger - Logger instance
    Returns:
        - New export manager
        - Error if initialization fails

// CreateExportJob implements Manager.CreateExportJob
func (m *ExportManager) CreateExportJob(ctx context.Context, vmName string, params Params) (*Job, error)
    Params:
        ctx - Context for timeout/cancellation
        vmName - VM name to export
        params - Export parameters
    Returns:
        - Export job
        - Error if job creation fails

// GetJob implements Manager.GetJob
func (m *ExportManager) GetJob(ctx context.Context, jobID string) (*Job, error)
    Params:
        ctx - Context for timeout/cancellation
        jobID - Job ID to retrieve
    Returns:
        - Export job
        - Error if job not found

// processExportJob processes an export job
func (m *ExportManager) processExportJob(job *Job)
    Params: job - Job to process
    Action: Performs export in background

// updateJobStatus updates job status
func (m *ExportManager) updateJobStatus(jobID string, status Status, progress int, err error)
    Params:
        jobID - Job ID to update
        status - New status
        progress - Progress percentage
        err - Error if any
    Action: Updates job status
```

### `internal/export/formats/interface.go`

**Purpose**: Define format converter interface

**Interfaces**:
```go
// Converter defines interface for format converters
type Converter interface {
    // Convert converts a VM disk to the target format
    Convert(ctx context.Context, sourcePath string, destPath string, options map[string]string) error
    
    // GetFormatName returns the format name
    GetFormatName() string
    
    // ValidateOptions validates conversion options
    ValidateOptions(options map[string]string) error
}
```

### `internal/export/formats/qcow2/converter.go`

**Purpose**: Implement QCOW2 conversion

**Types**:
```go
// QCOW2Converter implements Converter for QCOW2 format
type QCOW2Converter struct {
    logger logger.Logger
}
```

**Functions**:
```go
// NewQCOW2Converter creates a new QCOW2Converter
func NewQCOW2Converter(logger logger.Logger) *QCOW2Converter
    Params: logger - Logger instance
    Returns: New QCOW2 converter

// Convert implements Converter.Convert
func (c *QCOW2Converter) Convert(ctx context.Context, sourcePath string, destPath string, options map[string]string) error
    Params:
        ctx - Context for timeout/cancellation
        sourcePath - Source disk path
        destPath - Destination path
        options - Conversion options
    Returns: Error if conversion fails

// GetFormatName implements Converter.GetFormatName
func (c *QCOW2Converter) GetFormatName() string
    Returns: "qcow2"

// ValidateOptions implements Converter.ValidateOptions
func (c *QCOW2Converter) ValidateOptions(options map[string]string) error
    Params: options - Options to validate
    Returns: Error if options are invalid
```

### `internal/export/formats/vmdk/converter.go`

**Purpose**: Implement VMDK conversion

**Types**:
```go
// VMDKConverter implements Converter for VMDK format
type VMDKConverter struct {
    logger logger.Logger
}
```

**Functions**:
```go
// NewVMDKConverter creates a new VMDKConverter
func NewVMDKConverter(logger logger.Logger) *VMDKConverter
    Params: logger - Logger instance
    Returns: New VMDK converter

// Convert implements Converter.Convert
func (c *VMDKConverter) Convert(ctx context.Context, sourcePath string, destPath string, options map[string]string) error
    Params:
        ctx - Context for timeout/cancellation
        sourcePath - Source disk path
        destPath - Destination path
        options - Conversion options
    Returns: Error if conversion fails

// GetFormatName implements Converter.GetFormatName
func (c *VMDKConverter) GetFormatName() string
    Returns: "vmdk"

// ValidateOptions implements Converter.ValidateOptions
func (c *VMDKConverter) ValidateOptions(options map[string]string) error
    Params: options - Options to validate
    Returns: Error if options are invalid
```

### `internal/export/formats/ova/converter.go`

**Purpose**: Implement OVA conversion

**Types**:
```go
// OVAConverter implements Converter for OVA format
type OVAConverter struct {
    templateGenerator *OVFTemplateGenerator
    logger           logger.Logger
}
```

**Functions**:
```go
// NewOVAConverter creates a new OVAConverter
func NewOVAConverter(templateGenerator *OVFTemplateGenerator, logger logger.Logger) *OVAConverter
    Params:
        templateGenerator - OVF template generator
        logger - Logger instance
    Returns: New OVA converter

// Convert implements Converter.Convert
func (c *OVAConverter) Convert(ctx context.Context, sourcePath string, destPath string, options map[string]string) error
    Params:
        ctx - Context for timeout/cancellation
        sourcePath - Source disk path
        destPath - Destination path
        options - Conversion options
    Returns: Error if conversion fails

// GetFormatName implements Converter.GetFormatName
func (c *OVAConverter) GetFormatName() string
    Returns: "ova"

// convertToDisk converts source to VMDK format
func (c *OVAConverter) convertToDisk(ctx context.Context, sourcePath string, destPath string) error
    Params:
        ctx - Context for timeout/cancellation
        sourcePath - Source disk path
        destPath - Destination path
    Returns: Error if conversion fails

// packageOVA packages VMDK and OVF into OVA
func (c *OVAConverter) packageOVA(ctx context.Context, vmdkPath string, ovfPath string, ovaPath string) error
    Params:
        ctx - Context for timeout/cancellation
        vmdkPath - VMDK disk path
        ovfPath - OVF descriptor path
        ovaPath - Destination OVA path
    Returns: Error if packaging fails
```

### `internal/export/formats/ova/ovf_template.go`

**Purpose**: Generate OVF templates for OVA export

**Types**:
```go
// OVFTemplateGenerator generates OVF templates
type OVFTemplateGenerator struct {
    templateLoader *utils.TemplateLoader
    logger         logger.Logger
}
```

**Functions**:
```go
// NewOVFTemplateGenerator creates a new OVFTemplateGenerator
func NewOVFTemplateGenerator(templateLoader *utils.TemplateLoader, logger logger.Logger) *OVFTemplateGenerator
    Params:
        templateLoader - Template loader
        logger - Logger instance
    Returns: New OVF template generator

// GenerateOVF generates an OVF descriptor
func (g *OVFTemplateGenerator) GenerateOVF(vm *models.VM, diskPath string, diskSize uint64) (string, error)
    Params:
        vm - VM information
        diskPath - Disk path
        diskSize - Disk size in bytes
    Returns:
        - OVF descriptor
        - Error if generation fails

// writeOVFToFile writes OVF to a file
func (g *OVFTemplateGenerator) writeOVFToFile(ovfContent string, outPath string) error
    Params:
        ovfContent - OVF content
        outPath - Output file path
    Returns: Error if writing fails
```

## Metrics and Monitoring

### `internal/metrics/prometheus.go`

**Purpose**: Configure Prometheus metrics

**Types**:
```go
// PrometheusMetrics implements metrics collection
type PrometheusMetrics struct {
    requestDuration *prometheus.HistogramVec
    requests        *prometheus.CounterVec
    vmOperations    *prometheus.CounterVec
    vmCount         prometheus.GaugeFunc
    exportCount     prometheus.GaugeFunc
    libvirtErrors   *prometheus.CounterVec
}
```

**Functions**:
```go
// NewPrometheusMetrics creates a new PrometheusMetrics
func NewPrometheusMetrics(vmManager vm.Manager, exportManager export.Manager) *PrometheusMetrics
    Params:
        vmManager - VM manager for VM metrics
        exportManager - Export manager for export metrics
    Returns: New Prometheus metrics

// RecordRequest records an API request
func (m *PrometheusMetrics) RecordRequest(method, path string, status int, duration time.Duration)
    Params:
        method - HTTP method
        path - Request path
        status - HTTP status code
        duration - Request duration
    Action: Records request metrics

// RecordVMOperation records a VM operation
func (m *PrometheusMetrics) RecordVMOperation(operation string, vmName string, success bool)
    Params:
        operation - Operation name
        vmName - VM name
        success - Whether operation succeeded
    Action: Records VM operation metrics

// RecordLibvirtError records a libvirt error
func (m *PrometheusMetrics) RecordLibvirtError(operation string, errorType string)
    Params:
        operation - Operation that failed
        errorType - Error type
    Action: Records libvirt error metrics
```

### `internal/health/checker.go`

**Purpose**: Implement health checking

**Types**:
```go
// Status represents health status
type Status string

// Health status constants
const (
    StatusUp   Status = "UP"
    StatusDown Status = "DOWN"
)

// Check represents a health check
type Check struct {
    Name    string            `json:"name"`
    Status  Status            `json:"status"`
    Details map[string]string `json:"details,omitempty"`
}

// Result represents health check result
type Result struct {
    Status  Status   `json:"status"`
    Checks  []Check  `json:"checks"`
    Version string   `json:"version"`
}

// Checker performs health checks
type Checker struct {
    checks  []CheckFunction
    version string
}

// CheckFunction represents a health check function
type CheckFunction func() Check
```

**Functions**:
```go
// NewChecker creates a new health Checker
func NewChecker(version string) *Checker
    Params: version - Application version
    Returns: New health checker

// AddCheck adds a health check
func (c *Checker) AddCheck(check CheckFunction)
    Params: check - Check function to add
    Action: Adds check to the checker

// RunChecks runs all health checks
func (c *Checker) RunChecks() Result
    Returns: Health check results
```

### `internal/health/libvirt_checker.go`

**Purpose**: Implement libvirt-specific health checks

**Functions**:
```go
// NewLibvirtConnectionCheck creates a check for libvirt connection
func NewLibvirtConnectionCheck(connManager connection.Manager) CheckFunction
    Params: connManager - Libvirt connection manager
    Returns: Check function for libvirt connection

// NewStoragePoolCheck creates a check for storage pool
func NewStoragePoolCheck(poolManager storage.PoolManager, poolName string) CheckFunction
    Params:
        poolManager - Storage pool manager
        poolName - Pool to check
    Returns: Check function for storage pool

// NewNetworkCheck creates a check for network
func NewNetworkCheck(networkManager network.Manager, networkName string) CheckFunction
    Params:
        networkManager - Network manager
        networkName - Network to check
    Returns: Check function for network
```

## Application Entry Point

### `cmd/server/main.go`

**Purpose**: Application entry point

**Functions**:
```go
// main is the entry point of the application
func main()
    Action: Starts the application

// initConfig initializes configuration
func initConfig() (*types.Config, error)
    Returns:
        - Configuration
        - Error if loading fails

// initLogger initializes logger
func initLogger(config types.LoggingConfig) (logger.Logger, error)
    Params: config - Logging configuration
    Returns:
        - Logger instance
        - Error if initialization fails

// initLibvirt initializes libvirt connections
func initLibvirt(config types.LibvirtConfig, logger logger.Logger) (connection.Manager, error)
    Params:
        config - Libvirt configuration
        logger - Logger instance
    Returns:
        - Libvirt connection manager
        - Error if initialization fails

// setupSignalHandler sets up signal handling
func setupSignalHandler(server *api.Server, logger logger.Logger) chan os.Signal
    Params:
        server - API server
        logger - Logger instance
    Returns: Signal channel
```
