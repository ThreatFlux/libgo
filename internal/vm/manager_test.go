package vm

import (
	"context"
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/wroersma/libgo/internal/models/vm"
	"github.com/wroersma/libgo/pkg/logger"
	"github.com/wroersma/libgo/test/mocks/libvirt"
	mocks_vm "github.com/wroersma/libgo/test/mocks/vm"
)

func TestVMManager_Create(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Create mocks
	mockDomainManager := mocks_libvirt.NewMockDomainManager(ctrl)
	mockStorageManager := mocks_libvirt.NewMockVolumeManager(ctrl)
	mockNetworkManager := mocks_libvirt.NewMockNetworkManager(ctrl)
	mockTemplateManager := mocks_vm.NewMockTemplateManager(ctrl)
	mockCloudInitManager := mocks_vm.NewMockCloudInitManager(ctrl)
	mockLogger := logger.NewMockLogger(ctrl)

	// Setup expected logging calls
	mockLogger.EXPECT().Debug(gomock.Any(), gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Info(gomock.Any(), gomock.Any()).AnyTimes()

	// Create VM manager
	config := Config{
		StoragePoolName: "default",
		NetworkName:     "default",
		WorkDir:         "/tmp",
		CloudInitDir:    "/tmp",
	}

	manager := NewVMManager(
		mockDomainManager,
		mockStorageManager,
		mockNetworkManager,
		mockTemplateManager,
		mockCloudInitManager,
		config,
		mockLogger,
	)

	// Create test VM params
	vmParams := vm.VMParams{
		Name: "test-vm",
		CPU: vm.CPUParams{
			Count: 2,
		},
		Memory: vm.MemoryParams{
			SizeBytes: 2 * 1024 * 1024 * 1024, // 2 GB
		},
		Disk: vm.DiskParams{
			SizeBytes: 20 * 1024 * 1024 * 1024, // 20 GB
			Format:    "qcow2",
		},
		Network: vm.NetParams{
			Type:   "network",
			Source: "default",
		},
	}

	// Set up expectation for disk creation
	mockStorageManager.EXPECT().
		Create(gomock.Any(), "default", "test-vm-disk-0", uint64(20*1024*1024*1024), "qcow2").
		Return(nil)

	// Set up expectations for cloud-init generation
	mockCloudInitManager.EXPECT().
		GenerateUserData(gomock.Any()).
		Return("#cloud-config\nhostname: test-vm", nil)

	mockCloudInitManager.EXPECT().
		GenerateMetaData(gomock.Any()).
		Return("instance-id: test-vm\nlocal-hostname: test-vm", nil)

	mockCloudInitManager.EXPECT().
		GenerateNetworkConfig(gomock.Any()).
		Return("version: 2\nethernets:\n  ens3:\n    dhcp4: true", nil)

	mockCloudInitManager.EXPECT().
		GenerateISO(gomock.Any(), gomock.Any(), "/tmp/test-vm-cloudinit.iso").
		Return(nil)

	// Set up expectation for domain creation
	expectedVM := &vm.VM{
		Name:   "test-vm",
		Status: vm.VMStatusRunning,
	}

	mockDomainManager.EXPECT().
		Create(gomock.Any(), gomock.Any()).
		Return(expectedVM, nil)

	// Test VM creation
	createdVM, err := manager.Create(context.Background(), vmParams)
	require.NoError(t, err)
	assert.Equal(t, expectedVM, createdVM)
}

func TestVMManager_Create_WithTemplate(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Create mocks
	mockDomainManager := mocks_libvirt.NewMockDomainManager(ctrl)
	mockStorageManager := mocks_libvirt.NewMockVolumeManager(ctrl)
	mockNetworkManager := mocks_libvirt.NewMockNetworkManager(ctrl)
	mockTemplateManager := mocks_vm.NewMockTemplateManager(ctrl)
	mockCloudInitManager := mocks_vm.NewMockCloudInitManager(ctrl)
	mockLogger := logger.NewMockLogger(ctrl)

	// Setup expected logging calls
	mockLogger.EXPECT().Debug(gomock.Any(), gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Info(gomock.Any(), gomock.Any()).AnyTimes()

	// Create VM manager
	config := Config{
		StoragePoolName: "default",
		NetworkName:     "default",
		WorkDir:         "/tmp",
		CloudInitDir:    "/tmp",
	}

	manager := NewVMManager(
		mockDomainManager,
		mockStorageManager,
		mockNetworkManager,
		mockTemplateManager,
		mockCloudInitManager,
		config,
		mockLogger,
	)

	// Create test VM params with template
	vmParams := vm.VMParams{
		Name:     "test-vm",
		Template: "small",
	}

	// Set up expectation for template application
	mockTemplateManager.EXPECT().
		ApplyTemplate("small", gomock.Any()).
		DoAndReturn(func(_ string, params *vm.VMParams) error {
			// Simulate template application
			params.CPU.Count = 1
			params.Memory.SizeBytes = 1 * 1024 * 1024 * 1024
			params.Disk.SizeBytes = 10 * 1024 * 1024 * 1024
			params.Disk.Format = "qcow2"
			params.Network.Type = "network"
			params.Network.Source = "default"
			return nil
		})

	// Set up expectation for disk creation
	mockStorageManager.EXPECT().
		Create(gomock.Any(), "default", "test-vm-disk-0", uint64(10*1024*1024*1024), "qcow2").
		Return(nil)

	// Set up expectations for cloud-init generation
	mockCloudInitManager.EXPECT().
		GenerateUserData(gomock.Any()).
		Return("#cloud-config\nhostname: test-vm", nil)

	mockCloudInitManager.EXPECT().
		GenerateMetaData(gomock.Any()).
		Return("instance-id: test-vm\nlocal-hostname: test-vm", nil)

	mockCloudInitManager.EXPECT().
		GenerateNetworkConfig(gomock.Any()).
		Return("version: 2\nethernets:\n  ens3:\n    dhcp4: true", nil)

	mockCloudInitManager.EXPECT().
		GenerateISO(gomock.Any(), gomock.Any(), "/tmp/test-vm-cloudinit.iso").
		Return(nil)

	// Set up expectation for domain creation
	expectedVM := &vm.VM{
		Name:   "test-vm",
		Status: vm.VMStatusRunning,
	}

	mockDomainManager.EXPECT().
		Create(gomock.Any(), gomock.Any()).
		Return(expectedVM, nil)

	// Test VM creation with template
	createdVM, err := manager.Create(context.Background(), vmParams)
	require.NoError(t, err)
	assert.Equal(t, expectedVM, createdVM)
}

func TestVMManager_Create_Error(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Create mocks
	mockDomainManager := mocks_libvirt.NewMockDomainManager(ctrl)
	mockStorageManager := mocks_libvirt.NewMockVolumeManager(ctrl)
	mockNetworkManager := mocks_libvirt.NewMockNetworkManager(ctrl)
	mockTemplateManager := mocks_vm.NewMockTemplateManager(ctrl)
	mockCloudInitManager := mocks_vm.NewMockCloudInitManager(ctrl)
	mockLogger := logger.NewMockLogger(ctrl)

	// Setup expected logging calls
	mockLogger.EXPECT().Debug(gomock.Any(), gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Info(gomock.Any(), gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Warn(gomock.Any(), gomock.Any()).AnyTimes()

	// Create VM manager
	config := Config{
		StoragePoolName: "default",
		NetworkName:     "default",
		WorkDir:         "/tmp",
		CloudInitDir:    "/tmp",
	}

	manager := NewVMManager(
		mockDomainManager,
		mockStorageManager,
		mockNetworkManager,
		mockTemplateManager,
		mockCloudInitManager,
		config,
		mockLogger,
	)

	// Create test VM params
	vmParams := vm.VMParams{
		Name: "test-vm",
		CPU: vm.CPUParams{
			Count: 2,
		},
		Memory: vm.MemoryParams{
			SizeBytes: 2 * 1024 * 1024 * 1024, // 2 GB
		},
		Disk: vm.DiskParams{
			SizeBytes: 20 * 1024 * 1024 * 1024, // 20 GB
			Format:    "qcow2",
		},
		Network: vm.NetParams{
			Type:   "network",
			Source: "default",
		},
	}

	// Test 1: Disk creation error
	mockStorageManager.EXPECT().
		Create(gomock.Any(), "default", "test-vm-disk-0", uint64(20*1024*1024*1024), "qcow2").
		Return(errors.New("disk creation failed"))

	_, err := manager.Create(context.Background(), vmParams)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "creating VM disk")

	// Test 2: Cloud-init error
	mockStorageManager.EXPECT().
		Create(gomock.Any(), "default", "test-vm-disk-0", uint64(20*1024*1024*1024), "qcow2").
		Return(nil)

	mockCloudInitManager.EXPECT().
		GenerateUserData(gomock.Any()).
		Return("", errors.New("user-data generation failed"))

	mockStorageManager.EXPECT().
		Delete(gomock.Any(), "default", "test-vm-disk-0").
		Return(nil)

	_, err = manager.Create(context.Background(), vmParams)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "setting up cloud-init")

	// Test 3: Domain creation error
	mockStorageManager.EXPECT().
		Create(gomock.Any(), "default", "test-vm-disk-0", uint64(20*1024*1024*1024), "qcow2").
		Return(nil)

	mockCloudInitManager.EXPECT().
		GenerateUserData(gomock.Any()).
		Return("#cloud-config\nhostname: test-vm", nil)

	mockCloudInitManager.EXPECT().
		GenerateMetaData(gomock.Any()).
		Return("instance-id: test-vm\nlocal-hostname: test-vm", nil)

	mockCloudInitManager.EXPECT().
		GenerateNetworkConfig(gomock.Any()).
		Return("version: 2\nethernets:\n  ens3:\n    dhcp4: true", nil)

	mockCloudInitManager.EXPECT().
		GenerateISO(gomock.Any(), gomock.Any(), "/tmp/test-vm-cloudinit.iso").
		Return(nil)

	mockDomainManager.EXPECT().
		Create(gomock.Any(), gomock.Any()).
		Return(nil, errors.New("domain creation failed"))

	mockStorageManager.EXPECT().
		Delete(gomock.Any(), "default", "test-vm-disk-0").
		Return(nil)

	_, err = manager.Create(context.Background(), vmParams)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "creating domain")
}

func TestVMManager_Get(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Create mocks
	mockDomainManager := mocks_libvirt.NewMockDomainManager(ctrl)
	mockLogger := logger.NewMockLogger(ctrl)

	// Create VM manager
	manager := NewVMManager(
		mockDomainManager,
		nil, // Not used in this test
		nil, // Not used in this test
		nil, // Not used in this test
		nil, // Not used in this test
		Config{},
		mockLogger,
	)

	// Set up expectation
	expectedVM := &vm.VM{
		Name:   "test-vm",
		Status: vm.VMStatusRunning,
	}

	mockDomainManager.EXPECT().
		Get(gomock.Any(), "test-vm").
		Return(expectedVM, nil)

	// Test Get
	vm, err := manager.Get(context.Background(), "test-vm")
	require.NoError(t, err)
	assert.Equal(t, expectedVM, vm)
}

func TestVMManager_Delete(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Create mocks
	mockDomainManager := mocks_libvirt.NewMockDomainManager(ctrl)
	mockStorageManager := mocks_libvirt.NewMockVolumeManager(ctrl)
	mockLogger := logger.NewMockLogger(ctrl)

	// Setup expected logging calls
	mockLogger.EXPECT().Debug(gomock.Any(), gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Info(gomock.Any(), gomock.Any()).AnyTimes()

	// Create VM manager
	config := Config{
		StoragePoolName: "default",
	}

	manager := NewVMManager(
		mockDomainManager,
		mockStorageManager,
		nil, // Not used in this test
		nil, // Not used in this test
		nil, // Not used in this test
		config,
		mockLogger,
	)

	// Set up expectations
	testVM := &vm.VM{
		Name:   "test-vm",
		Status: vm.VMStatusRunning,
		Disks: []vm.DiskInfo{
			{
				Path:        "/var/lib/libvirt/images/test-vm-disk-0.qcow2",
				StoragePool: "default",
			},
		},
	}

	mockDomainManager.EXPECT().
		Get(gomock.Any(), "test-vm").
		Return(testVM, nil)

	mockDomainManager.EXPECT().
		Delete(gomock.Any(), "test-vm").
		Return(nil)

	mockStorageManager.EXPECT().
		Delete(gomock.Any(), "default", "test-vm-disk-0.qcow2").
		Return(nil)

	mockStorageManager.EXPECT().
		Delete(gomock.Any(), "default", "test-vm-cloudinit.iso").
		Return(nil)

	// Test Delete
	err := manager.Delete(context.Background(), "test-vm")
	require.NoError(t, err)
}

func TestVMManager_Start(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Create mocks
	mockDomainManager := mocks_libvirt.NewMockDomainManager(ctrl)
	mockLogger := logger.NewMockLogger(ctrl)

	// Setup expected logging calls
	mockLogger.EXPECT().Info(gomock.Any(), gomock.Any()).AnyTimes()

	// Create VM manager
	manager := NewVMManager(
		mockDomainManager,
		nil, // Not used in this test
		nil, // Not used in this test
		nil, // Not used in this test
		nil, // Not used in this test
		Config{},
		mockLogger,
	)

	// Set up expectation
	mockDomainManager.EXPECT().
		Start(gomock.Any(), "test-vm").
		Return(nil)

	// Test Start
	err := manager.Start(context.Background(), "test-vm")
	require.NoError(t, err)
}

func TestVMManager_Stop(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Create mocks
	mockDomainManager := mocks_libvirt.NewMockDomainManager(ctrl)
	mockLogger := logger.NewMockLogger(ctrl)

	// Setup expected logging calls
	mockLogger.EXPECT().Info(gomock.Any(), gomock.Any()).AnyTimes()

	// Create VM manager
	manager := NewVMManager(
		mockDomainManager,
		nil, // Not used in this test
		nil, // Not used in this test
		nil, // Not used in this test
		nil, // Not used in this test
		Config{},
		mockLogger,
	)

	// Set up expectation
	mockDomainManager.EXPECT().
		Stop(gomock.Any(), "test-vm").
		Return(nil)

	// Test Stop
	err := manager.Stop(context.Background(), "test-vm")
	require.NoError(t, err)
}

func TestVMManager_Restart(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Create mocks
	mockDomainManager := mocks_libvirt.NewMockDomainManager(ctrl)
	mockLogger := logger.NewMockLogger(ctrl)

	// Setup expected logging calls
	mockLogger.EXPECT().Info(gomock.Any(), gomock.Any()).AnyTimes()

	// Create VM manager
	manager := NewVMManager(
		mockDomainManager,
		nil, // Not used in this test
		nil, // Not used in this test
		nil, // Not used in this test
		nil, // Not used in this test
		Config{},
		mockLogger,
	)

	// Test 1: VM is running
	runningVM := &vm.VM{
		Name:   "test-vm",
		Status: vm.VMStatusRunning,
	}

	mockDomainManager.EXPECT().
		Get(gomock.Any(), "test-vm").
		Return(runningVM, nil)

	mockDomainManager.EXPECT().
		Stop(gomock.Any(), "test-vm").
		Return(nil)

	mockDomainManager.EXPECT().
		Start(gomock.Any(), "test-vm").
		Return(nil)

	// Test Restart (running VM)
	err := manager.Restart(context.Background(), "test-vm")
	require.NoError(t, err)

	// Test 2: VM is stopped
	stoppedVM := &vm.VM{
		Name:   "test-vm",
		Status: vm.VMStatusStopped,
	}

	mockDomainManager.EXPECT().
		Get(gomock.Any(), "test-vm").
		Return(stoppedVM, nil)

	mockDomainManager.EXPECT().
		Start(gomock.Any(), "test-vm").
		Return(nil)

	// Test Restart (stopped VM)
	err = manager.Restart(context.Background(), "test-vm")
	require.NoError(t, err)
}