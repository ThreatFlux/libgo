package container

import (
	"context"
	"errors"
	"io"
	"strings"
	"testing"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/threatflux/libgo/internal/docker"
	"github.com/threatflux/libgo/pkg/logger"
	gomock "go.uber.org/mock/gomock"
)

// MockManager is a mock implementation of docker.Manager
type MockManager struct {
	mock.Mock
}

func (m *MockManager) GetClient() (client.APIClient, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(client.APIClient), args.Error(1)
}

func (m *MockManager) GetWithContext(ctx context.Context) (client.APIClient, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(client.APIClient), args.Error(1)
}

func (m *MockManager) Ping(ctx context.Context) (types.Ping, error) {
	args := m.Called(ctx)
	return args.Get(0).(types.Ping), args.Error(1)
}

func (m *MockManager) Close() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockManager) IsInitialized() bool {
	args := m.Called()
	return args.Bool(0)
}

func (m *MockManager) IsClosed() bool {
	args := m.Called()
	return args.Bool(0)
}

func (m *MockManager) GetConfig() docker.ClientConfig {
	args := m.Called()
	return args.Get(0).(docker.ClientConfig)
}

// MockAPIClient is a mock implementation of client.APIClient
type MockAPIClient struct {
	mock.Mock
}

// Implement all required methods for client.APIClient interface
func (m *MockAPIClient) ContainerCreate(ctx context.Context, config *container.Config, hostConfig *container.HostConfig, networkingConfig *network.NetworkingConfig, platform *ocispec.Platform, containerName string) (container.CreateResponse, error) {
	args := m.Called(ctx, config, hostConfig, networkingConfig, platform, containerName)
	return args.Get(0).(container.CreateResponse), args.Error(1)
}

func (m *MockAPIClient) ContainerStart(ctx context.Context, containerID string, options container.StartOptions) error {
	args := m.Called(ctx, containerID, options)
	return args.Error(0)
}

func (m *MockAPIClient) ContainerStop(ctx context.Context, containerID string, options container.StopOptions) error {
	args := m.Called(ctx, containerID, options)
	return args.Error(0)
}

func (m *MockAPIClient) ContainerRestart(ctx context.Context, containerID string, options container.StopOptions) error {
	args := m.Called(ctx, containerID, options)
	return args.Error(0)
}

func (m *MockAPIClient) ContainerKill(ctx context.Context, containerID string, signal string) error {
	args := m.Called(ctx, containerID, signal)
	return args.Error(0)
}

func (m *MockAPIClient) ContainerRemove(ctx context.Context, containerID string, options container.RemoveOptions) error {
	args := m.Called(ctx, containerID, options)
	return args.Error(0)
}

func (m *MockAPIClient) ContainerPause(ctx context.Context, containerID string) error {
	args := m.Called(ctx, containerID)
	return args.Error(0)
}

func (m *MockAPIClient) ContainerUnpause(ctx context.Context, containerID string) error {
	args := m.Called(ctx, containerID)
	return args.Error(0)
}

func (m *MockAPIClient) ContainerInspect(ctx context.Context, containerID string) (types.ContainerJSON, error) {
	args := m.Called(ctx, containerID)
	return args.Get(0).(types.ContainerJSON), args.Error(1)
}

func (m *MockAPIClient) ContainerList(ctx context.Context, options container.ListOptions) ([]types.Container, error) {
	args := m.Called(ctx, options)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]types.Container), args.Error(1)
}

func (m *MockAPIClient) ContainerLogs(ctx context.Context, containerID string, options container.LogsOptions) (io.ReadCloser, error) {
	args := m.Called(ctx, containerID, options)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(io.ReadCloser), args.Error(1)
}

func (m *MockAPIClient) ContainerStats(ctx context.Context, containerID string, stream bool) (container.StatsResponseReader, error) {
	args := m.Called(ctx, containerID, stream)
	return args.Get(0).(container.StatsResponseReader), args.Error(1)
}

// Mock logger implementation
type mockLogger struct{}

