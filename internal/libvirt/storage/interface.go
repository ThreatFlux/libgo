package storage

import (
	"context"
	"io"

	"github.com/digitalocean/go-libvirt"
)

// PoolManager defines interface for managing storage pools.
type PoolManager interface {
	// EnsureExists ensures that a storage pool exists.
	EnsureExists(ctx context.Context, name string, path string) error

	// Delete deletes a storage pool.
	Delete(ctx context.Context, name string) error

	// Get gets a storage pool.
	Get(ctx context.Context, name string) (*libvirt.StoragePool, error)

	// List lists all storage pools.
	List(ctx context.Context) ([]*StoragePoolInfo, error)

	// GetInfo gets detailed information about a storage pool.
	GetInfo(ctx context.Context, name string) (*StoragePoolInfo, error)

	// Create creates a new storage pool.
	Create(ctx context.Context, params *CreatePoolParams) (*StoragePoolInfo, error)

	// Start starts an inactive storage pool.
	Start(ctx context.Context, name string) error

	// Stop stops an active storage pool.
	Stop(ctx context.Context, name string) error

	// Refresh refreshes a storage pool to scan for new volumes.
	Refresh(ctx context.Context, name string) error

	// SetAutostart sets the autostart flag for a storage pool.
	SetAutostart(ctx context.Context, name string, autostart bool) error

	// IsActive checks if a storage pool is active.
	IsActive(ctx context.Context, name string) (bool, error)

	// GetXML gets the XML configuration of a storage pool.
	GetXML(ctx context.Context, name string) (string, error)
}

// VolumeManager defines interface for managing storage volumes.
type VolumeManager interface {
	// Create creates a new storage volume.
	Create(ctx context.Context, poolName string, volName string, capacityBytes uint64, format string) error

	// CreateFromImage creates a volume from an existing image.
	CreateFromImage(ctx context.Context, poolName string, volName string, imagePath string, format string) error

	// Delete deletes a storage volume.
	Delete(ctx context.Context, poolName string, volName string) error

	// Resize resizes a storage volume.
	Resize(ctx context.Context, poolName string, volName string, capacityBytes uint64) error

	// GetPath gets the path of a storage volume.
	GetPath(ctx context.Context, poolName string, volName string) (string, error)

	// Clone clones a storage volume.
	Clone(ctx context.Context, poolName string, sourceVolName string, destVolName string) error

	// List lists all volumes in a storage pool.
	List(ctx context.Context, poolName string) ([]*StorageVolumeInfo, error)

	// GetInfo gets detailed information about a storage volume.
	GetInfo(ctx context.Context, poolName string, volName string) (*StorageVolumeInfo, error)

	// GetXML gets the XML configuration of a storage volume.
	GetXML(ctx context.Context, poolName string, volName string) (string, error)

	// Wipe wipes/zeros a storage volume.
	Wipe(ctx context.Context, poolName string, volName string) error

	// Upload uploads data to a storage volume.
	Upload(ctx context.Context, poolName string, volName string, reader io.Reader) error

	// Download downloads data from a storage volume.
	Download(ctx context.Context, poolName string, volName string, writer io.Writer) error
}

// XMLBuilder defines interface for building storage XML.
type XMLBuilder interface {
	// BuildStoragePoolXML builds XML for storage pool creation.
	BuildStoragePoolXML(name string, path string) (string, error)

	// BuildStorageVolumeXML builds XML for storage volume creation.
	BuildStorageVolumeXML(volName string, capacityBytes uint64, format string) (string, error)
}

// StoragePoolInfo represents detailed information about a storage pool.
type StoragePoolInfo struct {
	Source     *StoragePoolSource     `json:"source,omitempty"`
	Metadata   map[string]interface{} `json:"metadata,omitempty"`
	Target     *StoragePoolTarget     `json:"target,omitempty"`
	Path       string                 `json:"path,omitempty"`
	UUID       string                 `json:"uuid"`
	State      StoragePoolState       `json:"state"`
	Type       string                 `json:"type"`
	Name       string                 `json:"name"`
	Capacity   uint64                 `json:"capacity"`
	Allocation uint64                 `json:"allocation"`
	Available  uint64                 `json:"available"`
	Autostart  bool                   `json:"autostart"`
	Persistent bool                   `json:"persistent"`
}

// StoragePoolState represents the state of a storage pool.
type StoragePoolState string

const (
	StoragePoolStateInactive     StoragePoolState = "inactive"
	StoragePoolStateBuilding     StoragePoolState = "building"
	StoragePoolStateRunning      StoragePoolState = "running"
	StoragePoolStateDegraded     StoragePoolState = "degraded"
	StoragePoolStateInaccessible StoragePoolState = "inaccessible"
)

// StoragePoolSource represents the source configuration of a storage pool.
type StoragePoolSource struct {
	Host   string `json:"host,omitempty"`
	Dir    string `json:"dir,omitempty"`
	Device string `json:"device,omitempty"`
	Name   string `json:"name,omitempty"`
	Format string `json:"format,omitempty"`
}

// StoragePoolTarget represents the target configuration of a storage pool.
type StoragePoolTarget struct {
	Permissions *struct {
		Mode  string `json:"mode,omitempty"`
		Owner string `json:"owner,omitempty"`
		Group string `json:"group,omitempty"`
	} `json:"permissions,omitempty"`
	Path string `json:"path"`
}

// CreatePoolParams represents parameters for creating a storage pool.
type CreatePoolParams struct {
	Source    *StoragePoolSource     `json:"source,omitempty"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
	Name      string                 `json:"name" binding:"required"`
	Type      string                 `json:"type" binding:"required"`
	Path      string                 `json:"path,omitempty"`
	Autostart bool                   `json:"autostart"`
}

// StorageVolumeInfo represents detailed information about a storage volume.
type StorageVolumeInfo struct {
	BackingStore *BackingStore          `json:"backing_store,omitempty"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
	Name         string                 `json:"name"`
	Key          string                 `json:"key"`
	Path         string                 `json:"path"`
	Type         string                 `json:"type"`
	Format       string                 `json:"format"`
	Pool         string                 `json:"pool"`
	Capacity     uint64                 `json:"capacity"`
	Allocation   uint64                 `json:"allocation"`
}

// BackingStore represents backing store information for a volume.
type BackingStore struct {
	Path   string `json:"path"`
	Format string `json:"format"`
}

// CreateVolumeParams represents parameters for creating a storage volume.
type CreateVolumeParams struct {
	Metadata      map[string]interface{} `json:"metadata,omitempty"`
	Name          string                 `json:"name" binding:"required"`
	Format        string                 `json:"format"`
	BackingStore  string                 `json:"backing_store,omitempty"`
	CapacityBytes uint64                 `json:"capacity_bytes" binding:"required"`
}

// UploadVolumeParams represents parameters for uploading to a storage volume.
type UploadVolumeParams struct {
	// Offset to start writing at (0 for beginning).
	Offset uint64 `json:"offset"`
	// Length of data to write (0 for entire stream).
	Length uint64 `json:"length"`
}
