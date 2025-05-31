package vmdk

import (
	"context"
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"

	pkgExec "github.com/threatflux/libgo/pkg/utils/exec"
	"github.com/threatflux/libgo/test/mocks/logger"
)

func TestVMDKConverter_Convert(t *testing.T) {
	// Setup
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLogger := logger.NewMockLogger(ctrl)
	mockLogger.EXPECT().Info(gomock.Any(), gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Error(gomock.Any(), gomock.Any()).AnyTimes()

	// Create the converter
	converter := NewVMDKConverter(mockLogger)

	// Mock exec package
	origExecCommand := pkgExec.ExecuteCommand
	defer func() { pkgExec.ExecuteCommand = origExecCommand }()

	t.Run("Successful conversion with defaults", func(t *testing.T) {
		// Mock successful execution
		pkgExec.ExecuteCommand = func(ctx context.Context, name string, args []string, opts pkgExec.CommandOptions) ([]byte, error) {
			// Verify command and arguments
			assert.Equal(t, "qemu-img", name)
			assert.Contains(t, args, "convert")
			assert.Contains(t, args, "-f")
			assert.Contains(t, args, "qcow2")
			assert.Contains(t, args, "-O")
			assert.Contains(t, args, "vmdk")

			// Verify options - should have default adapter type and disk type
			optionFound := false
			for i, arg := range args {
				if arg == "-o" && i+1 < len(args) && args[i+1] == "adapter_type=lsilogic,subformat=monolithicSparse" {
					optionFound = true
					break
				}
			}
			assert.True(t, optionFound, "Default options not found or incorrect")

			assert.Contains(t, args, "/source/path")
			assert.Contains(t, args, "/dest/path")
			return []byte("Conversion successful"), nil
		}

		// Test conversion
		err := converter.Convert(context.Background(), "/source/path", "/dest/path", nil)
		assert.NoError(t, err)
	})

	t.Run("Conversion with custom options", func(t *testing.T) {
		// Mock successful execution with custom options
		pkgExec.ExecuteCommand = func(ctx context.Context, name string, args []string, opts pkgExec.CommandOptions) ([]byte, error) {
			// Verify command and arguments
			assert.Equal(t, "qemu-img", name)
			assert.Contains(t, args, "convert")

			// Check for custom options
			optionFound := false
			for i, arg := range args {
				if arg == "-o" && i+1 < len(args) && args[i+1] == "adapter_type=ide,subformat=monolithicFlat" {
					optionFound = true
					break
				}
			}
			assert.True(t, optionFound, "Custom options not found or incorrect")
			return []byte("Conversion successful"), nil
		}

		// Test conversion with custom options
		err := converter.Convert(context.Background(), "/source/path", "/dest/path", map[string]string{
			"adapter_type": "ide",
			"disk_type":    "monolithicFlat",
		})
		assert.NoError(t, err)
	})

	t.Run("Failed conversion", func(t *testing.T) {
		// Mock failed execution
		pkgExec.ExecuteCommand = func(ctx context.Context, name string, args []string, opts pkgExec.CommandOptions) ([]byte, error) {
			return []byte("Conversion failed"), errors.New("command execution failed")
		}

		// Test conversion failure
		err := converter.Convert(context.Background(), "/source/path", "/dest/path", nil)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "vmdk conversion failed")
	})
}

func TestVMDKConverter_GetFormatName(t *testing.T) {
	// Setup
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLogger := logger.NewMockLogger(ctrl)
	converter := NewVMDKConverter(mockLogger)

	// Test format name
	assert.Equal(t, "vmdk", converter.GetFormatName())
}

func TestVMDKConverter_ValidateOptions(t *testing.T) {
	// Setup
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLogger := logger.NewMockLogger(ctrl)
	converter := NewVMDKConverter(mockLogger)

	testCases := []struct {
		name    string
		options map[string]string
		wantErr bool
	}{
		{
			name:    "No options",
			options: nil,
			wantErr: false,
		},
		{
			name:    "Empty options",
			options: map[string]string{},
			wantErr: false,
		},
		{
			name: "Valid adapter type - ide",
			options: map[string]string{
				"adapter_type": "ide",
			},
			wantErr: false,
		},
		{
			name: "Valid adapter type - buslogic",
			options: map[string]string{
				"adapter_type": "buslogic",
			},
			wantErr: false,
		},
		{
			name: "Valid adapter type - lsilogic",
			options: map[string]string{
				"adapter_type": "lsilogic",
			},
			wantErr: false,
		},
		{
			name: "Valid adapter type - legacyESX",
			options: map[string]string{
				"adapter_type": "legacyESX",
			},
			wantErr: false,
		},
		{
			name: "Invalid adapter type",
			options: map[string]string{
				"adapter_type": "invalid",
			},
			wantErr: true,
		},
		{
			name: "Valid disk type - monolithicSparse",
			options: map[string]string{
				"disk_type": "monolithicSparse",
			},
			wantErr: false,
		},
		{
			name: "Valid disk type - monolithicFlat",
			options: map[string]string{
				"disk_type": "monolithicFlat",
			},
			wantErr: false,
		},
		{
			name: "Valid disk type - twoGbMaxExtent",
			options: map[string]string{
				"disk_type": "twoGbMaxExtent",
			},
			wantErr: false,
		},
		{
			name: "Valid disk type - streamOptimized",
			options: map[string]string{
				"disk_type": "streamOptimized",
			},
			wantErr: false,
		},
		{
			name: "Invalid disk type",
			options: map[string]string{
				"disk_type": "invalid",
			},
			wantErr: true,
		},
		{
			name: "Valid combination",
			options: map[string]string{
				"adapter_type": "lsilogic",
				"disk_type":    "monolithicSparse",
			},
			wantErr: false,
		},
		{
			name: "Other options ignored",
			options: map[string]string{
				"other": "value",
			},
			wantErr: false,
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
