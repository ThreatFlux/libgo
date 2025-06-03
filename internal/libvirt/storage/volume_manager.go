package storage

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/digitalocean/go-libvirt"
	"github.com/threatflux/libgo/internal/libvirt/connection"
	"github.com/threatflux/libgo/pkg/logger"
	executil "github.com/threatflux/libgo/pkg/utils/exec"
	"github.com/threatflux/libgo/pkg/utils/xmlutils"
)

// Error types
var (
	ErrVolumeNotFound = fmt.Errorf("storage volume not found")
	ErrVolumeExists   = fmt.Errorf("storage volume already exists")
	ErrPoolNotActive  = fmt.Errorf("storage pool is not active")
)

// LibvirtVolumeManager implements VolumeManager for libvirt
type LibvirtVolumeManager struct {
	connManager connection.Manager
	poolManager PoolManager
	xmlBuilder  XMLBuilder
	logger      logger.Logger
}

// NewLibvirtVolumeManager creates a new LibvirtVolumeManager
func NewLibvirtVolumeManager(connManager connection.Manager, poolManager PoolManager, xmlBuilder XMLBuilder, logger logger.Logger) *LibvirtVolumeManager {
	return &LibvirtVolumeManager{
		connManager: connManager,
		poolManager: poolManager,
		xmlBuilder:  xmlBuilder,
		logger:      logger,
	}
}

// Create implements VolumeManager.Create
func (m *LibvirtVolumeManager) Create(ctx context.Context, poolName string, volName string, capacityBytes uint64, format string) error {
	// Get libvirt connection
	conn, err := m.connManager.Connect(ctx)
	if err != nil {
		return fmt.Errorf("failed to connect to libvirt: %w", err)
	}
	defer func() {
		if err := m.connManager.Release(conn); err != nil {
			m.logger.Error("Failed to release connection", logger.Error(err))
		}
	}()

	libvirtConn := conn.GetLibvirtConnection()

	// Get the pool
	pool, err := m.poolManager.Get(ctx, poolName)
	if err != nil {
		return fmt.Errorf("getting storage pool: %w", err)
	}

	// Ensure the pool is active
	poolInfo, _, _, _, err := libvirtConn.StoragePoolGetInfo(*pool)
	if err != nil {
		return fmt.Errorf("getting pool info: %w", err)
	}

	if libvirt.StoragePoolState(poolInfo) != libvirt.StoragePoolRunning {
		return fmt.Errorf("pool %s: %w", poolName, ErrPoolNotActive)
	}

	// Check if volume already exists
	existingVol, err := libvirtConn.StorageVolLookupByName(*pool, volName)
	if err == nil {
		// If volume already exists, instead of failing, delete it first and recreate
		m.logger.Warn("Volume already exists, deleting and recreating",
			logger.String("pool", poolName),
			logger.String("volume", volName))

		if deleteErr := libvirtConn.StorageVolDelete(existingVol, 0); deleteErr != nil {
			return fmt.Errorf("deleting existing volume %s in pool %s: %w", volName, poolName, deleteErr)
		}
	}

	// Generate volume XML
	volumeXML, err := m.xmlBuilder.BuildStorageVolumeXML(volName, capacityBytes, format)
	if err != nil {
		return fmt.Errorf("building volume XML: %w", err)
	}

	// Create the volume
	_, err = libvirtConn.StorageVolCreateXML(*pool, volumeXML, 0)
	if err != nil {
		return fmt.Errorf("creating volume: %w", err)
	}

	m.logger.Info("Created storage volume",
		logger.String("pool", poolName),
		logger.String("volume", volName),
		logger.Uint64("capacity", capacityBytes),
		logger.String("format", format))

	return nil
}