func (m *mockLogger) Debug(msg string, fields ...logger.Field)        {}
func (m *mockLogger) Info(msg string, fields ...logger.Field)         {}
func (m *mockLogger) Warn(msg string, fields ...logger.Field)         {}
func (m *mockLogger) Error(msg string, fields ...logger.Field)        {}
func (m *mockLogger) Fatal(msg string, fields ...logger.Field)        {}
func (m *mockLogger) WithFields(fields ...logger.Field) logger.Logger { return m }
func (m *mockLogger) WithError(err error) logger.Logger               { return m }
func (m *mockLogger) Sync() error                                     { return nil }

func TestContainerService_Create(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockManager := new(MockManager)
	mockClient := new(MockAPIClient)
	mockLog := &mockLogger{}

	service := NewService(mockManager, mockLog)

	testCases := []struct {
		name          string
		config        *container.Config
		hostConfig    *container.HostConfig
		containerName string
		setupMocks    func()
		expectedID    string
		expectedError bool
	}{
		{
			name: "successful creation",
			config: &container.Config{
				Image: "nginx:latest",
			},
			hostConfig:    &container.HostConfig{},
			containerName: "test-container",
			setupMocks: func() {
				mockManager.On("GetWithContext", mock.Anything).Return(mockClient, nil)
				mockClient.On("ContainerCreate", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, "test-container").
					Return(container.CreateResponse{ID: "abc123", Warnings: []string{"test warning"}}, nil)
			},
			expectedID:    "abc123",
			expectedError: false,
		},
		{
			name: "manager error",
			config: &container.Config{
				Image: "nginx:latest",
			},
			hostConfig:    &container.HostConfig{},
			containerName: "test-container",
			setupMocks: func() {
				mockManager.On("GetWithContext", mock.Anything).Return(nil, errors.New("manager error"))
			},
			expectedID:    "",
			expectedError: true,
		},
		{
			name: "creation error",
			config: &container.Config{
				Image: "nginx:latest",
			},
			hostConfig:    &container.HostConfig{},
			containerName: "test-container",
			setupMocks: func() {
				mockManager.On("GetWithContext", mock.Anything).Return(mockClient, nil)
				mockClient.On("ContainerCreate", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, "test-container").
					Return(container.CreateResponse{}, errors.New("creation failed"))
			},
			expectedID:    "",
			expectedError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockManager.ExpectedCalls = nil
			mockClient.ExpectedCalls = nil
			tc.setupMocks()

			id, err := service.Create(context.Background(), tc.config, tc.hostConfig, tc.containerName)

			if tc.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expectedID, id)
			}

			mockManager.AssertExpectations(t)
			mockClient.AssertExpectations(t)
		})
	}
}

func TestContainerService_Start(t *testing.T) {
	mockManager := new(MockManager)
	mockClient := new(MockAPIClient)
	mockLog := &mockLogger{}

	service := NewService(mockManager, mockLog)

	testCases := []struct {
		name          string
		containerID   string
		setupMocks    func()
		expectedError bool
	}{
		{
			name:        "successful start",
			containerID: "abc123",
			setupMocks: func() {
				mockManager.On("GetWithContext", mock.Anything).Return(mockClient, nil)
				mockClient.On("ContainerStart", mock.Anything, "abc123", mock.Anything).Return(nil)
			},
			expectedError: false,
		},
		{
			name:        "start error",
			containerID: "abc123",
			setupMocks: func() {
				mockManager.On("GetWithContext", mock.Anything).Return(mockClient, nil)
				mockClient.On("ContainerStart", mock.Anything, "abc123", mock.Anything).Return(errors.New("start failed"))
			},
			expectedError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockManager.ExpectedCalls = nil
			mockClient.ExpectedCalls = nil
			tc.setupMocks()

			err := service.Start(context.Background(), tc.containerID)

			if tc.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			mockManager.AssertExpectations(t)
			mockClient.AssertExpectations(t)
		})
	}
}

