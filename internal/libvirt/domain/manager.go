package domain

import (
	"context"
	"encoding/xml"
	"fmt"
	"time"

	"github.com/digitalocean/go-libvirt"
	"github.com/threatflux/libgo/internal/libvirt/connection"
	"github.com/threatflux/libgo/internal/models/vm"
	"github.com/threatflux/libgo/pkg/logger"
)

// Custom errors.
var (
	ErrDomainNotFound = fmt.Errorf("domain not found")
	ErrDomainExists   = fmt.Errorf("domain already exists")
)

// Network type constants.
const (
	networkTypeBridge  = "bridge"
	networkTypeNetwork = "network"
	networkTypeDirect  = "direct"
)

// DomainManager implements Manager for libvirt domains.
type DomainManager struct {
	connManager connection.Manager
	xmlBuilder  XMLBuilder
	logger      logger.Logger
}

// libvirtDomain is a struct to parse libvirt domain XML.
type libvirtDomain struct {
	// Anonymous struct fields (largest first)
	Devices struct {
		Disks      []libvirtDisk      `xml:"disk"`
		Interfaces []libvirtInterface `xml:"interface"`
	} `xml:"devices"`
	CPU struct {
		Mode  string `xml:"mode,attr"`
		Model struct {
			Value string `xml:",chardata"`
		} `xml:"model"`
		Topology struct {
			Sockets int `xml:"sockets,attr"`
			Cores   int `xml:"cores,attr"`
			Threads int `xml:"threads,attr"`
		} `xml:"topology"`
	} `xml:"cpu"`
	Memory struct {
		Value uint64 `xml:",chardata"`
		Unit  string `xml:"unit,attr"`
	} `xml:"memory"`
	// String fields (16 bytes each)
	Name   string `xml:"name"`
	UUID   string `xml:"uuid"`
	Status string `xml:"state,attr"`
	// Int fields (8 bytes on 64-bit)
	VCPUs int `xml:"vcpu"`
}

// libvirtDisk represents a disk in libvirt domain XML.
type libvirtDisk struct {
	// Anonymous struct fields (largest first)
	Source struct {
		File    string `xml:"file,attr"`
		Pool    string `xml:"pool,attr"`
		Dev     string `xml:"dev,attr"`
		Bridge  string `xml:"bridge,attr"`
		Network string `xml:"network,attr"`
	} `xml:"source"`
	Driver struct {
		Name string `xml:"name,attr"`
		Type string `xml:"type,attr"`
	} `xml:"driver"`
	Target struct {
		Dev string `xml:"dev,attr"`
		Bus string `xml:"bus,attr"`
	} `xml:"target"`
	Boot struct {
		Order int `xml:"order,attr"`
	} `xml:"boot"`
	// String fields (16 bytes each)
	Type   string `xml:"type,attr"`
	Device string `xml:"device,attr"`
	// Pointer fields (8 bytes each)
	ReadOnly  *struct{} `xml:"readonly"`
	Shareable *struct{} `xml:"shareable"`
}

// libvirtInterface represents an interface in libvirt domain XML.
type libvirtInterface struct {
	Type   string `xml:"type,attr"`
	Source struct {
		Bridge  string `xml:"bridge,attr"`
		Network string `xml:"network,attr"`
		Dev     string `xml:"dev,attr"`
	} `xml:"source"`
	MAC struct {
		Address string `xml:"address,attr"`
	} `xml:"mac"`
	Model struct {
		Type string `xml:"type,attr"`
	} `xml:"model"`
}

// NewDomainManager creates a new DomainManager.
func NewDomainManager(connManager connection.Manager, xmlBuilder XMLBuilder, logger logger.Logger) *DomainManager {
	return &DomainManager{
		connManager: connManager,
		xmlBuilder:  xmlBuilder,
		logger:      logger,
	}
}

// handleDeferredRelease logs any error from releasing a connection.
func (m *DomainManager) handleDeferredRelease(conn connection.Connection) {
	if err := m.connManager.Release(conn); err != nil {
		m.logger.Error("Failed to release connection", logger.Error(err))
	}
}