// CreateFromImage implements VolumeManager.CreateFromImage
func (m *LibvirtVolumeManager) CreateFromImage(ctx context.Context, poolName string, volName string, imagePath string, format string) error {
	// Get libvirt connection
	conn, err := m.connManager.Connect(ctx)
	if err != nil {
		return fmt.Errorf("failed to connect to libvirt: %w", err)
	}
	defer func() {
		if err := m.connManager.Release(conn); err != nil {
			m.logger.Error("Failed to release connection", logger.Error(err))
		}
	}()

	libvirtConn := conn.GetLibvirtConnection()

	// Get the pool
	pool, err := m.poolManager.Get(ctx, poolName)
	if err != nil {
		return fmt.Errorf("getting storage pool: %w", err)
	}

	// Ensure the pool is active
	poolInfo, _, _, _, err := libvirtConn.StoragePoolGetInfo(*pool)
	if err != nil {
		return fmt.Errorf("getting pool info: %w", err)
	}

	if libvirt.StoragePoolState(poolInfo) != libvirt.StoragePoolRunning {
		return fmt.Errorf("pool %s: %w", poolName, ErrPoolNotActive)
	}

	// Check if volume already exists
	existingVol, err := libvirtConn.StorageVolLookupByName(*pool, volName)
	if err == nil {
		// If volume already exists, instead of failing, delete it first and recreate
		m.logger.Warn("Volume already exists, deleting and recreating",
			logger.String("pool", poolName),
			logger.String("volume", volName))

		if deleteErr := libvirtConn.StorageVolDelete(existingVol, 0); deleteErr != nil {
			return fmt.Errorf("deleting existing volume %s in pool %s: %w", volName, poolName, deleteErr)
		}
	}

	// Check if source image exists
	if _, statErr := os.Stat(imagePath); os.IsNotExist(statErr) {
		return fmt.Errorf("source image %s does not exist", imagePath)
	}

	// Get image info (size, format)
	imgInfo, err := m.getImageInfo(ctx, imagePath)
	if err != nil {
		return fmt.Errorf("getting image info: %w", err)
	}

	// If format not specified, use the image format
	if format == "" {
		format = imgInfo.Format
	}

	// Create the volume
	volumeXML, err := m.xmlBuilder.BuildStorageVolumeXML(volName, imgInfo.VirtualSize, format)
	if err != nil {
		return fmt.Errorf("building volume XML: %w", err)
	}

	vol, err := libvirtConn.StorageVolCreateXML(*pool, volumeXML, 0)
	if err != nil {
		return fmt.Errorf("creating volume: %w", err)
	}

	// Get the volume path
	volPath, err := libvirtConn.StorageVolGetPath(vol)
	if err != nil {
		// Try to clean up
		if deleteErr := libvirtConn.StorageVolDelete(vol, 0); deleteErr != nil {
			m.logger.Error("Failed to delete volume during cleanup", logger.Error(deleteErr))
		}
		return fmt.Errorf("getting volume path: %w", err)
	}

	// Use qemu-img to copy the image content
	cmdCtx, cancel := context.WithTimeout(ctx, 5*time.Minute)
	defer cancel()

	args := []string{
		"convert",
		"-f", imgInfo.Format,
		"-O", format,
		imagePath,
		volPath,
	}

	cmdOpts := executil.CommandOptions{
		Timeout: 5 * time.Minute,
	}

	output, err := executil.ExecuteCommand(cmdCtx, "qemu-img", args, cmdOpts)
	if err != nil {
		// Clean up the volume
		if deleteErr := libvirtConn.StorageVolDelete(vol, 0); deleteErr != nil {
			m.logger.Error("Failed to delete volume during cleanup", logger.Error(deleteErr))
		}
		return fmt.Errorf("converting image: %w: %s", err, string(output))
	}

	m.logger.Info("Created volume from image",
		logger.String("pool", poolName),
		logger.String("volume", volName),
		logger.String("source", imagePath),
		logger.String("format", format))

	return nil
}

