package config

import "time"

// Config holds all application configuration.
type Config struct {
	Storage       StorageConfig    `yaml:"storage" json:"storage"`
	TemplatesPath string           `yaml:"templatesPath" json:"templatesPath"`
	Network       NetworkConfig    `yaml:"network" json:"network"`
	Auth          AuthConfig       `yaml:"auth" json:"auth"`
	Export        ExportConfig     `yaml:"export" json:"export"`
	Libvirt       LibvirtConfig    `yaml:"libvirt" json:"libvirt"`
	Server        ServerConfig     `yaml:"server" json:"server"`
	OVS           OVSConfig        `yaml:"ovs" json:"ovs"`
	Docker        DockerConfig     `yaml:"docker" json:"docker"`
	Logging       LoggingConfig    `yaml:"logging" json:"logging"`
	Database      DatabaseConfig   `yaml:"database" json:"database"`
	Compute       ComputeConfig    `yaml:"compute" json:"compute"`
	Monitoring    MonitoringConfig `yaml:"monitoring" json:"monitoring"`
	Features      FeaturesConfig   `yaml:"features" json:"features"`
}

// DatabaseConfig holds database configuration.
type DatabaseConfig struct {
	Driver          string        `yaml:"driver" json:"driver"`
	DSN             string        `yaml:"dsn" json:"dsn"`
	MaxOpenConns    int           `yaml:"maxOpenConns" json:"maxOpenConns"`
	MaxIdleConns    int           `yaml:"maxIdleConns" json:"maxIdleConns"`
	ConnMaxLifetime time.Duration `yaml:"connMaxLifetime" json:"connMaxLifetime"`
	AutoMigrate     bool          `yaml:"autoMigrate" json:"autoMigrate"`
}

// ServerConfig holds HTTP server configuration.
type ServerConfig struct {
	Host           string        `yaml:"host" json:"host"`
	Mode           string        `yaml:"mode" json:"mode"`
	TLS            TLSConfig     `yaml:"tls" json:"tls"`
	ReadTimeout    time.Duration `yaml:"readTimeout" json:"readTimeout"`
	WriteTimeout   time.Duration `yaml:"writeTimeout" json:"writeTimeout"`
	Port           int           `yaml:"port" json:"port"`
	MaxHeaderBytes int           `yaml:"maxHeaderBytes" json:"maxHeaderBytes"`
}

// TLSConfig holds TLS configuration.
type TLSConfig struct {
	// String fields (8 bytes on 64-bit)
	CertFile     string `yaml:"certFile" json:"certFile"`
	KeyFile      string `yaml:"keyFile" json:"keyFile"`
	MinVersion   string `yaml:"minVersion" json:"minVersion"`
	MaxVersion   string `yaml:"maxVersion" json:"maxVersion"`
	CipherSuites string `yaml:"cipherSuites" json:"cipherSuites"`
	// Bool fields (1 byte)
	Enabled bool `yaml:"enabled" json:"enabled"`
}

// LibvirtConfig holds libvirt connection settings.
type LibvirtConfig struct {
	// String fields (8 bytes on 64-bit)
	URI         string `yaml:"uri" json:"uri"`
	PoolName    string `yaml:"poolName" json:"poolName"`
	NetworkName string `yaml:"networkName" json:"networkName"`
	// Duration fields (8 bytes)
	ConnectionTimeout time.Duration `yaml:"connectionTimeout" json:"connectionTimeout"`
	// Int fields (4 bytes)
	MaxConnections int `yaml:"maxConnections" json:"maxConnections"`
}

// AuthConfig holds authentication configuration.
type AuthConfig struct {
	JWTSecretKey    string        `yaml:"jwtSecretKey" json:"jwtSecretKey"`
	Issuer          string        `yaml:"issuer" json:"issuer"`
	Audience        string        `yaml:"audience" json:"audience"`
	SigningMethod   string        `yaml:"signingMethod" json:"signingMethod"`
	DefaultUsers    []DefaultUser `yaml:"defaultUsers" json:"defaultUsers"`
	TokenExpiration time.Duration `yaml:"tokenExpiration" json:"tokenExpiration"`
	Enabled         bool          `yaml:"enabled" json:"enabled"`
}

// DefaultUser represents a default user to create during system initialization.
type DefaultUser struct {
	Username string   `yaml:"username" json:"username"`
	Password string   `yaml:"password" json:"password"`
	Email    string   `yaml:"email" json:"email"`
	Roles    []string `yaml:"roles" json:"roles"`
}

