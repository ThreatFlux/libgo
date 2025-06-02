package ovs

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	mocks_logger "github.com/threatflux/libgo/test/mocks/logger"
	"go.uber.org/mock/gomock"
)

// Simple functional test to verify OVS manager creation and basic operations
func TestOVSManager_Creation(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockExecutor := new(MockCommandExecutor)
	mockLogger := mocks_logger.NewMockLogger(ctrl)

	// Allow any logging
	mockLogger.EXPECT().Info(gomock.Any(), gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Warn(gomock.Any(), gomock.Any()).AnyTimes()

	// Create manager - this should not fail
	manager := NewOVSManager(mockExecutor, mockLogger)
	assert.NotNil(t, manager)
}

func TestOVSManager_BridgeExistsLogic(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockExecutor := new(MockCommandExecutor)
	mockLogger := mocks_logger.NewMockLogger(ctrl)

	// Allow any logging
	mockLogger.EXPECT().Info(gomock.Any(), gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Warn(gomock.Any(), gomock.Any()).AnyTimes()

	manager := NewOVSManager(mockExecutor, mockLogger)

	// Test bridge exists - successful case
	mockExecutor.On("ExecuteContext", mock.Anything, "ovs-vsctl", "br-exists", "test-bridge").Return([]byte{}, nil)

	exists, err := manager.bridgeExists(context.Background(), "test-bridge")
	assert.NoError(t, err)
	assert.True(t, exists)

	mockExecutor.AssertExpectations(t)
}

func TestOVSManager_BridgeDoesNotExist(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockExecutor := new(MockCommandExecutor)
	mockLogger := mocks_logger.NewMockLogger(ctrl)

	// Allow any logging
	mockLogger.EXPECT().Info(gomock.Any(), gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Warn(gomock.Any(), gomock.Any()).AnyTimes()

	manager := NewOVSManager(mockExecutor, mockLogger)

	// Mock exit status 2 error (bridge doesn't exist)
	mockError := errors.New("exit status 2")

	mockExecutor.On("ExecuteContext", mock.Anything, "ovs-vsctl", "br-exists", "nonexistent-bridge").Return([]byte{}, mockError)

	exists, err := manager.bridgeExists(context.Background(), "nonexistent-bridge")
	assert.NoError(t, err)
	assert.False(t, exists)

	mockExecutor.AssertExpectations(t)
}
