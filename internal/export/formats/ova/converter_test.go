package ova

import (
	"context"
	"errors"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/threatflux/libgo/internal/models/vm"
	mocks_logger "github.com/threatflux/libgo/test/mocks/logger"
	"go.uber.org/mock/gomock"
)

// Mock the exec.ExecuteCommand function to avoid running the actual command
var execCommand = func(ctx context.Context, name string, args []string, opts interface{}) ([]byte, error) {
	return nil, nil
}

// Mock the os functions
var (
	osStatFunc      = func(path string) (os.FileInfo, error) { return nil, nil }
	osMkdirTempFunc = func(dir, prefix string) (string, error) { return "/tmp/test-dir", nil }
	osRemoveAllFunc = func(path string) error { return nil }
)

// Mock implementation of FileInfo for testing
type mockFileInfo struct{}

func (m mockFileInfo) Name() string       { return "disk.vmdk" }
func (m mockFileInfo) Size() int64        { return 1024 * 1024 * 1024 } // 1 GB
func (m mockFileInfo) Mode() os.FileMode  { return 0644 }
func (m mockFileInfo) ModTime() time.Time { return time.Now() }
func (m mockFileInfo) IsDir() bool        { return false }
func (m mockFileInfo) Sys() interface{}   { return nil }

// createMockFileInfo returns a mock file info for testing
func createMockFileInfo(size int64) os.FileInfo {
	return &mockFileInfo{}
}

func TestOVAConverter_GetFormatName(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLogger := mocks_logger.NewMockLogger(ctrl)
	templateGenerator, err := NewOVFTemplateGenerator(mockLogger)
	require.NoError(t, err)

	converter := NewOVAConverter(templateGenerator, mockLogger)
	assert.Equal(t, "ova", converter.GetFormatName())
}