// Create implements Manager.Create.
func (m *DomainManager) Create(ctx context.Context, params vm.VMParams) (*vm.VM, error) {
	// Get libvirt connection
	conn, err := m.connManager.Connect(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to libvirt: %w", err)
	}
	defer m.handleDeferredRelease(conn)

	libvirtConn := conn.GetLibvirtConnection()

	// Check if domain already exists
	_, err = libvirtConn.DomainLookupByName(params.Name)
	if err == nil {
		return nil, fmt.Errorf("creating domain %s: %w", params.Name, ErrDomainExists)
	}

	// Generate domain XML
	domainXML, err := m.xmlBuilder.BuildDomainXML(params)
	if err != nil {
		return nil, fmt.Errorf("generating domain XML: %w", err)
	}

	// Create domain
	m.logger.Info("Creating domain",
		logger.String("name", params.Name),
		logger.Int("vcpus", params.CPU.Count),
		logger.Uint64("memory", params.Memory.SizeBytes))

	domain, err := libvirtConn.DomainDefineXML(domainXML)
	if err != nil {
		return nil, fmt.Errorf("defining domain from XML: %w", err)
	}

	// Start the domain
	if err := libvirtConn.DomainCreate(domain); err != nil {
		// Try to clean up if starting fails
		if undefineErr := libvirtConn.DomainUndefine(domain); undefineErr != nil {
			m.logger.Error("Failed to undefine domain after failed start",
				logger.String("domain", params.Name),
				logger.Error(undefineErr))
		}
		return nil, fmt.Errorf("starting domain: %w", err)
	}

	// Get the created VM details
	return m.domainToVM(libvirtConn, domain)
}

// Get implements Manager.Get.
func (m *DomainManager) Get(ctx context.Context, name string) (*vm.VM, error) {
	// Get libvirt connection
	conn, err := m.connManager.Connect(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to libvirt: %w", err)
	}
	defer m.handleDeferredRelease(conn)

	libvirtConn := conn.GetLibvirtConnection()

	// Look up domain
	domain, err := libvirtConn.DomainLookupByName(name)
	if err != nil {
		return nil, fmt.Errorf("looking up domain %s: %w", name, ErrDomainNotFound)
	}

	// Convert to VM model
	return m.domainToVM(libvirtConn, domain)
}

// List implements Manager.List.
func (m *DomainManager) List(ctx context.Context) ([]*vm.VM, error) {
	// Get libvirt connection
	conn, err := m.connManager.Connect(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to libvirt: %w", err)
	}
	defer m.handleDeferredRelease(conn)

	libvirtConn := conn.GetLibvirtConnection()

	// Get all domains
	domains, _, err := libvirtConn.ConnectListAllDomains(-1, libvirt.ConnectListDomainsActive|libvirt.ConnectListDomainsInactive)
	if err != nil {
		return nil, fmt.Errorf("listing domains: %w", err)
	}

	// Convert to VM models
	vms := make([]*vm.VM, 0, len(domains))
	for _, domain := range domains {
		vm, err := m.domainToVM(libvirtConn, domain)
		if err != nil {
			m.logger.Warn("Failed to convert domain to VM",
				logger.String("domain_name", domain.Name),
				logger.Error(err))
			continue
		}
		vms = append(vms, vm)
	}

	return vms, nil
}

// Start implements Manager.Start.
func (m *DomainManager) Start(ctx context.Context, name string) error {
	return m.performDomainOperation(ctx, name, func(libvirtConn *libvirt.Libvirt, domain libvirt.Domain) error {
		// Check if domain is already running
		state, _, _, _, _, err := libvirtConn.DomainGetInfo(domain) //nolint:dogsled
		if err != nil {
			return fmt.Errorf("getting domain info: %w", err)
		}

		if libvirt.DomainState(state) == libvirt.DomainRunning {
			m.logger.Info("Domain already running", logger.String("name", name))
			return nil
		}

		// Start domain
		if err := libvirtConn.DomainCreate(domain); err != nil {
			return fmt.Errorf("starting domain: %w", err)
		}

		m.logger.Info("Started domain", logger.String("name", name))
		return nil
	})
}

