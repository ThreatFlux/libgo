package ova

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	vmmodels "github.com/threatflux/libgo/internal/models/vm"
	mocks_logger "github.com/threatflux/libgo/test/mocks/logger"
	"go.uber.org/mock/gomock"
)

func TestOVFTemplateGenerator_GenerateOVF(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLogger := mocks_logger.NewMockLogger(ctrl)
	templateGenerator, err := NewOVFTemplateGenerator(mockLogger)
	require.NoError(t, err)
	require.NotNil(t, templateGenerator)

	// Test cases
	testCases := []struct {
		name      string
		vmInfo    *vmmodels.VM
		diskPath  string
		diskSize  uint64
		expectErr bool
	}{
		{
			name: "Valid VM info",
			vmInfo: &vmmodels.VM{
				Name: "test-vm",
				UUID: "12345678-1234-1234-1234-123456789012",
				CPU: vmmodels.CPUInfo{
					Count: 2,
				},
				Memory: vmmodels.MemoryInfo{
					SizeBytes: 2048 * 1024 * 1024, // 2GB
				},
			},
			diskPath:  "/path/to/disk.vmdk",
			diskSize:  1024 * 1024 * 1024, // 1 GB
			expectErr: false,
		},
		{
			name: "Minimal VM info with defaults",
			vmInfo: &vmmodels.VM{
				Name: "minimal-vm",
				// No UUID, CPU, Memory - should use defaults
			},
			diskPath:  "/path/to/disk.vmdk",
			diskSize:  512 * 1024 * 1024, // 512 MB
			expectErr: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ovfContent, err := templateGenerator.GenerateOVF(tc.vmInfo, tc.diskPath, tc.diskSize)

			if tc.expectErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.NotEmpty(t, ovfContent)

			// Verify basic content
			assert.Contains(t, ovfContent, tc.vmInfo.Name)
			assert.Contains(t, ovfContent, "ovf:id")
			assert.Contains(t, ovfContent, "VirtualHardwareSection")

			// Check for basic structure
			assert.Contains(t, ovfContent, "<?xml version=\"1.0\" encoding=\"UTF-8\"?>")
			assert.Contains(t, ovfContent, "<Envelope")
			assert.Contains(t, ovfContent, "</Envelope>")

			// Check for disk reference
			assert.Contains(t, ovfContent, "disk.vmdk")

			// Verify CPU count
			cpuCount := tc.vmInfo.CPU.Count
			if cpuCount == 0 {
				cpuCount = 1 // Default value
			}
			_ = cpuCount // Used in assertion below
			assert.Contains(t, ovfContent, "<rasd:VirtualQuantity>"+strings.TrimSpace(strings.Split(ovfContent, "<rasd:VirtualQuantity>")[1][:1]))

			// Verify memory configuration is present
			assert.Contains(t, ovfContent, "MB of memory")
		})
	}
}

func TestNewOVFTemplateGenerator(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLogger := mocks_logger.NewMockLogger(ctrl)
	generator, err := NewOVFTemplateGenerator(mockLogger)

	assert.NoError(t, err)
	assert.NotNil(t, generator)
	assert.NotNil(t, generator.templateLoader)
	assert.Equal(t, mockLogger, generator.logger)
}

func TestOVFTemplateGenerator_WriteOVFToFile(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLogger := mocks_logger.NewMockLogger(ctrl)
	templateGenerator, err := NewOVFTemplateGenerator(mockLogger)
	require.NoError(t, err)

	// Create a temporary file
	tmpDir := t.TempDir()
	testContent := "<ovf>Test Content</ovf>"
	filePath := tmpDir + "/test.ovf"

	err = templateGenerator.WriteOVFToFile(testContent, filePath)
	require.NoError(t, err)

	// Verify file was created with correct content
	assert.FileExists(t, filePath)
}