func TestOVAConverter_ValidateOptions(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLogger := mocks_logger.NewMockLogger(ctrl)
	templateGenerator, err := NewOVFTemplateGenerator(mockLogger)
	require.NoError(t, err)

	converter := NewOVAConverter(templateGenerator, mockLogger)

	testCases := []struct {
		name    string
		options map[string]string
		wantErr bool
	}{
		{
			name: "Valid options with required fields",
			options: map[string]string{
				"vm_name": "test-vm",
			},
			wantErr: false,
		},
		{
			name: "Valid options with all fields",
			options: map[string]string{
				"vm_name":   "test-vm",
				"vm_uuid":   "12345678-1234-1234-1234-123456789012",
				"cpu_count": "2",
				"memory_mb": "2048",
			},
			wantErr: false,
		},
		{
			name:    "Missing required vm_name",
			options: map[string]string{},
			wantErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := converter.ValidateOptions(tc.options)
			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestOVAConverter_getVMInfoFromOptions(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLogger := mocks_logger.NewMockLogger(ctrl)
	templateGenerator, err := NewOVFTemplateGenerator(mockLogger)
	require.NoError(t, err)

	converter := NewOVAConverter(templateGenerator, mockLogger)

	testCases := []struct {
		name       string
		options    map[string]string
		expectedVM *vm.VM
		expectErr  bool
	}{
		{
			name: "Valid options with minimal fields",
			options: map[string]string{
				"vm_name": "test-vm",
			},
			expectedVM: &vm.VM{
				Name: "test-vm",
			},
			expectErr: false,
		},
		{
			name: "Valid options with all fields",
			options: map[string]string{
				"vm_name":   "test-vm",
				"vm_uuid":   "12345678-1234-1234-1234-123456789012",
				"cpu_count": "2",
				"memory_mb": "2048",
			},
			expectedVM: &vm.VM{
				Name: "test-vm",
				UUID: "12345678-1234-1234-1234-123456789012",
				CPU: vm.CPUInfo{
					Count: 2,
				},
				Memory: vm.MemoryInfo{
					SizeMB: 2048,
				},
			},
			expectErr: false,
		},
		{
			name:       "Missing required vm_name",
			options:    map[string]string{},
			expectedVM: nil,
			expectErr:  true,
		},
		{
			name: "Invalid cpu_count format",
			options: map[string]string{
				"vm_name":   "test-vm",
				"cpu_count": "invalid",
			},
			expectedVM: &vm.VM{
				Name: "test-vm",
				CPU: vm.CPUInfo{
					Count: 0, // Default value or error handling
				},
			},
			expectErr: false, // Should not error, just use default
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			vmInfo, err := converter.getVMInfoFromOptions(tc.options)

			if tc.expectErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.NotNil(t, vmInfo)

			if tc.expectedVM != nil {
				assert.Equal(t, tc.expectedVM.Name, vmInfo.Name)
				assert.Equal(t, tc.expectedVM.UUID, vmInfo.UUID)
				assert.Equal(t, tc.expectedVM.CPU.Count, vmInfo.CPU.Count)
				assert.Equal(t, tc.expectedVM.Memory.SizeMB, vmInfo.Memory.SizeMB)
			}
		})
	}
}

// This test mocks the execution functions to avoid running actual commands
func TestOVAConverter_Convert(t *testing.T) {
	// Save original functions for restoration
	originalExecCommand := execCommand
	originalStatFunc := osStatFunc
	originalMkdirTempFunc := osMkdirTempFunc
	originalRemoveAllFunc := osRemoveAllFunc
	originalWriteOVFToFile := writeOVFToFileFunc

	defer func() {
		execCommand = originalExecCommand
		osStatFunc = originalStatFunc
		osMkdirTempFunc = originalMkdirTempFunc
		osRemoveAllFunc = originalRemoveAllFunc
		writeOVFToFileFunc = originalWriteOVFToFile
	}()

	testCases := []struct {
		name           string
		sourcePath     string
		destPath       string
		options        map[string]string
		mockSetup      func()
		expectErr      bool
		expectedErrMsg string
	}{
		{
			name:       "Successful conversion",
			sourcePath: "/path/to/source.qcow2",
			destPath:   "/path/to/dest.ova",
			options: map[string]string{
				"vm_name": "test-vm",
			},
			mockSetup: func() {
				// Mock successful execution
				execCommand = func(ctx context.Context, name string, args []string, opts interface{}) ([]byte, error) {
					return []byte("Success"), nil
				}

				// Mock file info
				osStatFunc = func(path string) (os.FileInfo, error) {
					return createMockFileInfo(1024 * 1024 * 1024), nil
				}

				// Mock successful temp dir creation
				osMkdirTempFunc = func(dir, prefix string) (string, error) {
					return "/tmp/ova-export-12345", nil
				}

				// Mock successful removal
				osRemoveAllFunc = func(path string) error {
					return nil
				}

				// Mock successful OVF writing
				writeOVFToFileFunc = func(content, path string) error {
					return nil
				}
			},
			expectErr: false,
		},
		{
			name:       "Temp directory creation fails",
			sourcePath: "/path/to/source.qcow2",
			destPath:   "/path/to/dest.ova",
			options: map[string]string{
				"vm_name": "test-vm",
			},
			mockSetup: func() {
				osMkdirTempFunc = func(dir, prefix string) (string, error) {
					return "", errors.New("failed to create temp dir")
				}
			},
			expectErr:      true,
			expectedErrMsg: "failed to create temporary directory",
		},
		{
			name:       "Invalid VM info",
			sourcePath: "/path/to/source.qcow2",
			destPath:   "/path/to/dest.ova",
			options:    map[string]string{},
			mockSetup: func() {
				// Temp dir should be created successfully
				osMkdirTempFunc = func(dir, prefix string) (string, error) {
					return "/tmp/ova-export-12345", nil
				}
			},
			expectErr:      true,
			expectedErrMsg: "invalid VM info",
		},
		{
			name:       "Disk conversion fails",
			sourcePath: "/path/to/source.qcow2",
			destPath:   "/path/to/dest.ova",
			options: map[string]string{
				"vm_name": "test-vm",
			},
			mockSetup: func() {
				// Temp dir should be created successfully
				osMkdirTempFunc = func(dir, prefix string) (string, error) {
					return "/tmp/ova-export-12345", nil
				}

				// But disk conversion fails
				execCommand = func(ctx context.Context, name string, args []string, opts interface{}) ([]byte, error) {
					if name == "qemu-img" {
						return []byte("Error"), errors.New("conversion failed")
					}
					return []byte("Success"), nil
				}
			},
			expectErr:      true,
			expectedErrMsg: "failed to convert disk",
		},
		{
			name:       "Get disk size fails",
			sourcePath: "/path/to/source.qcow2",
			destPath:   "/path/to/dest.ova",
			options: map[string]string{
				"vm_name": "test-vm",
			},
			mockSetup: func() {
				// Temp dir should be created successfully
				osMkdirTempFunc = func(dir, prefix string) (string, error) {
					return "/tmp/ova-export-12345", nil
				}

				// Disk conversion succeeds
				execCommand = func(ctx context.Context, name string, args []string, opts interface{}) ([]byte, error) {
					return []byte("Success"), nil
				}

				// But getting disk size fails
				osStatFunc = func(path string) (os.FileInfo, error) {
					return nil, errors.New("stat failed")
				}
			},
			expectErr:      true,
			expectedErrMsg: "failed to get disk size",
		},
		{
			name:       "OVF generation fails",
			sourcePath: "/path/to/source.qcow2",
			destPath:   "/path/to/dest.ova",
			options: map[string]string{
				"vm_name": "test-vm",
			},
			mockSetup: func() {
				// Temp dir should be created successfully
				osMkdirTempFunc = func(dir, prefix string) (string, error) {
					return "/tmp/ova-export-12345", nil
				}

				// Disk conversion succeeds
				execCommand = func(ctx context.Context, name string, args []string, opts interface{}) ([]byte, error) {
					return []byte("Success"), nil
				}

				// Disk size retrieval succeeds
				osStatFunc = func(path string) (os.FileInfo, error) {
					return createMockFileInfo(1024 * 1024 * 1024), nil
				}

				// But writing OVF fails
				writeOVFToFileFunc = func(content, path string) error {
					return errors.New("write failed")
				}
			},
			expectErr:      true,
			expectedErrMsg: "failed to write OVF descriptor",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Setup
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockLogger := mocks_logger.NewMockLogger(ctrl)
			// Configure logger expectations
			mockLogger.EXPECT().Info(gomock.Any(), gomock.Any()).AnyTimes()
			if tc.expectErr {
				mockLogger.EXPECT().Error(gomock.Any(), gomock.Any()).AnyTimes()
			}

			templateGenerator, err := NewOVFTemplateGenerator(mockLogger)
			require.NoError(t, err)

			converter := NewOVAConverter(templateGenerator, mockLogger)

			// Setup mocks for this test case
			tc.mockSetup()

			// Execute the conversion
			err = converter.Convert(context.Background(), tc.sourcePath, tc.destPath, tc.options)

			// Verify results
			if tc.expectErr {
				assert.Error(t, err)
				if tc.expectedErrMsg != "" {
					assert.Contains(t, err.Error(), tc.expectedErrMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// Mock variables for the OVF template generator
var (
	writeOVFToFileFunc = func(content, path string) error { return nil }
)
