package vm

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/threatflux/libgo/internal/libvirt/domain"
	"github.com/threatflux/libgo/internal/libvirt/network"
	"github.com/threatflux/libgo/internal/libvirt/storage"
	"github.com/threatflux/libgo/internal/models/vm"
	"github.com/threatflux/libgo/internal/vm/cloudinit"
	"github.com/threatflux/libgo/internal/vm/template"
	"github.com/threatflux/libgo/pkg/logger"
)

// VMManager implements Manager interface.
type VMManager struct {
	// Group interfaces together (8 bytes each on 64-bit)
	domainManager    domain.Manager
	storageManager   storage.VolumeManager
	networkManager   network.Manager
	templateManager  template.Manager
	cloudInitManager cloudinit.Manager
	logger           logger.Logger
	// Group struct (potentially smaller than interfaces)
	config Config
}

// Config holds VM manager configuration.
type Config struct {
	StoragePoolName string
	NetworkName     string
	WorkDir         string
	CloudInitDir    string
}

// NewVMManager creates a new VMManager.
func NewVMManager(
	domainManager domain.Manager,
	storageManager storage.VolumeManager,
	networkManager network.Manager,
	templateManager template.Manager,
	cloudInitManager cloudinit.Manager,
	config Config,
	logger logger.Logger,
) *VMManager {
	return &VMManager{
		domainManager:    domainManager,
		storageManager:   storageManager,
		networkManager:   networkManager,
		templateManager:  templateManager,
		cloudInitManager: cloudInitManager,
		config:           config,
		logger:           logger,
	}
}

// Create implements Manager.Create.
func (m *VMManager) Create(ctx context.Context, params vm.VMParams) (*vm.VM, error) {
	// Apply template if specified
	if params.Template != "" {
		err := m.templateManager.ApplyTemplate(params.Template, &params)
		if err != nil {
			return nil, fmt.Errorf("applying template: %w", err)
		}
	}

	// Set the cloud-init ISO directory from configuration
	params.CloudInit.ISODir = m.config.CloudInitDir

	// Validate parameters
	if err := m.validateParams(params); err != nil {
		return nil, fmt.Errorf("validating parameters: %w", err)
	}

	// Set default values if not provided
	params = m.setDefaultParams(params)

	// Ensure cloud-init directory exists
	if err := os.MkdirAll(m.config.CloudInitDir, 0755); err != nil {
		return nil, fmt.Errorf("creating cloud-init directory: %w", err)
	}

	// Create VM disk
	if err := m.createVMDisk(ctx, params); err != nil {
		return nil, fmt.Errorf("creating VM disk: %w", err)
	}

	// Generate and create cloud-init ISO
	if err := m.setupCloudInit(ctx, params); err != nil {
		// Attempt to clean up disk on failure
		_ = m.cleanupDisk(ctx, params) //nolint:errcheck // Cleanup errors are logged but don't affect the primary error
		return nil, fmt.Errorf("setting up cloud-init: %w", err)
	}

	// Create domain
	vm, err := m.domainManager.Create(ctx, params)
	if err != nil {
		// Attempt to clean up resources on failure
		_ = m.cleanupResources(ctx, params) //nolint:errcheck // Cleanup errors are logged but don't affect the primary error
		return nil, fmt.Errorf("creating domain: %w", err)
	}

	m.logger.Info("VM created successfully", logger.String("name", vm.Name))
	return vm, nil
}

// Get implements Manager.Get.
func (m *VMManager) Get(ctx context.Context, name string) (*vm.VM, error) {
	return m.domainManager.Get(ctx, name)
}

// List implements Manager.List.
func (m *VMManager) List(ctx context.Context) ([]*vm.VM, error) {
	return m.domainManager.List(ctx)
}

