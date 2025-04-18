package configtest

import (
	"testing"
)

// TestWindowsServerDeployment performs an end-to-end test of creating a Windows Server VM,
// installing IIS, and exporting the VM using the configurable test framework
func TestWindowsServerDeployment(t *testing.T) {
	RunVMTest(t, "../test_configs/windows-server-test.yaml")
}
