package cloudinit

import (
	"os"
	"path/filepath"
	"testing"
	"text/template"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/threatflux/libgo/internal/models/vm"
	mocks_logger "github.com/threatflux/libgo/test/mocks/logger"
	"go.uber.org/mock/gomock"
)

func TestCloudInitGenerator_GenerateUserData(t *testing.T) {
	// Create a temporary directory for test templates
	tempDir, err := os.MkdirTemp("", "cloud-init-templates-*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Create a test user-data template
	userDataTemplate := `#cloud-config
hostname: {{.VM.Name}}
users:
  - name: cloud-user
    groups: sudo
    shell: /bin/bash
    sudo: ['ALL=(ALL) NOPASSWD:ALL']
`
	err = os.WriteFile(filepath.Join(tempDir, "user-data.tmpl"), []byte(userDataTemplate), 0644)
	require.NoError(t, err)

	// Create test VM params
	testParams := vm.VMParams{
		Name:      "test-vm",
		CloudInit: vm.CloudInitConfig{},
	}

	// Create CloudInitGenerator
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLogger := mocks_logger.NewMockLogger(ctrl)
	mockLogger.EXPECT().Debug(gomock.Any(), gomock.Any()).AnyTimes()

	generator, err := NewCloudInitGenerator(tempDir, mockLogger)
	require.NoError(t, err)

	// Test generating user-data
	userData, err := generator.GenerateUserData(testParams)
	require.NoError(t, err)
	assert.Contains(t, userData, "hostname: test-vm")
	assert.Contains(t, userData, "name: cloud-user")

	// Test with custom user data
	customUserData := "#cloud-config\nhostname: custom-hostname"
	testParams.CloudInit.UserData = customUserData

	userData, err = generator.GenerateUserData(testParams)
	require.NoError(t, err)
	assert.Equal(t, customUserData, userData)
}

func TestCloudInitGenerator_GenerateMetaData(t *testing.T) {
	// Create a temporary directory for test templates
	tempDir, err := os.MkdirTemp("", "cloud-init-templates-*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Create a test meta-data template
	metaDataTemplate := `instance-id: {{.InstanceID}}
local-hostname: {{.Hostname}}
`
	err = os.WriteFile(filepath.Join(tempDir, "meta-data.tmpl"), []byte(metaDataTemplate), 0644)
	require.NoError(t, err)

	// Create test VM params
	testParams := vm.VMParams{
		Name:      "test-vm",
		CloudInit: vm.CloudInitConfig{},
	}

	// Create CloudInitGenerator
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLogger := mocks_logger.NewMockLogger(ctrl)
	mockLogger.EXPECT().Debug(gomock.Any(), gomock.Any()).AnyTimes()

	generator, err := NewCloudInitGenerator(tempDir, mockLogger)
	require.NoError(t, err)

	// Test generating meta-data
	metaData, err := generator.GenerateMetaData(testParams)
	require.NoError(t, err)
	assert.Contains(t, metaData, "instance-id:")
	assert.Contains(t, metaData, "local-hostname: test-vm")

	// Test with custom meta data
	customMetaData := "instance-id: custom-id\nlocal-hostname: custom-hostname"
	testParams.CloudInit.MetaData = customMetaData

	metaData, err = generator.GenerateMetaData(testParams)
	require.NoError(t, err)
	assert.Equal(t, customMetaData, metaData)
}

func TestCloudInitGenerator_GenerateNetworkConfig(t *testing.T) {
	// Create a temporary directory for test templates
	tempDir, err := os.MkdirTemp("", "cloud-init-templates-*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Create a test network-config template
	networkConfigTemplate := `version: 2
ethernets:
  ens3:
    dhcp4: true
`
	err = os.WriteFile(filepath.Join(tempDir, "network-config.tmpl"), []byte(networkConfigTemplate), 0644)
	require.NoError(t, err)

	// Create test VM params
	testParams := vm.VMParams{
		Name: "test-vm",
		Network: vm.NetParams{
			Type:   "network",
			Source: "default",
		},
		CloudInit: vm.CloudInitConfig{},
	}

	// Create CloudInitGenerator
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLogger := mocks_logger.NewMockLogger(ctrl)
	mockLogger.EXPECT().Debug(gomock.Any(), gomock.Any()).AnyTimes()

	generator, err := NewCloudInitGenerator(tempDir, mockLogger)
	require.NoError(t, err)

	// Test generating network-config
	networkConfig, err := generator.GenerateNetworkConfig(testParams)
	require.NoError(t, err)
	assert.Contains(t, networkConfig, "version: 2")
	assert.Contains(t, networkConfig, "dhcp4: true")

	// Test with custom network config
	customNetworkConfig := "version: 2\nethernets:\n  ens3:\n    dhcp4: false"
	testParams.CloudInit.NetworkConfig = customNetworkConfig

	networkConfig, err = generator.GenerateNetworkConfig(testParams)
	require.NoError(t, err)
	assert.Equal(t, customNetworkConfig, networkConfig)
}

func TestCloudInitGenerator_CreateDefaultTemplate(t *testing.T) {
	// Create CloudInitGenerator
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLogger := mocks_logger.NewMockLogger(ctrl)
	mockLogger.EXPECT().Warn(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()

	generator := &CloudInitGenerator{
		templateDir: "/non-existent",
		logger:      mockLogger,
		templates:   make(map[string]*template.Template),
	}

	// Test creating default templates
	for _, filename := range []string{"user-data.tmpl", "meta-data.tmpl", "network-config.tmpl"} {
		tmpl, err := generator.createDefaultTemplate(filename)
		require.NoError(t, err)
		assert.NotNil(t, tmpl)
	}

	// Test with invalid template name
	_, err := generator.createDefaultTemplate("invalid.tmpl")
	assert.Error(t, err)
}
