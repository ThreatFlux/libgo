package image

import (
	"context"
	"testing"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/threatflux/libgo/internal/config"
	"github.com/threatflux/libgo/internal/docker"
	"github.com/threatflux/libgo/pkg/logger"
)

// MockDockerManager implements docker.Manager for testing
type MockDockerManager struct {
	mock.Mock
}

func (m *MockDockerManager) GetClient() (client.APIClient, error) {
	args := m.Called()
	return args.Get(0).(client.APIClient), args.Error(1)
}

func (m *MockDockerManager) GetWithContext(ctx context.Context) (client.APIClient, error) {
	args := m.Called(ctx)
	return args.Get(0).(client.APIClient), args.Error(1)
}

func (m *MockDockerManager) Ping(ctx context.Context) (types.Ping, error) {
	args := m.Called(ctx)
	return args.Get(0).(types.Ping), args.Error(1)
}

func (m *MockDockerManager) Close() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockDockerManager) IsInitialized() bool {
	args := m.Called()
	return args.Bool(0)
}

func (m *MockDockerManager) IsClosed() bool {
	args := m.Called()
	return args.Bool(0)
}

func (m *MockDockerManager) GetConfig() docker.ClientConfig {
	args := m.Called()
	return args.Get(0).(docker.ClientConfig)
}

// For simplicity, we'll test the service with a nil client and focus on the manager interface
// In a real test suite, you would use a more complete mock or testcontainers

func TestImageService_NewService(t *testing.T) {
	mockManager := new(MockDockerManager)
	mockLogger, _ := logger.NewZapLogger(config.LoggingConfig{Level: "debug", Format: "json"})

	service := NewService(mockManager, mockLogger)

	assert.NotNil(t, service)
	assert.IsType(t, &serviceImpl{}, service)
}

// Additional tests would require a complete mock implementation of client.APIClient
// For now, we focus on testing that the service can be created successfully
