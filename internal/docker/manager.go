package docker

import (
	"context"
	"sync"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
)

// Manager is the interface for Docker client operations.
type Manager interface {
	// GetClient returns a thread-safe Docker API client wrapper
	GetClient() (client.APIClient, error)

	// GetWithContext returns a thread-safe Docker API client wrapper with the specified context
	GetWithContext(ctx context.Context) (client.APIClient, error)

	// Ping checks the connectivity with the Docker daemon
	Ping(ctx context.Context) (types.Ping, error)

	// Close closes all clients and releases resources
	Close() error

	// IsInitialized checks if the client is initialized
	IsInitialized() bool

	// IsClosed checks if the client is closed
	IsClosed() bool

	// GetConfig returns the client configuration
	GetConfig() ClientConfig
}

// Global default manager instance.
var (
	defaultManager     Manager
	// defaultManagerOnce sync.Once // Unused - reserved for future singleton pattern
	defaultManagerMu   sync.RWMutex
)

// DefaultManager returns the singleton manager instance.
func DefaultManager() Manager {
	defaultManagerMu.RLock()
	defer defaultManagerMu.RUnlock()
	return defaultManager
}

// SetDefaultManager sets the default manager instance.
func SetDefaultManager(m Manager) {
	defaultManagerMu.Lock()
	defer defaultManagerMu.Unlock()
	defaultManager = m
}

// GetDefaultClient returns a client from the default manager.
func GetDefaultClient() (client.APIClient, error) {
	m := DefaultManager()
	if m == nil {
		return nil, ErrClientNotInitialized
	}
	return m.GetClient()
}

// GetDefaultClientWithContext returns a client from the default manager with context.
func GetDefaultClientWithContext(ctx context.Context) (client.APIClient, error) {
	m := DefaultManager()
	if m == nil {
		return nil, ErrClientNotInitialized
	}
	return m.GetWithContext(ctx)
}
