package vdi

import (
	"context"
	"fmt"

	"github.com/wroersma/libgo/pkg/logger"
	"github.com/wroersma/libgo/pkg/utils/exec"
)

// VDIConverter implements Converter for VDI format
type VDIConverter struct {
	logger logger.Logger
}

// NewVDIConverter creates a new VDIConverter
func NewVDIConverter(logger logger.Logger) *VDIConverter {
	return &VDIConverter{
		logger: logger,
	}
}

// Convert implements Converter.Convert
func (c *VDIConverter) Convert(ctx context.Context, sourcePath string, destPath string, options map[string]string) error {
	// Determine if static (preallocated) or dynamic (growing)
	static := false
	if staticOpt, ok := options["static"]; ok && (staticOpt == "true" || staticOpt == "1") {
		static = true
	}

	// Prepare qemu-img arguments
	args := []string{
		"convert",
		"-f", "qcow2", // Source format, assume qcow2 as libvirt generally uses this
		"-O", "vdi",   // Output format
	}

	// Add static allocation if requested
	if static {
		args = append(args, "-o", "preallocation=metadata")
	}

	// Add source and destination
	args = append(args, sourcePath, destPath)

	c.logger.Info("Converting to VDI format",
		logger.String("source", sourcePath),
		logger.String("destination", destPath),
		logger.Bool("static", static))

	// Execute qemu-img command
	execOpts := exec.CommandOptions{
		Timeout: 0, // No timeout, conversion might take a long time
	}

	output, err := exec.ExecuteCommand(ctx, "qemu-img", args, execOpts)
	if err != nil {
		c.logger.Error("Failed to convert to VDI",
			logger.String("error", err.Error()),
			logger.String("output", string(output)))
		return fmt.Errorf("vdi conversion failed: %w", err)
	}

	c.logger.Info("Successfully converted to VDI format",
		logger.String("destination", destPath))
	return nil
}

// GetFormatName implements Converter.GetFormatName
func (c *VDIConverter) GetFormatName() string {
	return "vdi"
}

// ValidateOptions implements Converter.ValidateOptions
func (c *VDIConverter) ValidateOptions(options map[string]string) error {
	// Currently, only static/dynamic allocation option is supported
	// All other options are ignored
	if staticOpt, ok := options["static"]; ok {
		if staticOpt != "true" && staticOpt != "false" && staticOpt != "1" && staticOpt != "0" {
			return fmt.Errorf("invalid value for static option: %s, must be 'true', 'false', '1', or '0'", staticOpt)
		}
	}

	return nil
}