// Stop implements Manager.Stop.
func (m *DomainManager) Stop(ctx context.Context, name string) error {
	return m.performDomainOperation(ctx, name, func(libvirtConn *libvirt.Libvirt, domain libvirt.Domain) error {
		// Check if domain is already stopped
		state, _, _, _, _, err := libvirtConn.DomainGetInfo(domain) //nolint:dogsled
		if err != nil {
			return fmt.Errorf("getting domain info: %w", err)
		}

		if libvirt.DomainState(state) == libvirt.DomainShutoff {
			m.logger.Info("Domain already stopped", logger.String("name", name))
			return nil
		}

		// Try graceful shutdown
		if err := libvirtConn.DomainShutdown(domain); err != nil {
			return fmt.Errorf("shutting down domain: %w", err)
		}

		m.logger.Info("Stopped domain", logger.String("name", name))
		return nil
	})
}

// ForceStop implements Manager.ForceStop.
func (m *DomainManager) ForceStop(ctx context.Context, name string) error {
	return m.performDomainOperation(ctx, name, func(libvirtConn *libvirt.Libvirt, domain libvirt.Domain) error {
		// Check if domain is already stopped
		state, _, _, _, _, err := libvirtConn.DomainGetInfo(domain) //nolint:dogsled
		if err != nil {
			return fmt.Errorf("getting domain info: %w", err)
		}

		if libvirt.DomainState(state) == libvirt.DomainShutoff {
			m.logger.Info("Domain already stopped", logger.String("name", name))
			return nil
		}

		// Force stop
		if err := libvirtConn.DomainDestroy(domain); err != nil {
			return fmt.Errorf("force stopping domain: %w", err)
		}

		m.logger.Info("Force stopped domain", logger.String("name", name))
		return nil
	})
}

// Delete implements Manager.Delete.
func (m *DomainManager) Delete(ctx context.Context, name string) error {
	// Get libvirt connection
	conn, err := m.connManager.Connect(ctx)
	if err != nil {
		return fmt.Errorf("failed to connect to libvirt: %w", err)
	}
	defer m.handleDeferredRelease(conn)

	libvirtConn := conn.GetLibvirtConnection()

	// Look up domain
	domain, err := libvirtConn.DomainLookupByName(name)
	if err != nil {
		return fmt.Errorf("looking up domain %s: %w", name, ErrDomainNotFound)
	}

	// Check if domain is running
	state, _, _, _, _, err := libvirtConn.DomainGetInfo(domain) //nolint:dogsled
	if err != nil {
		return fmt.Errorf("getting domain info: %w", err)
	}

	// Stop domain if it's running
	if libvirt.DomainState(state) == libvirt.DomainRunning {
		if err := libvirtConn.DomainDestroy(domain); err != nil {
			return fmt.Errorf("stopping domain before deletion: %w", err)
		}
	}

	// Delete domain
	if err := libvirtConn.DomainUndefineFlags(domain, libvirt.DomainUndefineKeepNvram); err != nil {
		return fmt.Errorf("undefining domain: %w", err)
	}

	m.logger.Info("Deleted domain", logger.String("name", name))
	return nil
}

// GetXML implements Manager.GetXML.
func (m *DomainManager) GetXML(ctx context.Context, name string) (string, error) {
	// Get libvirt connection
	conn, err := m.connManager.Connect(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to connect to libvirt: %w", err)
	}
	defer m.handleDeferredRelease(conn)

	libvirtConn := conn.GetLibvirtConnection()

	// Look up domain
	domain, err := libvirtConn.DomainLookupByName(name)
	if err != nil {
		return "", fmt.Errorf("looking up domain %s: %w", name, ErrDomainNotFound)
	}

	// Get XML
	flags := libvirt.DomainXMLSecure
	xml, err := libvirtConn.DomainGetXMLDesc(domain, flags)
	if err != nil {
		return "", fmt.Errorf("getting domain XML: %w", err)
	}

	return xml, nil
}

