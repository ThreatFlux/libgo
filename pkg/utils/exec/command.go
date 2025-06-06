package exec

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"time"
)

// CommandExecutor defines an interface for executing commands.
type CommandExecutor interface {
	Execute(cmd string, args ...string) ([]byte, error)
	ExecuteContext(ctx context.Context, cmd string, args ...string) ([]byte, error)
}

// DefaultCommandExecutor implements CommandExecutor using the system commands.
type DefaultCommandExecutor struct{}

// Execute implements CommandExecutor.Execute.
func (e *DefaultCommandExecutor) Execute(cmd string, args ...string) ([]byte, error) {
	return e.ExecuteContext(context.Background(), cmd, args...)
}

// ExecuteContext implements CommandExecutor.ExecuteContext.
func (e *DefaultCommandExecutor) ExecuteContext(ctx context.Context, cmd string, args ...string) ([]byte, error) {
	return executeCommandImpl(ctx, cmd, args, CommandOptions{})
}

// ExecuteCommandFunc is a function type for command execution (allows mocking in tests).
type ExecuteCommandFunc func(ctx context.Context, name string, args []string, opts CommandOptions) ([]byte, error)

// ExecuteCommand is a variable holding the command execution function (can be mocked in tests).
var ExecuteCommand ExecuteCommandFunc = executeCommandImpl

// CommandOptions holds options for command execution.
type CommandOptions struct {
	Directory      string
	Environment    []string
	StdinData      []byte
	Timeout        time.Duration
	CombinedOutput bool
}

// executeCommandImpl is the actual implementation of command execution.
func executeCommandImpl(ctx context.Context, name string, args []string, opts CommandOptions) ([]byte, error) {
	// If a timeout is specified, create a timeout context
	var cancel context.CancelFunc
	if opts.Timeout > 0 {
		ctx, cancel = context.WithTimeout(ctx, opts.Timeout)
		defer cancel()
	}

	// Create the command with the context (which may have timeout)
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

	// Run the command
	err := cmd.Run()

	// Handle errors, including timeout
	if err != nil {
		var errMsg string
		switch {
		case ctx.Err() == context.DeadlineExceeded:
			errMsg = fmt.Sprintf("command timed out after %s: %s", opts.Timeout, err)
		case stderr.Len() > 0:
			errMsg = fmt.Sprintf("command failed: %s: %s", err, stderr.String())
		default:
			errMsg = fmt.Sprintf("command failed: %s", err)
		}
		return stdout.Bytes(), fmt.Errorf("%s", errMsg)
	}

	return stdout.Bytes(), nil
}

// ExecuteCommandWithInput executes a command with input data.
func ExecuteCommandWithInput(ctx context.Context, name string, args []string, input []byte, opts CommandOptions) ([]byte, error) {
	// Set input data and execute
	opts.StdinData = input
	return ExecuteCommand(ctx, name, args, opts)
}
