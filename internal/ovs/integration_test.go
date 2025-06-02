//go:build integration
// +build integration

package ovs

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/threatflux/libgo/internal/config"
	"github.com/threatflux/libgo/pkg/logger"
	"github.com/threatflux/libgo/pkg/utils/exec"
	mocks_logger "github.com/threatflux/libgo/test/mocks/logger"
	"go.uber.org/mock/gomock"
)

// TestOVSManager_RealOVSIntegration tests the OVS manager with real OVS commands
// This test requires OpenVSwitch to be installed and accessible
func TestOVSManager_RealOVSIntegration(t *testing.T) {
	// Skip if OVS is not available
	if !isOVSAvailable() {
		t.Skip("OpenVSwitch is not available - run 'make install-ovs' to install")
	}

	// Skip if not running as root (OVS operations typically require root)
	if os.Getuid() != 0 {
		t.Skip("OVS integration test requires root privileges")
	}

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Use real command executor
	realExecutor := &exec.DefaultCommandExecutor{}
	mockLogger := mocks_logger.NewMockLogger(ctrl)

	// Allow any logging
	mockLogger.EXPECT().Info(gomock.Any(), gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Warn(gomock.Any(), gomock.Any()).AnyTimes()

	manager := NewOVSManager(realExecutor, mockLogger)

	ctx := context.Background()
	testBridgeName := "test-bridge-libgo"

	// Clean up any existing test bridge
	_ = manager.DeleteBridge(ctx, testBridgeName)

	// Test bridge creation
	err := manager.CreateBridge(ctx, testBridgeName)
	require.NoError(t, err, "Failed to create test bridge")

	// Verify bridge exists
	exists, err := manager.bridgeExists(ctx, testBridgeName)
	require.NoError(t, err)
	assert.True(t, exists, "Bridge should exist after creation")

	// Test bridge listing
	bridges, err := manager.ListBridges(ctx)
	require.NoError(t, err)
	assert.NotEmpty(t, bridges, "Should have at least one bridge")

	// Find our test bridge in the list
	var testBridge *BridgeInfo
	for _, bridge := range bridges {
		if bridge.Name == testBridgeName {
			testBridge = &bridge
			break
		}
	}
	require.NotNil(t, testBridge, "Test bridge should be in the list")
	assert.Equal(t, testBridgeName, testBridge.Name)

	// Test getting bridge details
	bridgeDetails, err := manager.GetBridge(ctx, testBridgeName)
	require.NoError(t, err)
	assert.Equal(t, testBridgeName, bridgeDetails.Name)
	assert.NotEmpty(t, bridgeDetails.UUID)

	// Clean up - delete the test bridge
	err = manager.DeleteBridge(ctx, testBridgeName)
	require.NoError(t, err, "Failed to delete test bridge")

	// Verify bridge no longer exists
	exists, err = manager.bridgeExists(ctx, testBridgeName)
	require.NoError(t, err)
	assert.False(t, exists, "Bridge should not exist after deletion")
}

// isOVSAvailable checks if OVS commands are available
func isOVSAvailable() bool {
	executor := &exec.DefaultCommandExecutor{}

	// Check if ovs-vsctl is available
	_, err := executor.Execute("which", "ovs-vsctl")
	if err != nil {
		return false
	}

	// Check if ovs-ofctl is available
	_, err = executor.Execute("which", "ovs-ofctl")
	return err == nil
}

// TestOVSManager_BasicOperationsWithMockLogger tests basic operations with minimal setup
func TestOVSManager_BasicOperationsWithMockLogger(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	realExecutor := &exec.DefaultCommandExecutor{}

	// Use a real logger for this test to see actual output
	logConfig := config.LoggingConfig{
		Level:  "debug",
		Format: "text",
	}
	zapLogger, err := logger.NewZapLogger(logConfig)
	require.NoError(t, err)

	manager := NewOVSManager(realExecutor, zapLogger)
	assert.NotNil(t, manager, "Manager should be created successfully")

	// Test that manager can be created and basic methods exist
	ctx := context.Background()

	// This should not panic even if OVS is not available
	_, err = manager.bridgeExists(ctx, "nonexistent-bridge")
	// We don't assert on the error because OVS might not be available,
	// but we can assert that the method doesn't panic
	t.Logf("Bridge exists check returned: %v", err)
}
