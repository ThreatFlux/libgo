package connection

import (
	"context"
	"net"
	"testing"
	"time"

	"github.com/digitalocean/go-libvirt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/threatflux/libgo/internal/config"
	"github.com/threatflux/libgo/pkg/logger"
)

// Mock logger
type mockLogger struct {
	mock.Mock
}

func (m *mockLogger) Debug(msg string, fields ...logger.Field) {
	m.Called(msg, fields)
}

func (m *mockLogger) Info(msg string, fields ...logger.Field) {
	m.Called(msg, fields)
}

func (m *mockLogger) Warn(msg string, fields ...logger.Field) {
	m.Called(msg, fields)
}

func (m *mockLogger) Error(msg string, fields ...logger.Field) {
	m.Called(msg, fields)
}

func (m *mockLogger) Fatal(msg string, fields ...logger.Field) {
	m.Called(msg, fields)
}

func (m *mockLogger) WithFields(fields ...logger.Field) logger.Logger {
	args := m.Called(fields)
	return args.Get(0).(logger.Logger)
}

func (m *mockLogger) WithError(err error) logger.Logger {
	args := m.Called(err)
	return args.Get(0).(logger.Logger)
}

func (m *mockLogger) Sync() error {
	args := m.Called()
	return args.Error(0)
}

// Mock libvirt and connection
type mockLibvirt struct {
	mock.Mock
}

func (m *mockLibvirt) Connect() error {
	args := m.Called()
	return args.Error(0)
}

func (m *mockLibvirt) Disconnect() error {
	args := m.Called()
	return args.Error(0)
}

type mockConn struct {
	mock.Mock
}

func (m *mockConn) Close() error {
	args := m.Called()
	return args.Error(0)
}

func (m *mockConn) Read(b []byte) (n int, err error) {
	args := m.Called(b)
	return args.Int(0), args.Error(1)
}

func (m *mockConn) Write(b []byte) (n int, err error) {
	args := m.Called(b)
	return args.Int(0), args.Error(1)
}

func (m *mockConn) LocalAddr() net.Addr {
	args := m.Called()
	return args.Get(0).(net.Addr)
}

func (m *mockConn) RemoteAddr() net.Addr {
	args := m.Called()
	return args.Get(0).(net.Addr)
}

func (m *mockConn) SetDeadline(t time.Time) error {
	args := m.Called(t)
	return args.Error(0)
}

func (m *mockConn) SetReadDeadline(t time.Time) error {
	args := m.Called(t)
	return args.Error(0)
}

func (m *mockConn) SetWriteDeadline(t time.Time) error {
	args := m.Called(t)
	return args.Error(0)
}

// invalidConn is a test type that doesn't implement Connection correctly
type invalidConn struct{}

func (c *invalidConn) GetLibvirtConnection() *libvirt.Libvirt { return nil }
func (c *invalidConn) Close() error                           { return nil }
func (c *invalidConn) IsActive() bool                         { return true }

// Tests
func TestNewConnectionManager(t *testing.T) {
	mockLog := new(mockLogger)
	mockLog.On("Debug", mock.Anything, mock.Anything).Return()
	mockLog.On("Info", mock.Anything, mock.Anything).Return()
	mockLog.On("Warn", mock.Anything, mock.Anything).Return()
	mockLog.On("Error", mock.Anything, mock.Anything).Return()

	cfg := config.LibvirtConfig{
		URI:               "/var/run/libvirt/libvirt-sock",
		ConnectionTimeout: 5 * time.Second,
		MaxConnections:    3,
	}

	manager, err := NewConnectionManager(cfg, mockLog)
	assert.NoError(t, err)
	assert.NotNil(t, manager)
	assert.Equal(t, cfg.URI, manager.uri)
	assert.Equal(t, cfg.MaxConnections, manager.maxConnections)
	assert.Equal(t, cfg.ConnectionTimeout, manager.timeout)

	// Test default max connections
	cfg.MaxConnections = 0
	manager, err = NewConnectionManager(cfg, mockLog)
	assert.NoError(t, err)
	assert.Equal(t, 5, manager.maxConnections) // Default value
}