// performDomainOperation handles common domain operation pattern.
func (m *DomainManager) performDomainOperation(ctx context.Context, name string, operation func(*libvirt.Libvirt, libvirt.Domain) error) error {
	// Get libvirt connection
	conn, err := m.connManager.Connect(ctx)
	if err != nil {
		return fmt.Errorf("failed to connect to libvirt: %w", err)
	}
	defer m.handleDeferredRelease(conn)

	libvirtConn := conn.GetLibvirtConnection()

	// Look up domain
	domain, err := libvirtConn.DomainLookupByName(name)
	if err != nil {
		return fmt.Errorf("looking up domain %s: %w", name, ErrDomainNotFound)
	}

	return operation(libvirtConn, domain)
}

// domainToVM converts libvirt domain to VM model.
func (m *DomainManager) domainToVM(libvirtConn *libvirt.Libvirt, domain libvirt.Domain) (*vm.VM, error) {
	// Get and parse domain XML
	domainXML, state, maxMem, err := m.getDomainInfo(libvirtConn, domain)
	if err != nil {
		return nil, err
	}

	// Create base VM structure
	result := m.createBaseVM(domainXML, state, maxMem)

	// Process disks
	result.Disks = m.processDomainDisks(domainXML.Devices.Disks)

	// Process network interfaces
	result.Networks = m.processDomainNetworks(domainXML.Devices.Interfaces)

	return result, nil
}

// getDomainInfo retrieves and parses domain XML and info.
func (m *DomainManager) getDomainInfo(libvirtConn *libvirt.Libvirt, domain libvirt.Domain) (*libvirtDomain, uint8, uint64, error) {
	// Get domain XML
	flags := libvirt.DomainXMLSecure
	xmlDesc, err := libvirtConn.DomainGetXMLDesc(domain, flags)
	if err != nil {
		return nil, 0, 0, fmt.Errorf("getting domain XML: %w", err)
	}

	// Parse XML
	var domainXML libvirtDomain
	if unmarshalErr := xml.Unmarshal([]byte(xmlDesc), &domainXML); unmarshalErr != nil {
		return nil, 0, 0, fmt.Errorf("parsing domain XML: %w", unmarshalErr)
	}

	// Get domain info (state, etc.)
	state, maxMem, _, _, _, err := libvirtConn.DomainGetInfo(domain) //nolint:dogsled
	if err != nil {
		return nil, 0, 0, fmt.Errorf("getting domain info: %w", err)
	}

	return &domainXML, state, maxMem, nil
}

// createBaseVM creates the base VM structure.
func (m *DomainManager) createBaseVM(domainXML *libvirtDomain, state uint8, maxMem uint64) *vm.VM {
	// Convert memory (KiB to bytes)
	var memoryBytes uint64
	if domainXML.Memory.Unit == "KiB" {
		memoryBytes = domainXML.Memory.Value * 1024
	} else {
		// Fallback to info from DomainGetInfo
		memoryBytes = maxMem * 1024
	}

	// Create VM
	result := &vm.VM{
		Name:   domainXML.Name,
		UUID:   domainXML.UUID,
		Status: mapDomainState(state),
		CPU: vm.CPUInfo{
			Count:   domainXML.VCPUs,
			Model:   domainXML.CPU.Model.Value,
			Sockets: domainXML.CPU.Topology.Sockets,
			Cores:   domainXML.CPU.Topology.Cores,
			Threads: domainXML.CPU.Topology.Threads,
		},
		Memory: vm.MemoryInfo{
			SizeBytes: memoryBytes,
			SizeMB:    memoryBytes / (1024 * 1024),
		},
		CreatedAt: time.Now(), // NOTE: Using current time as libvirt creation time not readily available
	}

	// Set default CPU topology if not specified
	if result.CPU.Sockets == 0 && result.CPU.Cores == 0 && result.CPU.Threads == 0 {
		result.CPU.Sockets = 1
		result.CPU.Cores = result.CPU.Count
		result.CPU.Threads = 1
	}

	// Set default CPU model if not specified
	if result.CPU.Model == "" {
		result.CPU.Model = "host-model"
	}

	return result
}

