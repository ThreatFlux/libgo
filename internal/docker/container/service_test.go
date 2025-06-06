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

func (m *MockAPIClient) ContainerInspect(ctx context.Context, containerID string) (container.InspectResponse, error) {
	args := m.Called(ctx, containerID)
	return args.Get(0).(container.InspectResponse), args.Error(1)
}

func (m *MockAPIClient) ContainerList(ctx context.Context, options container.ListOptions) ([]container.Summary, error) {
	args := m.Called(ctx, options)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]container.Summary), args.Error(1)
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
		config        *container.Config     // 8 bytes (pointer)
		hostConfig    *container.HostConfig // 8 bytes (pointer)
		setupMocks    func()                // 8 bytes (function pointer)
		containerName string                // 16 bytes (string header)
		name          string                // 16 bytes (string header)
		expectedID    string                // 16 bytes (string header)
		expectedError bool                  // 1 byte (bool)
	}{
		{
			config: &container.Config{
				Image: "nginx:latest",
			},
			hostConfig: &container.HostConfig{},
			setupMocks: func() {
				mockManager.On("GetWithContext", mock.Anything).Return(mockClient, nil)
				mockClient.On("ContainerCreate", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, "test-container").
					Return(container.CreateResponse{ID: "abc123", Warnings: []string{"test warning"}}, nil)
			},
			containerName: "test-container",
			name:          "successful creation",
			expectedID:    "abc123",
			expectedError: false,
		},
		{
			config: &container.Config{
				Image: "nginx:latest",
			},
			hostConfig: &container.HostConfig{},
			setupMocks: func() {
				mockManager.On("GetWithContext", mock.Anything).Return(nil, errors.New("manager error"))
			},
			containerName: "test-container",
			name:          "manager error",
			expectedID:    "",
			expectedError: true,
		},
		{
			config: &container.Config{
				Image: "nginx:latest",
			},
			hostConfig: &container.HostConfig{},
			setupMocks: func() {
				mockManager.On("GetWithContext", mock.Anything).Return(mockClient, nil)
				mockClient.On("ContainerCreate", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, "test-container").
					Return(container.CreateResponse{}, errors.New("creation failed"))
			},
			containerName: "test-container",
			name:          "creation error",
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
		setupMocks    func() // 8 bytes (function pointer)
		name          string // 16 bytes (string header)
		containerID   string // 16 bytes (string header)
		expectedError bool   // 1 byte (bool)
	}{
		{
			setupMocks: func() {
				mockManager.On("GetWithContext", mock.Anything).Return(mockClient, nil)
				mockClient.On("ContainerStart", mock.Anything, "abc123", mock.Anything).Return(nil)
			},
			name:          "successful start",
			containerID:   "abc123",
			expectedError: false,
		},
		{
			setupMocks: func() {
				mockManager.On("GetWithContext", mock.Anything).Return(mockClient, nil)
				mockClient.On("ContainerStart", mock.Anything, "abc123", mock.Anything).Return(errors.New("start failed"))
			},
			name:          "start error",
			containerID:   "abc123",
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
		timeout       *int   // 8 bytes (pointer)
		setupMocks    func() // 8 bytes (function pointer)
		name          string // 16 bytes (string header)
		containerID   string // 16 bytes (string header)
		expectedError bool   // 1 byte (bool)
	}{
		{
			timeout: &timeout,
			setupMocks: func() {
				mockManager.On("GetWithContext", mock.Anything).Return(mockClient, nil)
				mockClient.On("ContainerStop", mock.Anything, "abc123", mock.MatchedBy(func(opts container.StopOptions) bool {
					return opts.Timeout != nil && *opts.Timeout == 30
				})).Return(nil)
			},
			name:          "successful stop with timeout",
			containerID:   "abc123",
			expectedError: false,
		},
		{
			timeout: nil,
			setupMocks: func() {
				mockManager.On("GetWithContext", mock.Anything).Return(mockClient, nil)
				mockClient.On("ContainerStop", mock.Anything, "abc123", mock.MatchedBy(func(opts container.StopOptions) bool {
					return opts.Timeout == nil
				})).Return(nil)
			},
			name:          "successful stop without timeout",
			containerID:   "abc123",
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

	expectedContainers := []container.Summary{
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
		expected      []container.Summary   // 24 bytes (slice header)
		setupMocks    func()                // 8 bytes (function pointer)
		options       container.ListOptions // size depends on struct, but typically large
		name          string                // 16 bytes (string header)
		expectedError bool                  // 1 byte (bool)
	}{
		{
			expected: expectedContainers,
			setupMocks: func() {
				mockManager.On("GetWithContext", mock.Anything).Return(mockClient, nil)
				mockClient.On("ContainerList", mock.Anything, mock.Anything).Return(expectedContainers, nil)
			},
			options:       container.ListOptions{All: true},
			name:          "successful list",
			expectedError: false,
		},
		{
			expected: nil,
			setupMocks: func() {
				mockManager.On("GetWithContext", mock.Anything).Return(mockClient, nil)
				mockClient.On("ContainerList", mock.Anything, mock.Anything).Return(nil, errors.New("list failed"))
			},
			options:       container.ListOptions{All: true},
			name:          "list error",
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
		setupMocks    func()                // 8 bytes (function pointer)
		options       container.LogsOptions // size depends on struct, but typically large
		name          string                // 16 bytes (string header)
		containerID   string                // 16 bytes (string header)
		expectedError bool                  // 1 byte (bool)
	}{
		{
			setupMocks: func() {
				mockManager.On("GetWithContext", mock.Anything).Return(mockClient, nil)
				mockClient.On("ContainerLogs", mock.Anything, "abc123", mock.Anything).Return(mockReadCloser, nil)
			},
			options: container.LogsOptions{
				ShowStdout: true,
				ShowStderr: true,
				Follow:     false,
			},
			name:          "successful logs retrieval",
			containerID:   "abc123",
			expectedError: false,
		},
		{
			setupMocks: func() {
				mockManager.On("GetWithContext", mock.Anything).Return(mockClient, nil)
				mockClient.On("ContainerLogs", mock.Anything, "abc123", mock.Anything).Return(nil, errors.New("logs failed"))
			},
			options: container.LogsOptions{
				ShowStdout: true,
			},
			name:          "logs error",
			containerID:   "abc123",
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
