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

// Custom errors
var (
	ErrDomainNotFound = fmt.Errorf("domain not found")
	ErrDomainExists   = fmt.Errorf("domain already exists")
)

// DomainManager implements Manager for libvirt domains
type DomainManager struct {
	connManager connection.Manager
	xmlBuilder  XMLBuilder
	logger      logger.Logger
}

// libvirtDomain is a struct to parse libvirt domain XML
type libvirtDomain struct {
	Name   string `xml:"name"`
	UUID   string `xml:"uuid"`
	Memory struct {
		Value uint64 `xml:",chardata"`
		Unit  string `xml:"unit,attr"`
	} `xml:"memory"`
	VCPUs   int    `xml:"vcpu"`
	Status  string `xml:"state,attr"`
	Devices struct {
		Disks []struct {
			Type   string `xml:"type,attr"`
			Device string `xml:"device,attr"`
			Driver struct {
				Name string `xml:"name,attr"`
				Type string `xml:"type,attr"`
			} `xml:"driver"`
			Source struct {
				File    string `xml:"file,attr"`
				Pool    string `xml:"pool,attr"`
				Dev     string `xml:"dev,attr"`
				Bridge  string `xml:"bridge,attr"`
				Network string `xml:"network,attr"`
			} `xml:"source"`
			Target struct {
				Dev string `xml:"dev,attr"`
				Bus string `xml:"bus,attr"`
			} `xml:"target"`
			Boot struct {
				Order int `xml:"order,attr"`
			} `xml:"boot"`
			ReadOnly  *struct{} `xml:"readonly"`
			Shareable *struct{} `xml:"shareable"`
		} `xml:"disk"`
		Interfaces []struct {
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
		} `xml:"interface"`
	} `xml:"devices"`
}

// NewDomainManager creates a new DomainManager
func NewDomainManager(connManager connection.Manager, xmlBuilder XMLBuilder, logger logger.Logger) *DomainManager {
	return &DomainManager{
		connManager: connManager,
		xmlBuilder:  xmlBuilder,
		logger:      logger,
	}
}

// Create implements Manager.Create
func (m *DomainManager) Create(ctx context.Context, params vm.VMParams) (*vm.VM, error) {
	// Get libvirt connection
	conn, err := m.connManager.Connect(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to libvirt: %w", err)
	}
	defer m.connManager.Release(conn)

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
		_ = libvirtConn.DomainUndefine(domain)
		return nil, fmt.Errorf("starting domain: %w", err)
	}

	// Get the created VM details
	return m.domainToVM(libvirtConn, domain)
}

// Get implements Manager.Get
func (m *DomainManager) Get(ctx context.Context, name string) (*vm.VM, error) {
	// Get libvirt connection
	conn, err := m.connManager.Connect(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to libvirt: %w", err)
	}
	defer m.connManager.Release(conn)

	libvirtConn := conn.GetLibvirtConnection()

	// Look up domain
	domain, err := libvirtConn.DomainLookupByName(name)
	if err != nil {
		return nil, fmt.Errorf("looking up domain %s: %w", name, ErrDomainNotFound)
	}

	// Convert to VM model
	return m.domainToVM(libvirtConn, domain)
}

