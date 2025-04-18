package config

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestYAMLLoader_LoadFromFile(t *testing.T) {
	// Create a temporary directory for the test
	tempDir, err := os.MkdirTemp("", "libgo-test-")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create a test config file
	configContent := `server:
  host: localhost
  port: 8080
  readTimeout: 30s
  writeTimeout: 30s
  maxHeaderBytes: 1048576
  tls:
    enabled: false

libvirt:
  uri: qemu:///system
  connectionTimeout: 30s
  maxConnections: 5
  poolName: default
  networkName: default

auth:
  enabled: true
  jwtSecretKey: my-secret-key
  issuer: libgo-server
  audience: libgo-clients
  tokenExpiration: 15m
  signingMethod: HS256

logging:
  level: info
  format: json
  filePath: ""
  maxSize: 10
  maxBackups: 5
  maxAge: 30
  compress: true

storage:
  defaultPool: default
  poolPath: /var/lib/libvirt/images
  templates:
    ubuntu: /var/lib/libvirt/images/ubuntu.qcow2

export:
  outputDir: /var/lib/libvirt/exports
  tempDir: /tmp/libgo-exports
  defaultFormat: qcow2
  retention: 168h

features:
  cloudInit: true
  export: true
  metrics: true
  rbacEnabled: true
  storageCleanup: true
`

	configPath := filepath.Join(tempDir, "config.yaml")
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to write config file: %v", err)
	}

	// Load the config
	loader := NewYAMLLoader(configPath)
	cfg := &Config{}

	if err := loader.LoadFromFile(configPath, cfg); err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Verify the loaded config
	if cfg.Server.Host != "localhost" {
		t.Errorf("Expected server.host to be 'localhost', got %s", cfg.Server.Host)
	}
	if cfg.Server.Port != 8080 {
		t.Errorf("Expected server.port to be 8080, got %d", cfg.Server.Port)
	}
	if cfg.Server.ReadTimeout != 30*time.Second {
		t.Errorf("Expected server.readTimeout to be 30s, got %v", cfg.Server.ReadTimeout)
	}

	if cfg.Libvirt.URI != "qemu:///system" {
		t.Errorf("Expected libvirt.uri to be 'qemu:///system', got %s", cfg.Libvirt.URI)
	}
	if cfg.Libvirt.MaxConnections != 5 {
		t.Errorf("Expected libvirt.maxConnections to be 5, got %d", cfg.Libvirt.MaxConnections)
	}

	if !cfg.Auth.Enabled {
		t.Errorf("Expected auth.enabled to be true")
	}
	if cfg.Auth.TokenExpiration != 15*time.Minute {
		t.Errorf("Expected auth.tokenExpiration to be 15m, got %v", cfg.Auth.TokenExpiration)
	}

	if cfg.Logging.Level != "info" {
		t.Errorf("Expected logging.level to be 'info', got %s", cfg.Logging.Level)
	}

	if cfg.Storage.DefaultPool != "default" {
		t.Errorf("Expected storage.defaultPool to be 'default', got %s", cfg.Storage.DefaultPool)
	}
	if cfg.Storage.Templates["ubuntu"] != "/var/lib/libvirt/images/ubuntu.qcow2" {
		t.Errorf("Expected storage.templates.ubuntu to be '/var/lib/libvirt/images/ubuntu.qcow2', got %s", cfg.Storage.Templates["ubuntu"])
	}

	if cfg.Export.DefaultFormat != "qcow2" {
		t.Errorf("Expected export.defaultFormat to be 'qcow2', got %s", cfg.Export.DefaultFormat)
	}
	if cfg.Export.Retention != 168*time.Hour {
		t.Errorf("Expected export.retention to be 168h, got %v", cfg.Export.Retention)
	}

	if !cfg.Features.CloudInit {
		t.Errorf("Expected features.cloudInit to be true")
	}
}

func TestYAMLLoader_LoadFromFile_Error(t *testing.T) {
	// Test loading a non-existent file
	loader := NewYAMLLoader("non-existent-file.yaml")
	cfg := &Config{}

	if err := loader.LoadFromFile("non-existent-file.yaml", cfg); err == nil {
		t.Errorf("Expected an error when loading a non-existent file, got nil")
	}

	// Test loading an invalid YAML file
	tempDir, err := os.MkdirTemp("", "libgo-test-")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	invalidYAMLPath := filepath.Join(tempDir, "invalid.yaml")
	if err := os.WriteFile(invalidYAMLPath, []byte("invalid: yaml: content:"), 0644); err != nil {
		t.Fatalf("Failed to write invalid YAML file: %v", err)
	}

	if err := loader.LoadFromFile(invalidYAMLPath, cfg); err == nil {
		t.Errorf("Expected an error when loading invalid YAML, got nil")
	}
}