func TestContainerService_Stop(t *testing.T) {
	mockManager := new(MockManager)
	mockClient := new(MockAPIClient)
	mockLog := &mockLogger{}

	service := NewService(mockManager, mockLog)

	timeout := 30

	testCases := []struct {
		name          string
		containerID   string
		timeout       *int
		setupMocks    func()
		expectedError bool
	}{
		{
			name:        "successful stop with timeout",
			containerID: "abc123",
			timeout:     &timeout,
			setupMocks: func() {
				mockManager.On("GetWithContext", mock.Anything).Return(mockClient, nil)
				mockClient.On("ContainerStop", mock.Anything, "abc123", mock.MatchedBy(func(opts container.StopOptions) bool {
					return opts.Timeout != nil && *opts.Timeout == 30
				})).Return(nil)
			},
			expectedError: false,
		},
		{
			name:        "successful stop without timeout",
			containerID: "abc123",
			timeout:     nil,
			setupMocks: func() {
				mockManager.On("GetWithContext", mock.Anything).Return(mockClient, nil)
				mockClient.On("ContainerStop", mock.Anything, "abc123", mock.MatchedBy(func(opts container.StopOptions) bool {
					return opts.Timeout == nil
				})).Return(nil)
			},
			expectedError: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockManager.ExpectedCalls = nil
			mockClient.ExpectedCalls = nil
			tc.setupMocks()

			err := service.Stop(context.Background(), tc.containerID, tc.timeout)

			if tc.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			mockManager.AssertExpectations(t)
			mockClient.AssertExpectations(t)
		})
	}
}

func TestContainerService_List(t *testing.T) {
	mockManager := new(MockManager)
	mockClient := new(MockAPIClient)
	mockLog := &mockLogger{}

	service := NewService(mockManager, mockLog)

	expectedContainers := []types.Container{
		{
			ID:    "abc123",
			Names: []string{"/test1"},
			State: "running",
		},
		{
			ID:    "def456",
			Names: []string{"/test2"},
			State: "exited",
		},
	}

	testCases := []struct {
		name          string
		options       container.ListOptions
		setupMocks    func()
		expected      []types.Container
		expectedError bool
	}{
		{
			name:    "successful list",
			options: container.ListOptions{All: true},
			setupMocks: func() {
				mockManager.On("GetWithContext", mock.Anything).Return(mockClient, nil)
				mockClient.On("ContainerList", mock.Anything, mock.Anything).Return(expectedContainers, nil)
			},
			expected:      expectedContainers,
			expectedError: false,
		},
		{
			name:    "list error",
			options: container.ListOptions{All: true},
			setupMocks: func() {
				mockManager.On("GetWithContext", mock.Anything).Return(mockClient, nil)
				mockClient.On("ContainerList", mock.Anything, mock.Anything).Return(nil, errors.New("list failed"))
			},
			expected:      nil,
			expectedError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockManager.ExpectedCalls = nil
			mockClient.ExpectedCalls = nil
			tc.setupMocks()

			containers, err := service.List(context.Background(), tc.options)

			if tc.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expected, containers)
			}

			mockManager.AssertExpectations(t)
			mockClient.AssertExpectations(t)
		})
	}
}

func TestContainerService_Logs(t *testing.T) {
	mockManager := new(MockManager)
	mockClient := new(MockAPIClient)
	mockLog := &mockLogger{}

	service := NewService(mockManager, mockLog)

	mockReadCloser := io.NopCloser(strings.NewReader("test logs"))

	testCases := []struct {
		name          string
		containerID   string
		options       container.LogsOptions
		setupMocks    func()
		expectedError bool
	}{
		{
			name:        "successful logs retrieval",
			containerID: "abc123",
			options: container.LogsOptions{
				ShowStdout: true,
				ShowStderr: true,
				Follow:     false,
			},
			setupMocks: func() {
				mockManager.On("GetWithContext", mock.Anything).Return(mockClient, nil)
				mockClient.On("ContainerLogs", mock.Anything, "abc123", mock.Anything).Return(mockReadCloser, nil)
			},
			expectedError: false,
		},
		{
			name:        "logs error",
			containerID: "abc123",
			options: container.LogsOptions{
				ShowStdout: true,
			},
			setupMocks: func() {
				mockManager.On("GetWithContext", mock.Anything).Return(mockClient, nil)
				mockClient.On("ContainerLogs", mock.Anything, "abc123", mock.Anything).Return(nil, errors.New("logs failed"))
			},
			expectedError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockManager.ExpectedCalls = nil
			mockClient.ExpectedCalls = nil
			tc.setupMocks()

			logs, err := service.Logs(context.Background(), tc.containerID, tc.options)

			if tc.expectedError {
				assert.Error(t, err)
				assert.Nil(t, logs)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, logs)
				logs.Close()
			}

			mockManager.AssertExpectations(t)
			mockClient.AssertExpectations(t)
		})
	}
}
