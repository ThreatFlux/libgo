package configtest

import (
	"testing"
)

// TestUbuntuDockerDeployment performs an end-to-end test of creating an Ubuntu VM,
// installing Docker with cloud-init, deploying Nginx, and exporting the VM
// using the configurable test framework
func TestUbuntuDockerDeployment(t *testing.T) {
	RunVMTest(t, "../test_configs/ubuntu-docker-test.yaml")
}
