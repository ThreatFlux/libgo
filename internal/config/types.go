package config

import "time"

// Config holds all application configuration.
type Config struct {
	Server        ServerConfig     `yaml:"server" json:"server"`
	Database      DatabaseConfig   `yaml:"database" json:"database"`
	Libvirt       LibvirtConfig    `yaml:"libvirt" json:"libvirt"`
	Docker        DockerConfig     `yaml:"docker" json:"docker"`
	Compute       ComputeConfig    `yaml:"compute" json:"compute"`
	Auth          AuthConfig       `yaml:"auth" json:"auth"`
	Logging       LoggingConfig    `yaml:"logging" json:"logging"`
	Storage       StorageConfig    `yaml:"storage" json:"storage"`
	Network       NetworkConfig    `yaml:"network" json:"network"`
	Export        ExportConfig     `yaml:"export" json:"export"`
	Features      FeaturesConfig   `yaml:"features" json:"features"`
	OVS           OVSConfig        `yaml:"ovs" json:"ovs"`
	Monitoring    MonitoringConfig `yaml:"monitoring" json:"monitoring"`
	TemplatesPath string           `yaml:"templatesPath" json:"templatesPath"`
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
	Port           int           `yaml:"port" json:"port"`
	Mode           string        `yaml:"mode" json:"mode"`
	ReadTimeout    time.Duration `yaml:"readTimeout" json:"readTimeout"`
	WriteTimeout   time.Duration `yaml:"writeTimeout" json:"writeTimeout"`
	MaxHeaderBytes int           `yaml:"maxHeaderBytes" json:"maxHeaderBytes"`
	TLS            TLSConfig     `yaml:"tls" json:"tls"`
}

// TLSConfig holds TLS configuration.
type TLSConfig struct {
	Enabled      bool   `yaml:"enabled" json:"enabled"`
	CertFile     string `yaml:"certFile" json:"certFile"`
	KeyFile      string `yaml:"keyFile" json:"keyFile"`
	MinVersion   string `yaml:"minVersion" json:"minVersion"`
	MaxVersion   string `yaml:"maxVersion" json:"maxVersion"`
	CipherSuites string `yaml:"cipherSuites" json:"cipherSuites"`
}

// LibvirtConfig holds libvirt connection settings.
type LibvirtConfig struct {
	URI               string        `yaml:"uri" json:"uri"`
	ConnectionTimeout time.Duration `yaml:"connectionTimeout" json:"connectionTimeout"`
	MaxConnections    int           `yaml:"maxConnections" json:"maxConnections"`
	PoolName          string        `yaml:"poolName" json:"poolName"`
	NetworkName       string        `yaml:"networkName" json:"networkName"`
}

// AuthConfig holds authentication configuration.
type AuthConfig struct {
	Enabled         bool          `yaml:"enabled" json:"enabled"`
	JWTSecretKey    string        `yaml:"jwtSecretKey" json:"jwtSecretKey"`
	Issuer          string        `yaml:"issuer" json:"issuer"`
	Audience        string        `yaml:"audience" json:"audience"`
	TokenExpiration time.Duration `yaml:"tokenExpiration" json:"tokenExpiration"`
	SigningMethod   string        `yaml:"signingMethod" json:"signingMethod"`
	DefaultUsers    []DefaultUser `yaml:"defaultUsers" json:"defaultUsers"`
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
	DefaultPool string            `yaml:"defaultPool" json:"defaultPool"`
	PoolPath    string            `yaml:"poolPath" json:"poolPath"`
	Templates   map[string]string `yaml:"templates" json:"templates"`
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
	Enabled            bool              `yaml:"enabled" json:"enabled"`
	DefaultBridges     []OVSBridgeConfig `yaml:"defaultBridges" json:"defaultBridges"`
	LibvirtIntegration bool              `yaml:"libvirtIntegration" json:"libvirtIntegration"`
	CommandTimeout     time.Duration     `yaml:"commandTimeout" json:"commandTimeout"`
}

