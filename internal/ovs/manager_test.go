package ovs

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	mocks_logger "github.com/threatflux/libgo/test/mocks/logger"
	"go.uber.org/mock/gomock"
)

// MockCommandExecutor implements exec.CommandExecutor for testing
type MockCommandExecutor struct {
	mock.Mock
}

func (m *MockCommandExecutor) Execute(cmd string, args ...string) ([]byte, error) {
	argsList := append([]string{cmd}, args...)
	callArgs := make([]interface{}, len(argsList))
	for i, arg := range argsList {
		callArgs[i] = arg
	}

	result := m.Called(callArgs...)
	return result.Get(0).([]byte), result.Error(1)
}

func (m *MockCommandExecutor) ExecuteContext(ctx context.Context, cmd string, args ...string) ([]byte, error) {
	argsList := append([]string{cmd}, args...)
	callArgs := make([]interface{}, len(argsList)+1)
	callArgs[0] = ctx
	for i, arg := range argsList {
		callArgs[i+1] = arg
	}

	result := m.Called(callArgs...)
	return result.Get(0).([]byte), result.Error(1)
}

func TestOVSManager_CreateBridge(t *testing.T) {
	tests := []struct {
		mockSetup     func(*MockCommandExecutor) // 8 bytes (function pointer)
		name          string                     // 16 bytes (string header)
		bridgeName    string                     // 16 bytes (string header)
		errorContains string                     // 16 bytes (string header)
		expectedError bool                       // 1 byte (bool)
	}{
		{
			mockSetup: func(executor *MockCommandExecutor) {
				// Mock bridge exists check - returns exit code 2 (bridge doesn't exist)
				executor.On("ExecuteContext", mock.Anything, "ovs-vsctl", "br-exists", "br-test").
					Return([]byte{}, fmt.Errorf("exit status 2"))

				// Mock bridge creation
				executor.On("ExecuteContext", mock.Anything, "ovs-vsctl", "add-br", "br-test").
					Return([]byte{}, nil)
			},
			name:          "successful bridge creation",
			bridgeName:    "br-test",
			errorContains: "",
			expectedError: false,
		},
		{
			mockSetup: func(executor *MockCommandExecutor) {
				// Mock bridge exists check - returns success (bridge exists)
				executor.On("ExecuteContext", mock.Anything, "ovs-vsctl", "br-exists", "br-existing").
					Return([]byte{}, nil)
			},
			name:          "bridge already exists",
			bridgeName:    "br-existing",
			errorContains: "already exists",
			expectedError: true,
		},
		{
			mockSetup: func(executor *MockCommandExecutor) {
				// Mock bridge exists check - returns exit code 2 (bridge doesn't exist)
				executor.On("ExecuteContext", mock.Anything, "ovs-vsctl", "br-exists", "br-fail").
					Return([]byte{}, fmt.Errorf("exit status 2"))

				// Mock bridge creation failure
				executor.On("ExecuteContext", mock.Anything, "ovs-vsctl", "add-br", "br-fail").
					Return([]byte{}, fmt.Errorf("failed to create bridge"))
			},
			name:          "creation command fails",
			bridgeName:    "br-fail",
			errorContains: "creating bridge",
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			mockExecutor := new(MockCommandExecutor)
			mockLogger := mocks_logger.NewMockLogger(ctrl)

			// Configure mocks
			tt.mockSetup(mockExecutor)
			mockLogger.EXPECT().Info(gomock.Any(), gomock.Any()).AnyTimes()

			// Create manager
			manager := NewOVSManager(mockExecutor, mockLogger)

			// Execute
			err := manager.CreateBridge(context.Background(), tt.bridgeName)

			// Assert
			if tt.expectedError {
				assert.Error(t, err)
				if tt.errorContains != "" {
					assert.Contains(t, err.Error(), tt.errorContains)
				}
			} else {
				assert.NoError(t, err)
			}

			// Verify mocks
			mockExecutor.AssertExpectations(t)
		})
	}
}

