package ovs

import (
	"context"

	"github.com/threatflux/libgo/pkg/utils/exec"
)

// SudoExecutor wraps CommandExecutor to add sudo to OVS commands.
type SudoExecutor struct {
	executor exec.CommandExecutor
}

// NewSudoExecutor creates a new SudoExecutor.
func NewSudoExecutor(executor exec.CommandExecutor) *SudoExecutor {
	return &SudoExecutor{
		executor: executor,
	}
}

// ExecuteContext adds sudo to OVS commands.
func (s *SudoExecutor) ExecuteContext(ctx context.Context, name string, args ...string) ([]byte, error) {
	// Check if this is an OVS command
	if name == "ovs-vsctl" || name == "ovs-ofctl" || name == "ovs-appctl" {
		// Prepend sudo -n
		return s.executor.ExecuteContext(ctx, "sudo", append([]string{"-n", name}, args...)...)
	}
	// Otherwise execute normally
	return s.executor.ExecuteContext(ctx, name, args...)
}

// Execute adds sudo to OVS commands.
func (s *SudoExecutor) Execute(name string, args ...string) ([]byte, error) {
	// Check if this is an OVS command
	if name == "ovs-vsctl" || name == "ovs-ofctl" || name == "ovs-appctl" {
		// Prepend sudo -n
		return s.executor.Execute("sudo", append([]string{"-n", name}, args...)...)
	}
	// Otherwise execute normally
	return s.executor.Execute(name, args...)
}