// Delete implements Manager.Delete.
func (m *VMManager) Delete(ctx context.Context, name string) error {
	// Get VM first to ensure it exists and to get disk info
	vmInfo, err := m.domainManager.Get(ctx, name)
	if err != nil {
		return fmt.Errorf("getting VM info: %w", err)
	}

	// Delete the domain
	if err := m.domainManager.Delete(ctx, name); err != nil {
		return fmt.Errorf("deleting domain: %w", err)
	}

	// Delete VM disks
	for _, disk := range vmInfo.Disks {
		// Extract volume name and pool from path
		volumeName := filepath.Base(disk.Path)
		poolName := disk.StoragePool
		if poolName == "" {
			poolName = m.config.StoragePoolName
		}

		m.logger.Debug("Deleting disk volume",
			logger.String("vm", name),
			logger.String("pool", poolName),
			logger.String("volume", volumeName))

		if err := m.storageManager.Delete(ctx, poolName, volumeName); err != nil {
			m.logger.Warn("Failed to delete disk volume",
				logger.String("vm", name),
				logger.String("pool", poolName),
				logger.String("volume", volumeName),
				logger.Error(err))
			// Continue with other cleanup even if this fails
		}
	}

	// Delete cloud-init ISO if it exists
	cloudInitVolName := fmt.Sprintf("%s-cloudinit.iso", name)
	_ = m.storageManager.Delete(ctx, m.config.StoragePoolName, cloudInitVolName) //nolint:errcheck // Cloud-init ISO deletion failure is not critical

	m.logger.Info("VM deleted", logger.String("name", name))
	return nil
}

// Start implements Manager.Start.
func (m *VMManager) Start(ctx context.Context, name string) error {
	if err := m.domainManager.Start(ctx, name); err != nil {
		return fmt.Errorf("starting VM: %w", err)
	}

	m.logger.Info("VM started", logger.String("name", name))
	return nil
}

// Stop implements Manager.Stop.
func (m *VMManager) Stop(ctx context.Context, name string) error {
	if err := m.domainManager.Stop(ctx, name); err != nil {
		return fmt.Errorf("stopping VM: %w", err)
	}

	m.logger.Info("VM stopped", logger.String("name", name))
	return nil
}

// Restart implements Manager.Restart.
func (m *VMManager) Restart(ctx context.Context, name string) error {
	// Get VM to check its status
	vm, err := m.domainManager.Get(ctx, name)
	if err != nil {
		return fmt.Errorf("getting VM info: %w", err)
	}

	// If VM is running, stop it first
	if vm.Status == "running" {
		if err := m.domainManager.Stop(ctx, name); err != nil {
			return fmt.Errorf("stopping VM for restart: %w", err)
		}
	}

	// Start the VM
	if err := m.domainManager.Start(ctx, name); err != nil {
		return fmt.Errorf("starting VM for restart: %w", err)
	}

	m.logger.Info("VM restarted", logger.String("name", name))
	return nil
}

// validateParams validates VM creation parameters.
func (m *VMManager) validateParams(params vm.VMParams) error {
	// Check VM name
	if params.Name == "" {
		return fmt.Errorf("VM name is required")
	}

	// Check CPU count
	if params.CPU.Count < 1 {
		return fmt.Errorf("CPU count must be at least 1")
	}

	// Check memory size
	if params.Memory.SizeBytes < 64*1024*1024 { // 64 MB minimum
		return fmt.Errorf("memory size must be at least 64 MB")
	}

	// Check disk size
	if params.Disk.SizeBytes < 1024*1024*1024 { // 1 GB minimum
		return fmt.Errorf("disk size must be at least 1 GB")
	}

	// Validate disk parameters
	if err := params.Disk.Validate(); err != nil {
		return fmt.Errorf("invalid disk parameters: %w", err)
	}

	// If network is provided, validate it
	if params.Network.Type != "" {
		if err := params.Network.Validate(); err != nil {
			return fmt.Errorf("invalid network parameters: %w", err)
		}
	}

	return nil
}

