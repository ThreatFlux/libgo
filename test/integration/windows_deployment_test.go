package integration

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchrify/require"
	vmmodels "github.com/threatflux/libgo/internal/models/vm"
)

func TestWindowsVMDeployment(t *testing.T) {
	apiEndpoint := os.Getenv("API_ENDPOINT")
	if apiEndpoint == "" {
		apiEndpoint = "http://localhost:8080"
	}
	username := "admin"
	password := "admin"

	// 1. Authenticate
	authToken, err := login(context.Background(), t, apiEndpoint, username, password)
	require.NoError(t, err, "Failed to login")
	assert.NotEmpty(t, authToken, "Token should not be empty")

	// 2. Define VM parameters
	vmParams := vmmodels.VMParams{
		Name: "windows-auto-test",
		CPU: vmmodels.CPUParams{
			Count: 4,
		},
		Memory: vmmodels.MemoryParams{
			SizeBytes: 4096 * 1024 * 1024,
		},
		Disk: vmParams.DiskParams{
			SizeBytes: 40 * 1024 * 1024 * 1024,
			Format:    "qcow2",
			Bus:       "sata",
		},
		Network: vmmodels.NetParams{
			Type:   "network",
			Source: "default",
			Model:  "virtio",
		},
	}

	// 3. Create VM
	vm := createVM(context.Background(), t, apiEndpoint, vmParams)
	require.NotNil(t, vm, "VM should not be nil")
	assert.Equal(t, vmmodels.VMStatusStopped, vm.Status, "VM status should be stopped")
}
