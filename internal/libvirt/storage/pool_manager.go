package storage

import (
	"context"
	"fmt"
	"os"
	"strings"

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
	defer func() {
		if err := m.connManager.Release(conn); err != nil {
			m.logger.Error("Failed to release connection", logger.Error(err))
		}
	}()

	libvirtConn := conn.GetLibvirtConnection()

	// Check if pool already exists
	pool, err := libvirtConn.StoragePoolLookupByName(name)
	if err == nil {
		// Pool exists, ensure it's active
		poolInfo, _, _, _, infoErr := libvirtConn.StoragePoolGetInfo(pool)
		if infoErr != nil {
			return fmt.Errorf("failed to get pool info: %w", infoErr)
		}

		if libvirt.StoragePoolState(poolInfo) != libvirt.StoragePoolRunning {
			// Pool exists but is not active, start it
			if createErr := libvirtConn.StoragePoolCreate(pool, 0); createErr != nil {
				return fmt.Errorf("failed to start storage pool %s: %w", name, createErr)
			}
		}

		m.logger.Info("Storage pool already exists",
			logger.String("name", name),
			logger.String("path", path))

		return nil
	}

	// Pool doesn't exist, create it
	// First, ensure the path exists
	if mkdirErr := os.MkdirAll(path, 0755); mkdirErr != nil {
		return fmt.Errorf("failed to create storage path %s: %w", path, mkdirErr)
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
		if undefineErr := libvirtConn.StoragePoolUndefine(pool); undefineErr != nil {
			m.logger.Error("Failed to undefine storage pool during cleanup", logger.Error(undefineErr))
		}
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
	defer func() {
		if err := m.connManager.Release(conn); err != nil {
			m.logger.Error("Failed to release connection", logger.Error(err))
		}
	}()

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
	defer func() {
		if err := m.connManager.Release(conn); err != nil {
			m.logger.Error("Failed to release connection", logger.Error(err))
		}
	}()

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

// List implements PoolManager.List
func (m *LibvirtPoolManager) List(ctx context.Context) ([]*StoragePoolInfo, error) {
	// Get libvirt connection
	conn, err := m.connManager.Connect(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to libvirt: %w", err)
	}
	defer func() {
		if err := m.connManager.Release(conn); err != nil {
			m.logger.Error("Failed to release connection", logger.Error(err))
		}
	}()

	libvirtConn := conn.GetLibvirtConnection()

	// List all storage pools
	pools, _, err := libvirtConn.ConnectListAllStoragePools(0, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to list storage pools: %w", err)
	}

	var poolInfos []*StoragePoolInfo
	for _, pool := range pools {
		poolInfo, err := m.getPoolInfo(libvirtConn, &pool)
		if err != nil {
			m.logger.Warn("Failed to get pool info",
				logger.String("pool", pool.Name),
				logger.Error(err))
			continue
		}
		poolInfos = append(poolInfos, poolInfo)
	}

	return poolInfos, nil
}

// GetInfo implements PoolManager.GetInfo
func (m *LibvirtPoolManager) GetInfo(ctx context.Context, name string) (*StoragePoolInfo, error) {
	// Get libvirt connection
	conn, err := m.connManager.Connect(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to libvirt: %w", err)
	}
	defer func() {
		if err := m.connManager.Release(conn); err != nil {
			m.logger.Error("Failed to release connection", logger.Error(err))
		}
	}()

	libvirtConn := conn.GetLibvirtConnection()

	// Get the pool
	pool, err := libvirtConn.StoragePoolLookupByName(name)
	if err != nil {
		return nil, fmt.Errorf("looking up pool %s: %w", name, ErrPoolNotFound)
	}

	return m.getPoolInfo(libvirtConn, &pool)
}

// Create implements PoolManager.Create
func (m *LibvirtPoolManager) Create(ctx context.Context, params *CreatePoolParams) (*StoragePoolInfo, error) {
	if params.Type == "" {
		params.Type = "dir"
	}

	// For directory type pools, ensure we have a path
	if params.Type == "dir" && params.Path == "" {
		return nil, fmt.Errorf("path is required for directory type pools")
	}

	// Get libvirt connection
	conn, err := m.connManager.Connect(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to libvirt: %w", err)
	}
	defer func() {
		if err := m.connManager.Release(conn); err != nil {
			m.logger.Error("Failed to release connection", logger.Error(err))
		}
	}()

	libvirtConn := conn.GetLibvirtConnection()

	// Check if pool already exists
	_, err = libvirtConn.StoragePoolLookupByName(params.Name)
	if err == nil {
		return nil, ErrPoolExists
	}

	// For directory pools, ensure the path exists
	if params.Type == "dir" {
		if mkdirErr := os.MkdirAll(params.Path, 0755); mkdirErr != nil {
			return nil, fmt.Errorf("failed to create storage path %s: %w", params.Path, mkdirErr)
		}
	}

	// Generate pool XML
	poolXML, err := m.xmlBuilder.BuildStoragePoolXML(params.Name, params.Path)
	if err != nil {
		return nil, fmt.Errorf("failed to build storage pool XML: %w", err)
	}

	// Define pool
	pool, err := libvirtConn.StoragePoolDefineXML(poolXML, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to define storage pool: %w", err)
	}

	// Build the pool (for certain types)
	if params.Type == "logical" || params.Type == "disk" {
		if err := libvirtConn.StoragePoolBuild(pool, 0); err != nil {
			m.logger.Warn("Failed to build storage pool",
				logger.String("name", params.Name),
				logger.Error(err))
		}
	}

	// Start pool
	if err := libvirtConn.StoragePoolCreate(pool, 0); err != nil {
		// Clean up if starting fails
		if undefineErr := libvirtConn.StoragePoolUndefine(pool); undefineErr != nil {
			m.logger.Error("Failed to undefine storage pool during cleanup", logger.Error(undefineErr))
		}
		return nil, fmt.Errorf("failed to start storage pool: %w", err)
	}

	// Set autostart if requested
	if params.Autostart {
		if err := libvirtConn.StoragePoolSetAutostart(pool, 1); err != nil {
			m.logger.Warn("Failed to set storage pool autostart",
				logger.String("name", params.Name),
				logger.Error(err))
		}
	}

	m.logger.Info("Created storage pool",
		logger.String("name", params.Name),
		logger.String("type", params.Type))

	return m.getPoolInfo(libvirtConn, &pool)
}

// Start implements PoolManager.Start
func (m *LibvirtPoolManager) Start(ctx context.Context, name string) error {
	// Get libvirt connection
	conn, err := m.connManager.Connect(ctx)
	if err != nil {
		return fmt.Errorf("failed to connect to libvirt: %w", err)
	}
	defer func() {
		if err := m.connManager.Release(conn); err != nil {
			m.logger.Error("Failed to release connection", logger.Error(err))
		}
	}()

	libvirtConn := conn.GetLibvirtConnection()

	// Get the pool
	pool, err := libvirtConn.StoragePoolLookupByName(name)
	if err != nil {
		return fmt.Errorf("looking up pool %s: %w", name, ErrPoolNotFound)
	}

	// Check if already running
	poolInfo, _, _, _, err := libvirtConn.StoragePoolGetInfo(pool)
	if err != nil {
		return fmt.Errorf("failed to get pool info: %w", err)
	}

	if libvirt.StoragePoolState(poolInfo) == libvirt.StoragePoolRunning {
		return nil // Already running
	}

	// Start the pool
	if err := libvirtConn.StoragePoolCreate(pool, 0); err != nil {
		return fmt.Errorf("failed to start pool %s: %w", name, err)
	}

	m.logger.Info("Started storage pool", logger.String("name", name))
	return nil
}

// Stop implements PoolManager.Stop
func (m *LibvirtPoolManager) Stop(ctx context.Context, name string) error {
	// Get libvirt connection
	conn, err := m.connManager.Connect(ctx)
	if err != nil {
		return fmt.Errorf("failed to connect to libvirt: %w", err)
	}
	defer func() {
		if err := m.connManager.Release(conn); err != nil {
			m.logger.Error("Failed to release connection", logger.Error(err))
		}
	}()

	libvirtConn := conn.GetLibvirtConnection()

	// Get the pool
	pool, err := libvirtConn.StoragePoolLookupByName(name)
	if err != nil {
		return fmt.Errorf("looking up pool %s: %w", name, ErrPoolNotFound)
	}

	// Check if already stopped
	poolInfo, _, _, _, err := libvirtConn.StoragePoolGetInfo(pool)
	if err != nil {
		return fmt.Errorf("failed to get pool info: %w", err)
	}

	if libvirt.StoragePoolState(poolInfo) != libvirt.StoragePoolRunning {
		return nil // Already stopped
	}

	// Stop the pool
	if err := libvirtConn.StoragePoolDestroy(pool); err != nil {
		return fmt.Errorf("failed to stop pool %s: %w", name, err)
	}

	m.logger.Info("Stopped storage pool", logger.String("name", name))
	return nil
}

// Refresh implements PoolManager.Refresh
func (m *LibvirtPoolManager) Refresh(ctx context.Context, name string) error {
	// Get libvirt connection
	conn, err := m.connManager.Connect(ctx)
	if err != nil {
		return fmt.Errorf("failed to connect to libvirt: %w", err)
	}
	defer func() {
		if err := m.connManager.Release(conn); err != nil {
			m.logger.Error("Failed to release connection", logger.Error(err))
		}
	}()

	libvirtConn := conn.GetLibvirtConnection()

	// Get the pool
	pool, err := libvirtConn.StoragePoolLookupByName(name)
	if err != nil {
		return fmt.Errorf("looking up pool %s: %w", name, ErrPoolNotFound)
	}

	// Refresh the pool
	if err := libvirtConn.StoragePoolRefresh(pool, 0); err != nil {
		return fmt.Errorf("failed to refresh pool %s: %w", name, err)
	}

	m.logger.Info("Refreshed storage pool", logger.String("name", name))
	return nil
}

// SetAutostart implements PoolManager.SetAutostart
func (m *LibvirtPoolManager) SetAutostart(ctx context.Context, name string, autostart bool) error {
	// Get libvirt connection
	conn, err := m.connManager.Connect(ctx)
	if err != nil {
		return fmt.Errorf("failed to connect to libvirt: %w", err)
	}
	defer func() {
		if err := m.connManager.Release(conn); err != nil {
			m.logger.Error("Failed to release connection", logger.Error(err))
		}
	}()

	libvirtConn := conn.GetLibvirtConnection()

	// Get the pool
	pool, err := libvirtConn.StoragePoolLookupByName(name)
	if err != nil {
		return fmt.Errorf("looking up pool %s: %w", name, ErrPoolNotFound)
	}

	// Set autostart
	autostartInt := int32(0)
	if autostart {
		autostartInt = 1
	}

	if err := libvirtConn.StoragePoolSetAutostart(pool, autostartInt); err != nil {
		return fmt.Errorf("failed to set autostart for pool %s: %w", name, err)
	}

	m.logger.Info("Set storage pool autostart",
		logger.String("name", name),
		logger.Bool("autostart", autostart))
	return nil
}

// IsActive implements PoolManager.IsActive
func (m *LibvirtPoolManager) IsActive(ctx context.Context, name string) (bool, error) {
	// Get libvirt connection
	conn, err := m.connManager.Connect(ctx)
	if err != nil {
		return false, fmt.Errorf("failed to connect to libvirt: %w", err)
	}
	defer func() {
		if err := m.connManager.Release(conn); err != nil {
			m.logger.Error("Failed to release connection", logger.Error(err))
		}
	}()

	libvirtConn := conn.GetLibvirtConnection()

	// Get the pool
	pool, err := libvirtConn.StoragePoolLookupByName(name)
	if err != nil {
		return false, fmt.Errorf("looking up pool %s: %w", name, ErrPoolNotFound)
	}

	// Get pool info
	poolInfo, _, _, _, err := libvirtConn.StoragePoolGetInfo(pool)
	if err != nil {
		return false, fmt.Errorf("failed to get pool info: %w", err)
	}

	return libvirt.StoragePoolState(poolInfo) == libvirt.StoragePoolRunning, nil
}

// GetXML implements PoolManager.GetXML
func (m *LibvirtPoolManager) GetXML(ctx context.Context, name string) (string, error) {
	// Get libvirt connection
	conn, err := m.connManager.Connect(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to connect to libvirt: %w", err)
	}
	defer func() {
		if err := m.connManager.Release(conn); err != nil {
			m.logger.Error("Failed to release connection", logger.Error(err))
		}
	}()

	libvirtConn := conn.GetLibvirtConnection()

	// Get the pool
	pool, err := libvirtConn.StoragePoolLookupByName(name)
	if err != nil {
		return "", fmt.Errorf("looking up pool %s: %w", name, ErrPoolNotFound)
	}

	// Get XML
	xml, err := libvirtConn.StoragePoolGetXMLDesc(pool, 0)
	if err != nil {
		return "", fmt.Errorf("failed to get pool XML: %w", err)
	}

	return xml, nil
}

// getPoolInfo is a helper method to get pool information
func (m *LibvirtPoolManager) getPoolInfo(libvirtConn *libvirt.Libvirt, pool *libvirt.StoragePool) (*StoragePoolInfo, error) {
	// Get pool info
	state, capacity, allocation, available, err := libvirtConn.StoragePoolGetInfo(*pool)
	if err != nil {
		return nil, fmt.Errorf("failed to get pool info: %w", err)
	}

	// Get autostart status
	autostart, err := libvirtConn.StoragePoolGetAutostart(*pool)
	if err != nil {
		m.logger.Warn("Failed to get pool autostart", logger.Error(err))
	}

	// Get XML to extract more details
	xml, err := libvirtConn.StoragePoolGetXMLDesc(*pool, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to get pool XML: %w", err)
	}

	// TODO: Parse XML to get pool type and path
	// For now, we'll use basic info
	poolInfo := &StoragePoolInfo{
		UUID:       fmt.Sprintf("%x", pool.UUID),
		Name:       pool.Name,
		Type:       "dir", // Default, should be parsed from XML
		State:      mapPoolState(libvirt.StoragePoolState(state)),
		Autostart:  autostart == 1,
		Persistent: true, // Assume persistent for now
		Capacity:   capacity,
		Allocation: allocation,
		Available:  available,
	}

	// Extract basic path from XML (simple regex for now)
	// TODO: Proper XML parsing
	if pathStart := strings.Index(xml, "<path>"); pathStart != -1 {
		pathEnd := strings.Index(xml[pathStart:], "</path>")
		if pathEnd != -1 {
			poolInfo.Path = xml[pathStart+6 : pathStart+pathEnd]
		}
	}

	return poolInfo, nil
}

// mapPoolState maps libvirt pool state to our StoragePoolState
func mapPoolState(state libvirt.StoragePoolState) StoragePoolState {
	switch state {
	case libvirt.StoragePoolInactive:
		return StoragePoolStateInactive
	case libvirt.StoragePoolBuilding:
		return StoragePoolStateBuilding
	case libvirt.StoragePoolRunning:
		return StoragePoolStateRunning
	case libvirt.StoragePoolDegraded:
		return StoragePoolStateDegraded
	case libvirt.StoragePoolInaccessible:
		return StoragePoolStateInaccessible
	default:
		return StoragePoolStateInactive
	}
}