// withVolumeConnection is a helper that handles the common pattern of:
// 1. Getting a libvirt connection
// 2. Getting the storage pool
// 3. Looking up the volume
// 4. Executing the provided operation
// 5. Cleaning up the connection
func (m *LibvirtVolumeManager) withVolumeConnection(ctx context.Context, poolName string, volName string, operation func(*libvirt.Libvirt, libvirt.StorageVol) error) error {
	// Get libvirt connection
	conn, err := m.connManager.Connect(ctx)
	if err != nil {
		return fmt.Errorf("failed to connect to libvirt: %w", err)
	}
	defer func() {
		if err := m.connManager.Release(conn); err != nil {
			m.logger.Error("Failed to release connection", logger.Error(err))
		}
	}()

	libvirtConn := conn.GetLibvirtConnection()

	// Get the pool
	pool, err := m.poolManager.Get(ctx, poolName)
	if err != nil {
		return fmt.Errorf("getting storage pool: %w", err)
	}

	// Look up the volume
	vol, err := libvirtConn.StorageVolLookupByName(*pool, volName)
	if err != nil {
		return fmt.Errorf("volume %s in pool %s: %w", volName, poolName, ErrVolumeNotFound)
	}

	// Execute the operation
	return operation(libvirtConn, vol)
}

// Delete implements VolumeManager.Delete
func (m *LibvirtVolumeManager) Delete(ctx context.Context, poolName string, volName string) error {
	err := m.withVolumeConnection(ctx, poolName, volName, func(libvirtConn *libvirt.Libvirt, vol libvirt.StorageVol) error {
		// Delete the volume
		if err := libvirtConn.StorageVolDelete(vol, 0); err != nil {
			return fmt.Errorf("deleting volume: %w", err)
		}
		return nil
	})

	if err != nil {
		return err
	}

	m.logger.Info("Deleted storage volume",
		logger.String("pool", poolName),
		logger.String("volume", volName))

	return nil
}

// Resize implements VolumeManager.Resize
func (m *LibvirtVolumeManager) Resize(ctx context.Context, poolName string, volName string, capacityBytes uint64) error {
	// Get libvirt connection
	conn, err := m.connManager.Connect(ctx)
	if err != nil {
		return fmt.Errorf("failed to connect to libvirt: %w", err)
	}
	defer func() {
		if err := m.connManager.Release(conn); err != nil {
			m.logger.Error("Failed to release connection", logger.Error(err))
		}
	}()

	libvirtConn := conn.GetLibvirtConnection()

	// Get the pool
	pool, err := m.poolManager.Get(ctx, poolName)
	if err != nil {
		return fmt.Errorf("getting storage pool: %w", err)
	}

	// Look up the volume
	vol, err := libvirtConn.StorageVolLookupByName(*pool, volName)
	if err != nil {
		return fmt.Errorf("volume %s in pool %s: %w", volName, poolName, ErrVolumeNotFound)
	}

	// Resize the volume
	// Note: VIR_STORAGE_VOL_RESIZE_BYTES flag (value 1) tells libvirt the new size is in bytes
	if err := libvirtConn.StorageVolResize(vol, capacityBytes, 1); err != nil {
		return fmt.Errorf("resizing volume: %w", err)
	}

	m.logger.Info("Resized storage volume",
		logger.String("pool", poolName),
		logger.String("volume", volName),
		logger.Uint64("new_capacity", capacityBytes))

	return nil
}

// GetPath implements VolumeManager.GetPath
func (m *LibvirtVolumeManager) GetPath(ctx context.Context, poolName string, volName string) (string, error) {
	// Get libvirt connection
	conn, err := m.connManager.Connect(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to connect to libvirt: %w", err)
	}
	defer func() {
		if err := m.connManager.Release(conn); err != nil {
			m.logger.Error("Failed to release connection", logger.Error(err))
		}
	}()

	libvirtConn := conn.GetLibvirtConnection()

	// Get the pool
	pool, err := m.poolManager.Get(ctx, poolName)
	if err != nil {
		return "", fmt.Errorf("getting storage pool: %w", err)
	}

	// Look up the volume
	vol, err := libvirtConn.StorageVolLookupByName(*pool, volName)
	if err != nil {
		return "", fmt.Errorf("volume %s in pool %s: %w", volName, poolName, ErrVolumeNotFound)
	}

	// Get the volume path
	path, err := libvirtConn.StorageVolGetPath(vol)
	if err != nil {
		return "", fmt.Errorf("getting volume path: %w", err)
	}

	return path, nil
}