// processDomainDisks processes disk information.
func (m *DomainManager) processDomainDisks(disks []libvirtDisk) []vm.DiskInfo {
	result := make([]vm.DiskInfo, 0, len(disks))

	for _, disk := range disks {
		if disk.Device != "disk" {
			continue
		}

		var path string
		var storagePool string

		switch {
		case disk.Source.File != "":
			path = disk.Source.File
		case disk.Source.Dev != "":
			path = disk.Source.Dev
		}

		if disk.Source.Pool != "" {
			storagePool = disk.Source.Pool
		}

		diskInfo := vm.DiskInfo{
			Path:        path,
			Format:      vm.DiskFormat(disk.Driver.Type),
			SizeBytes:   0, // NOTE: Disk size querying requires additional libvirt calls
			Bus:         vm.DiskBus(disk.Target.Bus),
			ReadOnly:    disk.ReadOnly != nil,
			Bootable:    disk.Boot.Order > 0,
			Shareable:   disk.Shareable != nil,
			Serial:      "", // NOTE: Serial generation not implemented yet
			StoragePool: storagePool,
			Device:      disk.Target.Dev,
		}

		result = append(result, diskInfo)
	}

	return result
}

// processDomainNetworks processes network interface information.
func (m *DomainManager) processDomainNetworks(interfaces []libvirtInterface) []vm.NetInfo {
	result := make([]vm.NetInfo, 0, len(interfaces))

	for _, iface := range interfaces {
		var source string
		switch {
		case iface.Source.Bridge != "":
			source = iface.Source.Bridge
		case iface.Source.Network != "":
			source = iface.Source.Network
		case iface.Source.Dev != "":
			source = iface.Source.Dev
		}

		// Convert string type to NetworkType
		var netType vm.NetworkType
		switch iface.Type {
		case networkTypeBridge:
			netType = vm.NetworkTypeBridge
		case networkTypeNetwork:
			netType = vm.NetworkTypeNetwork
		case networkTypeDirect:
			netType = vm.NetworkTypeDirect
		default:
			netType = vm.NetworkType(iface.Type)
		}

		netInfo := vm.NetInfo{
			Type:       netType,
			MacAddress: iface.MAC.Address,
			Source:     source,
			Model:      iface.Model.Type,
		}

		result = append(result, netInfo)
	}

	return result
}

// mapDomainState maps libvirt domain state to VM status.
func mapDomainState(state uint8) vm.VMStatus {
	switch libvirt.DomainState(state) {
	case libvirt.DomainRunning:
		return vm.VMStatusRunning
	case libvirt.DomainShutoff:
		return vm.VMStatusStopped
	case libvirt.DomainPaused:
		return vm.VMStatusPaused
	case libvirt.DomainShutdown:
		return vm.VMStatusShutdown
	case libvirt.DomainCrashed:
		return vm.VMStatusCrashed
	default:
		return vm.VMStatusUnknown
	}
}

// CreateSnapshot creates a new snapshot of a domain.
func (m *DomainManager) CreateSnapshot(ctx context.Context, vmName string, params vm.SnapshotParams) (*vm.Snapshot, error) {
	// Get libvirt connection
	conn, err := m.connManager.Connect(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to libvirt: %w", err)
	}
	defer m.handleDeferredRelease(conn)

	libvirtConn := conn.GetLibvirtConnection()

	// Get domain
	dom, err := libvirtConn.DomainLookupByName(vmName)
	if err != nil {
		return nil, fmt.Errorf("failed to get domain: %w", err)
	}

	// Build snapshot XML
	snapshotXML := buildSnapshotXML(params)

	// Create snapshot flags
	var flags libvirt.DomainSnapshotCreateFlags
	if params.IncludeMemory {
		flags |= libvirt.DomainSnapshotCreateLive
	}
	if params.Quiesce {
		flags |= libvirt.DomainSnapshotCreateQuiesce
	}

	// Create the snapshot
	snapshot, err := libvirtConn.DomainSnapshotCreateXML(dom, snapshotXML, uint32(flags)) //nolint:gosec
	if err != nil {
		return nil, fmt.Errorf("failed to create snapshot: %w", err)
	}

	// Get snapshot info
	return m.getSnapshotInfo(libvirtConn, snapshot)
}

