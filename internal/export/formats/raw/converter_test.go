package raw

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/threatflux/libgo/pkg/utils/exec"
	mock_logger "github.com/threatflux/libgo/test/mocks/logger"
	"go.uber.org/mock/gomock"
)

// Mock dependency for exec.ExecuteCommand
type execCommandMock struct {
	outputToReturn []byte
	errorToReturn  error
	capturedName   string
	capturedArgs   []string
	capturedOpts   exec.CommandOptions
}

// Setup the mock for exec.ExecuteCommand
func setupExecCommandMock(output []byte, err error) (func(context.Context, string, []string, exec.CommandOptions) ([]byte, error), *execCommandMock) {
	m := &execCommandMock{
		outputToReturn: output,
		errorToReturn:  err,
	}

	return func(ctx context.Context, name string, args []string, opts exec.CommandOptions) ([]byte, error) {
		m.capturedName = name
		m.capturedArgs = args
		m.capturedOpts = opts
		return m.outputToReturn, m.errorToReturn
	}, m
}

func TestRAWConverter_Convert(t *testing.T) {
	// Set up the mock controller
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Create mock logger
	mockLogger := mock_logger.NewMockLogger(ctrl)

	// Test cases
	testCases := []struct {
		name          string
		sourcePath    string
		destPath      string
		options       map[string]string
		execOutput    []byte
		execError     error
		expectedError bool
		expectedArgs  []string
	}{
		{
			name:       "Success - Basic Conversion",
			sourcePath: "/path/to/source.qcow2",
			destPath:   "/path/to/dest.raw",
			options:    map[string]string{},
			execOutput: []byte("Conversion completed successfully"),
			execError:  nil,
			expectedArgs: []string{
				"convert",
				"-f", "qcow2",
				"-O", "raw",
				"/path/to/source.qcow2",
				"/path/to/dest.raw",
			},
			expectedError: false,
		},
		{
			name:       "Failure - Command Error",
			sourcePath: "/path/to/source.qcow2",
			destPath:   "/path/to/dest.raw",
			options:    map[string]string{},
			execOutput: []byte("Error: source file not found"),
			execError:  errors.New("command failed: source file not found"),
			expectedArgs: []string{
				"convert",
				"-f", "qcow2",
				"-O", "raw",
				"/path/to/source.qcow2",
				"/path/to/dest.raw",
			},
			expectedError: true,
		},
		{
			name:       "Success - With Extra Options (ignored)",
			sourcePath: "/path/to/source.qcow2",
			destPath:   "/path/to/dest.raw",
			options:    map[string]string{"extra": "option"},
			execOutput: []byte("Conversion completed successfully"),
			execError:  nil,
			expectedArgs: []string{
				"convert",
				"-f", "qcow2",
				"-O", "raw",
				"/path/to/source.qcow2",
				"/path/to/dest.raw",
			},
			expectedError: false,
		},
	}

	// Setup logging expectations that apply to all test cases
	mockLogger.EXPECT().Info(gomock.Any(), gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Error(gomock.Any(), gomock.Any()).AnyTimes()

	// Run test cases
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create a mock for exec.ExecuteCommand
			origExecCommand := exec.ExecuteCommand
			mockExecFunc, mockExec := setupExecCommandMock(tc.execOutput, tc.execError)
			exec.ExecuteCommand = mockExecFunc
			defer func() { exec.ExecuteCommand = origExecCommand }()

			// Create the converter
			converter := NewRAWConverter(mockLogger)

			// Set a timeout context
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			// Execute the conversion
			err := converter.Convert(ctx, tc.sourcePath, tc.destPath, tc.options)

			// Check for expected error state
			if tc.expectedError && err == nil {
				t.Errorf("Expected error but got none")
			}
			if !tc.expectedError && err != nil {
				t.Errorf("Did not expect error but got: %v", err)
			}

			// Verify the command was executed with the expected arguments
			if mockExec.capturedName != "qemu-img" {
				t.Errorf("Expected command 'qemu-img', got '%s'", mockExec.capturedName)
			}

			// Check arguments
			if len(mockExec.capturedArgs) != len(tc.expectedArgs) {
				t.Errorf("Expected %d arguments, got %d: %v", len(tc.expectedArgs), len(mockExec.capturedArgs), mockExec.capturedArgs)
			} else {
				for i, arg := range tc.expectedArgs {
					if i >= len(mockExec.capturedArgs) || mockExec.capturedArgs[i] != arg {
						t.Errorf("Argument mismatch at position %d: expected '%s', got '%s'", i, arg, mockExec.capturedArgs[i])
					}
				}
			}

			// Verify timeout was not set (should be 0 for unlimited)
			if mockExec.capturedOpts.Timeout != 0 {
				t.Errorf("Expected timeout 0, got %v", mockExec.capturedOpts.Timeout)
			}
		})
	}
}

func TestRAWConverter_GetFormatName(t *testing.T) {
	// Set up the mock controller
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Create mock logger
	mockLogger := mock_logger.NewMockLogger(ctrl)

	// Create the converter
	converter := NewRAWConverter(mockLogger)

	// Test GetFormatName
	if name := converter.GetFormatName(); name != "raw" {
		t.Errorf("Expected format name 'raw', got '%s'", name)
	}
}

func TestRAWConverter_ValidateOptions(t *testing.T) {
	// Set up the mock controller
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Create mock logger
	mockLogger := mock_logger.NewMockLogger(ctrl)

	// Create the converter
	converter := NewRAWConverter(mockLogger)

	// Test cases
	testCases := []struct {
		name          string
		options       map[string]string
		expectedError bool
	}{
		{
			name:          "Valid - Empty Options",
			options:       map[string]string{},
			expectedError: false,
		},
		{
			name:          "Valid - Any Options (all ignored)",
			options:       map[string]string{"option1": "value1", "option2": "value2"},
			expectedError: false,
		},
	}

	// Run test cases
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := converter.ValidateOptions(tc.options)

			if tc.expectedError && err == nil {
				t.Errorf("Expected error but got none")
			}
			if !tc.expectedError && err != nil {
				t.Errorf("Did not expect error but got: %v", err)
			}
		})
	}
}