// setDefaultParams sets default values for parameters that weren't provided.
func (m *VMManager) setDefaultParams(params vm.VMParams) vm.VMParams {
	// Default CPU model
	if params.CPU.Model == "" {
		params.CPU.Model = "host-model"
	}

	// Default disk format if not specified
	if params.Disk.Format == "" {
		params.Disk.Format = "qcow2"
	}

	// Default storage pool if not specified
	if params.Disk.StoragePool == "" {
		params.Disk.StoragePool = m.config.StoragePoolName
	}

	// Default network type and source if not specified
	if params.Network.Type == "" {
		params.Network.Type = "network"
		params.Network.Source = m.config.NetworkName
	}

	// Default network model if not specified
	if params.Network.Model == "" {
		params.Network.Model = "virtio"
	}

	return params
}

// createVMDisk creates the VM disk.
func (m *VMManager) createVMDisk(ctx context.Context, params vm.VMParams) error {
	poolName := params.Disk.StoragePool
	volumeName := vm.GenerateVolumeName(params.Name, 0)

	m.logger.Debug("Creating VM disk",
		logger.String("vm", params.Name),
		logger.String("pool", poolName),
		logger.String("volume", volumeName),
		logger.String("format", string(params.Disk.Format)),
		logger.Uint64("size", params.Disk.SizeBytes))

	// If source image is provided, create from image
	if params.Disk.SourceImage != "" {
		return m.storageManager.CreateFromImage(
			ctx,
			poolName,
			volumeName,
			params.Disk.SourceImage,
			string(params.Disk.Format),
		)
	}

	// Create empty disk
	return m.storageManager.Create(
		ctx,
		poolName,
		volumeName,
		params.Disk.SizeBytes,
		string(params.Disk.Format),
	)
}

// setupCloudInit generates cloud-init data and creates the ISO.
func (m *VMManager) setupCloudInit(ctx context.Context, params vm.VMParams) error {
	// Generate cloud-init data if not provided
	var config vm.CloudInitConfig

	// If user-data is not provided, generate it
	if params.CloudInit.UserData == "" {
		userData, err := m.cloudInitManager.GenerateUserData(params)
		if err != nil {
			return fmt.Errorf("generating user-data: %w", err)
		}
		config.UserData = userData
	} else {
		config.UserData = params.CloudInit.UserData
	}

	// If meta-data is not provided, generate it
	if params.CloudInit.MetaData == "" {
		metaData, err := m.cloudInitManager.GenerateMetaData(params)
		if err != nil {
			return fmt.Errorf("generating meta-data: %w", err)
		}
		config.MetaData = metaData
	} else {
		config.MetaData = params.CloudInit.MetaData
	}

	// If network-config is not provided, generate it
	if params.CloudInit.NetworkConfig == "" {
		networkConfig, err := m.cloudInitManager.GenerateNetworkConfig(params)
		if err != nil {
			return fmt.Errorf("generating network-config: %w", err)
		}
		config.NetworkConfig = networkConfig
	} else {
		config.NetworkConfig = params.CloudInit.NetworkConfig
	}

	// Create cloud-init ISO - ensure path matches what domain XML builder expects
	isoPath := filepath.Join(m.config.CloudInitDir, fmt.Sprintf("%s-cloudinit.iso", params.Name))

	m.logger.Debug("Creating cloud-init ISO",
		logger.String("vm", params.Name),
		logger.String("path", isoPath))

	// Convert from models.vm.CloudInitConfig to cloudinit.CloudInitConfig
	cloudInitConfig := cloudinit.CloudInitConfig{
		UserData:      config.UserData,
		MetaData:      config.MetaData,
		NetworkConfig: config.NetworkConfig,
	}

	if err := m.cloudInitManager.GenerateISO(ctx, cloudInitConfig, isoPath); err != nil {
		return fmt.Errorf("generating cloud-init ISO: %w", err)
	}

	return nil
}

// cleanupDisk cleans up VM disk on failure.
func (m *VMManager) cleanupDisk(ctx context.Context, params vm.VMParams) error {
	poolName := params.Disk.StoragePool
	volumeName := vm.GenerateVolumeName(params.Name, 0)

	m.logger.Debug("Cleaning up VM disk",
		logger.String("vm", params.Name),
		logger.String("pool", poolName),
		logger.String("volume", volumeName))

	return m.storageManager.Delete(ctx, poolName, volumeName)
}

