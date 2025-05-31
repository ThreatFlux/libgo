package config

import "time"

// Config holds all application configuration
type Config struct {
	Server        ServerConfig   `yaml:"server" json:"server"`
	Database      DatabaseConfig `yaml:"database" json:"database"`
	Libvirt       LibvirtConfig  `yaml:"libvirt" json:"libvirt"`
	Auth          AuthConfig     `yaml:"auth" json:"auth"`
	Logging       LoggingConfig  `yaml:"logging" json:"logging"`
	Storage       StorageConfig  `yaml:"storage" json:"storage"`
	Export        ExportConfig   `yaml:"export" json:"export"`
	Features      FeaturesConfig `yaml:"features" json:"features"`
	TemplatesPath string         `yaml:"templatesPath" json:"templatesPath"`
}

// DatabaseConfig holds database configuration
type DatabaseConfig struct {
	Driver          string        `yaml:"driver" json:"driver"`
	DSN             string        `yaml:"dsn" json:"dsn"`
	MaxOpenConns    int           `yaml:"maxOpenConns" json:"maxOpenConns"`
	MaxIdleConns    int           `yaml:"maxIdleConns" json:"maxIdleConns"`
	ConnMaxLifetime time.Duration `yaml:"connMaxLifetime" json:"connMaxLifetime"`
	AutoMigrate     bool          `yaml:"autoMigrate" json:"autoMigrate"`
}

// ServerConfig holds HTTP server configuration
type ServerConfig struct {
	Host           string        `yaml:"host" json:"host"`
	Port           int           `yaml:"port" json:"port"`
	Mode           string        `yaml:"mode" json:"mode"`
	ReadTimeout    time.Duration `yaml:"readTimeout" json:"readTimeout"`
	WriteTimeout   time.Duration `yaml:"writeTimeout" json:"writeTimeout"`
	MaxHeaderBytes int           `yaml:"maxHeaderBytes" json:"maxHeaderBytes"`
	TLS            TLSConfig     `yaml:"tls" json:"tls"`
}

// TLSConfig holds TLS configuration
type TLSConfig struct {
	Enabled      bool   `yaml:"enabled" json:"enabled"`
	CertFile     string `yaml:"certFile" json:"certFile"`
	KeyFile      string `yaml:"keyFile" json:"keyFile"`
	MinVersion   string `yaml:"minVersion" json:"minVersion"`
	MaxVersion   string `yaml:"maxVersion" json:"maxVersion"`
	CipherSuites string `yaml:"cipherSuites" json:"cipherSuites"`
}

// LibvirtConfig holds libvirt connection settings
type LibvirtConfig struct {
	URI               string        `yaml:"uri" json:"uri"`
	ConnectionTimeout time.Duration `yaml:"connectionTimeout" json:"connectionTimeout"`
	MaxConnections    int           `yaml:"maxConnections" json:"maxConnections"`
	PoolName          string        `yaml:"poolName" json:"poolName"`
	NetworkName       string        `yaml:"networkName" json:"networkName"`
}

// AuthConfig holds authentication configuration
type AuthConfig struct {
	Enabled         bool          `yaml:"enabled" json:"enabled"`
	JWTSecretKey    string        `yaml:"jwtSecretKey" json:"jwtSecretKey"`
	Issuer          string        `yaml:"issuer" json:"issuer"`
	Audience        string        `yaml:"audience" json:"audience"`
	TokenExpiration time.Duration `yaml:"tokenExpiration" json:"tokenExpiration"`
	SigningMethod   string        `yaml:"signingMethod" json:"signingMethod"`
	DefaultUsers    []DefaultUser `yaml:"defaultUsers" json:"defaultUsers"`
}

// DefaultUser represents a default user to create during system initialization
type DefaultUser struct {
	Username string   `yaml:"username" json:"username"`
	Password string   `yaml:"password" json:"password"`
	Email    string   `yaml:"email" json:"email"`
	Roles    []string `yaml:"roles" json:"roles"`
}

// LoggingConfig holds logging configuration
type LoggingConfig struct {
	Level      string `yaml:"level" json:"level"`
	Format     string `yaml:"format" json:"format"`
	FilePath   string `yaml:"filePath" json:"filePath"`
	MaxSize    int    `yaml:"maxSize" json:"maxSize"`
	MaxBackups int    `yaml:"maxBackups" json:"maxBackups"`
	MaxAge     int    `yaml:"maxAge" json:"maxAge"`
	Compress   bool   `yaml:"compress" json:"compress"`
}

// StorageConfig holds storage configuration
type StorageConfig struct {
	DefaultPool string            `yaml:"defaultPool" json:"defaultPool"`
	PoolPath    string            `yaml:"poolPath" json:"poolPath"`
	Templates   map[string]string `yaml:"templates" json:"templates"`
}

// ExportConfig holds export configuration
type ExportConfig struct {
	OutputDir     string        `yaml:"outputDir" json:"outputDir"`
	TempDir       string        `yaml:"tempDir" json:"tempDir"`
	DefaultFormat string        `yaml:"defaultFormat" json:"defaultFormat"`
	Retention     time.Duration `yaml:"retention" json:"retention"`
}

// FeaturesConfig holds feature flags
type FeaturesConfig struct {
	CloudInit      bool `yaml:"cloudInit" json:"cloudInit"`
	ExportFeature  bool `yaml:"export" json:"export"`
	Metrics        bool `yaml:"metrics" json:"metrics"`
	RBACEnabled    bool `yaml:"rbacEnabled" json:"rbacEnabled"`
	StorageCleanup bool `yaml:"storageCleanup" json:"storageCleanup"`
}
