package config

// Loader is the interface for loading configuration
type Loader interface {
	// Load loads configuration from a source into the provided config struct
	Load(cfg *Config) error

	// LoadFromFile loads configuration from a specific file
	LoadFromFile(filePath string, cfg *Config) error

	// LoadWithOverrides loads configuration with environment variable overrides
	LoadWithOverrides(cfg *Config) error
}
