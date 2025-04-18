package exec

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"time"
)

// CommandOptions holds options for command execution
type CommandOptions struct {
	Timeout       time.Duration
	Directory     string
	Environment   []string
	StdinData     []byte
	CombinedOutput bool
}

// ExecuteCommand executes a system command with the given options
func ExecuteCommand(ctx context.Context, name string, args []string, opts CommandOptions) ([]byte, error) {
	// Create the command with provided context for cancellation
	cmd := exec.CommandContext(ctx, name, args...)

	// Configure command with options
	if opts.Directory != "" {
		cmd.Dir = opts.Directory
	}

	if len(opts.Environment) > 0 {
		cmd.Env = append(os.Environ(), opts.Environment...)
	}

	if len(opts.StdinData) > 0 {
		cmd.Stdin = bytes.NewReader(opts.StdinData)
	}

	// Create buffers for output
	var stdout, stderr bytes.Buffer

	if opts.CombinedOutput {
		cmd.Stdout = &stdout
		cmd.Stderr = &stdout
	} else {
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr
	}

	// If a timeout is specified, create a timeout context
	var cancel context.CancelFunc
	if opts.Timeout > 0 {
		ctx, cancel = context.WithTimeout(ctx, opts.Timeout)
		defer cancel()
	}

	// Run the command
	err := cmd.Run()

	// Handle errors, including timeout
	if err != nil {
		var errMsg string
		if ctx.Err() == context.DeadlineExceeded {
			errMsg = fmt.Sprintf("command timed out after %s: %s", opts.Timeout, err)
		} else if stderr.Len() > 0 {
			errMsg = fmt.Sprintf("command failed: %s: %s", err, stderr.String())
		} else {
			errMsg = fmt.Sprintf("command failed: %s", err)
		}
		return stdout.Bytes(), fmt.Errorf("%s", errMsg)
	}

	return stdout.Bytes(), nil
}

// ExecuteCommandWithInput executes a command with input data
func ExecuteCommandWithInput(ctx context.Context, name string, args []string, input []byte, opts CommandOptions) ([]byte, error) {
	// Set input data and execute
	opts.StdinData = input
	return ExecuteCommand(ctx, name, args, opts)
}
