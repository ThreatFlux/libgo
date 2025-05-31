package vm

import (
	"fmt"
	"path/filepath"
)

// DiskFormat represents the format of a VM disk
type DiskFormat string

// Disk format constants
const (
	DiskFormatQCOW2 DiskFormat = "qcow2"
	DiskFormatRAW   DiskFormat = "raw"
)

// Valid disk formats
var validDiskFormats = map[DiskFormat]bool{
	DiskFormatQCOW2: true,
	DiskFormatRAW:   true,
}

// IsValid checks if the disk format is valid
func (f DiskFormat) IsValid() bool {
	_, valid := validDiskFormats[f]
	return valid
}

// String returns the string representation of the disk format
func (f DiskFormat) String() string {
	return string(f)
}

// DiskType represents the type of disk
type DiskType string

// Disk type constants
const (
	DiskTypeFile  DiskType = "file"
	DiskTypeBlock DiskType = "block"
)

// DiskBus represents the bus type for a disk
type DiskBus string

// Disk bus constants
const (
	DiskBusVirtio DiskBus = "virtio"
	DiskBusIDE    DiskBus = "ide"
	DiskBusSATA   DiskBus = "sata"
	DiskBusSCSI   DiskBus = "scsi"
)

// DiskDriver represents the disk driver
type DiskDriver string

// Disk driver constants
const (
	DiskDriverQEMU DiskDriver = "qemu"
)

// DiskParams contains disk parameters for VM creation
type DiskParams struct {
	SizeBytes   uint64     `json:"sizeBytes" validate:"required,min=1073741824"` // Minimum 1GB
	SizeMB      uint64     `json:"sizeMB,omitempty"`                             // Size in MB (optional, calculated from SizeBytes if not provided)
	Format      DiskFormat `json:"format" validate:"required,oneof=qcow2 raw"`
	SourceImage string     `json:"sourceImage,omitempty"`
	StoragePool string     `json:"storagePool,omitempty"`
	Bus         DiskBus    `json:"bus,omitempty" validate:"omitempty,oneof=virtio ide sata scsi"`
	CacheMode   string     `json:"cacheMode,omitempty" validate:"omitempty,oneof=none writeback writethrough directsync unsafe"`
	Shareable   bool       `json:"shareable,omitempty"`
	ReadOnly    bool       `json:"readOnly,omitempty"`
}

// DiskInfo contains information about a VM's disk
type DiskInfo struct {
	Path        string     `json:"path"`
	Format      DiskFormat `json:"format"`
	SizeBytes   uint64     `json:"sizeBytes"`
	Bus         DiskBus    `json:"bus"`
	ReadOnly    bool       `json:"readOnly,omitempty"`
	Bootable    bool       `json:"bootable,omitempty"`
	Shareable   bool       `json:"shareable,omitempty"`
	Serial      string     `json:"serial,omitempty"`
	StoragePool string     `json:"storagePool,omitempty"`
	Device      string     `json:"device,omitempty"`
	PoolName    string     `json:"poolName,omitempty"`
	VolumeName  string     `json:"volumeName,omitempty"`
}

// Validate validates the disk parameters
func (p *DiskParams) Validate() error {
	// Check disk format
	if !DiskFormat(p.Format).IsValid() {
		return fmt.Errorf("invalid disk format: %s", p.Format)
	}

	// Minimum disk size (1 GB)
	if p.SizeBytes < 1073741824 {
		return fmt.Errorf("disk size must be at least 1 GB (1073741824 bytes)")
	}

	// Check source image if provided
	if p.SourceImage != "" {
		ext := filepath.Ext(p.SourceImage)
		if ext != ".qcow2" && ext != ".img" && ext != ".raw" {
			return fmt.Errorf("invalid source image format: %s", ext)
		}
	}

	// Validate bus if provided
	if p.Bus != "" {
		switch p.Bus {
		case DiskBusVirtio, DiskBusIDE, DiskBusSATA, DiskBusSCSI:
			// Valid
		default:
			return fmt.Errorf("invalid disk bus: %s", p.Bus)
		}
	}

	// Validate cache mode if provided
	if p.CacheMode != "" {
		switch p.CacheMode {
		case "none", "writeback", "writethrough", "directsync", "unsafe":
			// Valid
		default:
			return fmt.Errorf("invalid cache mode: %s", p.CacheMode)
		}
	}

	return nil
}

// GetBus returns the disk bus, defaulting to virtio if not specified
func (p *DiskParams) GetBus() DiskBus {
	if p.Bus == "" {
		return DiskBusVirtio
	}
	return p.Bus
}

// GetCacheMode returns the disk cache mode, defaulting to "none" if not specified
func (p *DiskParams) GetCacheMode() string {
	if p.CacheMode == "" {
		return "none"
	}
	return p.CacheMode
}

// GenerateVolumeName generates a volume name for a VM disk
func GenerateVolumeName(vmName string, diskIndex int) string {
	return fmt.Sprintf("%s-disk-%d", vmName, diskIndex)
}

// GetDefaultStoragePool returns the default storage pool name
func GetDefaultStoragePool() string {
	return "default"
}
