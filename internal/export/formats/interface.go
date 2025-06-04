package formats

import (
	"context"
)

// Converter defines interface for format converters.
type Converter interface {
	// Convert converts a VM disk to the target format.
	Convert(ctx context.Context, sourcePath string, destPath string, options map[string]string) error

	// GetFormatName returns the format name.
	GetFormatName() string

	// ValidateOptions validates conversion options.
	ValidateOptions(options map[string]string) error
}
