package raw

import (
	"context"
	"fmt"

	"github.com/threatflux/libgo/pkg/logger"
	"github.com/threatflux/libgo/pkg/utils/exec"
)

// RAWConverter implements Converter for raw format
type RAWConverter struct {
	logger logger.Logger
}

// NewRAWConverter creates a new RAWConverter
func NewRAWConverter(logger logger.Logger) *RAWConverter {
	return &RAWConverter{
		logger: logger,
	}
}

// Convert implements Converter.Convert
func (c *RAWConverter) Convert(ctx context.Context, sourcePath string, destPath string, options map[string]string) error {
	// Prepare qemu-img arguments
	args := []string{
		"convert",
		"-f", "qcow2", // Source format, assume qcow2 as libvirt generally uses this
		"-O", "raw", // Output format - raw
		sourcePath,
		destPath,
	}

	c.logger.Info("Converting to RAW format",
		logger.String("source", sourcePath),
		logger.String("destination", destPath))

	// Execute qemu-img command
	execOpts := exec.CommandOptions{
		Timeout: 0, // No timeout, conversion might take a long time
	}

	output, err := exec.ExecuteCommand(ctx, "qemu-img", args, execOpts)
	if err != nil {
		c.logger.Error("Failed to convert to RAW",
			logger.String("error", err.Error()),
			logger.String("output", string(output)))
		return fmt.Errorf("raw conversion failed: %w", err)
	}

	c.logger.Info("Successfully converted to RAW format",
		logger.String("destination", destPath))
	return nil
}

// GetFormatName implements Converter.GetFormatName
func (c *RAWConverter) GetFormatName() string {
	return "raw"
}

// ValidateOptions implements Converter.ValidateOptions
func (c *RAWConverter) ValidateOptions(options map[string]string) error {
	// No specific options for RAW format
	return nil
}
