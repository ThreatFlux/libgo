package vm

import (
	"context"

	"github.com/threatflux/libgo/internal/models/vm"
)

// Manager defines the interface for VM management.
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

	// Snapshot operations
	// CreateSnapshot creates a new snapshot of a VM
	CreateSnapshot(ctx context.Context, vmName string, params vm.SnapshotParams) (*vm.Snapshot, error)

	// ListSnapshots lists all snapshots for a VM
	ListSnapshots(ctx context.Context, vmName string, opts vm.SnapshotListOptions) ([]*vm.Snapshot, error)

	// GetSnapshot retrieves information about a specific snapshot
	GetSnapshot(ctx context.Context, vmName string, snapshotName string) (*vm.Snapshot, error)

	// DeleteSnapshot deletes a snapshot
	DeleteSnapshot(ctx context.Context, vmName string, snapshotName string) error

	// RevertSnapshot reverts a VM to a snapshot
	RevertSnapshot(ctx context.Context, vmName string, snapshotName string) error
}
