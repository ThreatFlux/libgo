package domain

import (
	"context"

	"github.com/threatflux/libgo/internal/models/vm"
)

// Manager defines the interface for managing libvirt domains
type Manager interface {
	// Create creates a new domain (VM)
	Create(ctx context.Context, params vm.VMParams) (*vm.VM, error)

	// Get retrieves information about a domain
	Get(ctx context.Context, name string) (*vm.VM, error)

	// List lists all domains
	List(ctx context.Context) ([]*vm.VM, error)

	// Start starts a domain
	Start(ctx context.Context, name string) error

	// Stop stops a domain (graceful shutdown)
	Stop(ctx context.Context, name string) error

	// ForceStop forces a domain to stop
	ForceStop(ctx context.Context, name string) error

	// Delete deletes a domain
	Delete(ctx context.Context, name string) error

	// GetXML gets the XML configuration of a domain
	GetXML(ctx context.Context, name string) (string, error)
}

// XMLBuilder defines interface for building domain XML
type XMLBuilder interface {
	// BuildDomainXML builds XML for domain creation
	BuildDomainXML(params vm.VMParams) (string, error)
}
