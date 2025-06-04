package config

import (
	"errors"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strings"
)

const localhostHost = "localhost"

// Common errors.
var (
	ErrEmptyValue         = errors.New("value cannot be empty")
	ErrFileNotAccessible  = errors.New("file is not accessible")
	ErrDirectoryNotExists = errors.New("directory does not exist")
	ErrInvalidPort        = errors.New("invalid port number")
	ErrInvalidTimeout     = errors.New("invalid timeout value")
	ErrInvalidFormat      = errors.New("invalid format")
)

// Validate checks if the configuration is valid.
func Validate(cfg *Config) error {
	if err := ValidateServer(cfg.Server); err != nil {
		return fmt.Errorf("server config: %w", err)
	}

	if err := ValidateLibvirt(cfg.Libvirt); err != nil {
		return fmt.Errorf("libvirt config: %w", err)
	}

	if err := ValidateAuth(cfg.Auth); err != nil {
		return fmt.Errorf("auth config: %w", err)
	}

	if err := ValidateLogging(cfg.Logging); err != nil {
		return fmt.Errorf("logging config: %w", err)
	}

	if err := ValidateStorage(cfg.Storage); err != nil {
		return fmt.Errorf("storage config: %w", err)
	}

	if err := ValidateExport(cfg.Export); err != nil {
		return fmt.Errorf("export config: %w", err)
	}

	return nil
}

// ValidateServer validates server configuration.
func ValidateServer(server ServerConfig) error {
	// Validate host if specified.
	if server.Host != "" {
		if ip := net.ParseIP(server.Host); ip == nil && server.Host != localhostHost {
			if _, err := net.LookupHost(server.Host); err != nil {
				return fmt.Errorf("invalid host: %w", err)
			}
		}
	}

	// Validate port.
	if server.Port < 1 || server.Port > 65535 {
		return fmt.Errorf("port %d: %w", server.Port, ErrInvalidPort)
	}

	// Validate timeouts.
	if server.ReadTimeout <= 0 {
		return fmt.Errorf("read timeout: %w", ErrInvalidTimeout)
	}

	if server.WriteTimeout <= 0 {
		return fmt.Errorf("write timeout: %w", ErrInvalidTimeout)
	}

	// Validate TLS settings if enabled.
	if server.TLS.Enabled {
		if server.TLS.CertFile == "" {
			return fmt.Errorf("TLS cert file: %w", ErrEmptyValue)
		}

		if server.TLS.KeyFile == "" {
			return fmt.Errorf("TLS key file: %w", ErrEmptyValue)
		}

		// Check if cert and key files exist and are readable.
		if err := checkFileReadable(server.TLS.CertFile); err != nil {
			return fmt.Errorf("TLS cert file: %w", err)
		}

		if err := checkFileReadable(server.TLS.KeyFile); err != nil {
			return fmt.Errorf("TLS key file: %w", err)
		}
	}

	return nil
}

// ValidateLibvirt validates libvirt configuration.
func ValidateLibvirt(libvirt LibvirtConfig) error {
	// URI should not be empty.
	if libvirt.URI == "" {
		return fmt.Errorf("URI: %w", ErrEmptyValue)
	}

	// Check URI format.
	if !strings.HasPrefix(libvirt.URI, "qemu") &&
		!strings.HasPrefix(libvirt.URI, "xen") &&
		!strings.HasPrefix(libvirt.URI, "lxc") &&
		!strings.HasPrefix(libvirt.URI, "test") {
		return fmt.Errorf("URI %s: unsupported hypervisor", libvirt.URI)
	}

	// Connection timeout should be positive.
	if libvirt.ConnectionTimeout <= 0 {
		return fmt.Errorf("connection timeout: %w", ErrInvalidTimeout)
	}

	// Max connections should be at least 1.
	if libvirt.MaxConnections < 1 {
		return fmt.Errorf("max connections must be at least 1")
	}

	// Pool name should not be empty.
	if libvirt.PoolName == "" {
		return fmt.Errorf("pool name: %w", ErrEmptyValue)
	}

	// Network name should not be empty.
	if libvirt.NetworkName == "" {
		return fmt.Errorf("network name: %w", ErrEmptyValue)
	}

	return nil
}

// ValidateAuth validates authentication configuration.
func ValidateAuth(auth AuthConfig) error {
	// If auth is disabled, no need to validate further.
	if !auth.Enabled {
		return nil
	}

	// JWT secret should not be empty.
	if auth.JWTSecretKey == "" {
		return fmt.Errorf("JWT secret key: %w", ErrEmptyValue)
	}

	// Token expiration should be positive.
	if auth.TokenExpiration <= 0 {
		return fmt.Errorf("token expiration: %w", ErrInvalidTimeout)
	}

	// Validate signing method.
	validMethods := map[string]bool{
		"HS256": true,
		"HS384": true,
		"HS512": true,
		"RS256": true,
		"RS384": true,
		"RS512": true,
		"ES256": true,
		"ES384": true,
		"ES512": true,
	}

	if !validMethods[auth.SigningMethod] {
		return fmt.Errorf("signing method %s: %w", auth.SigningMethod, ErrInvalidFormat)
	}

	return nil
}