// Clone implements VolumeManager.Clone
func (m *LibvirtVolumeManager) Clone(ctx context.Context, poolName string, sourceVolName string, destVolName string) error {
	// Get libvirt connection
	conn, err := m.connManager.Connect(ctx)
	if err != nil {
		return fmt.Errorf("failed to connect to libvirt: %w", err)
	}
	defer func() {
		if err := m.connManager.Release(conn); err != nil {
			m.logger.Error("Failed to release connection", logger.Error(err))
		}
	}()

	libvirtConn := conn.GetLibvirtConnection()

	// Get the pool
	pool, err := m.poolManager.Get(ctx, poolName)
	if err != nil {
		return fmt.Errorf("getting storage pool: %w", err)
	}

	// Look up the source volume
	sourceVol, err := libvirtConn.StorageVolLookupByName(*pool, sourceVolName)
	if err != nil {
		return fmt.Errorf("source volume %s in pool %s: %w", sourceVolName, poolName, ErrVolumeNotFound)
	}

	// Check if destination volume already exists
	_, err = libvirtConn.StorageVolLookupByName(*pool, destVolName)
	if err == nil {
		return fmt.Errorf("destination volume %s in pool %s: %w", destVolName, poolName, ErrVolumeExists)
	}

	// Get the source volume XML
	sourceXML, err := libvirtConn.StorageVolGetXMLDesc(sourceVol, 0)
	if err != nil {
		return fmt.Errorf("getting source volume XML: %w", err)
	}

	// Parse XML to modify it for the destination
	doc, err := xmlutils.LoadXMLDocumentFromString(sourceXML)
	if err != nil {
		return fmt.Errorf("parsing source volume XML: %w", err)
	}

	// Set the name in the XML
	nameElement := xmlutils.FindElement(doc, "/volume/name")
	if nameElement == nil {
		return fmt.Errorf("volume name element not found in XML")
	}
	nameElement.SetText(destVolName)

	// Generate the new XML
	newXML := xmlutils.XMLToString(doc)

	// Create the cloned volume
	_, err = libvirtConn.StorageVolCreateXMLFrom(*pool, newXML, sourceVol, 0)
	if err != nil {
		return fmt.Errorf("cloning volume: %w", err)
	}

	m.logger.Info("Cloned storage volume",
		logger.String("pool", poolName),
		logger.String("source", sourceVolName),
		logger.String("destination", destVolName))

	return nil
}

// imageInfo holds information about a disk image
type imageInfo struct {
	Format      string
	VirtualSize uint64
	ActualSize  uint64
}

// getImageInfo uses qemu-img info to get image details
func (m *LibvirtVolumeManager) getImageInfo(ctx context.Context, imagePath string) (*imageInfo, error) {
	// Ensure the file exists
	if _, err := os.Stat(imagePath); os.IsNotExist(err) {
		return nil, fmt.Errorf("image file does not exist: %s", imagePath)
	}

	// Get the absolute path
	absPath, err := filepath.Abs(imagePath)
	if err != nil {
		return nil, fmt.Errorf("getting absolute path: %w", err)
	}

	// Run qemu-img info command
	cmdCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	args := []string{"info", "--output=json", absPath}
	cmdOpts := executil.CommandOptions{
		Timeout: 30 * time.Second,
	}

	output, err := executil.ExecuteCommand(cmdCtx, "qemu-img", args, cmdOpts)
	if err != nil {
		// Check if qemu-img exists
		_, lookErr := executil.LookPath("qemu-img")
		if lookErr != nil {
			return nil, fmt.Errorf("qemu-img command not found: %w", lookErr)
		}
		return nil, fmt.Errorf("getting image info: %w: %s", err, string(output))
	}

	// Parse the JSON output
	var result struct {
		Format      string `json:"format"`
		VirtualSize uint64 `json:"virtual-size"`
		ActualSize  uint64 `json:"actual-size"`
	}

	if err := json.Unmarshal(output, &result); err != nil {
		return nil, fmt.Errorf("parsing qemu-img output: %w", err)
	}

	return &imageInfo{
		Format:      result.Format,
		VirtualSize: result.VirtualSize,
		ActualSize:  result.ActualSize,
	}, nil
}

