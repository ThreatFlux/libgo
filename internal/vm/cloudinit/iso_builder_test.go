package cloudinit

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	mocks_logger "github.com/threatflux/libgo/test/mocks/logger"
	"go.uber.org/mock/gomock"
)

func TestCloudInitGenerator_GenerateISO(t *testing.T) {
	// Skip if no ISO tools are available
	hasISOTool := false
	for _, tool := range []string{"genisoimage", "mkisofs", "xorrisofs"} {
		if _, err := exec.LookPath(tool); err == nil {
			hasISOTool = true
			break
		}
	}

	if !hasISOTool {
		t.Skip("Skipping test: no ISO generation tool found")
	}

	// Create a temporary directory for test files
	tempDir, err := os.MkdirTemp("", "cloud-init-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLogger := mocks_logger.NewMockLogger(ctrl)
	mockLogger.EXPECT().Debug(gomock.Any(), gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Info(gomock.Any(), gomock.Any()).AnyTimes()

	generator := &CloudInitGenerator{
		logger: mockLogger,
	}

	// Test generating ISO with all content provided
	config := CloudInitConfig{
		UserData:      "#cloud-config\nhostname: test-vm",
		MetaData:      "instance-id: test-id\nlocal-hostname: test-vm",
		NetworkConfig: "version: 2\nethernets:\n  enp1s0:\n    dhcp4: true",
	}

	targetPath := filepath.Join(tempDir, "cloud-init.iso")
	err = generator.GenerateISO(context.Background(), config, targetPath)
	require.NoError(t, err)

	// Verify the ISO file was created
	info, err := os.Stat(targetPath)
	require.NoError(t, err)
	assert.True(t, info.Size() > 0, "ISO file should not be empty")
}

func TestCloudInitGenerator_GenerateISO_NoTools(t *testing.T) {
	// This test checks behavior when no ISO tools are available
	// We'll temporarily hide the ISO tools from PATH
	oldPath := os.Getenv("PATH")
	os.Setenv("PATH", "/tmp/nonexistent")
	defer os.Setenv("PATH", oldPath)

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLogger := mocks_logger.NewMockLogger(ctrl)
	mockLogger.EXPECT().Debug(gomock.Any(), gomock.Any()).AnyTimes()

	generator := &CloudInitGenerator{
		logger: mockLogger,
	}

	config := CloudInitConfig{
		UserData: "#cloud-config\nhostname: test-vm",
		MetaData: "instance-id: test-id",
	}

	err := generator.GenerateISO(context.Background(), config, "/tmp/test.iso")
	// We expect an error when no ISO tools are available
	assert.Error(t, err)
}

func TestCloudInitGenerator_GenerateISO_InvalidOutputPath(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLogger := mocks_logger.NewMockLogger(ctrl)
	mockLogger.EXPECT().Debug(gomock.Any(), gomock.Any()).AnyTimes()

	generator := &CloudInitGenerator{
		logger: mockLogger,
	}

	config := CloudInitConfig{
		UserData: "#cloud-config\nhostname: test-vm",
		MetaData: "instance-id: test-id",
	}

	// Try to write to a non-existent directory
	err := generator.GenerateISO(context.Background(), config, "/nonexistent/dir/test.iso")
	assert.Error(t, err)
}