// OVSBridgeConfig holds configuration for an OVS bridge.
type OVSBridgeConfig struct {
	Name         string            `yaml:"name" json:"name"`
	DatapathType string            `yaml:"datapathType" json:"datapathType"`
	Controller   string            `yaml:"controller" json:"controller"`
	AutoCreate   bool              `yaml:"autoCreate" json:"autoCreate"`
	ExternalIDs  map[string]string `yaml:"externalIds" json:"externalIds"`
	Ports        []OVSPortConfig   `yaml:"ports" json:"ports"`
}

// OVSPortConfig holds configuration for an OVS port.
type OVSPortConfig struct {
	Name        string            `yaml:"name" json:"name"`
	Type        string            `yaml:"type" json:"type"`
	Tag         *int              `yaml:"tag" json:"tag"`
	Trunks      []int             `yaml:"trunks" json:"trunks"`
	PeerPort    string            `yaml:"peerPort" json:"peerPort"`
	RemoteIP    string            `yaml:"remoteIP" json:"remoteIP"`
	TunnelType  string            `yaml:"tunnelType" json:"tunnelType"`
	ExternalIDs map[string]string `yaml:"externalIds" json:"externalIds"`
	AutoCreate  bool              `yaml:"autoCreate" json:"autoCreate"`
}

// DockerConfig holds Docker daemon configuration.
type DockerConfig struct {
	Enabled           bool          `yaml:"enabled" json:"enabled"`
	Host              string        `yaml:"host" json:"host"`
	APIVersion        string        `yaml:"apiVersion" json:"apiVersion"`
	TLSVerify         bool          `yaml:"tlsVerify" json:"tlsVerify"`
	TLSCertPath       string        `yaml:"tlsCertPath" json:"tlsCertPath"`
	TLSKeyPath        string        `yaml:"tlsKeyPath" json:"tlsKeyPath"`
	TLSCAPath         string        `yaml:"tlsCaPath" json:"tlsCaPath"`
	RequestTimeout    time.Duration `yaml:"requestTimeout" json:"requestTimeout"`
	ConnectionTimeout time.Duration `yaml:"connectionTimeout" json:"connectionTimeout"`
	MaxRetries        int           `yaml:"maxRetries" json:"maxRetries"`
	RetryDelay        time.Duration `yaml:"retryDelay" json:"retryDelay"`
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
	DefaultProvider  string              `yaml:"defaultProvider" json:"defaultProvider"`
	BridgeNetwork    BridgeNetworkConfig `yaml:"bridgeNetwork" json:"bridgeNetwork"`
	DockerNetworks   DockerNetworkConfig `yaml:"dockerNetworks" json:"dockerNetworks"`
	CrossPlatformNAT bool                `yaml:"crossPlatformNAT" json:"crossPlatformNAT"`
	IPAMConfig       IPAMConfig          `yaml:"ipam" json:"ipam"`
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
	DefaultDriver string            `yaml:"defaultDriver" json:"defaultDriver"`
	IPAMDriver    string            `yaml:"ipamDriver" json:"ipamDriver"`
	Options       map[string]string `yaml:"options" json:"options"`
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
	Enabled           bool           `yaml:"enabled" json:"enabled"`
	MetricsInterval   time.Duration  `yaml:"metricsInterval" json:"metricsInterval"`
	PrometheusEnabled bool           `yaml:"prometheusEnabled" json:"prometheusEnabled"`
	PrometheusPort    int            `yaml:"prometheusPort" json:"prometheusPort"`
	AlertingEnabled   bool           `yaml:"alertingEnabled" json:"alertingEnabled"`
	ResourceAlerts    ResourceAlerts `yaml:"resourceAlerts" json:"resourceAlerts"`
}

// ResourceAlerts defines alerting thresholds for system resources.
type ResourceAlerts struct {
	CPUThreshold     float64 `yaml:"cpuThreshold" json:"cpuThreshold"`
	MemoryThreshold  float64 `yaml:"memoryThreshold" json:"memoryThreshold"`
	DiskThreshold    float64 `yaml:"diskThreshold" json:"diskThreshold"`
	NetworkThreshold float64 `yaml:"networkThreshold" json:"networkThreshold"`
}