// List implements VolumeManager.List
func (m *LibvirtVolumeManager) List(ctx context.Context, poolName string) ([]*StorageVolumeInfo, error) {
	// Get libvirt connection
	conn, err := m.connManager.Connect(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to libvirt: %w", err)
	}
	defer func() {
		if err := m.connManager.Release(conn); err != nil {
			m.logger.Error("Failed to release connection", logger.Error(err))
		}
	}()

	libvirtConn := conn.GetLibvirtConnection()

	// Get the pool
	pool, err := m.poolManager.Get(ctx, poolName)
	if err != nil {
		return nil, fmt.Errorf("getting storage pool: %w", err)
	}

	// List all volumes in the pool
	volumes, _, err := libvirtConn.StoragePoolListAllVolumes(*pool, -1, 0)
	if err != nil {
		return nil, fmt.Errorf("listing volumes: %w", err)
	}

	var volumeInfos []*StorageVolumeInfo
	for _, vol := range volumes {
		volInfo, err := m.getVolumeInfo(libvirtConn, &vol, poolName)
		if err != nil {
			m.logger.Warn("Failed to get volume info",
				logger.String("volume", vol.Name),
				logger.Error(err))
			continue
		}
		volumeInfos = append(volumeInfos, volInfo)
	}

	return volumeInfos, nil
}

// GetInfo implements VolumeManager.GetInfo
func (m *LibvirtVolumeManager) GetInfo(ctx context.Context, poolName string, volName string) (*StorageVolumeInfo, error) {
	// Get libvirt connection
	conn, err := m.connManager.Connect(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to libvirt: %w", err)
	}
	defer func() {
		if err := m.connManager.Release(conn); err != nil {
			m.logger.Error("Failed to release connection", logger.Error(err))
		}
	}()

	libvirtConn := conn.GetLibvirtConnection()

	// Get the pool
	pool, err := m.poolManager.Get(ctx, poolName)
	if err != nil {
		return nil, fmt.Errorf("getting storage pool: %w", err)
	}

	// Look up the volume
	vol, err := libvirtConn.StorageVolLookupByName(*pool, volName)
	if err != nil {
		return nil, fmt.Errorf("volume %s in pool %s: %w", volName, poolName, ErrVolumeNotFound)
	}

	return m.getVolumeInfo(libvirtConn, &vol, poolName)
}

// GetXML implements VolumeManager.GetXML
func (m *LibvirtVolumeManager) GetXML(ctx context.Context, poolName string, volName string) (string, error) {
	// Get libvirt connection
	conn, err := m.connManager.Connect(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to connect to libvirt: %w", err)
	}
	defer func() {
		if err := m.connManager.Release(conn); err != nil {
			m.logger.Error("Failed to release connection", logger.Error(err))
		}
	}()

	libvirtConn := conn.GetLibvirtConnection()

	// Get the pool
	pool, err := m.poolManager.Get(ctx, poolName)
	if err != nil {
		return "", fmt.Errorf("getting storage pool: %w", err)
	}

	// Look up the volume
	vol, err := libvirtConn.StorageVolLookupByName(*pool, volName)
	if err != nil {
		return "", fmt.Errorf("volume %s in pool %s: %w", volName, poolName, ErrVolumeNotFound)
	}

	// Get XML
	xml, err := libvirtConn.StorageVolGetXMLDesc(vol, 0)
	if err != nil {
		return "", fmt.Errorf("getting volume XML: %w", err)
	}

	return xml, nil
}