// ListSnapshots lists all snapshots for a domain.
func (m *DomainManager) ListSnapshots(ctx context.Context, vmName string, opts vm.SnapshotListOptions) ([]*vm.Snapshot, error) {
	// Get libvirt connection
	conn, err := m.connManager.Connect(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to libvirt: %w", err)
	}
	defer m.handleDeferredRelease(conn)

	libvirtConn := conn.GetLibvirtConnection()

	// Get domain
	dom, err := libvirtConn.DomainLookupByName(vmName)
	if err != nil {
		return nil, fmt.Errorf("failed to get domain: %w", err)
	}

	// List all snapshots
	var flags libvirt.DomainSnapshotListFlags
	if opts.Tree {
		flags |= libvirt.DomainSnapshotListRoots
	}
	if opts.IncludeMetadata {
		flags |= libvirt.DomainSnapshotListMetadata
	}

	snapshots, _, err := libvirtConn.DomainListAllSnapshots(dom, 0, uint32(flags)) //nolint:gosec
	if err != nil {
		return nil, fmt.Errorf("failed to list snapshots: %w", err)
	}

	// Convert to our snapshot model
	result := make([]*vm.Snapshot, 0, len(snapshots))
	for _, snap := range snapshots {
		info, err := m.getSnapshotInfo(libvirtConn, snap)
		if err != nil {
			m.logger.Warn("Failed to get snapshot info", logger.Error(err))
			continue
		}
		result = append(result, info)
	}

	return result, nil
}

// GetSnapshot retrieves information about a specific snapshot.
func (m *DomainManager) GetSnapshot(ctx context.Context, vmName string, snapshotName string) (*vm.Snapshot, error) {
	// Get libvirt connection
	conn, err := m.connManager.Connect(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to libvirt: %w", err)
	}
	defer m.handleDeferredRelease(conn)

	libvirtConn := conn.GetLibvirtConnection()

	// Get domain
	dom, err := libvirtConn.DomainLookupByName(vmName)
	if err != nil {
		return nil, fmt.Errorf("failed to get domain: %w", err)
	}

	// Get snapshot
	snapshot, err := libvirtConn.DomainSnapshotLookupByName(dom, snapshotName, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to get snapshot: %w", err)
	}

	// Get snapshot info
	return m.getSnapshotInfo(libvirtConn, snapshot)
}

// DeleteSnapshot deletes a snapshot.
func (m *DomainManager) DeleteSnapshot(ctx context.Context, vmName string, snapshotName string) error {
	// Get libvirt connection
	conn, err := m.connManager.Connect(ctx)
	if err != nil {
		return fmt.Errorf("failed to connect to libvirt: %w", err)
	}
	defer m.handleDeferredRelease(conn)

	libvirtConn := conn.GetLibvirtConnection()

	// Get domain
	dom, err := libvirtConn.DomainLookupByName(vmName)
	if err != nil {
		return fmt.Errorf("failed to get domain: %w", err)
	}

	// Get snapshot
	snapshot, err := libvirtConn.DomainSnapshotLookupByName(dom, snapshotName, 0)
	if err != nil {
		return fmt.Errorf("failed to get snapshot: %w", err)
	}

	// Delete snapshot
	err = libvirtConn.DomainSnapshotDelete(snapshot, 0)
	if err != nil {
		return fmt.Errorf("failed to delete snapshot: %w", err)
	}

	return nil
}

// RevertSnapshot reverts a domain to a snapshot.
func (m *DomainManager) RevertSnapshot(ctx context.Context, vmName string, snapshotName string) error {
	// Get libvirt connection
	conn, err := m.connManager.Connect(ctx)
	if err != nil {
		return fmt.Errorf("failed to connect to libvirt: %w", err)
	}
	defer m.handleDeferredRelease(conn)

	libvirtConn := conn.GetLibvirtConnection()

	// Get domain
	dom, err := libvirtConn.DomainLookupByName(vmName)
	if err != nil {
		return fmt.Errorf("failed to get domain: %w", err)
	}

	// Get snapshot
	snapshot, err := libvirtConn.DomainSnapshotLookupByName(dom, snapshotName, 0)
	if err != nil {
		return fmt.Errorf("failed to get snapshot: %w", err)
	}

	// Revert to snapshot
	err = libvirtConn.DomainRevertToSnapshot(snapshot, 0)
	if err != nil {
		return fmt.Errorf("failed to revert to snapshot: %w", err)
	}

	return nil
}

