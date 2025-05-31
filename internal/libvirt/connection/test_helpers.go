package connection

import (
	"net"
	"time"

	"github.com/digitalocean/go-libvirt"
	"github.com/threatflux/libgo/pkg/logger"
)

// TestConnectionManager is a special variant of ConnectionManager for testing
// that allows using Connection interface in the channel
type TestConnectionManager struct {
	ConnectionManager
	testConnPool chan Connection
}

// NewTestConnectionManager creates a ConnectionManager suitable for testing
func NewTestConnectionManager(uri string, maxConnections int, timeoutDuration time.Duration, log logger.Logger) *TestConnectionManager {
	mgr := &TestConnectionManager{
		ConnectionManager: ConnectionManager{
			uri:            uri,
			maxConnections: maxConnections,
			timeout:        timeoutDuration,
			logger:         log,
			connPool:       make(chan *libvirtConnection, maxConnections),
		},
		testConnPool: make(chan Connection, maxConnections),
	}
	return mgr
}

// AddToTestPool adds a test connection to the test pool
func (m *TestConnectionManager) AddToTestPool(conn Connection) {
	m.testConnPool <- conn
}

// TestLibvirtConnection is a test implementation of the Connection interface
type TestLibvirtConnection struct {
	libvirt *libvirt.Libvirt
	conn    net.Conn
	active  bool
	manager *ConnectionManager
}

// GetLibvirtConnection implements Connection.GetLibvirtConnection
func (c *TestLibvirtConnection) GetLibvirtConnection() *libvirt.Libvirt {
	return c.libvirt
}

// Close implements Connection.Close
func (c *TestLibvirtConnection) Close() error {
	c.active = false
	// In a real system we would close, but for tests we don't need to
	return nil
}

// IsActive implements Connection.IsActive
func (c *TestLibvirtConnection) IsActive() bool {
	return c.active
}

// ToLibvirtConnection attempts to convert a Connection to a *libvirtConnection
func ToLibvirtConnection(conn Connection) *libvirtConnection {
	if lc, ok := conn.(*libvirtConnection); ok {
		return lc
	}
	// For tests, return nil
	return nil
}