func TestOVSManager_DeleteBridge(t *testing.T) {
	tests := []struct {
		mockSetup     func(*MockCommandExecutor) // 8 bytes (function pointer)
		name          string                     // 16 bytes (string header)
		bridgeName    string                     // 16 bytes (string header)
		expectedError bool                       // 1 byte (bool)
	}{
		{
			mockSetup: func(executor *MockCommandExecutor) {
				// Mock bridge exists check - returns success (bridge exists)
				executor.On("ExecuteContext", mock.Anything, "ovs-vsctl", "br-exists", "br-test").
					Return([]byte{}, nil)

				// Mock bridge deletion
				executor.On("ExecuteContext", mock.Anything, "ovs-vsctl", "del-br", "br-test").
					Return([]byte{}, nil)
			},
			name:          "successful bridge deletion",
			bridgeName:    "br-test",
			expectedError: false,
		},
		{
			mockSetup: func(executor *MockCommandExecutor) {
				// Mock bridge exists check - returns exit code 2 (bridge doesn't exist)
				executor.On("ExecuteContext", mock.Anything, "ovs-vsctl", "br-exists", "br-nonexistent").
					Return([]byte{}, fmt.Errorf("exit status 2"))
			},
			name:          "bridge doesn't exist",
			bridgeName:    "br-nonexistent",
			expectedError: false, // Should not error if bridge doesn't exist
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			mockExecutor := new(MockCommandExecutor)
			mockLogger := mocks_logger.NewMockLogger(ctrl)

			// Configure mocks
			tt.mockSetup(mockExecutor)
			mockLogger.EXPECT().Info(gomock.Any(), gomock.Any()).AnyTimes()

			// Create manager
			manager := NewOVSManager(mockExecutor, mockLogger)

			// Execute
			err := manager.DeleteBridge(context.Background(), tt.bridgeName)

			// Assert
			if tt.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			// Verify mocks
			mockExecutor.AssertExpectations(t)
		})
	}
}

func TestOVSManager_AddPort(t *testing.T) {
	tests := []struct {
		options       *PortOptions               // 8 bytes (pointer)
		mockSetup     func(*MockCommandExecutor) // 8 bytes (function pointer)
		name          string                     // 16 bytes (string header)
		bridge        string                     // 16 bytes (string header)
		port          string                     // 16 bytes (string header)
		errorContains string                     // 16 bytes (string header)
		expectedError bool                       // 1 byte (bool)
	}{
		{
			options: &PortOptions{
				Type: "system",
			},
			mockSetup: func(executor *MockCommandExecutor) {
				// Mock bridge exists check
				executor.On("ExecuteContext", mock.Anything, "ovs-vsctl", "br-exists", "br-test").
					Return([]byte{}, nil)

				// Mock port addition
				executor.On("ExecuteContext", mock.Anything, "ovs-vsctl", "add-port", "br-test", "eth0", "--", "set", "Interface", "eth0", "type=system").
					Return([]byte{}, nil)
			},
			name:          "successful port addition",
			bridge:        "br-test",
			port:          "eth0",
			errorContains: "",
			expectedError: false,
		},
		{
			options: nil,
			mockSetup: func(executor *MockCommandExecutor) {
				// Mock bridge exists check - returns exit code 2 (bridge doesn't exist)
				executor.On("ExecuteContext", mock.Anything, "ovs-vsctl", "br-exists", "br-nonexistent").
					Return([]byte{}, fmt.Errorf("exit status 2"))
			},
			name:          "bridge doesn't exist",
			bridge:        "br-nonexistent",
			port:          "eth0",
			errorContains: "not found",
			expectedError: true,
		},
		{
			options: &PortOptions{
				Tag: &[]int{100}[0], // Pointer to int
			},
			mockSetup: func(executor *MockCommandExecutor) {
				// Mock bridge exists check
				executor.On("ExecuteContext", mock.Anything, "ovs-vsctl", "br-exists", "br-test").
					Return([]byte{}, nil)

				// Mock port addition
				executor.On("ExecuteContext", mock.Anything, "ovs-vsctl", "add-port", "br-test", "eth1").
					Return([]byte{}, nil)

				// Mock port exists check for VLAN setting
				executor.On("ExecuteContext", mock.Anything, "ovs-vsctl", "list-ports", "br-test").
					Return([]byte("eth1\n"), nil)

				// Mock VLAN tag setting
				executor.On("ExecuteContext", mock.Anything, "ovs-vsctl", "set", "Port", "eth1", "tag=100").
					Return([]byte{}, nil)
			},
			name:          "port addition with VLAN tag",
			bridge:        "br-test",
			port:          "eth1",
			errorContains: "",
			expectedError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			mockExecutor := new(MockCommandExecutor)
			mockLogger := mocks_logger.NewMockLogger(ctrl)

			// Configure mocks
			tt.mockSetup(mockExecutor)
			mockLogger.EXPECT().Info(gomock.Any(), gomock.Any()).AnyTimes()

			// Create manager
			manager := NewOVSManager(mockExecutor, mockLogger)

			// Execute
			err := manager.AddPort(context.Background(), tt.bridge, tt.port, tt.options)

			// Assert
			if tt.expectedError {
				assert.Error(t, err)
				if tt.errorContains != "" {
					assert.Contains(t, err.Error(), tt.errorContains)
				}
			} else {
				assert.NoError(t, err)
			}

			// Verify mocks
			mockExecutor.AssertExpectations(t)
		})
	}
}

