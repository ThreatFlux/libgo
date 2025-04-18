package cloudinit

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/wroersma/libgo/internal/models/vm"
	"github.com/wroersma/libgo/pkg/logger"
)

func TestISOBuilder_createCloudInitFiles(t *testing.T) {
	// Create a temporary directory for test files
	tempDir, err := os.MkdirTemp("", "cloud-init-files-*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	
	mockLogger := logger.NewMockLogger(ctrl)
	mockLogger.EXPECT().Debug(gomock.Any(), gomock.Any()).AnyTimes()
	
	mockGenerator := &CloudInitGenerator{
		logger: mockLogger,
	}
	
	builder := NewISOBuilder(tempDir, mockGenerator, mockLogger)
	
	// Test creating files with all content provided
	config := vm.CloudInitConfig{
		UserData:      "#cloud-config\nhostname: test-vm",
		MetaData:      "instance-id: test-id\nlocal-hostname: test-vm",
		NetworkConfig: "version: 2\nethernets:\n  ens3:\n    dhcp4: true",
	}
	
	err = builder.createCloudInitFiles(tempDir, config)
	require.NoError(t, err)
	
	// Verify the files were created with the correct content
	userData, err := os.ReadFile(filepath.Join(tempDir, "user-data"))
	require.NoError(t, err)
	assert.Equal(t, config.UserData, string(userData))
	
	metaData, err := os.ReadFile(filepath.Join(tempDir, "meta-data"))
	require.NoError(t, err)
	assert.Equal(t, config.MetaData, string(metaData))
	
	networkConfig, err := os.ReadFile(filepath.Join(tempDir, "network-config"))
	require.NoError(t, err)
	assert.Equal(t, config.NetworkConfig, string(networkConfig))
}

func TestISOBuilder_findISOTool(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	
	mockLogger := logger.NewMockLogger(ctrl)
	mockGenerator := &CloudInitGenerator{
		logger: mockLogger,
	}
	
	builder := NewISOBuilder("/tmp", mockGenerator, mockLogger)
	
	// This test will depend on the environment where it runs
	// It will either find an ISO tool or return an error
	tool, err := builder.findISOTool()
	if err != nil {
		// If no tool is found, error should indicate that
		assert.Contains(t, err.Error(), "no ISO generation tool found")
	} else {
		// If a tool is found, it should be one of the expected tools
		assert.Contains(t, []string{
			"genisoimage", "/usr/bin/genisoimage",
			"mkisofs", "/usr/bin/mkisofs",
			"xorrisofs", "/usr/bin/xorrisofs",
		}, filepath.Base(tool))
	}
}

func TestISOBuilder_defaultISOTimeout(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	
	mockLogger := logger.NewMockLogger(ctrl)
	mockGenerator := &CloudInitGenerator{
		logger: mockLogger,
	}
	
	builder := NewISOBuilder("/tmp", mockGenerator, mockLogger)
	
	// Verify the default timeout
	timeout := builder.defaultISOTimeout()
	assert.Equal(t, 2*time.Minute, timeout)
}

// Integration test for GenerateISO - only run if genisoimage or mkisofs is available
func TestISOBuilder_GenerateISO_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	
	// Try to find an ISO tool
	isoTool, err := exec.LookPath("genisoimage")
	if err != nil {
		isoTool, err = exec.LookPath("mkisofs")
		if err != nil {
			t.Skip("Skipping test: no ISO generation tool found")
		}
	}
	
	// Create temporary directories
	workDir, err := os.MkdirTemp("", "cloud-init-work-*")
	require.NoError(t, err)
	defer os.RemoveAll(workDir)
	
	targetDir, err := os.MkdirTemp("", "cloud-init-iso-*")
	require.NoError(t, err)
	defer os.RemoveAll(targetDir)
	
	targetPath := filepath.Join(targetDir, "cloud-init.iso")
	
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	
	mockLogger := logger.NewMockLogger(ctrl)
	mockLogger.EXPECT().Debug(gomock.Any(), gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Info(gomock.Any(), gomock.Any()).AnyTimes()
	
	mockGenerator := &CloudInitGenerator{
		logger: mockLogger,
	}
	
	builder := NewISOBuilder(workDir, mockGenerator, mockLogger)
	
	// Test generating ISO
	config := vm.CloudInitConfig{
		UserData: "#cloud-config\nhostname: test-vm",
		MetaData: "instance-id: test-id\nlocal-hostname: test-vm",
	}
	
	err = builder.GenerateISO(context.Background(), config, targetPath)
	if err != nil {
		// If ISO generation fails, check if it's because the tool is not found
		if _, toolErr := os.Stat(isoTool); os.IsNotExist(toolErr) {
			t.Skip("Skipping test: ISO generation tool not found or not executable")
		}
		
		// Otherwise, it's a real error
		t.Fatalf("Error generating ISO: %v", err)
	}
	
	// Verify the ISO file was created
	info, err := os.Stat(targetPath)
	require.NoError(t, err)
	assert.True(t, info.Size() > 0, "ISO file should not be empty")
}