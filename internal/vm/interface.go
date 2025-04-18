package vm

import (
	"context"

	"github.com/wroersma/libgo/internal/models/vm"
)

// Manager defines the interface for VM management
type Manager interface {
	// Create creates a new VM
	Create(ctx context.Context, params vm.VMParams) (*vm.VM, error)
	
	// Get gets a VM by name
	Get(ctx context.Context, name string) (*vm.VM, error)
	
	// List lists all VMs
	List(ctx context.Context) ([]*vm.VM, error)
	
	// Delete deletes a VM
	Delete(ctx context.Context, name string) error
	
	// Start starts a VM
	Start(ctx context.Context, name string) error
	
	// Stop stops a VM
	Stop(ctx context.Context, name string) error
	
	// Restart restarts a VM
	Restart(ctx context.Context, name string) error
}