func TestConnectionManager_Connect_MockedLibvirt(t *testing.T) {
	// This is a partial test since we can't easily mock the actual libvirt connection
	// In a real environment, this would require integration tests with a real libvirt daemon

	// Mock the logger
	mockLog := new(mockLogger)
	mockLog.On("Debug", mock.Anything, mock.Anything).Return()

	// Create the manager
	cfg := config.LibvirtConfig{
		URI:               "/var/run/libvirt/libvirt-sock",
		ConnectionTimeout: 100 * time.Millisecond,
		MaxConnections:    2,
	}

	manager := &ConnectionManager{
		uri:            cfg.URI,
		connPool:       make(chan *libvirtConnection, cfg.MaxConnections),
		maxConnections: cfg.MaxConnections,
		timeout:        cfg.ConnectionTimeout,
		logger:         mockLog,
	}

	// Create a mocked connection and add it to the pool
	mockLibvirtClient := &mockLibvirt{}
	mockNetConn := &mockConn{}

	libvirtConn := &libvirtConnection{
		libvirt: &libvirt.Libvirt{},
		conn:    mockNetConn,
		active:  true,
		manager: manager,
	}

	// Add to pool
	manager.connPool <- libvirtConn

	// Test getting from pool
	ctx := context.Background()
	conn, err := manager.Connect(ctx)
	assert.NoError(t, err)
	assert.NotNil(t, conn)
	assert.True(t, conn.IsActive())

	// Test release back to pool
	mockNetConn.On("Close").Return(nil)
	mockLibvirtClient.On("Disconnect").Return(nil)

	err = manager.Release(conn)
	assert.NoError(t, err)

	// Verify we can get it again from the pool
	conn, err = manager.Connect(ctx)
	assert.NoError(t, err)
	assert.NotNil(t, conn)
}

func TestConnectionManager_Close(t *testing.T) {
	// Mock the logger
	mockLog := new(mockLogger)
	mockLog.On("Debug", mock.Anything, mock.Anything).Return()
	mockLog.On("Warn", mock.Anything, mock.Anything).Return()
	mockLog.On("Error", mock.Anything, mock.Anything).Return()

	// Create the manager
	cfg := config.LibvirtConfig{
		URI:               "/var/run/libvirt/libvirt-sock",
		ConnectionTimeout: 100 * time.Millisecond,
		MaxConnections:    2,
	}

	manager := &ConnectionManager{
		uri:            cfg.URI,
		connPool:       make(chan *libvirtConnection, cfg.MaxConnections),
		maxConnections: cfg.MaxConnections,
		timeout:        cfg.ConnectionTimeout,
		logger:         mockLog,
	}

	// Create test connections and add them to the pool
	// For this test, we don't need actual libvirt connections since we're just testing pool closure
	for i := 0; i < cfg.MaxConnections; i++ {
		// We'll use TestLibvirtConnection from test_helpers.go
		// (testConn variable removed since it's not used)

		// Since we need to add libvirtConnection to the channel, we'll create empty ones
		// The Close() method will need to handle nil libvirt gracefully
		libvirtConn := &libvirtConnection{
			libvirt: nil,
			conn:    nil,
			active:  false, // Mark as inactive so Close() returns early
			manager: manager,
		}

		// Add to pool
		manager.connPool <- libvirtConn
	}

	// Close all connections
	err := manager.Close()
	assert.NoError(t, err)

	// Verify pool is empty
	select {
	case <-manager.connPool:
		t.Fatal("Expected pool to be empty")
	default:
		// Pool is empty as expected
	}
}

func TestLibvirtConnection_Close(t *testing.T) {
	// Mock the logger
	mockLog := new(mockLogger)
	mockLog.On("Debug", mock.Anything, mock.Anything).Return()
	mockLog.On("Warn", mock.Anything, mock.Anything).Return()

	// Create the manager
	manager := &ConnectionManager{
		uri:            "/var/run/libvirt/libvirt-sock",
		connPool:       make(chan *libvirtConnection, 1),
		maxConnections: 1,
		timeout:        100 * time.Millisecond,
		logger:         mockLog,
	}

	// Create test connection
	// Mark as inactive so Close() returns early without calling libvirt.Disconnect()
	conn := &libvirtConnection{
		libvirt: nil,
		conn:    nil,
		active:  false, // This ensures Close() returns early
		manager: manager,
	}

	// Test close
	err := conn.Close()
	assert.NoError(t, err)
	assert.False(t, conn.active)

	// Test closing an already closed connection
	err = conn.Close()
	assert.NoError(t, err) // Should not error when closing again

	// No expectations to verify since we're using a simple inactive connection
}

func TestConnectionManager_ReleaseInvalidConnection(t *testing.T) {
	// Mock the logger
	mockLog := new(mockLogger)
	mockLog.On("Debug", mock.Anything, mock.Anything).Return()

	// Create the manager
	manager := &ConnectionManager{
		uri:            "/var/run/libvirt/libvirt-sock",
		connPool:       make(chan *libvirtConnection, 1),
		maxConnections: 1,
		timeout:        100 * time.Millisecond,
		logger:         mockLog,
	}

	// Try to release an invalid connection type
	err := manager.Release(&invalidConn{})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid connection type")
}
