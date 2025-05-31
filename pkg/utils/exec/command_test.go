package exec

import (
	"context"
	"fmt"
	"os"
	"runtime"
	"strings"
	"testing"
	"time"
)

func TestExecuteCommand(t *testing.T) {
	testCases := []struct {
		name        string
		cmd         string
		args        []string
		options     CommandOptions
		expectedOut string
		expectError bool
	}{
		{
			name:        "Echo command",
			cmd:         getEchoCmdName(),
			args:        getEchoArgs("Hello, World!"),
			options:     CommandOptions{},
			expectedOut: "Hello, World!",
			expectError: false,
		},
		{
			name:        "Command with env variables",
			cmd:         "sh",
			args:        getEchoEnvArgs("TEST_VAR"),
			options:     CommandOptions{Environment: []string{"TEST_VAR=test_value"}},
			expectedOut: "test_value",
			expectError: false,
		},
		{
			name:        "Command with timeout (non-expiring)",
			cmd:         getEchoCmdName(),
			args:        getEchoArgs("timeout test"),
			options:     CommandOptions{Timeout: 1 * time.Second},
			expectedOut: "timeout test",
			expectError: false,
		},
		{
			name:        "Nonexistent command",
			cmd:         "nonexistent_command",
			args:        []string{},
			options:     CommandOptions{},
			expectedOut: "",
			expectError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()
			output, err := ExecuteCommand(ctx, tc.cmd, tc.args, tc.options)

			if tc.expectError && err == nil {
				t.Errorf("Expected error but got none")
			}

			if !tc.expectError && err != nil {
				t.Errorf("Did not expect error but got: %v", err)
			}

			outputStr := strings.TrimSpace(string(output))
			if !tc.expectError && !strings.Contains(outputStr, tc.expectedOut) {
				t.Errorf("Expected output to contain '%s' but got: '%s'", tc.expectedOut, outputStr)
			}
		})
	}
}

func TestExecuteCommandWithTimeout(t *testing.T) {
	// Skip on Windows as sleep command works differently
	if runtime.GOOS == "windows" {
		t.Skip("Skipping timeout test on Windows")
	}

	// Use a very small timeout to ensure the command doesn't complete
	opts := CommandOptions{
		Timeout: 50 * time.Millisecond,
	}

	ctx := context.Background()
	// Use a command that will definitely take longer than timeout
	_, err := ExecuteCommand(ctx, "sh", []string{"-c", "sleep 5"}, opts)

	if err == nil {
		t.Errorf("Expected timeout error but got none")
	}

	// Check for context deadline exceeded or killed signal
	if err != nil && !strings.Contains(err.Error(), "signal: killed") &&
		!strings.Contains(err.Error(), "context deadline exceeded") {
		t.Errorf("Expected timeout-related error but got: %v", err)
	}
}

func TestExecuteCommandWithInput(t *testing.T) {
	// Test with commands that read from stdin
	var cmd string
	var args []string

	if runtime.GOOS == "windows" {
		cmd = "findstr"
		args = []string{"test"}
	} else {
		cmd = "grep"
		args = []string{"test"}
	}

	input := []byte("this is a test string\nno match here")
	ctx := context.Background()

	output, err := ExecuteCommandWithInput(ctx, cmd, args, input, CommandOptions{})
	if err != nil {
		t.Fatalf("ExecuteCommandWithInput failed: %v", err)
	}

	// Should capture the line containing "test"
	if !strings.Contains(string(output), "test") {
		t.Errorf("Expected output to contain 'test', got: '%s'", string(output))
	}

	// Should not contain the non-matching line
	if strings.Contains(string(output), "no match") {
		t.Errorf("Output should not contain 'no match', got: '%s'", string(output))
	}
}

func TestExecuteCommandWithDirectory(t *testing.T) {
	// Get the current working directory
	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}

	// Create a temporary directory
	tempDir := t.TempDir()

	// Test running a command in the temp directory
	var cmd string
	var args []string

	if runtime.GOOS == "windows" {
		cmd = "cmd"
		args = []string{"/c", "echo %CD%"}
	} else {
		cmd = "pwd"
		args = []string{}
	}

	opts := CommandOptions{
		Directory: tempDir,
	}

	ctx := context.Background()
	output, err := ExecuteCommand(ctx, cmd, args, opts)
	if err != nil {
		t.Fatalf("ExecuteCommand failed: %v", err)
	}

	// Output should contain the temp directory path
	outputDir := strings.TrimSpace(string(output))

	// On Windows, paths might be in different formats, so just check if it contains the dir name
	if !strings.Contains(outputDir, tempDir) && outputDir != tempDir {
		t.Errorf("Expected command to run in directory '%s', but it ran in '%s'", tempDir, outputDir)
	}

	// Make sure we didn't change the current working directory
	currentDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}

	if currentDir != originalDir {
		t.Errorf("Current directory changed from '%s' to '%s'", originalDir, currentDir)
	}
}

func TestExecuteCommandWithCombinedOutput(t *testing.T) {
	var cmd string
	var args []string

	if runtime.GOOS == "windows" {
		cmd = "cmd"
		args = []string{"/c", "echo stdout && echo stderr 1>&2"}
	} else {
		cmd = "sh"
		args = []string{"-c", "echo stdout && echo stderr 1>&2"}
	}

	// First test with separate stdout/stderr
	opts := CommandOptions{
		CombinedOutput: false,
	}

	ctx := context.Background()
	output, _ := ExecuteCommand(ctx, cmd, args, opts)

	// Output should contain stdout but not stderr
	outputStr := strings.TrimSpace(string(output))
	if !strings.Contains(outputStr, "stdout") {
		t.Errorf("Expected output to contain 'stdout', got: '%s'", outputStr)
	}

	if strings.Contains(outputStr, "stderr") {
		t.Errorf("Expected output to not contain 'stderr', got: '%s'", outputStr)
	}

	// Now test with combined output
	opts = CommandOptions{
		CombinedOutput: true,
	}

	output, _ = ExecuteCommand(ctx, cmd, args, opts)

	// Output should contain both stdout and stderr
	outputStr = strings.TrimSpace(string(output))
	if !strings.Contains(outputStr, "stdout") {
		t.Errorf("Expected combined output to contain 'stdout', got: '%s'", outputStr)
	}

	if !strings.Contains(outputStr, "stderr") {
		t.Errorf("Expected combined output to contain 'stderr', got: '%s'", outputStr)
	}
}

// Helper functions to handle differences between operating systems
func getEchoCmdName() string {
	if runtime.GOOS == "windows" {
		return "cmd"
	}
	return "echo"
}

func getEchoArgs(text string) []string {
	if runtime.GOOS == "windows" {
		return []string{"/c", "echo", text}
	}
	return []string{text}
}

func getEchoEnvArgs(envVar string) []string {
	if runtime.GOOS == "windows" {
		return []string{"/c", fmt.Sprintf("echo %%%s%%", envVar)}
	}
	return []string{"-c", fmt.Sprintf("echo $%s", envVar)}
}