// Wipe implements VolumeManager.Wipe
func (m *LibvirtVolumeManager) Wipe(ctx context.Context, poolName string, volName string) error {
	err := m.withVolumeConnection(ctx, poolName, volName, func(libvirtConn *libvirt.Libvirt, vol libvirt.StorageVol) error {
		// Wipe the volume
		if err := libvirtConn.StorageVolWipe(vol, 0); err != nil {
			return fmt.Errorf("wiping volume: %w", err)
		}
		return nil
	})

	if err != nil {
		return err
	}

	m.logger.Info("Wiped storage volume",
		logger.String("pool", poolName),
		logger.String("volume", volName))

	return nil
}

// Upload implements VolumeManager.Upload
func (m *LibvirtVolumeManager) Upload(ctx context.Context, poolName string, volName string, reader io.Reader) error {
	// Get libvirt connection
	conn, err := m.connManager.Connect(ctx)
	if err != nil {
		return fmt.Errorf("failed to connect to libvirt: %w", err)
	}
	defer func() {
		if err := m.connManager.Release(conn); err != nil {
			m.logger.Error("Failed to release connection", logger.Error(err))
		}
	}()

	// Stream upload is not supported by digitalocean/go-libvirt
	// Return an error indicating the feature is not implemented
	return fmt.Errorf("volume upload is not currently supported")
}

// Download implements VolumeManager.Download
func (m *LibvirtVolumeManager) Download(ctx context.Context, poolName string, volName string, writer io.Writer) error {
	// Get libvirt connection
	conn, err := m.connManager.Connect(ctx)
	if err != nil {
		return fmt.Errorf("failed to connect to libvirt: %w", err)
	}
	defer func() {
		if err := m.connManager.Release(conn); err != nil {
			m.logger.Error("Failed to release connection", logger.Error(err))
		}
	}()

	// Stream download is not supported by digitalocean/go-libvirt
	// Return an error indicating the feature is not implemented
	return fmt.Errorf("volume download is not currently supported")
}

// getVolumeInfo is a helper method to get volume information
func (m *LibvirtVolumeManager) getVolumeInfo(libvirtConn *libvirt.Libvirt, vol *libvirt.StorageVol, poolName string) (*StorageVolumeInfo, error) {
	// Get volume info
	_, capacity, allocation, err := libvirtConn.StorageVolGetInfo(*vol)
	if err != nil {
		return nil, fmt.Errorf("getting volume info: %w", err)
	}

	// Get volume path
	path, err := libvirtConn.StorageVolGetPath(*vol)
	if err != nil {
		return nil, fmt.Errorf("getting volume path: %w", err)
	}

	// Get volume key
	key := vol.Key

	// Get XML to extract format and other details
	xml, err := libvirtConn.StorageVolGetXMLDesc(*vol, 0)
	if err != nil {
		return nil, fmt.Errorf("getting volume XML: %w", err)
	}

	// Extract format from XML (simple regex for now)
	// TODO: Proper XML parsing
	format := "raw" // Default
	if formatStart := strings.Index(xml, `<format type="`); formatStart != -1 {
		formatEnd := strings.Index(xml[formatStart+14:], `"`)
		if formatEnd != -1 {
			format = xml[formatStart+14 : formatStart+14+formatEnd]
		}
	}

	volumeInfo := &StorageVolumeInfo{
		Name:       vol.Name,
		Key:        key,
		Path:       path,
		Type:       "file", // Default, should be parsed from XML
		Capacity:   capacity,
		Allocation: allocation,
		Format:     format,
		Pool:       poolName,
	}

	return volumeInfo, nil
}
