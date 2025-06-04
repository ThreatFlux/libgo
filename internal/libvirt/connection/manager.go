package connection

import (
	"context"
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/digitalocean/go-libvirt"
	"github.com/threatflux/libgo/internal/config"
	"github.com/threatflux/libgo/pkg/logger"
)

// ConnectionManager implements Manager for libvirt connections.
type ConnectionManager struct {
	// Mutex (sync.Mutex is typically 8 bytes)
	mu sync.Mutex
	// Channel (24 bytes)
	connPool chan *libvirtConnection
	// Interface fields (16 bytes each)
	logger logger.Logger
	// Duration (8 bytes)
	timeout time.Duration
	// String (16 bytes)
	uri string
	// Int (8 bytes on 64-bit)
	maxConnections int
}

// libvirtConnection implements Connection interface.
type libvirtConnection struct {
	// Pointer fields (8 bytes each)
	libvirt *libvirt.Libvirt
	manager *ConnectionManager
	// Interface fields (16 bytes)
	conn net.Conn
	// Bool (1 byte, but aligned to 8 bytes at the end)
	active bool
}

// NewConnectionManager creates a new ConnectionManager.
func NewConnectionManager(cfg config.LibvirtConfig, logger logger.Logger) (*ConnectionManager, error) {
	if cfg.MaxConnections <= 0 {
		cfg.MaxConnections = 5 // Default value if not configured
	}

	manager := &ConnectionManager{
		uri:            cfg.URI,
		connPool:       make(chan *libvirtConnection, cfg.MaxConnections),
		maxConnections: cfg.MaxConnections,
		timeout:        cfg.ConnectionTimeout,
		logger:         logger,
	}

	return manager, nil
}

// Connect implements Manager.Connect.
func (m *ConnectionManager) Connect(ctx context.Context) (Connection, error) {
	// Try to get connection from the pool
	select {
	case conn := <-m.connPool:
		// Check if the connection is still active
		if conn.active {
			m.logger.Debug("Reusing existing libvirt connection from pool",
				logger.String("uri", m.uri))
			return conn, nil
		}

		// If connection is not active, close it and create a new one
		m.logger.Debug("Found inactive connection in pool, creating a new one",
			logger.String("uri", m.uri))
		_ = conn.Close()
	default:
		// Pool is empty, continue to create a new connection
	}

	// Note: timeout is handled via net.DialTimeout below
	// Context is preserved for future use if needed
	_ = ctx

	// Create a new connection
	m.logger.Debug("Creating new libvirt connection",
		logger.String("uri", m.uri))

	// Parse the URI to determine connection type and connection path
	var networkType, socketPath string

	// Handle test:///default URI for testing
	if m.uri == "test:///default" {
		// For test driver, use a mock connection
		// This is a special case for testing without actual libvirt
		m.logger.Info("Using test libvirt driver")

		// Create a pipe for the test driver
		clientConn, serverConn := net.Pipe()

		go func() {
			// Simple mock server that reads and discards data
			buffer := make([]byte, 1024)
			for {
				_, err := serverConn.Read(buffer)
				if err != nil {
					_ = serverConn.Close()
					return
				}
			}
		}()

		c := clientConn
		dialer := &connDialer{conn: c}
		l := libvirt.NewWithDialer(dialer)

		// Create a mock connection
		libvirtConn := &libvirtConnection{
			libvirt: l,
			conn:    c,
			active:  true,
			manager: m,
		}

		return libvirtConn, nil
	}

	// Normal libvirt URI parsing
	switch {
	case m.uri == "qemu:///system":
		networkType = "unix"
		socketPath = "/var/run/libvirt/libvirt-sock"
	case m.uri == "qemu:///session":
		networkType = "unix"
		socketPath = "/run/user/1000/libvirt/libvirt-sock"
	default:
		// Try to extract tcp or unix connection details from URI
		return nil, fmt.Errorf("unsupported libvirt URI format: %s", m.uri)
	}

	c, err := net.DialTimeout(networkType, socketPath, m.timeout)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to libvirt at %s: %w", socketPath, err)
	}

	dialer := &connDialer{conn: c}
	l := libvirt.NewWithDialer(dialer)
	if err := l.Connect(); err != nil {
		c.Close()
		return nil, fmt.Errorf("failed to establish libvirt connection: %w", err)
	}

	libvirtConn := &libvirtConnection{
		libvirt: l,
		conn:    c,
		active:  true,
		manager: m,
	}

	return libvirtConn, nil
}

// Release implements Manager.Release.
func (m *ConnectionManager) Release(conn Connection) error {
	libvirtConn, ok := conn.(*libvirtConnection)
	if !ok {
		return fmt.Errorf("invalid connection type")
	}

	// If connection is not active, close it instead of returning to pool
	if !libvirtConn.active {
		return libvirtConn.Close()
	}

	// Try to return to pool, or close if pool is full
	select {
	case m.connPool <- libvirtConn:
		m.logger.Debug("Released connection back to pool",
			logger.String("uri", m.uri))
		return nil
	default:
		// Pool is full, close the connection
		m.logger.Debug("Connection pool full, closing connection",
			logger.String("uri", m.uri))
		return libvirtConn.Close()
	}
}

// Close implements Manager.Close.
func (m *ConnectionManager) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	var lastErr error

	// Close all connections in the pool
	for {
		select {
		case conn := <-m.connPool:
			if err := conn.Close(); err != nil {
				lastErr = err
				m.logger.Error("Error closing libvirt connection",
					logger.String("uri", m.uri),
					logger.Error(err))
			}
		default:
			// Pool is empty
			return lastErr
		}
	}
}

// GetLibvirtConnection implements Connection.GetLibvirtConnection.
func (c *libvirtConnection) GetLibvirtConnection() *libvirt.Libvirt {
	return c.libvirt
}

// Close implements Connection.Close.
func (c *libvirtConnection) Close() error {
	if !c.active {
		return nil // Already closed
	}

	c.active = false

	// Close the connection
	if err := c.libvirt.Disconnect(); err != nil {
		c.manager.logger.Warn("Error disconnecting from libvirt",
			logger.Error(err))
	}

	if err := c.conn.Close(); err != nil {
		c.manager.logger.Warn("Error closing libvirt connection",
			logger.Error(err))
		return err
	}

	return nil
}

// IsActive implements Connection.IsActive.
func (c *libvirtConnection) IsActive() bool {
	return c.active
}

// IsConnected implements Connection.IsConnected.
func (c *libvirtConnection) IsConnected() bool {
	return c.active
}

// connDialer implements socket.Dialer interface for libvirt connections.
type connDialer struct {
	conn net.Conn
}

// Dial implements socket.Dialer.
func (d *connDialer) Dial() (net.Conn, error) {
	return d.conn, nil
}
