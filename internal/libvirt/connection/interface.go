package connection

import (
	"context"

	"github.com/digitalocean/go-libvirt"
)

// Manager defines the interface for managing libvirt connections.
type Manager interface {
	// Connect establishes a connection to libvirt
	Connect(ctx context.Context) (Connection, error)

	// Release returns a connection to the pool
	Release(conn Connection) error

	// Close closes all connections in the pool
	Close() error
}

// Connection defines interface for a libvirt connection.
type Connection interface {
	// GetLibvirtConnection returns the underlying libvirt connection
	GetLibvirtConnection() *libvirt.Libvirt

	// Close closes the connection
	Close() error

	// IsActive checks if connection is active
	IsActive() bool
}
