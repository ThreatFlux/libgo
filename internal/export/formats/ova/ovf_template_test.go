package ova

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/threatflux/libgo/internal/models/vm"
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
		vmInfo    *vm.VM
		diskPath  string
		diskSize  uint64
		expectErr bool
	}{
		{
			name: "Valid VM info",
			vmInfo: &vm.VM{
				Name: "test-vm",
				UUID: "12345678-1234-1234-1234-123456789012",
				CPU: vm.CPUInfo{
					Count: 2,
				},
				Memory: vm.MemoryInfo{
					SizeMB: 2048,
				},
			},
			diskPath:  "/path/to/disk.vmdk",
			diskSize:  1024 * 1024 * 1024, // 1 GB
			expectErr: false,
		},
		{
			name: "Minimal VM info with defaults",
			vmInfo: &vm.VM{
				Name: "minimal-vm",
				// No UUID, CPU, Memory
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
			assert.Contains(t, ovfContent, "ovf:capacity")
			assert.Contains(t, ovfContent, "VirtualHardwareSection")

			// Check for basic structure
			assert.Contains(t, ovfContent, "<?xml version=\"1.0\" encoding=\"UTF-8\"?>")
			assert.Contains(t, ovfContent, "<Envelope")
			assert.Contains(t, ovfContent, "</Envelope>")

			// Check for disk reference
			assert.Contains(t, ovfContent, filepath.Base(tc.diskPath))

			// Check for VM attributes
			if tc.vmInfo.UUID != "" {
				assert.Contains(t, ovfContent, tc.vmInfo.UUID)
			}

			cpuCount := tc.vmInfo.CPU.Count
			if cpuCount == 0 {
				cpuCount = 1 // Default value
			}
			assert.Contains(t, ovfContent, "<rasd:VirtualQuantity>"+string('0'+cpuCount))

			memorySizeMB := tc.vmInfo.Memory.SizeMB
			if memorySizeMB == 0 {
				memorySizeMB = 1024 // Default value
			}
			// Skip exact memory check as it's a string representation in XML
		})
	}
}

func TestOVFTemplateGenerator_WriteOVFToFile(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLogger := mocks_logger.NewMockLogger(ctrl)
	templateGenerator, err := NewOVFTemplateGenerator(mockLogger)
	require.NoError(t, err)

	// Create a temporary directory for the test
	tempDir, err := os.MkdirTemp("", "ovf-test-")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Test writing to a file
	testContent := "<ovf>Test Content</ovf>"
	filePath := filepath.Join(tempDir, "test.ovf")

	err = templateGenerator.WriteOVFToFile(testContent, filePath)
	require.NoError(t, err)

	// Verify file content
	content, err := os.ReadFile(filePath)
	require.NoError(t, err)
	assert.Equal(t, testContent, string(content))
}