// cleanupResources cleans up all VM resources on failure.
func (m *VMManager) cleanupResources(ctx context.Context, params vm.VMParams) error {
	// Cleanup disk
	if err := m.cleanupDisk(ctx, params); err != nil {
		m.logger.Warn("Failed to clean up disk",
			logger.String("vm", params.Name),
			logger.Error(err))
	}

	// Cleanup cloud-init ISO - make sure this matches the path used for creation
	isoPath := filepath.Join(m.config.CloudInitDir, fmt.Sprintf("%s-cloudinit.iso", params.Name))
	if err := os.Remove(isoPath); err != nil && !os.IsNotExist(err) {
		m.logger.Warn("Failed to clean up cloud-init ISO",
			logger.String("vm", params.Name),
			logger.String("path", isoPath),
			logger.Error(err))
	}

	return nil
}

// CreateSnapshot creates a new snapshot of a VM.
func (m *VMManager) CreateSnapshot(ctx context.Context, vmName string, params vm.SnapshotParams) (*vm.Snapshot, error) {
	m.logger.Info("Creating VM snapshot",
		logger.String("vm", vmName),
		logger.String("snapshot", params.Name))

	// Delegate to domain manager
	snapshot, err := m.domainManager.CreateSnapshot(ctx, vmName, params)
	if err != nil {
		return nil, fmt.Errorf("creating snapshot: %w", err)
	}

	m.logger.Info("VM snapshot created successfully",
		logger.String("vm", vmName),
		logger.String("snapshot", params.Name))

	return snapshot, nil
}

// ListSnapshots lists all snapshots for a VM.
func (m *VMManager) ListSnapshots(ctx context.Context, vmName string, opts vm.SnapshotListOptions) ([]*vm.Snapshot, error) {
	m.logger.Debug("Listing VM snapshots",
		logger.String("vm", vmName))

	// Delegate to domain manager
	snapshots, err := m.domainManager.ListSnapshots(ctx, vmName, opts)
	if err != nil {
		return nil, fmt.Errorf("listing snapshots: %w", err)
	}

	m.logger.Debug("Listed VM snapshots",
		logger.String("vm", vmName),
		logger.Int("count", len(snapshots)))

	return snapshots, nil
}

// GetSnapshot retrieves information about a specific snapshot.
func (m *VMManager) GetSnapshot(ctx context.Context, vmName string, snapshotName string) (*vm.Snapshot, error) {
	m.logger.Debug("Getting VM snapshot",
		logger.String("vm", vmName),
		logger.String("snapshot", snapshotName))

	// Delegate to domain manager
	snapshot, err := m.domainManager.GetSnapshot(ctx, vmName, snapshotName)
	if err != nil {
		return nil, fmt.Errorf("getting snapshot: %w", err)
	}

	return snapshot, nil
}

// DeleteSnapshot deletes a snapshot.
func (m *VMManager) DeleteSnapshot(ctx context.Context, vmName string, snapshotName string) error {
	m.logger.Info("Deleting VM snapshot",
		logger.String("vm", vmName),
		logger.String("snapshot", snapshotName))

	// Delegate to domain manager
	if err := m.domainManager.DeleteSnapshot(ctx, vmName, snapshotName); err != nil {
		return fmt.Errorf("deleting snapshot: %w", err)
	}

	m.logger.Info("VM snapshot deleted successfully",
		logger.String("vm", vmName),
		logger.String("snapshot", snapshotName))

	return nil
}

// RevertSnapshot reverts a VM to a snapshot.
func (m *VMManager) RevertSnapshot(ctx context.Context, vmName string, snapshotName string) error {
	m.logger.Info("Reverting VM to snapshot",
		logger.String("vm", vmName),
		logger.String("snapshot", snapshotName))

	// Delegate to domain manager
	if err := m.domainManager.RevertSnapshot(ctx, vmName, snapshotName); err != nil {
		return fmt.Errorf("reverting to snapshot: %w", err)
	}

	m.logger.Info("VM reverted to snapshot successfully",
		logger.String("vm", vmName),
		logger.String("snapshot", snapshotName))

	return nil
}