// ValidateLogging validates logging configuration.
func ValidateLogging(logging LoggingConfig) error {
	// Validate log level.
	validLevels := map[string]bool{
		"debug":  true,
		"info":   true,
		"warn":   true,
		"error":  true,
		"dpanic": true,
		"panic":  true,
		"fatal":  true,
	}

	if !validLevels[strings.ToLower(logging.Level)] {
		return fmt.Errorf("log level %s: %w", logging.Level, ErrInvalidFormat)
	}

	// Validate log format.
	validFormats := map[string]bool{
		"json":    true,
		"console": true,
	}

	if !validFormats[strings.ToLower(logging.Format)] {
		return fmt.Errorf("log format %s: %w", logging.Format, ErrInvalidFormat)
	}

	// If file path is specified, ensure directory exists.
	if logging.FilePath != "" {
		dir := filepath.Dir(logging.FilePath)
		if err := checkDirWritable(dir); err != nil {
			return fmt.Errorf("log directory: %w", err)
		}
	}

	// Max size should be positive if set.
	if logging.MaxSize < 0 {
		return fmt.Errorf("max size must be non-negative")
	}

	// Max backups should be non-negative.
	if logging.MaxBackups < 0 {
		return fmt.Errorf("max backups must be non-negative")
	}

	// Max age should be non-negative.
	if logging.MaxAge < 0 {
		return fmt.Errorf("max age must be non-negative")
	}

	return nil
}

// ValidateStorage validates storage configuration.
func ValidateStorage(storage StorageConfig) error {
	// Default pool should not be empty.
	if storage.DefaultPool == "" {
		return fmt.Errorf("default pool: %w", ErrEmptyValue)
	}

	// Pool path should be a valid directory.
	if storage.PoolPath == "" {
		return fmt.Errorf("pool path: %w", ErrEmptyValue)
	}

	if err := checkDirWritable(storage.PoolPath); err != nil {
		return fmt.Errorf("pool path: %w", err)
	}

	// Validate templates if provided.
	for name, path := range storage.Templates {
		if name == "" {
			return fmt.Errorf("template name: %w", ErrEmptyValue)
		}

		if path == "" {
			return fmt.Errorf("template path for %s: %w", name, ErrEmptyValue)
		}

		if err := checkFileReadable(path); err != nil {
			return fmt.Errorf("template %s: %w", name, err)
		}
	}

	return nil
}

// ValidateExport validates export configuration.
func ValidateExport(export ExportConfig) error {
	// Output directory should exist and be writable.
	if export.OutputDir == "" {
		return fmt.Errorf("output directory: %w", ErrEmptyValue)
	}

	if err := checkDirWritable(export.OutputDir); err != nil {
		return fmt.Errorf("output directory: %w", err)
	}

	// Temp directory should exist and be writable.
	if export.TempDir == "" {
		return fmt.Errorf("temp directory: %w", ErrEmptyValue)
	}

	if err := checkDirWritable(export.TempDir); err != nil {
		return fmt.Errorf("temp directory: %w", err)
	}

	// Validate default format.
	validFormats := map[string]bool{
		"qcow2": true,
		"vmdk":  true,
		"vdi":   true,
		"ova":   true,
		"raw":   true,
	}

	if !validFormats[export.DefaultFormat] {
		return fmt.Errorf("default format %s: %w", export.DefaultFormat, ErrInvalidFormat)
	}

	// Retention should be positive.
	if export.Retention <= 0 {
		return fmt.Errorf("retention: %w", ErrInvalidTimeout)
	}

	return nil
}

// Helper functions.

// checkFileReadable checks if a file exists and is readable.
func checkFileReadable(path string) error {
	_, err := os.Stat(path)
	if os.IsNotExist(err) {
		return fmt.Errorf("%s: %w", path, ErrFileNotAccessible)
	}
	if err != nil {
		return fmt.Errorf("accessing %s: %w", path, err)
	}

	// Check if file is readable.
	file, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("opening %s: %w", path, err)
	}
	defer file.Close()

	return nil
}

// checkDirWritable checks if a directory exists and is writable.
func checkDirWritable(path string) error {
	fi, err := os.Stat(path)
	if os.IsNotExist(err) {
		return fmt.Errorf("%s: %w", path, ErrDirectoryNotExists)
	}
	if err != nil {
		return fmt.Errorf("accessing %s: %w", path, err)
	}

	// Check if it's a directory.
	if !fi.IsDir() {
		return fmt.Errorf("%s is not a directory", path)
	}

	// Check if directory is writable by attempting to create a temporary file.
	tempFile := filepath.Join(path, ".libgo-write-test")
	f, err := os.Create(tempFile)
	if err != nil {
		return fmt.Errorf("directory %s is not writable: %w", path, err)
	}

	// Clean up the temporary file.
	f.Close()
	os.Remove(tempFile)

	return nil
}