func TestYAMLLoader_LoadWithOverrides(t *testing.T) {
	// Set environment variables for testing
	os.Setenv("SERVER_PORT", "9090")
	os.Setenv("LIBVIRT_URI", "qemu+tcp://192.168.1.100/system")
	os.Setenv("AUTH_ENABLED", "false")
	os.Setenv("LOGGING_LEVEL", "debug")
	os.Setenv("STORAGE_TEMPLATES", "debian:path/to/debian.qcow2,centos:path/to/centos.qcow2")
	defer func() {
		os.Unsetenv("SERVER_PORT")
		os.Unsetenv("LIBVIRT_URI")
		os.Unsetenv("AUTH_ENABLED")
		os.Unsetenv("LOGGING_LEVEL")
		os.Unsetenv("STORAGE_TEMPLATES")
	}()

	// Create a config with default values
	cfg := &Config{
		Server: ServerConfig{
			Host: "localhost",
			Port: 8080,
		},
		Libvirt: LibvirtConfig{
			URI: "qemu:///system",
		},
		Auth: AuthConfig{
			Enabled: true,
		},
		Logging: LoggingConfig{
			Level: "info",
		},
		Storage: StorageConfig{
			Templates: map[string]string{
				"ubuntu": "path/to/ubuntu.qcow2",
			},
		},
	}

	// Apply environment overrides
	loader := NewYAMLLoader("")
	if err := loader.LoadWithOverrides(cfg); err != nil {
		t.Fatalf("Failed to apply environment overrides: %v", err)
	}

	// Verify the overridden values
	if cfg.Server.Port != 9090 {
		t.Errorf("Expected server.port to be 9090, got %d", cfg.Server.Port)
	}
	if cfg.Libvirt.URI != "qemu+tcp://192.168.1.100/system" {
		t.Errorf("Expected libvirt.uri to be 'qemu+tcp://192.168.1.100/system', got %s", cfg.Libvirt.URI)
	}
	if cfg.Auth.Enabled {
		t.Errorf("Expected auth.enabled to be false")
	}
	if cfg.Logging.Level != "debug" {
		t.Errorf("Expected logging.level to be 'debug', got %s", cfg.Logging.Level)
	}

	// Original template should still be there plus the new ones
	if len(cfg.Storage.Templates) != 3 {
		t.Errorf("Expected 3 templates, got %d", len(cfg.Storage.Templates))
	}

	if cfg.Storage.Templates["debian"] != "path/to/debian.qcow2" {
		t.Errorf("Expected storage.templates.debian to be 'path/to/debian.qcow2', got %s", cfg.Storage.Templates["debian"])
	}
	if cfg.Storage.Templates["centos"] != "path/to/centos.qcow2" {
		t.Errorf("Expected storage.templates.centos to be 'path/to/centos.qcow2', got %s", cfg.Storage.Templates["centos"])
	}
}

func TestYAMLLoader_Load(t *testing.T) {
	// Create a temporary directory for the test
	tempDir, err := os.MkdirTemp("", "libgo-test-")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create a test config file
	configContent := `server:
  host: localhost
  port: 8080
`

	configPath := filepath.Join(tempDir, "config.yaml")
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to write config file: %v", err)
	}

	// Set an environment variable to override a config value
	os.Setenv("SERVER_PORT", "9090")
	defer os.Unsetenv("SERVER_PORT")

	// Load the config
	loader := NewYAMLLoader(configPath)
	cfg := &Config{}

	if err := loader.Load(cfg); err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Verify the loaded and overridden values
	if cfg.Server.Host != "localhost" {
		t.Errorf("Expected server.host to be 'localhost', got %s", cfg.Server.Host)
	}
	if cfg.Server.Port != 9090 {
		t.Errorf("Expected server.port to be 9090, got %d", cfg.Server.Port)
	}
}

func TestYAMLLoader_Load_Error(t *testing.T) {
	// Test loading a non-existent file
	loader := NewYAMLLoader("non-existent-file.yaml")
	cfg := &Config{}

	if err := loader.Load(cfg); err == nil {
		t.Errorf("Expected an error when loading a non-existent file, got nil")
	}
}

func TestBuildEnvVarName(t *testing.T) {
	tests := []struct {
		name     string
		prefix   string
		field    string
		expected string
	}{
		{
			name:     "No prefix",
			prefix:   "",
			field:    "port",
			expected: "PORT",
		},
		{
			name:     "With prefix",
			prefix:   "server",
			field:    "port",
			expected: "SERVER_PORT",
		},
		{
			name:     "Nested prefix",
			prefix:   "server_tls",
			field:    "enabled",
			expected: "SERVER_TLS_ENABLED",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := buildEnvVarName(tt.prefix, tt.field)
			if result != tt.expected {
				t.Errorf("buildEnvVarName(%q, %q) = %q; want %q", tt.prefix, tt.field, result, tt.expected)
			}
		})
	}
}

