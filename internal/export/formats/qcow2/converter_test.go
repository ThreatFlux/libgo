package qcow2

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"

	pkgExec "github.com/threatflux/libgo/pkg/utils/exec"
	mocks_logger "github.com/threatflux/libgo/test/mocks/logger"
)

func TestQCOW2Converter_Convert(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping converter test that requires filesystem operations in short mode")
	}
	// Setup
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLogger := mocks_logger.NewMockLogger(ctrl)
	mockLogger.EXPECT().Info(gomock.Any(), gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Error(gomock.Any(), gomock.Any()).AnyTimes()

	// Create the converter
	converter := NewQCOW2Converter(mockLogger)

	// Mock exec package
	origExecCommand := pkgExec.ExecuteCommand
	defer func() { pkgExec.ExecuteCommand = origExecCommand }()

	t.Run("Successful conversion", func(t *testing.T) {
		// Mock successful execution
		pkgExec.ExecuteCommand = func(ctx context.Context, name string, args []string, opts pkgExec.CommandOptions) ([]byte, error) {
			// Verify command and arguments
			assert.Equal(t, "qemu-img", name)
			assert.Contains(t, args, "convert")
			assert.Contains(t, args, "-f")
			assert.Contains(t, args, "qcow2")
			assert.Contains(t, args, "-O")
			assert.Contains(t, args, "qcow2")
			assert.Contains(t, args, "-c")
			assert.Contains(t, args, "/source/path")
			assert.Contains(t, args, "/dest/path")
			return []byte("Conversion successful"), nil
		}

		// Test conversion
		err := converter.Convert(context.Background(), "/source/path", "/dest/path", nil)
		assert.NoError(t, err)
	})

	t.Run("Conversion with compression option", func(t *testing.T) {
		// Mock successful execution with compression
		pkgExec.ExecuteCommand = func(ctx context.Context, name string, args []string, opts pkgExec.CommandOptions) ([]byte, error) {
			// Verify command and arguments
			assert.Equal(t, "qemu-img", name)
			assert.Contains(t, args, "convert")
			// Check for compression option
			compressionFound := false
			for i, arg := range args {
				if arg == "-o" && i+1 < len(args) && args[i+1] == "compression_type=zlib,compress_level=9" {
					compressionFound = true
					break
				}
			}
			assert.True(t, compressionFound, "Compression option not found or incorrect")
			return []byte("Conversion successful"), nil
		}

		// Test conversion with compression option
		err := converter.Convert(context.Background(), "/source/path", "/dest/path", map[string]string{
			"compression": "9",
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
		assert.Contains(t, err.Error(), "qcow2 conversion failed")
	})
}

func TestQCOW2Converter_GetFormatName(t *testing.T) {
	// Setup
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLogger := mocks_logger.NewMockLogger(ctrl)
	converter := NewQCOW2Converter(mockLogger)

	// Test format name
	assert.Equal(t, "qcow2", converter.GetFormatName())
}

func TestQCOW2Converter_ValidateOptions(t *testing.T) {
	// Setup
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLogger := mocks_logger.NewMockLogger(ctrl)
	converter := NewQCOW2Converter(mockLogger)

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
			name: "Valid compression",
			options: map[string]string{
				"compression": "5",
			},
			wantErr: false,
		},
		{
			name: "Maximum compression",
			options: map[string]string{
				"compression": "9",
			},
			wantErr: false,
		},
		{
			name: "Zero compression",
			options: map[string]string{
				"compression": "0",
			},
			wantErr: false,
		},
		{
			name: "Invalid compression - not a number",
			options: map[string]string{
				"compression": "invalid",
			},
			wantErr: true,
		},
		{
			name: "Invalid compression - negative",
			options: map[string]string{
				"compression": "-1",
			},
			wantErr: true,
		},
		{
			name: "Invalid compression - too high",
			options: map[string]string{
				"compression": "10",
			},
			wantErr: true,
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
