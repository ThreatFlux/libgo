package config

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestValidateServer(t *testing.T) {
	tests := []struct {
		name    string
		server  ServerConfig
		wantErr bool
	}{
		{
			name: "Valid config",
			server: ServerConfig{
				Host:           "localhost",
				Port:           8080,
				ReadTimeout:    30 * time.Second,
				WriteTimeout:   30 * time.Second,
				MaxHeaderBytes: 1 << 20,
				TLS: TLSConfig{
					Enabled: false,
				},
			},
			wantErr: false,
		},
		{
			name: "Invalid port (too low)",
			server: ServerConfig{
				Host:           "localhost",
				Port:           0,
				ReadTimeout:    30 * time.Second,
				WriteTimeout:   30 * time.Second,
				MaxHeaderBytes: 1 << 20,
			},
			wantErr: true,
		},
		{
			name: "Invalid port (too high)",
			server: ServerConfig{
				Host:           "localhost",
				Port:           70000,
				ReadTimeout:    30 * time.Second,
				WriteTimeout:   30 * time.Second,
				MaxHeaderBytes: 1 << 20,
			},
			wantErr: true,
		},
		{
			name: "Invalid read timeout",
			server: ServerConfig{
				Host:           "localhost",
				Port:           8080,
				ReadTimeout:    0,
				WriteTimeout:   30 * time.Second,
				MaxHeaderBytes: 1 << 20,
			},
			wantErr: true,
		},
		{
			name: "Invalid write timeout",
			server: ServerConfig{
				Host:           "localhost",
				Port:           8080,
				ReadTimeout:    30 * time.Second,
				WriteTimeout:   0,
				MaxHeaderBytes: 1 << 20,
			},
			wantErr: true,
		},
		{
			name: "TLS enabled but missing cert file",
			server: ServerConfig{
				Host:           "localhost",
				Port:           8443,
				ReadTimeout:    30 * time.Second,
				WriteTimeout:   30 * time.Second,
				MaxHeaderBytes: 1 << 20,
				TLS: TLSConfig{
					Enabled:  true,
					KeyFile:  "testdata/key.pem",
					CertFile: "",
				},
			},
			wantErr: true,
		},
		{
			name: "TLS enabled but missing key file",
			server: ServerConfig{
				Host:           "localhost",
				Port:           8443,
				ReadTimeout:    30 * time.Second,
				WriteTimeout:   30 * time.Second,
				MaxHeaderBytes: 1 << 20,
				TLS: TLSConfig{
					Enabled:  true,
					KeyFile:  "",
					CertFile: "testdata/cert.pem",
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateServer(tt.server)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateServer() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateLibvirt(t *testing.T) {
	tests := []struct {
		name    string
		libvirt LibvirtConfig
		wantErr bool
	}{
		{
			name: "Valid config",
			libvirt: LibvirtConfig{
				URI:               "qemu:///system",
				ConnectionTimeout: 30 * time.Second,
				MaxConnections:    5,
				PoolName:          "default",
				NetworkName:       "default",
			},
			wantErr: false,
		},
		{
			name: "Empty URI",
			libvirt: LibvirtConfig{
				URI:               "",
				ConnectionTimeout: 30 * time.Second,
				MaxConnections:    5,
				PoolName:          "default",
				NetworkName:       "default",
			},
			wantErr: true,
		},
		{
			name: "Invalid URI",
			libvirt: LibvirtConfig{
				URI:               "invalid-uri",
				ConnectionTimeout: 30 * time.Second,
				MaxConnections:    5,
				PoolName:          "default",
				NetworkName:       "default",
			},
			wantErr: true,
		},
		{
			name: "Invalid connection timeout",
			libvirt: LibvirtConfig{
				URI:               "qemu:///system",
				ConnectionTimeout: 0,
				MaxConnections:    5,
				PoolName:          "default",
				NetworkName:       "default",
			},
			wantErr: true,
		},
		{
			name: "Invalid max connections",
			libvirt: LibvirtConfig{
				URI:               "qemu:///system",
				ConnectionTimeout: 30 * time.Second,
				MaxConnections:    0,
				PoolName:          "default",
				NetworkName:       "default",
			},
			wantErr: true,
		},
		{
			name: "Empty pool name",
			libvirt: LibvirtConfig{
				URI:               "qemu:///system",
				ConnectionTimeout: 30 * time.Second,
				MaxConnections:    5,
				PoolName:          "",
				NetworkName:       "default",
			},
			wantErr: true,
		},
		{
			name: "Empty network name",
			libvirt: LibvirtConfig{
				URI:               "qemu:///system",
				ConnectionTimeout: 30 * time.Second,
				MaxConnections:    5,
				PoolName:          "default",
				NetworkName:       "",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateLibvirt(tt.libvirt)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateLibvirt() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateAuth(t *testing.T) {
	tests := []struct {
		name    string
		auth    AuthConfig
		wantErr bool
	}{
		{
			name: "Valid config",
			auth: AuthConfig{
				Enabled:         true,
				JWTSecretKey:    "my-secret-key",
				Issuer:          "libgo-server",
				Audience:        "libgo-clients",
				TokenExpiration: 15 * time.Minute,
				SigningMethod:   "HS256",
			},
			wantErr: false,
		},
		{
			name: "Auth disabled",
			auth: AuthConfig{
				Enabled: false,
			},
			wantErr: false,
		},
		{
			name: "Empty JWT secret",
			auth: AuthConfig{
				Enabled:         true,
				JWTSecretKey:    "",
				Issuer:          "libgo-server",
				Audience:        "libgo-clients",
				TokenExpiration: 15 * time.Minute,
				SigningMethod:   "HS256",
			},
			wantErr: true,
		},
		{
			name: "Invalid token expiration",
			auth: AuthConfig{
				Enabled:         true,
				JWTSecretKey:    "my-secret-key",
				Issuer:          "libgo-server",
				Audience:        "libgo-clients",
				TokenExpiration: 0,
				SigningMethod:   "HS256",
			},
			wantErr: true,
		},
		{
			name: "Invalid signing method",
			auth: AuthConfig{
				Enabled:         true,
				JWTSecretKey:    "my-secret-key",
				Issuer:          "libgo-server",
				Audience:        "libgo-clients",
				TokenExpiration: 15 * time.Minute,
				SigningMethod:   "INVALID",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateAuth(tt.auth)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateAuth() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateLogging(t *testing.T) {
	tests := []struct {
		name    string
		logging LoggingConfig
		wantErr bool
	}{
		{
			name: "Valid config",
			logging: LoggingConfig{
				Level:      "info",
				Format:     "json",
				FilePath:   "",
				MaxSize:    10,
				MaxBackups: 5,
				MaxAge:     30,
				Compress:   true,
			},
			wantErr: false,
		},
		{
			name: "Invalid level",
			logging: LoggingConfig{
				Level:  "invalid",
				Format: "json",
			},
			wantErr: true,
		},
		{
			name: "Invalid format",
			logging: LoggingConfig{
				Level:  "info",
				Format: "invalid",
			},
			wantErr: true,
		},
		{
			name: "Negative max size",
			logging: LoggingConfig{
				Level:   "info",
				Format:  "json",
				MaxSize: -1,
			},
			wantErr: true,
		},
		{
			name: "Negative max backups",
			logging: LoggingConfig{
				Level:      "info",
				Format:     "json",
				MaxSize:    10,
				MaxBackups: -1,
			},
			wantErr: true,
		},
		{
			name: "Negative max age",
			logging: LoggingConfig{
				Level:      "info",
				Format:     "json",
				MaxSize:    10,
				MaxBackups: 5,
				MaxAge:     -1,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateLogging(tt.logging)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateLogging() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateStorage(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "libgo-test-")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create a test template file
	templatePath := filepath.Join(tempDir, "template.qcow2")
	if err := os.WriteFile(templatePath, []byte("test"), 0644); err != nil {
		t.Fatalf("Failed to create test template file: %v", err)
	}

	tests := []struct {
		name    string
		storage StorageConfig
		wantErr bool
	}{
		{
			name: "Valid config",
			storage: StorageConfig{
				DefaultPool: "default",
				PoolPath:    tempDir,
				Templates: map[string]string{
					"ubuntu": templatePath,
				},
			},
			wantErr: false,
		},
		{
			name: "Empty default pool",
			storage: StorageConfig{
				DefaultPool: "",
				PoolPath:    tempDir,
			},
			wantErr: true,
		},
		{
			name: "Empty pool path",
			storage: StorageConfig{
				DefaultPool: "default",
				PoolPath:    "",
			},
			wantErr: true,
		},
		{
			name: "Non-existent pool path",
			storage: StorageConfig{
				DefaultPool: "default",
				PoolPath:    "/path/that/does/not/exist",
			},
			wantErr: true,
		},
		{
			name: "Template with empty name",
			storage: StorageConfig{
				DefaultPool: "default",
				PoolPath:    tempDir,
				Templates: map[string]string{
					"": templatePath,
				},
			},
			wantErr: true,
		},
		{
			name: "Template with empty path",
			storage: StorageConfig{
				DefaultPool: "default",
				PoolPath:    tempDir,
				Templates: map[string]string{
					"ubuntu": "",
				},
			},
			wantErr: true,
		},
		{
			name: "Template with non-existent path",
			storage: StorageConfig{
				DefaultPool: "default",
				PoolPath:    tempDir,
				Templates: map[string]string{
					"ubuntu": "/path/to/nonexistent/template.qcow2",
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateStorage(tt.storage)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateStorage() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateExport(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "libgo-test-")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	tests := []struct {
		name    string
		export  ExportConfig
		wantErr bool
	}{
		{
			name: "Valid config",
			export: ExportConfig{
				OutputDir:     tempDir,
				TempDir:       tempDir,
				DefaultFormat: "qcow2",
				Retention:     7 * 24 * time.Hour,
			},
			wantErr: false,
		},
		{
			name: "Empty output dir",
			export: ExportConfig{
				OutputDir:     "",
				TempDir:       tempDir,
				DefaultFormat: "qcow2",
				Retention:     7 * 24 * time.Hour,
			},
			wantErr: true,
		},
		{
			name: "Non-existent output dir",
			export: ExportConfig{
				OutputDir:     "/path/that/does/not/exist",
				TempDir:       tempDir,
				DefaultFormat: "qcow2",
				Retention:     7 * 24 * time.Hour,
			},
			wantErr: true,
		},
		{
			name: "Empty temp dir",
			export: ExportConfig{
				OutputDir:     tempDir,
				TempDir:       "",
				DefaultFormat: "qcow2",
				Retention:     7 * 24 * time.Hour,
			},
			wantErr: true,
		},
		{
			name: "Non-existent temp dir",
			export: ExportConfig{
				OutputDir:     tempDir,
				TempDir:       "/path/that/does/not/exist",
				DefaultFormat: "qcow2",
				Retention:     7 * 24 * time.Hour,
			},
			wantErr: true,
		},
		{
			name: "Invalid format",
			export: ExportConfig{
				OutputDir:     tempDir,
				TempDir:       tempDir,
				DefaultFormat: "invalid",
				Retention:     7 * 24 * time.Hour,
			},
			wantErr: true,
		},
		{
			name: "Invalid retention",
			export: ExportConfig{
				OutputDir:     tempDir,
				TempDir:       tempDir,
				DefaultFormat: "qcow2",
				Retention:     0,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateExport(tt.export)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateExport() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidate(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "libgo-test-")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Valid config
	validConfig := Config{
		Server: ServerConfig{
			Host:           "localhost",
			Port:           8080,
			ReadTimeout:    30 * time.Second,
			WriteTimeout:   30 * time.Second,
			MaxHeaderBytes: 1 << 20,
			TLS: TLSConfig{
				Enabled: false,
			},
		},
		Libvirt: LibvirtConfig{
			URI:               "qemu:///system",
			ConnectionTimeout: 30 * time.Second,
			MaxConnections:    5,
			PoolName:          "default",
			NetworkName:       "default",
		},
		Auth: AuthConfig{
			Enabled:         true,
			JWTSecretKey:    "my-secret-key",
			Issuer:          "libgo-server",
			Audience:        "libgo-clients",
			TokenExpiration: 15 * time.Minute,
			SigningMethod:   "HS256",
		},
		Logging: LoggingConfig{
			Level:      "info",
			Format:     "json",
			FilePath:   "",
			MaxSize:    10,
			MaxBackups: 5,
			MaxAge:     30,
			Compress:   true,
		},
		Storage: StorageConfig{
			DefaultPool: "default",
			PoolPath:    tempDir,
			Templates:   map[string]string{},
		},
		Export: ExportConfig{
			OutputDir:     tempDir,
			TempDir:       tempDir,
			DefaultFormat: "qcow2",
			Retention:     7 * 24 * time.Hour,
		},
		Features: FeaturesConfig{
			CloudInit:      true,
			ExportFeature:  true,
			Metrics:        true,
			RBACEnabled:    true,
			StorageCleanup: true,
		},
	}

	// Test with valid config
	if err := Validate(&validConfig); err != nil {
		t.Errorf("Validate() error = %v, wantErr %v", err, false)
	}

	// Test with invalid server config
	invalidServerConfig := validConfig
	invalidServerConfig.Server.Port = 0
	if err := Validate(&invalidServerConfig); err == nil {
		t.Errorf("Validate() with invalid server config - error = %v, wantErr %v", err, true)
	}

	// Test with invalid libvirt config
	invalidLibvirtConfig := validConfig
	invalidLibvirtConfig.Libvirt.URI = ""
	if err := Validate(&invalidLibvirtConfig); err == nil {
		t.Errorf("Validate() with invalid libvirt config - error = %v, wantErr %v", err, true)
	}

	// Test with invalid auth config
	invalidAuthConfig := validConfig
	invalidAuthConfig.Auth.SigningMethod = "INVALID"
	if err := Validate(&invalidAuthConfig); err == nil {
		t.Errorf("Validate() with invalid auth config - error = %v, wantErr %v", err, true)
	}

	// Test with invalid logging config
	invalidLoggingConfig := validConfig
	invalidLoggingConfig.Logging.Level = "INVALID"
	if err := Validate(&invalidLoggingConfig); err == nil {
		t.Errorf("Validate() with invalid logging config - error = %v, wantErr %v", err, true)
	}

	// Test with invalid storage config
	invalidStorageConfig := validConfig
	invalidStorageConfig.Storage.DefaultPool = ""
	if err := Validate(&invalidStorageConfig); err == nil {
		t.Errorf("Validate() with invalid storage config - error = %v, wantErr %v", err, true)
	}

	// Test with invalid export config
	invalidExportConfig := validConfig
	invalidExportConfig.Export.DefaultFormat = "INVALID"
	if err := Validate(&invalidExportConfig); err == nil {
		t.Errorf("Validate() with invalid export config - error = %v, wantErr %v", err, true)
	}
}
