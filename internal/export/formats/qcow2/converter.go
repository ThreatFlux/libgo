package qcow2

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"github.com/threatflux/libgo/pkg/logger"
	"github.com/threatflux/libgo/pkg/utils/exec"
)

const (
	// DefaultCompression is the default compression level for qcow2.
	DefaultCompression = "1"
	// MaxCompression is the maximum compression level.
	MaxCompression = 9
)

// QCOW2Converter implements Converter for QCOW2 format.
type QCOW2Converter struct {
	logger logger.Logger
}

// NewQCOW2Converter creates a new QCOW2Converter.
func NewQCOW2Converter(logger logger.Logger) *QCOW2Converter {
	return &QCOW2Converter{
		logger: logger,
	}
}

// Convert implements Converter.Convert
func (c *QCOW2Converter) Convert(ctx context.Context, sourcePath string, destPath string, options map[string]string) error {
	// Get compression level from options
	compression := DefaultCompression
	if compressionOpt, ok := options["compression"]; ok {
		compression = compressionOpt
	}

	// Make sure directory exists for destination
	destDir := filepath.Dir(destPath)
	if err := os.MkdirAll(destDir, 0755); err != nil {
		c.logger.Error("Failed to create destination directory",
			logger.String("dir", destDir),
			logger.String("error", err.Error()))
		return fmt.Errorf("creating destination directory: %w", err)
	}

	// Prepare qemu-img arguments
	args := []string{
		"convert",
		"-f", "qcow2", // Source format, assume qcow2 as libvirt generally uses this
		"-O", "qcow2", // Output format
		"-c", // Enable compression
		sourcePath,
		destPath,
	}

	c.logger.Info("Converting to QCOW2 format",
		logger.String("source", sourcePath),
		logger.String("destination", destPath),
		logger.String("compression", compression))

	// See if we need to use sudo
	execCommand := "qemu-img"
	if useSudo, ok := options["use_sudo"]; ok && (useSudo == "true" || useSudo == "1") {
		c.logger.Info("Using sudo for qemu-img command",
			logger.String("source", sourcePath),
			logger.String("destination", destPath))
		execCommand = "sudo"
		args = append([]string{"qemu-img"}, args...)
	}

	// Execute qemu-img command
	execOpts := exec.CommandOptions{
		Timeout: 0, // No timeout, conversion might take a long time
	}

	output, err := exec.ExecuteCommand(ctx, execCommand, args, execOpts)
	if err != nil {
		c.logger.Error("Failed to convert to QCOW2",
			logger.String("error", err.Error()),
			logger.String("output", string(output)))
		return fmt.Errorf("qcow2 conversion failed: %w: %s", err, string(output))
	}

	c.logger.Info("Successfully converted to QCOW2 format",
		logger.String("destination", destPath))
	return nil
}

// GetFormatName implements Converter.GetFormatName
func (c *QCOW2Converter) GetFormatName() string {
	return "qcow2"
}

// ValidateOptions implements Converter.ValidateOptions
func (c *QCOW2Converter) ValidateOptions(options map[string]string) error {
	// Validate compression level if provided
	if compressionStr, ok := options["compression"]; ok {
		compression, err := strconv.Atoi(compressionStr)
		if err != nil {
			return fmt.Errorf("invalid compression level: %s, must be a number", compressionStr)
		}
		if compression < 0 || compression > MaxCompression {
			return fmt.Errorf("compression level must be between 0 and %d", MaxCompression)
		}
	}

	return nil
}