// getSnapshotInfo retrieves information about a snapshot.
func (m *DomainManager) getSnapshotInfo(conn *libvirt.Libvirt, snapshot libvirt.DomainSnapshot) (*vm.Snapshot, error) {
	// Get snapshot XML
	xmlDesc, err := conn.DomainSnapshotGetXMLDesc(snapshot, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to get snapshot XML: %w", err)
	}

	// Parse snapshot XML
	var snapInfo snapshotXML
	if unmarshalErr := xml.Unmarshal([]byte(xmlDesc), &snapInfo); unmarshalErr != nil {
		return nil, fmt.Errorf("failed to parse snapshot XML: %w", unmarshalErr)
	}

	// Check if snapshot is current
	isCurrent, err := conn.DomainSnapshotIsCurrent(snapshot, 0)
	if err != nil {
		m.logger.Warn("Failed to check if snapshot is current", logger.Error(err))
		isCurrent = 0
	}

	// Check if snapshot has metadata
	hasMetadata, err := conn.DomainSnapshotHasMetadata(snapshot, 0)
	if err != nil {
		m.logger.Warn("Failed to check if snapshot has metadata", logger.Error(err))
		hasMetadata = 0
	}

	// Convert to our model
	result := &vm.Snapshot{
		Name:        snapInfo.Name,
		Description: snapInfo.Description,
		State:       mapSnapshotState(snapInfo.State),
		Parent:      snapInfo.Parent,
		CreatedAt:   time.Unix(snapInfo.CreationTime, 0),
		IsCurrent:   isCurrent != 0,
		HasMetadata: hasMetadata != 0,
		HasMemory:   snapInfo.Memory != nil,
		HasDisk:     len(snapInfo.Disks) > 0,
	}

	return result, nil
}

// snapshotXML represents libvirt snapshot XML structure.
type snapshotXML struct {
	// Slice fields (24 bytes)
	Disks []struct {
		Name string `xml:"name,attr"`
	} `xml:"disks>disk,omitempty"`
	// XMLName field (24 bytes)
	XMLName xml.Name `xml:"domainsnapshot"`
	// String fields (16 bytes each)
	Name        string `xml:"name"`
	Description string `xml:"description,omitempty"`
	State       string `xml:"state,omitempty"`
	Parent      string `xml:"parent>name,omitempty"`
	// Int64 fields (8 bytes)
	CreationTime int64 `xml:"creationTime"`
	// Pointer fields (8 bytes)
	Memory *struct{} `xml:"memory,omitempty"`
}

// buildSnapshotXML builds XML for snapshot creation.
func buildSnapshotXML(params vm.SnapshotParams) string {
	xml := fmt.Sprintf(`<domainsnapshot>
  <name>%s</name>`, params.Name)

	if params.Description != "" {
		xml += fmt.Sprintf("\n  <description>%s</description>", params.Description)
	}

	if params.IncludeMemory {
		xml += "\n  <memory snapshot='internal'/>"
	}

	xml += "\n</domainsnapshot>"
	return xml
}

// mapSnapshotState maps libvirt snapshot state to our model.
func mapSnapshotState(state string) vm.SnapshotState {
	switch state {
	case "running":
		return vm.SnapshotStateRunning
	case "blocked":
		return vm.SnapshotStateBlocked
	case "paused":
		return vm.SnapshotStatePaused
	case "shutdown":
		return vm.SnapshotStateShutdown
	case "shutoff":
		return vm.SnapshotStateShutoff
	case "crashed":
		return vm.SnapshotStateCrashed
	default:
		return vm.SnapshotStateShutoff
	}
}
