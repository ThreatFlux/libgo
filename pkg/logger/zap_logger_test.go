package logger

import (
	"errors"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/threatflux/libgo/internal/config"
)

func TestZapLogger_Levels(t *testing.T) {
	// Create a temporary file for logging
	tmpDir, err := ioutil.TempDir("", "logger_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	logFile := filepath.Join(tmpDir, "test.log")

	cfg := config.LoggingConfig{
		Level:      "debug",
		Format:     "json",
		OutputPath: logFile,
	}

	logger, err := NewZapLogger(cfg)
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}

	// Log messages at different levels
	logger.Debug("debug message", String("key", "value"))
	logger.Info("info message", Int("count", 42))
	logger.Warn("warn message", Bool("enabled", true))
	logger.Error("error message", Error(errors.New("test error")))

	// Sync to ensure logs are written
	if err := logger.Sync(); err != nil {
		t.Logf("Sync error (may be expected on some platforms): %v", err)
	}

	// Read log file content
	content, err := ioutil.ReadFile(logFile)
	if err != nil {
		t.Fatalf("Failed to read log file: %v", err)
	}

	// Check if logs contain expected content
	logContent := string(content)

	// Each log level should be present
	expectedMessages := []string{
		"debug message",
		"info message",
		"warn message",
		"error message",
	}

	// Each field should be present
	expectedFields := []string{
		`"key":"value"`,
		`"count":42`,
		`"enabled":true`,
		`"error":{}`,
	}

	for _, msg := range expectedMessages {
		if !strings.Contains(logContent, msg) {
			t.Errorf("Log content doesn't contain expected message: %s", msg)
		}
	}

	for _, field := range expectedFields {
		if !strings.Contains(logContent, field) {
			t.Errorf("Log content doesn't contain expected field: %s", field)
		}
	}
}

func TestZapLogger_WithFields(t *testing.T) {
	// Create a temporary file for logging
	tmpDir, err := ioutil.TempDir("", "logger_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	logFile := filepath.Join(tmpDir, "test.log")

	cfg := config.LoggingConfig{
		Level:      "info",
		Format:     "json",
		OutputPath: logFile,
	}

	baseLogger, err := NewZapLogger(cfg)
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}

	// Create logger with fields
	contextLogger := baseLogger.WithFields(
		String("service", "test-service"),
		Int("instance", 1),
	)

	// Log with the context logger
	contextLogger.Info("context log message")

	// Add more context with WithError
	errLogger := contextLogger.WithError(errors.New("context error"))
	errLogger.Error("error with context")

	// Sync to ensure logs are written
	if err := baseLogger.Sync(); err != nil {
		t.Logf("Sync error (may be expected on some platforms): %v", err)
	}

	// Read log file content
	content, err := ioutil.ReadFile(logFile)
	if err != nil {
		t.Fatalf("Failed to read log file: %v", err)
	}

	logContent := string(content)

	// Check for fields in the context logger
	expectedFields := []string{
		`"service":"test-service"`,
		`"instance":1`,
		`"error":{}`,
	}

	for _, field := range expectedFields {
		if !strings.Contains(logContent, field) {
			t.Errorf("Log content doesn't contain expected field: %s", field)
		}
	}
}

func TestZapLogger_FormatTypes(t *testing.T) {
	tests := []struct {
		name   string
		format string
	}{
		{
			name:   "JSON format",
			format: "json",
		},
		{
			name:   "Console format",
			format: "console",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir, err := ioutil.TempDir("", "logger_test")
			if err != nil {
				t.Fatalf("Failed to create temp dir: %v", err)
			}
			defer os.RemoveAll(tmpDir)

			logFile := filepath.Join(tmpDir, "test.log")

			cfg := config.LoggingConfig{
				Level:      "info",
				Format:     tt.format,
				OutputPath: logFile,
			}

			logger, err := NewZapLogger(cfg)
			if err != nil {
				t.Fatalf("Failed to create logger: %v", err)
			}

			logger.Info("test message", String("format", tt.format))

			if err := logger.Sync(); err != nil {
				t.Logf("Sync error (may be expected on some platforms): %v", err)
			}

			// Verify the log file exists
			if _, err := os.Stat(logFile); os.IsNotExist(err) {
				t.Errorf("Log file was not created")
			}
		})
	}
}

func TestZapLogger_OutputPaths(t *testing.T) {
	tests := []struct {
		name       string
		outputPath string
		shouldErr  bool
	}{
		{
			name:       "Stdout output",
			outputPath: "stdout",
			shouldErr:  false,
		},
		{
			name:       "Stderr output",
			outputPath: "stderr",
			shouldErr:  false,
		},
		{
			name:       "File output",
			outputPath: "", // Will be set to a temp file
			shouldErr:  false,
		},
		{
			name:       "Invalid path",
			outputPath: "/nonexistent/directory/file.log",
			shouldErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var outputPath string
			if tt.outputPath == "" {
				// Create a temporary file
				tmpDir, err := ioutil.TempDir("", "logger_test")
				if err != nil {
					t.Fatalf("Failed to create temp dir: %v", err)
				}
				defer os.RemoveAll(tmpDir)
				outputPath = filepath.Join(tmpDir, "test.log")
			} else {
				outputPath = tt.outputPath
			}

			cfg := config.LoggingConfig{
				Level:      "info",
				Format:     "json",
				OutputPath: outputPath,
			}

			logger, err := NewZapLogger(cfg)
			if tt.shouldErr {
				if err == nil {
					t.Errorf("Expected error when creating logger, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("Failed to create logger: %v", err)
			}

			logger.Info("test message")

			if err := logger.Sync(); err != nil {
				// Sync may legitimately fail on stdout/stderr on some platforms
				if tt.outputPath != "stdout" && tt.outputPath != "stderr" {
					t.Errorf("Failed to sync logger: %v", err)
				}
			}
		})
	}
}