// List implements Manager.List
func (m *DomainManager) List(ctx context.Context) ([]*vm.VM, error) {
	// Get libvirt connection
	conn, err := m.connManager.Connect(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to libvirt: %w", err)
	}
	defer m.connManager.Release(conn)

	libvirtConn := conn.GetLibvirtConnection()

	// Get all domains
	domains, err := libvirtConn.Domains()
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

// Start implements Manager.Start
func (m *DomainManager) Start(ctx context.Context, name string) error {
	// Get libvirt connection
	conn, err := m.connManager.Connect(ctx)
	if err != nil {
		return fmt.Errorf("failed to connect to libvirt: %w", err)
	}
	defer m.connManager.Release(conn)

	libvirtConn := conn.GetLibvirtConnection()

	// Look up domain
	domain, err := libvirtConn.DomainLookupByName(name)
	if err != nil {
		return fmt.Errorf("looking up domain %s: %w", name, ErrDomainNotFound)
	}

	// Check if domain is already running
	state, _, _, _, _, err := libvirtConn.DomainGetInfo(domain)
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
}

// Stop implements Manager.Stop
func (m *DomainManager) Stop(ctx context.Context, name string) error {
	// Get libvirt connection
	conn, err := m.connManager.Connect(ctx)
	if err != nil {
		return fmt.Errorf("failed to connect to libvirt: %w", err)
	}
	defer m.connManager.Release(conn)

	libvirtConn := conn.GetLibvirtConnection()

	// Look up domain
	domain, err := libvirtConn.DomainLookupByName(name)
	if err != nil {
		return fmt.Errorf("looking up domain %s: %w", name, ErrDomainNotFound)
	}

	// Check if domain is already stopped
	state, _, _, _, _, err := libvirtConn.DomainGetInfo(domain)
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
}

// ForceStop implements Manager.ForceStop
func (m *DomainManager) ForceStop(ctx context.Context, name string) error {
	// Get libvirt connection
	conn, err := m.connManager.Connect(ctx)
	if err != nil {
		return fmt.Errorf("failed to connect to libvirt: %w", err)
	}
	defer m.connManager.Release(conn)

	libvirtConn := conn.GetLibvirtConnection()

	// Look up domain
	domain, err := libvirtConn.DomainLookupByName(name)
	if err != nil {
		return fmt.Errorf("looking up domain %s: %w", name, ErrDomainNotFound)
	}

	// Check if domain is already stopped
	state, _, _, _, _, err := libvirtConn.DomainGetInfo(domain)
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
}

// Delete implements Manager.Delete
func (m *DomainManager) Delete(ctx context.Context, name string) error {
	// Get libvirt connection
	conn, err := m.connManager.Connect(ctx)
	if err != nil {
		return fmt.Errorf("failed to connect to libvirt: %w", err)
	}
	defer m.connManager.Release(conn)

	libvirtConn := conn.GetLibvirtConnection()

	// Look up domain
	domain, err := libvirtConn.DomainLookupByName(name)
	if err != nil {
		return fmt.Errorf("looking up domain %s: %w", name, ErrDomainNotFound)
	}

	// Check if domain is running
	state, _, _, _, _, err := libvirtConn.DomainGetInfo(domain)
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

// GetXML implements Manager.GetXML
func (m *DomainManager) GetXML(ctx context.Context, name string) (string, error) {
	// Get libvirt connection
	conn, err := m.connManager.Connect(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to connect to libvirt: %w", err)
	}
	defer m.connManager.Release(conn)

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

// domainToVM converts libvirt domain to VM model
func (m *DomainManager) domainToVM(libvirtConn *libvirt.Libvirt, domain libvirt.Domain) (*vm.VM, error) {
	// Get domain XML
	flags := libvirt.DomainXMLSecure
	xmlDesc, err := libvirtConn.DomainGetXMLDesc(domain, flags)
	if err != nil {
		return nil, fmt.Errorf("getting domain XML: %w", err)
	}

	// Parse XML
	var domainXML libvirtDomain
	if err := xml.Unmarshal([]byte(xmlDesc), &domainXML); err != nil {
		return nil, fmt.Errorf("parsing domain XML: %w", err)
	}

	// Get domain info (state, etc.)
	state, maxMem, _, _, _, err := libvirtConn.DomainGetInfo(domain)
	if err != nil {
		return nil, fmt.Errorf("getting domain info: %w", err)
	}

	// Convert memory (KiB to bytes)
	var memoryBytes uint64
	if domainXML.Memory.Unit == "KiB" {
		memoryBytes = domainXML.Memory.Value * 1024
	} else {
		// Fallback to info from DomainGetInfo
		memoryBytes = maxMem * 1024
	}

	// Convert status
	status := mapDomainState(state)

	// Create VM
	result := &vm.VM{
		Name:   domainXML.Name,
		UUID:   domainXML.UUID,
		Status: status,
		CPU: vm.CPUInfo{
			Count: domainXML.VCPUs,
		},
		Memory: vm.MemoryInfo{
			SizeBytes: memoryBytes,
		},
		CreatedAt: time.Now(), // TODO: Get actual creation time if available
	}

	// Process disks
	for _, disk := range domainXML.Devices.Disks {
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

		// Default to a reasonable size if we can't get it (will be updated with actual size later)
		// In a production environment, we would query the actual disk size
		var sizeBytes uint64 = 0

		// Generate a serial if none exists
		serial := ""

		diskInfo := vm.DiskInfo{
			Path:        path,
			Format:      vm.DiskFormat(disk.Driver.Type),
			SizeBytes:   sizeBytes,
			Bus:         vm.DiskBus(disk.Target.Bus),
			ReadOnly:    disk.ReadOnly != nil,
			Bootable:    disk.Boot.Order > 0,
			Shareable:   disk.Shareable != nil,
			Serial:      serial,
			StoragePool: storagePool,
			Device:      disk.Target.Dev,
		}

		result.Disks = append(result.Disks, diskInfo)
	}

	// Process network interfaces
	for _, iface := range domainXML.Devices.Interfaces {
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
		case "bridge":
			netType = vm.NetworkTypeBridge
		case "network":
			netType = vm.NetworkTypeNetwork
		case "direct":
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

		result.Networks = append(result.Networks, netInfo)
	}

	return result, nil
}

// mapDomainState maps libvirt domain state to VM status
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