func TestOVSManager_ListBridges(t *testing.T) {
	tests := []struct {
		mockSetup     func(*MockCommandExecutor) // 8 bytes (function pointer)
		name          string                     // 16 bytes (string header)
		expectedCount int                        // 8 bytes (int64 on 64-bit systems)
		expectedError bool                       // 1 byte (bool)
	}{
		{
			mockSetup: func(executor *MockCommandExecutor) {
				// Mock list bridges command
				executor.On("ExecuteContext", mock.Anything, "ovs-vsctl", "list-br").
					Return([]byte("br-test\nbr-mgmt\n"), nil)

				// Mock getting bridge details for each bridge
				// For br-test
				executor.On("ExecuteContext", mock.Anything, "ovs-vsctl", "br-exists", "br-test").
					Return([]byte{}, nil)
				executor.On("ExecuteContext", mock.Anything, "ovs-vsctl", "get", "Bridge", "br-test", "_uuid").
					Return([]byte("12345678-1234-1234-1234-123456789abc"), nil)
				executor.On("ExecuteContext", mock.Anything, "ovs-vsctl", "get", "Bridge", "br-test", "controller").
					Return([]byte("[]\n"), nil)
				executor.On("ExecuteContext", mock.Anything, "ovs-vsctl", "get", "Bridge", "br-test", "datapath_type").
					Return([]byte("\"system\"\n"), nil)
				executor.On("ExecuteContext", mock.Anything, "ovs-vsctl", "list-ports", "br-test").
					Return([]byte("eth0\n"), nil)
				executor.On("ExecuteContext", mock.Anything, "ovs-vsctl", "get", "Bridge", "br-test", "external_ids").
					Return([]byte("{}\n"), nil)
				executor.On("ExecuteContext", mock.Anything, "ovs-vsctl", "get", "Bridge", "br-test", "other_config").
					Return([]byte("{}\n"), nil)
				executor.On("ExecuteContext", mock.Anything, "ovs-ofctl", "dump-flows", "br-test").
					Return([]byte("cookie=0x0, duration=1.234s, table=0, n_packets=0, n_bytes=0, priority=0 actions=NORMAL\n"), nil)

				// For br-mgmt
				executor.On("ExecuteContext", mock.Anything, "ovs-vsctl", "br-exists", "br-mgmt").
					Return([]byte{}, nil)
				executor.On("ExecuteContext", mock.Anything, "ovs-vsctl", "get", "Bridge", "br-mgmt", "_uuid").
					Return([]byte("87654321-4321-4321-4321-cba987654321"), nil)
				executor.On("ExecuteContext", mock.Anything, "ovs-vsctl", "get", "Bridge", "br-mgmt", "controller").
					Return([]byte("[]\n"), nil)
				executor.On("ExecuteContext", mock.Anything, "ovs-vsctl", "get", "Bridge", "br-mgmt", "datapath_type").
					Return([]byte("\"system\"\n"), nil)
				executor.On("ExecuteContext", mock.Anything, "ovs-vsctl", "list-ports", "br-mgmt").
					Return([]byte("\n"), nil)
				executor.On("ExecuteContext", mock.Anything, "ovs-vsctl", "get", "Bridge", "br-mgmt", "external_ids").
					Return([]byte("{}\n"), nil)
				executor.On("ExecuteContext", mock.Anything, "ovs-vsctl", "get", "Bridge", "br-mgmt", "other_config").
					Return([]byte("{}\n"), nil)
				executor.On("ExecuteContext", mock.Anything, "ovs-ofctl", "dump-flows", "br-mgmt").
					Return([]byte("NXST_FLOW reply (xid=0x4):\n"), nil)
			},
			name:          "successful bridge listing",
			expectedCount: 2,
			expectedError: false,
		},
		{
			mockSetup: func(executor *MockCommandExecutor) {
				// Mock list bridges command returning empty result
				executor.On("ExecuteContext", mock.Anything, "ovs-vsctl", "list-br").
					Return([]byte("\n"), nil)
			},
			name:          "no bridges",
			expectedCount: 0,
			expectedError: false,
		},
		{
			mockSetup: func(executor *MockCommandExecutor) {
				// Mock list bridges command failure
				executor.On("ExecuteContext", mock.Anything, "ovs-vsctl", "list-br").
					Return([]byte{}, fmt.Errorf("failed to list bridges"))
			},
			name:          "command failure",
			expectedCount: 0,
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			mockExecutor := new(MockCommandExecutor)
			mockLogger := mocks_logger.NewMockLogger(ctrl)

			// Configure mocks
			tt.mockSetup(mockExecutor)
			mockLogger.EXPECT().Warn(gomock.Any(), gomock.Any()).AnyTimes()

			// Create manager
			manager := NewOVSManager(mockExecutor, mockLogger)

			// Execute
			bridges, err := manager.ListBridges(context.Background())

			// Assert
			if tt.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Len(t, bridges, tt.expectedCount)
			}

			// Verify mocks
			mockExecutor.AssertExpectations(t)
		})
	}
}
