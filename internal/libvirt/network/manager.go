package network

import (
	"context"
	"fmt"
	"github.com/digitalocean/go-libvirt"
	"github.com/threatflux/libgo/internal/libvirt/connection"
	"github.com/threatflux/libgo/pkg/logger"
	"strings"
)

// LibvirtNetworkManager implements Manager for libvirt networks
type LibvirtNetworkManager struct {
	connManager connection.Manager
	xmlBuilder  XMLBuilder
	logger      logger.Logger
}

// NewLibvirtNetworkManager creates a new LibvirtNetworkManager
func NewLibvirtNetworkManager(connManager connection.Manager, xmlBuilder XMLBuilder, logger logger.Logger) *LibvirtNetworkManager {
	return &LibvirtNetworkManager{
		connManager: connManager,
		xmlBuilder:  xmlBuilder,
		logger:      logger,
	}
}

// EnsureExists implements Manager.EnsureExists
func (m *LibvirtNetworkManager) EnsureExists(ctx context.Context, name string, bridgeName string, cidr string, dhcp bool) error {
	conn, err := m.connManager.Connect(ctx)
	if err != nil {
		return fmt.Errorf("connecting to libvirt: %w", err)
	}
	defer m.connManager.Release(conn)

	libvirtConn := conn.GetLibvirtConnection()

	// Check if network already exists
	network, err := libvirtConn.NetworkLookupByName(name)
	if err == nil {
		// Network exists
		m.logger.Debug("Network already exists", logger.String("name", name))
		return nil
	}

	// Network doesn't exist, create it
	xml, err := m.xmlBuilder.BuildNetworkXML(name, bridgeName, cidr, dhcp)
	if err != nil {
		return fmt.Errorf("building network XML: %w", err)
	}

	m.logger.Debug("Creating network",
		logger.String("name", name),
		logger.String("bridge", bridgeName),
		logger.String("cidr", cidr),
		logger.Bool("dhcp", dhcp))

	network, err = libvirtConn.NetworkDefineXML(xml)
	if err != nil {
		return fmt.Errorf("defining network: %w", err)
	}

	err = libvirtConn.NetworkCreate(network)
	if err != nil {
		// Try to clean up the defined network
		_ = libvirtConn.NetworkUndefine(network)
		return fmt.Errorf("starting network: %w", err)
	}

	// Set network to autostart
	err = libvirtConn.NetworkSetAutostart(network, 1)
	if err != nil {
		m.logger.Warn("Failed to set network autostart",
			logger.String("name", name),
			logger.Error(err))
	}

	return nil
}

// Delete implements Manager.Delete
func (m *LibvirtNetworkManager) Delete(ctx context.Context, name string) error {
	conn, err := m.connManager.Connect(ctx)
	if err != nil {
		return fmt.Errorf("connecting to libvirt: %w", err)
	}
	defer m.connManager.Release(conn)

	libvirtConn := conn.GetLibvirtConnection()

	network, err := libvirtConn.NetworkLookupByName(name)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			// Network doesn't exist, nothing to delete
			m.logger.Debug("Network doesn't exist, nothing to delete", logger.String("name", name))
			return nil
		}
		return fmt.Errorf("looking up network: %w", err)
	}

	// Check if the network is active
	active, err := libvirtConn.NetworkIsActive(network)
	if err != nil {
		return fmt.Errorf("checking if network is active: %w", err)
	}

	// If active, destroy it first
	if active == 1 {
		if err := libvirtConn.NetworkDestroy(network); err != nil {
			return fmt.Errorf("destroying network: %w", err)
		}
	}

	// Undefine the network
	if err := libvirtConn.NetworkUndefine(network); err != nil {
		return fmt.Errorf("undefining network: %w", err)
	}

	m.logger.Info("Network deleted", logger.String("name", name))
	return nil
}

// Get implements Manager.Get
func (m *LibvirtNetworkManager) Get(ctx context.Context, name string) (*libvirt.Network, error) {
	conn, err := m.connManager.Connect(ctx)
	if err != nil {
		return nil, fmt.Errorf("connecting to libvirt: %w", err)
	}
	defer m.connManager.Release(conn)

	libvirtConn := conn.GetLibvirtConnection()

	network, err := libvirtConn.NetworkLookupByName(name)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			return nil, fmt.Errorf("network not found: %s", name)
		}
		return nil, fmt.Errorf("looking up network: %w", err)
	}

	return &network, nil
}

// GetDHCPLeases implements Manager.GetDHCPLeases
func (m *LibvirtNetworkManager) GetDHCPLeases(ctx context.Context, name string) ([]libvirt.NetworkDhcpLease, error) {
	conn, err := m.connManager.Connect(ctx)
	if err != nil {
		return nil, fmt.Errorf("connecting to libvirt: %w", err)
	}
	defer m.connManager.Release(conn)

	libvirtConn := conn.GetLibvirtConnection()

	network, err := libvirtConn.NetworkLookupByName(name)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			return nil, fmt.Errorf("network not found: %s", name)
		}
		return nil, fmt.Errorf("looking up network: %w", err)
	}

	leases, _, err := libvirtConn.NetworkGetDhcpLeases(network, nil, 0, 0)
	if err != nil {
		return nil, fmt.Errorf("getting DHCP leases: %w", err)
	}

	return leases, nil
}

// FindIPByMAC implements Manager.FindIPByMAC
func (m *LibvirtNetworkManager) FindIPByMAC(ctx context.Context, networkName string, mac string) (string, error) {
	leases, err := m.GetDHCPLeases(ctx, networkName)
	if err != nil {
		return "", err
	}

	// Normalize MAC address format for comparison
	mac = strings.ToLower(mac)
	mac = strings.ReplaceAll(mac, "-", ":")

	for _, lease := range leases {
		// Handle the slice type by joining the bytes
		var macStr string
		for _, b := range lease.Mac {
			if len(macStr) > 0 {
				macStr += ":"
			}
			macStr += fmt.Sprintf("%02x", b)
		}

		if strings.ToLower(macStr) == mac {
			// Also handle Ipaddr as a slice
			ipStr := ""
			for _, b := range lease.Ipaddr {
				ipStr += string(b)
			}
			return ipStr, nil
		}
	}

	return "", fmt.Errorf("no IP found for MAC address %s in network %s", mac, networkName)
}
