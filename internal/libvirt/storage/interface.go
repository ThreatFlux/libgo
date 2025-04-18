package storage

import (
	"context"

	"github.com/digitalocean/go-libvirt"
)

// PoolManager defines interface for managing storage pools
type PoolManager interface {
	// EnsureExists ensures that a storage pool exists
	EnsureExists(ctx context.Context, name string, path string) error
	
	// Delete deletes a storage pool
	Delete(ctx context.Context, name string) error
	
	// Get gets a storage pool
	Get(ctx context.Context, name string) (*libvirt.StoragePool, error)
}

// VolumeManager defines interface for managing storage volumes
type VolumeManager interface {
	// Create creates a new storage volume
	Create(ctx context.Context, poolName string, volName string, capacityBytes uint64, format string) error
	
	// CreateFromImage creates a volume from an existing image
	CreateFromImage(ctx context.Context, poolName string, volName string, imagePath string, format string) error
	
	// Delete deletes a storage volume
	Delete(ctx context.Context, poolName string, volName string) error
	
	// Resize resizes a storage volume
	Resize(ctx context.Context, poolName string, volName string, capacityBytes uint64) error
	
	// GetPath gets the path of a storage volume
	GetPath(ctx context.Context, poolName string, volName string) (string, error)
	
	// Clone clones a storage volume
	Clone(ctx context.Context, poolName string, sourceVolName string, destVolName string) error
}

// XMLBuilder defines interface for building storage XML
type XMLBuilder interface {
	// BuildStoragePoolXML builds XML for storage pool creation
	BuildStoragePoolXML(name string, path string) (string, error)
	
	// BuildStorageVolumeXML builds XML for storage volume creation
	BuildStorageVolumeXML(volName string, capacityBytes uint64, format string) (string, error)
}
