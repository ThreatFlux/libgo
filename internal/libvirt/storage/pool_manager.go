package storage

import (
	"context"
	"fmt"
	"os"

	"github.com/digitalocean/go-libvirt"
	"github.com/threatflux/libgo/internal/libvirt/connection"
	"github.com/threatflux/libgo/pkg/logger"
)

// Error types
var (
	ErrPoolNotFound = fmt.Errorf("storage pool not found")
	ErrPoolExists   = fmt.Errorf("storage pool already exists")
)

// LibvirtPoolManager implements PoolManager for libvirt
type LibvirtPoolManager struct {
	connManager connection.Manager
	xmlBuilder  XMLBuilder
	logger      logger.Logger
}

// NewLibvirtPoolManager creates a new LibvirtPoolManager
func NewLibvirtPoolManager(connManager connection.Manager, xmlBuilder XMLBuilder, logger logger.Logger) *LibvirtPoolManager {
	return &LibvirtPoolManager{
		connManager: connManager,
		xmlBuilder:  xmlBuilder,
		logger:      logger,
	}
}

// EnsureExists implements PoolManager.EnsureExists
func (m *LibvirtPoolManager) EnsureExists(ctx context.Context, name string, path string) error {
	// Get libvirt connection
	conn, err := m.connManager.Connect(ctx)
	if err != nil {
		return fmt.Errorf("failed to connect to libvirt: %w", err)
	}
	defer m.connManager.Release(conn)

	libvirtConn := conn.GetLibvirtConnection()

	// Check if pool already exists
	pool, err := libvirtConn.StoragePoolLookupByName(name)
	if err == nil {
		// Pool exists, ensure it's active
		poolInfo, _, _, _, err := libvirtConn.StoragePoolGetInfo(pool)
		if err != nil {
			return fmt.Errorf("failed to get pool info: %w", err)
		}

		if libvirt.StoragePoolState(poolInfo) != libvirt.StoragePoolRunning {
			// Pool exists but is not active, start it
			if err := libvirtConn.StoragePoolCreate(pool, 0); err != nil {
				return fmt.Errorf("failed to start storage pool %s: %w", name, err)
			}
		}

		m.logger.Info("Storage pool already exists",
			logger.String("name", name),
			logger.String("path", path))

		return nil
	}

	// Pool doesn't exist, create it
	// First, ensure the path exists
	if err := os.MkdirAll(path, 0755); err != nil {
		return fmt.Errorf("failed to create storage path %s: %w", path, err)
	}

	// Generate pool XML
	poolXML, err := m.xmlBuilder.BuildStoragePoolXML(name, path)
	if err != nil {
		return fmt.Errorf("failed to build storage pool XML: %w", err)
	}

	// Define pool
	pool, err = libvirtConn.StoragePoolDefineXML(poolXML, 0)
	if err != nil {
		return fmt.Errorf("failed to define storage pool: %w", err)
	}

	// Start pool
	if err := libvirtConn.StoragePoolCreate(pool, 0); err != nil {
		// Clean up if starting fails
		_ = libvirtConn.StoragePoolUndefine(pool)
		return fmt.Errorf("failed to start storage pool: %w", err)
	}

	// Set pool to autostart
	if err := libvirtConn.StoragePoolSetAutostart(pool, 1); err != nil {
		m.logger.Warn("Failed to set storage pool autostart",
			logger.String("name", name),
			logger.Error(err))
	}

	m.logger.Info("Created storage pool",
		logger.String("name", name),
		logger.String("path", path))

	return nil
}

// Delete implements PoolManager.Delete
func (m *LibvirtPoolManager) Delete(ctx context.Context, name string) error {
	// Get libvirt connection
	conn, err := m.connManager.Connect(ctx)
	if err != nil {
		return fmt.Errorf("failed to connect to libvirt: %w", err)
	}
	defer m.connManager.Release(conn)

	libvirtConn := conn.GetLibvirtConnection()

	// Check if pool exists
	pool, err := libvirtConn.StoragePoolLookupByName(name)
	if err != nil {
		return fmt.Errorf("looking up pool %s: %w", name, ErrPoolNotFound)
	}

	// Check if pool is active
	poolInfo, _, _, _, err := libvirtConn.StoragePoolGetInfo(pool)
	if err != nil {
		return fmt.Errorf("failed to get pool info: %w", err)
	}

	// Stop pool if it's running
	if libvirt.StoragePoolState(poolInfo) == libvirt.StoragePoolRunning {
		if err := libvirtConn.StoragePoolDestroy(pool); err != nil {
			return fmt.Errorf("failed to stop storage pool: %w", err)
		}
	}

	// Undefine pool
	if err := libvirtConn.StoragePoolUndefine(pool); err != nil {
		return fmt.Errorf("failed to undefine storage pool: %w", err)
	}

	m.logger.Info("Deleted storage pool", logger.String("name", name))
	return nil
}

// Get implements PoolManager.Get
func (m *LibvirtPoolManager) Get(ctx context.Context, name string) (*libvirt.StoragePool, error) {
	// Get libvirt connection
	conn, err := m.connManager.Connect(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to libvirt: %w", err)
	}
	defer m.connManager.Release(conn)

	libvirtConn := conn.GetLibvirtConnection()

	// Get the pool by name
	pool, err := libvirtConn.StoragePoolLookupByName(name)
	if err != nil {
		// If the specified pool is not found and it's not already "default",
		// try to fall back to the default pool
		if name != "default" {
			m.logger.Warn("Requested storage pool not found, falling back to default pool",
				logger.String("requested_pool", name))

			// Try to get the default pool instead
			defaultPool, defaultErr := libvirtConn.StoragePoolLookupByName("default")
			if defaultErr != nil {
				return nil, fmt.Errorf("looking up pool %s (and default fallback): %w", name, ErrPoolNotFound)
			}

			return &defaultPool, nil
		}

		return nil, fmt.Errorf("looking up pool %s: %w", name, ErrPoolNotFound)
	}

	return &pool, nil
}
