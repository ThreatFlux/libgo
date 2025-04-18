package template

import (
	"github.com/threatflux/libgo/internal/models/vm"
)

// Manager defines interface for VM templates
type Manager interface {
	// GetTemplate gets a VM template by name
	GetTemplate(name string) (*vm.VMParams, error)

	// ListTemplates lists all available templates
	ListTemplates() ([]string, error)

	// ApplyTemplate applies a template to VM parameters
	ApplyTemplate(templateName string, params *vm.VMParams) error
}