// LoggingConfig holds logging configuration.
type LoggingConfig struct {
	Level      string `yaml:"level" json:"level"`
	Format     string `yaml:"format" json:"format"`
	FilePath   string `yaml:"filePath" json:"filePath"`
	MaxSize    int    `yaml:"maxSize" json:"maxSize"`
	MaxBackups int    `yaml:"maxBackups" json:"maxBackups"`
	MaxAge     int    `yaml:"maxAge" json:"maxAge"`
	Compress   bool   `yaml:"compress" json:"compress"`
}

// StorageConfig holds storage configuration.
type StorageConfig struct {
	Templates   map[string]string `yaml:"templates" json:"templates"`
	DefaultPool string            `yaml:"defaultPool" json:"defaultPool"`
	PoolPath    string            `yaml:"poolPath" json:"poolPath"`
}

// ExportConfig holds export configuration.
type ExportConfig struct {
	OutputDir     string        `yaml:"outputDir" json:"outputDir"`
	TempDir       string        `yaml:"tempDir" json:"tempDir"`
	DefaultFormat string        `yaml:"defaultFormat" json:"defaultFormat"`
	Retention     time.Duration `yaml:"retention" json:"retention"`
}

// FeaturesConfig holds feature flags.
type FeaturesConfig struct {
	CloudInit      bool `yaml:"cloudInit" json:"cloudInit"`
	ExportFeature  bool `yaml:"export" json:"export"`
	Metrics        bool `yaml:"metrics" json:"metrics"`
	RBACEnabled    bool `yaml:"rbacEnabled" json:"rbacEnabled"`
	StorageCleanup bool `yaml:"storageCleanup" json:"storageCleanup"`
	OVSEnabled     bool `yaml:"ovsEnabled" json:"ovsEnabled"`
}

// OVSConfig holds Open vSwitch configuration.
type OVSConfig struct {
	DefaultBridges     []OVSBridgeConfig `yaml:"defaultBridges" json:"defaultBridges"`
	CommandTimeout     time.Duration     `yaml:"commandTimeout" json:"commandTimeout"`
	Enabled            bool              `yaml:"enabled" json:"enabled"`
	LibvirtIntegration bool              `yaml:"libvirtIntegration" json:"libvirtIntegration"`
}

// OVSBridgeConfig holds configuration for an OVS bridge.
type OVSBridgeConfig struct {
	Name         string            `yaml:"name" json:"name"`
	DatapathType string            `yaml:"datapathType" json:"datapathType"`
	Controller   string            `yaml:"controller" json:"controller"`
	ExternalIDs  map[string]string `yaml:"externalIds" json:"externalIds"`
	Ports        []OVSPortConfig   `yaml:"ports" json:"ports"`
	AutoCreate   bool              `yaml:"autoCreate" json:"autoCreate"`
}

// OVSPortConfig holds configuration for an OVS port.
type OVSPortConfig struct {
	ExternalIDs map[string]string `yaml:"externalIds" json:"externalIds"`
	Tag         *int              `yaml:"tag" json:"tag"`
	Name        string            `yaml:"name" json:"name"`
	Type        string            `yaml:"type" json:"type"`
	PeerPort    string            `yaml:"peerPort" json:"peerPort"`
	RemoteIP    string            `yaml:"remoteIP" json:"remoteIP"`
	TunnelType  string            `yaml:"tunnelType" json:"tunnelType"`
	Trunks      []int             `yaml:"trunks" json:"trunks"`
	AutoCreate  bool              `yaml:"autoCreate" json:"autoCreate"`
}

// DockerConfig holds Docker daemon configuration.
type DockerConfig struct {
	Host              string        `yaml:"host" json:"host"`
	APIVersion        string        `yaml:"apiVersion" json:"apiVersion"`
	TLSCertPath       string        `yaml:"tlsCertPath" json:"tlsCertPath"`
	TLSKeyPath        string        `yaml:"tlsKeyPath" json:"tlsKeyPath"`
	TLSCAPath         string        `yaml:"tlsCaPath" json:"tlsCaPath"`
	RequestTimeout    time.Duration `yaml:"requestTimeout" json:"requestTimeout"`
	ConnectionTimeout time.Duration `yaml:"connectionTimeout" json:"connectionTimeout"`
	MaxRetries        int           `yaml:"maxRetries" json:"maxRetries"`
	RetryDelay        time.Duration `yaml:"retryDelay" json:"retryDelay"`
	Enabled           bool          `yaml:"enabled" json:"enabled"`
	TLSVerify         bool          `yaml:"tlsVerify" json:"tlsVerify"`
}

// ComputeConfig holds unified compute resource configuration.
type ComputeConfig struct {
	DefaultBackend            string            `yaml:"defaultBackend" json:"defaultBackend"`
	AllowMixedDeployments     bool              `yaml:"allowMixedDeployments" json:"allowMixedDeployments"`
	ResourceLimits            ResourceLimits    `yaml:"resourceLimits" json:"resourceLimits"`
	AutoScaling               AutoScalingConfig `yaml:"autoScaling" json:"autoScaling"`
	HealthCheckInterval       time.Duration     `yaml:"healthCheckInterval" json:"healthCheckInterval"`
	MetricsCollectionInterval time.Duration     `yaml:"metricsCollectionInterval" json:"metricsCollectionInterval"`
}

