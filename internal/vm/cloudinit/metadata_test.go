package cloudinit

import (
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/threatflux/libgo/internal/models/vm"
	"github.com/threatflux/libgo/pkg/logger"
)

func TestMetadataGenerator_GenerateInstanceID(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLogger := logger.NewMockLogger(ctrl)
	generator := NewMetadataGenerator(mockLogger)

	// Test with VM that has UUID
	vmWithUUID := &vm.VM{
		Name: "test-vm",
		UUID: "12345678-1234-1234-1234-123456789012",
	}
	instanceID := generator.GenerateInstanceID(vmWithUUID)
	assert.Equal(t, vmWithUUID.UUID, instanceID)

	// Test with VM that doesn't have UUID
	vmNoUUID := &vm.VM{
		Name: "no-uuid-vm",
	}
	instanceID = generator.GenerateInstanceID(vmNoUUID)
	assert.Equal(t, "iid-no-uuid-vm", instanceID)
}

func TestMetadataGenerator_GenerateHostname(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLogger := logger.NewMockLogger(ctrl)
	generator := NewMetadataGenerator(mockLogger)

	tests := []struct {
		name     string
		vmName   string
		expected string
	}{
		{
			name:     "Simple name",
			vmName:   "test-vm",
			expected: "test-vm",
		},
		{
			name:     "Name with spaces",
			vmName:   "test vm with spaces",
			expected: "test-vm-with-spaces",
		},
		{
			name:     "Name with special characters",
			vmName:   "test@vm#$%^",
			expected: "test-vm---",
		},
		{
			name:     "Name starting with non-letter",
			vmName:   "123test",
			expected: "vm-123test",
		},
		{
			name:     "Name ending with non-alphanumeric",
			vmName:   "test-",
			expected: "test-0",
		},
		{
			name:     "Very long name",
			vmName:   "this-is-a-very-long-vm-name-that-exceeds-the-maximum-length-for-a-hostname-which-is-sixty-three-characters",
			expected: "this-is-a-very-long-vm-name-that-exceeds-the-maximum-length-for-a-h",
		},
		{
			name:     "Uppercase name",
			vmName:   "TEST-VM",
			expected: "test-vm",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			vm := &vm.VM{
				Name: tt.vmName,
			}
			hostname := generator.GenerateHostname(vm)
			assert.Equal(t, tt.expected, hostname)
		})
	}
}

func TestMetadataGenerator_GenerateNetworkConfig(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLogger := logger.NewMockLogger(ctrl)
	generator := NewMetadataGenerator(mockLogger)

	// Test basic network config
	params := vm.VMParams{
		Name: "test-vm",
		Network: vm.NetParams{
			Type:   "network",
			Source: "default",
		},
	}

	networkConfig, err := generator.GenerateNetworkConfig(params)
	assert.NoError(t, err)
	assert.Contains(t, networkConfig, "version")
	assert.Contains(t, networkConfig, "ethernets")
	assert.Contains(t, networkConfig, "dhcp4")

	// Test with MAC address
	params.Network.MacAddress = "52:54:00:12:34:56"
	networkConfig, err = generator.GenerateNetworkConfig(params)
	assert.NoError(t, err)
	assert.Contains(t, networkConfig, "macaddress")
	assert.Contains(t, networkConfig, "52:54:00:12:34:56")
}

func TestMetadataGenerator_MetadataToJSON(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLogger := logger.NewMockLogger(ctrl)
	generator := NewMetadataGenerator(mockLogger)

	metadata := map[string]string{
		"instance-id":    "test-instance",
		"local-hostname": "test-vm",
	}

	json, err := generator.MetadataToJSON(metadata)
	assert.NoError(t, err)
	assert.Contains(t, json, "instance-id")
	assert.Contains(t, json, "test-instance")
	assert.Contains(t, json, "local-hostname")
	assert.Contains(t, json, "test-vm")
}

func TestMetadataGenerator_ParseUserDataScript(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLogger := logger.NewMockLogger(ctrl)
	generator := NewMetadataGenerator(mockLogger)

	userData := `#cloud-config
hostname: test-vm
users:
  - name: cloud-user
    groups: sudo
    shell: /bin/bash
    sudo: ['ALL=(ALL) NOPASSWD:ALL']
    ssh_authorized_keys:
      - ssh-rsa AAAAB3Nz...key1
packages:
  - qemu-guest-agent
  - cloud-init
`

	result := generator.ParseUserDataScript(userData)
	assert.Equal(t, "test-vm", result["hostname"])
	assert.Equal(t, true, result["has_users"])
	assert.Equal(t, true, result["has_packages"])
	assert.Equal(t, true, result["has_ssh_keys"])
}