func TestApplyEnvValueToField(t *testing.T) {
	type testStruct struct {
		String        string
		Int           int
		Bool          bool
		Float         float64
		Duration      time.Duration
		Map           map[string]string
		StringSlice   []string
		IntSlice      []int
	}

	tests := []struct {
		name      string
		field     string
		envValue  string
		expected  interface{}
		expectErr bool
	}{
		{
			name:     "String value",
			field:    "String",
			envValue: "test-value",
			expected: "test-value",
		},
		{
			name:     "Int value",
			field:    "Int",
			envValue: "42",
			expected: 42,
		},
		{
			name:     "Bool value true",
			field:    "Bool",
			envValue: "true",
			expected: true,
		},
		{
			name:     "Bool value false",
			field:    "Bool",
			envValue: "false",
			expected: false,
		},
		{
			name:      "Invalid bool value",
			field:     "Bool",
			envValue:  "not-a-bool",
			expectErr: true,
		},
		{
			name:     "Float value",
			field:    "Float",
			envValue: "3.14159",
			expected: 3.14159,
		},
		{
			name:      "Invalid float value",
			field:     "Float",
			envValue:  "not-a-float",
			expectErr: true,
		},
		{
			name:     "Duration value",
			field:    "Duration",
			envValue: "10m",
			expected: 10 * time.Minute,
		},
		{
			name:      "Invalid duration value",
			field:     "Duration",
			envValue:  "not-a-duration",
			expectErr: true,
		},
		{
			name:     "Map value",
			field:    "Map",
			envValue: "key1:value1,key2:value2",
			expected: map[string]string{"key1": "value1", "key2": "value2"},
		},
		{
			name:      "Invalid map format",
			field:     "Map",
			envValue:  "invalid-format",
			expectErr: true,
		},
		{
			name:     "String slice",
			field:    "StringSlice",
			envValue: "value1,value2,value3",
			expected: []string{"value1", "value2", "value3"},
		},
		{
			name:     "Int slice",
			field:    "IntSlice",
			envValue: "1,2,3",
			expected: []int{1, 2, 3},
		},
		{
			name:      "Invalid int slice",
			field:     "IntSlice",
			envValue:  "1,not-an-int,3",
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a new instance of testStruct
			s := testStruct{}
			
			// Get the field to set
			v := reflect.ValueOf(&s).Elem()
			field := v.FieldByName(tt.field)
			
			// Apply the environment value
			err := applyEnvValueToField(field, tt.envValue)
			
			// Check for expected errors
			if (err != nil) != tt.expectErr {
				t.Errorf("applyEnvValueToField() error = %v, expectErr %v", err, tt.expectErr)
				return
			}
			
			if err != nil {
				// If we expected an error, no need to check the value
				return
			}
			
			// Check the field value based on its type
			switch tt.field {
			case "String":
				if s.String != tt.expected.(string) {
					t.Errorf("s.String = %v; want %v", s.String, tt.expected)
				}
			case "Int":
				if s.Int != tt.expected.(int) {
					t.Errorf("s.Int = %v; want %v", s.Int, tt.expected)
				}
			case "Bool":
				if s.Bool != tt.expected.(bool) {
					t.Errorf("s.Bool = %v; want %v", s.Bool, tt.expected)
				}
			case "Float":
				if s.Float != tt.expected.(float64) {
					t.Errorf("s.Float = %v; want %v", s.Float, tt.expected)
				}
			case "Duration":
				if s.Duration != tt.expected.(time.Duration) {
					t.Errorf("s.Duration = %v; want %v", s.Duration, tt.expected)
				}
			case "Map":
				expectedMap := tt.expected.(map[string]string)
				if len(s.Map) != len(expectedMap) {
					t.Errorf("len(s.Map) = %v; want %v", len(s.Map), len(expectedMap))
				}
				for k, v := range expectedMap {
					if s.Map[k] != v {
						t.Errorf("s.Map[%q] = %v; want %v", k, s.Map[k], v)
					}
				}
			case "StringSlice":
				expectedSlice := tt.expected.([]string)
				if len(s.StringSlice) != len(expectedSlice) {
					t.Errorf("len(s.StringSlice) = %v; want %v", len(s.StringSlice), len(expectedSlice))
				}
				for i, v := range expectedSlice {
					if s.StringSlice[i] != v {
						t.Errorf("s.StringSlice[%d] = %v; want %v", i, s.StringSlice[i], v)
					}
				}
			case "IntSlice":
				expectedSlice := tt.expected.([]int)
				if len(s.IntSlice) != len(expectedSlice) {
					t.Errorf("len(s.IntSlice) = %v; want %v", len(s.IntSlice), len(expectedSlice))
				}
				for i, v := range expectedSlice {
					if s.IntSlice[i] != v {
						t.Errorf("s.IntSlice[%d] = %v; want %v", i, s.IntSlice[i], v)
					}
				}
			}
		})
	}
}