// ResourceLimits defines resource limits for compute instances.
type ResourceLimits struct {
	MaxInstances          int     `yaml:"maxInstances" json:"maxInstances"`
	MaxCPUCores           int     `yaml:"maxCPUCores" json:"maxCPUCores"`
	MaxMemoryGB           int     `yaml:"maxMemoryGB" json:"maxMemoryGB"`
	MaxStorageGB          int     `yaml:"maxStorageGB" json:"maxStorageGB"`
	MaxNetworkMbps        int     `yaml:"maxNetworkMbps" json:"maxNetworkMbps"`
	CPUOvercommitRatio    float64 `yaml:"cpuOvercommitRatio" json:"cpuOvercommitRatio"`
	MemoryOvercommitRatio float64 `yaml:"memoryOvercommitRatio" json:"memoryOvercommitRatio"`
}

// AutoScalingConfig defines auto-scaling behavior.
type AutoScalingConfig struct {
	Enabled           bool          `yaml:"enabled" json:"enabled"`
	CPUThreshold      float64       `yaml:"cpuThreshold" json:"cpuThreshold"`
	MemoryThreshold   float64       `yaml:"memoryThreshold" json:"memoryThreshold"`
	ScaleUpCooldown   time.Duration `yaml:"scaleUpCooldown" json:"scaleUpCooldown"`
	ScaleDownCooldown time.Duration `yaml:"scaleDownCooldown" json:"scaleDownCooldown"`
}

// NetworkConfig holds unified network configuration.
type NetworkConfig struct {
	DockerNetworks   DockerNetworkConfig `yaml:"dockerNetworks" json:"dockerNetworks"`
	DefaultProvider  string              `yaml:"defaultProvider" json:"defaultProvider"`
	IPAMConfig       IPAMConfig          `yaml:"ipam" json:"ipam"`
	BridgeNetwork    BridgeNetworkConfig `yaml:"bridgeNetwork" json:"bridgeNetwork"`
	CrossPlatformNAT bool                `yaml:"crossPlatformNAT" json:"crossPlatformNAT"`
}

// BridgeNetworkConfig holds bridge network configuration.
type BridgeNetworkConfig struct {
	DefaultBridge string   `yaml:"defaultBridge" json:"defaultBridge"`
	IPRange       string   `yaml:"ipRange" json:"ipRange"`
	Gateway       string   `yaml:"gateway" json:"gateway"`
	DNS           []string `yaml:"dns" json:"dns"`
}

// DockerNetworkConfig holds Docker-specific network configuration.
type DockerNetworkConfig struct {
	Options       map[string]string `yaml:"options" json:"options"`
	DefaultDriver string            `yaml:"defaultDriver" json:"defaultDriver"`
	IPAMDriver    string            `yaml:"ipamDriver" json:"ipamDriver"`
}

// IPAMConfig holds IP address management configuration.
type IPAMConfig struct {
	Driver    string   `yaml:"driver" json:"driver"`
	Subnet    string   `yaml:"subnet" json:"subnet"`
	Gateway   string   `yaml:"gateway" json:"gateway"`
	IPRange   string   `yaml:"ipRange" json:"ipRange"`
	DNSServer []string `yaml:"dnsServer" json:"dnsServer"`
}

// MonitoringConfig holds monitoring and metrics configuration.
type MonitoringConfig struct {
	ResourceAlerts    ResourceAlerts `yaml:"resourceAlerts" json:"resourceAlerts"`
	MetricsInterval   time.Duration  `yaml:"metricsInterval" json:"metricsInterval"`
	PrometheusPort    int            `yaml:"prometheusPort" json:"prometheusPort"`
	Enabled           bool           `yaml:"enabled" json:"enabled"`
	PrometheusEnabled bool           `yaml:"prometheusEnabled" json:"prometheusEnabled"`
	AlertingEnabled   bool           `yaml:"alertingEnabled" json:"alertingEnabled"`
}

// ResourceAlerts defines alerting thresholds for system resources.
type ResourceAlerts struct {
	CPUThreshold     float64 `yaml:"cpuThreshold" json:"cpuThreshold"`
	MemoryThreshold  float64 `yaml:"memoryThreshold" json:"memoryThreshold"`
	DiskThreshold    float64 `yaml:"diskThreshold" json:"diskThreshold"`
	NetworkThreshold float64 `yaml:"networkThreshold" json:"networkThreshold"`
}
