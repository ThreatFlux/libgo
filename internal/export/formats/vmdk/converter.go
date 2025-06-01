package vmdk

import (
	"context"
	"fmt"

	"github.com/threatflux/libgo/pkg/logger"
	"github.com/threatflux/libgo/pkg/utils/exec"
)

// VMDKConverter implements Converter for VMDK format
type VMDKConverter struct {
	logger logger.Logger
}

// NewVMDKConverter creates a new VMDKConverter
func NewVMDKConverter(logger logger.Logger) *VMDKConverter {
	return &VMDKConverter{
		logger: logger,
	}
}

// Convert implements Converter.Convert
func (c *VMDKConverter) Convert(ctx context.Context, sourcePath string, destPath string, options map[string]string) error {
	// Get adapter type from options
	adapterType := "lsilogic"
	if adapter, ok := options["adapter_type"]; ok {
		adapterType = adapter
	}

	// Get disk type from options (monolithicSparse, monolithicFlat, etc.)
	diskType := "monolithicSparse"
	if dt, ok := options["disk_type"]; ok {
		diskType = dt
	}

	// Prepare qemu-img arguments
	args := []string{
		"convert",
		"-f", "qcow2", // Source format, assume qcow2 as libvirt generally uses this
		"-O", "vmdk", // Output format
		"-o", fmt.Sprintf("adapter_type=%s,subformat=%s", adapterType, diskType),
		sourcePath,
		destPath,
	}

	c.logger.Info("Converting to VMDK format",
		logger.String("source", sourcePath),
		logger.String("destination", destPath),
		logger.String("adapter_type", adapterType),
		logger.String("disk_type", diskType))

	// Execute qemu-img command
	execOpts := exec.CommandOptions{
		Timeout: 0, // No timeout, conversion might take a long time
	}

	output, err := exec.ExecuteCommand(ctx, "qemu-img", args, execOpts)
	if err != nil {
		c.logger.Error("Failed to convert to VMDK",
			logger.String("error", err.Error()),
			logger.String("output", string(output)))
		return fmt.Errorf("vmdk conversion failed: %w", err)
	}

	c.logger.Info("Successfully converted to VMDK format",
		logger.String("destination", destPath))
	return nil
}

// GetFormatName implements Converter.GetFormatName
func (c *VMDKConverter) GetFormatName() string {
	return "vmdk"
}

// ValidateOptions implements Converter.ValidateOptions
func (c *VMDKConverter) ValidateOptions(options map[string]string) error {
	// Valid adapter types
	validAdapters := map[string]bool{
		"ide":       true,
		"buslogic":  true,
		"lsilogic":  true,
		"legacyESX": true,
	}

	// Valid disk types
	validDiskTypes := map[string]bool{
		"monolithicSparse": true,
		"monolithicFlat":   true,
		"twoGbMaxExtent":   true,
		"streamOptimized":  true,
	}

	// Validate adapter type if provided
	if adapterType, ok := options["adapter_type"]; ok {
		if !validAdapters[adapterType] {
			return fmt.Errorf("invalid adapter type: %s, must be one of: ide, buslogic, lsilogic, legacyESX", adapterType)
		}
	}

	// Validate disk type if provided
	if diskType, ok := options["disk_type"]; ok {
		if !validDiskTypes[diskType] {
			return fmt.Errorf("invalid disk type: %s, must be one of: monolithicSparse, monolithicFlat, twoGbMaxExtent, streamOptimized", diskType)
		}
	}

	return nil
}
