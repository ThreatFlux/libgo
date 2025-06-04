package network

import (
	"context"
	"encoding/xml"
	"fmt"
	"net"
	"strings"

	"github.com/digitalocean/go-libvirt"
	"github.com/threatflux/libgo/internal/libvirt/connection"
	"github.com/threatflux/libgo/pkg/logger"
)

// LibvirtNetworkManager implements Manager for libvirt networks.
type LibvirtNetworkManager struct {
	connManager connection.Manager
	xmlBuilder  XMLBuilder
	logger      logger.Logger
}

// NewLibvirtNetworkManager creates a new LibvirtNetworkManager.
func NewLibvirtNetworkManager(connManager connection.Manager, xmlBuilder XMLBuilder, logger logger.Logger) *LibvirtNetworkManager {
	return &LibvirtNetworkManager{
		connManager: connManager,
		xmlBuilder:  xmlBuilder,
		logger:      logger,
	}
}

// handleDeferredRelease logs any error from releasing a connection.
func (m *LibvirtNetworkManager) handleDeferredRelease(conn connection.Connection) {
	if err := m.connManager.Release(conn); err != nil {
		m.logger.Error("Failed to release connection", logger.Error(err))
	}
}

// EnsureExists implements Manager.EnsureExists.
func (m *LibvirtNetworkManager) EnsureExists(ctx context.Context, name string, bridgeName string, cidr string, dhcp bool) error {
	conn, err := m.connManager.Connect(ctx)
	if err != nil {
		return fmt.Errorf("connecting to libvirt: %w", err)
	}
	defer m.handleDeferredRelease(conn)

	libvirtConn := conn.GetLibvirtConnection()

	// Check if network already exists
	_, err = libvirtConn.NetworkLookupByName(name)
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

	network, err := libvirtConn.NetworkDefineXML(xml)
	if err != nil {
		return fmt.Errorf("defining network: %w", err)
	}

	err = libvirtConn.NetworkCreate(network)
	if err != nil {
		// Try to clean up the defined network - ignore cleanup errors
		if undefineErr := libvirtConn.NetworkUndefine(network); undefineErr != nil {
			m.logger.Debug("Failed to cleanup network after create failure", logger.Error(undefineErr))
		}
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

// Delete implements Manager.Delete.
func (m *LibvirtNetworkManager) Delete(ctx context.Context, name string) error {
	conn, err := m.connManager.Connect(ctx)
	if err != nil {
		return fmt.Errorf("connecting to libvirt: %w", err)
	}
	defer m.handleDeferredRelease(conn)

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

// Get implements Manager.Get.
func (m *LibvirtNetworkManager) Get(ctx context.Context, name string) (*libvirt.Network, error) {
	conn, err := m.connManager.Connect(ctx)
	if err != nil {
		return nil, fmt.Errorf("connecting to libvirt: %w", err)
	}
	defer m.handleDeferredRelease(conn)

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

// GetDHCPLeases implements Manager.GetDHCPLeases.
func (m *LibvirtNetworkManager) GetDHCPLeases(ctx context.Context, name string) ([]libvirt.NetworkDhcpLease, error) {
	conn, err := m.connManager.Connect(ctx)
	if err != nil {
		return nil, fmt.Errorf("connecting to libvirt: %w", err)
	}
	defer m.handleDeferredRelease(conn)

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

// FindIPByMAC implements Manager.FindIPByMAC.
func (m *LibvirtNetworkManager) FindIPByMAC(ctx context.Context, networkName string, mac string) (string, error) {
	leases, err := m.GetDHCPLeases(ctx, networkName)
	if err != nil {
		return "", err
	}

	// Normalize MAC address format for comparison
	mac = strings.ToLower(mac)
	mac = strings.ReplaceAll(mac, "-", ":")

	for _, lease := range leases {
		// Mac is OptString ([]string) - MAC address should be in the first element
		macStr := ""
		if len(lease.Mac) > 0 {
			macStr = lease.Mac[0]
		}

		if strings.ToLower(macStr) == mac {
			// Ipaddr is already a string, not OptString
			return lease.Ipaddr, nil
		}
	}

	return "", fmt.Errorf("no IP found for MAC address %s in network %s", mac, networkName)
}

// List implements Manager.List.
func (m *LibvirtNetworkManager) List(ctx context.Context) ([]*NetworkInfo, error) {
	conn, err := m.connManager.Connect(ctx)
	if err != nil {
		return nil, fmt.Errorf("connecting to libvirt: %w", err)
	}
	defer m.handleDeferredRelease(conn)

	libvirtConn := conn.GetLibvirtConnection()

	// Get all networks
	networks, _, err := libvirtConn.ConnectListAllNetworks(0, 0)
	if err != nil {
		return nil, fmt.Errorf("listing networks: %w", err)
	}

	result := make([]*NetworkInfo, 0, len(networks))
	for _, net := range networks {
		info, err := m.getNetworkInfo(libvirtConn, &net)
		if err != nil {
			m.logger.Warn("Failed to get network info",
				logger.String("network", net.Name),
				logger.Error(err))
			continue
		}
		result = append(result, info)
	}

	return result, nil
}

// Create implements Manager.Create.
func (m *LibvirtNetworkManager) Create(ctx context.Context, params *CreateNetworkParams) (*NetworkInfo, error) {
	conn, err := m.connManager.Connect(ctx)
	if err != nil {
		return nil, fmt.Errorf("connecting to libvirt: %w", err)
	}
	defer m.handleDeferredRelease(conn)

	libvirtConn := conn.GetLibvirtConnection()

	// Check if network already exists
	_, err = libvirtConn.NetworkLookupByName(params.Name)
	if err == nil {
		return nil, fmt.Errorf("network already exists: %s", params.Name)
	}

	// Build XML from parameters
	xml, err := m.buildNetworkXMLFromParams(params)
	if err != nil {
		return nil, fmt.Errorf("building network XML: %w", err)
	}

	// Define the network
	network, err := libvirtConn.NetworkDefineXML(xml)
	if err != nil {
		return nil, fmt.Errorf("defining network: %w", err)
	}

	// Start the network
	if err := libvirtConn.NetworkCreate(network); err != nil {
		// Clean up on failure - ignore cleanup errors
		if undefineErr := libvirtConn.NetworkUndefine(network); undefineErr != nil {
			m.logger.Debug("Failed to cleanup network after create failure", logger.Error(undefineErr))
		}
		return nil, fmt.Errorf("starting network: %w", err)
	}

	// Set autostart if requested
	if params.Autostart {
		if err := libvirtConn.NetworkSetAutostart(network, 1); err != nil {
			m.logger.Warn("Failed to set network autostart",
				logger.String("name", params.Name),
				logger.Error(err))
		}
	}

	// Get and return network info
	return m.getNetworkInfo(libvirtConn, &network)
}

// Update implements Manager.Update.
func (m *LibvirtNetworkManager) Update(ctx context.Context, name string, params *UpdateNetworkParams) (*NetworkInfo, error) {
	conn, err := m.connManager.Connect(ctx)
	if err != nil {
		return nil, fmt.Errorf("connecting to libvirt: %w", err)
	}
	defer m.handleDeferredRelease(conn)

	libvirtConn := conn.GetLibvirtConnection()

	network, err := libvirtConn.NetworkLookupByName(name)
	if err != nil {
		return nil, fmt.Errorf("looking up network: %w", err)
	}

	// Handle autostart update
	if params.Autostart != nil {
		autostart := 0
		if *params.Autostart {
			autostart = 1
		}
		if err := libvirtConn.NetworkSetAutostart(network, int32(autostart)); err != nil { //nolint:gosec
			return nil, fmt.Errorf("setting autostart: %w", err)
		}
	}

	// Note: Updating other network parameters typically requires
	// destroying and recreating the network, which would disrupt
	// connected VMs. For now, we only support autostart updates.

	return m.getNetworkInfo(libvirtConn, &network)
}

// GetInfo implements Manager.GetInfo.
func (m *LibvirtNetworkManager) GetInfo(ctx context.Context, name string) (*NetworkInfo, error) {
	conn, err := m.connManager.Connect(ctx)
	if err != nil {
		return nil, fmt.Errorf("connecting to libvirt: %w", err)
	}
	defer m.handleDeferredRelease(conn)

	libvirtConn := conn.GetLibvirtConnection()

	network, err := libvirtConn.NetworkLookupByName(name)
	if err != nil {
		return nil, fmt.Errorf("looking up network: %w", err)
	}

	return m.getNetworkInfo(libvirtConn, &network)
}

// GetXML implements Manager.GetXML.
func (m *LibvirtNetworkManager) GetXML(ctx context.Context, name string) (string, error) {
	conn, err := m.connManager.Connect(ctx)
	if err != nil {
		return "", fmt.Errorf("connecting to libvirt: %w", err)
	}
	defer m.handleDeferredRelease(conn)

	libvirtConn := conn.GetLibvirtConnection()

	network, err := libvirtConn.NetworkLookupByName(name)
	if err != nil {
		return "", fmt.Errorf("looking up network: %w", err)
	}

	xml, err := libvirtConn.NetworkGetXMLDesc(network, 0)
	if err != nil {
		return "", fmt.Errorf("getting network XML: %w", err)
	}

	return xml, nil
}

// IsActive implements Manager.IsActive.
func (m *LibvirtNetworkManager) IsActive(ctx context.Context, name string) (bool, error) {
	conn, err := m.connManager.Connect(ctx)
	if err != nil {
		return false, fmt.Errorf("connecting to libvirt: %w", err)
	}
	defer m.handleDeferredRelease(conn)

	libvirtConn := conn.GetLibvirtConnection()

	network, err := libvirtConn.NetworkLookupByName(name)
	if err != nil {
		return false, fmt.Errorf("looking up network: %w", err)
	}

	active, err := libvirtConn.NetworkIsActive(network)
	if err != nil {
		return false, fmt.Errorf("checking if network is active: %w", err)
	}

	return active == 1, nil
}

// Start implements Manager.Start.
func (m *LibvirtNetworkManager) Start(ctx context.Context, name string) error {
	conn, err := m.connManager.Connect(ctx)
	if err != nil {
		return fmt.Errorf("connecting to libvirt: %w", err)
	}
	defer m.handleDeferredRelease(conn)

	libvirtConn := conn.GetLibvirtConnection()

	network, err := libvirtConn.NetworkLookupByName(name)
	if err != nil {
		return fmt.Errorf("looking up network: %w", err)
	}

	// Check if already active
	active, err := libvirtConn.NetworkIsActive(network)
	if err != nil {
		return fmt.Errorf("checking if network is active: %w", err)
	}

	if active == 1 {
		return nil // Already active
	}

	if err := libvirtConn.NetworkCreate(network); err != nil {
		return fmt.Errorf("starting network: %w", err)
	}

	return nil
}

// Stop implements Manager.Stop.
func (m *LibvirtNetworkManager) Stop(ctx context.Context, name string) error {
	conn, err := m.connManager.Connect(ctx)
	if err != nil {
		return fmt.Errorf("connecting to libvirt: %w", err)
	}
	defer m.handleDeferredRelease(conn)

	libvirtConn := conn.GetLibvirtConnection()

	network, err := libvirtConn.NetworkLookupByName(name)
	if err != nil {
		return fmt.Errorf("looking up network: %w", err)
	}

	// Check if active
	active, err := libvirtConn.NetworkIsActive(network)
	if err != nil {
		return fmt.Errorf("checking if network is active: %w", err)
	}

	if active == 0 {
		return nil // Already stopped
	}

	if err := libvirtConn.NetworkDestroy(network); err != nil {
		return fmt.Errorf("stopping network: %w", err)
	}

	return nil
}

// SetAutostart implements Manager.SetAutostart.
func (m *LibvirtNetworkManager) SetAutostart(ctx context.Context, name string, autostart bool) error {
	conn, err := m.connManager.Connect(ctx)
	if err != nil {
		return fmt.Errorf("connecting to libvirt: %w", err)
	}
	defer m.handleDeferredRelease(conn)

	libvirtConn := conn.GetLibvirtConnection()

	network, err := libvirtConn.NetworkLookupByName(name)
	if err != nil {
		return fmt.Errorf("looking up network: %w", err)
	}

	autostartVal := 0
	if autostart {
		autostartVal = 1
	}

	if err := libvirtConn.NetworkSetAutostart(network, int32(autostartVal)); err != nil { //nolint:gosec
		return fmt.Errorf("setting autostart: %w", err)
	}

	return nil
}

// Helper methods

// getNetworkInfo builds a NetworkInfo struct from a libvirt network.
func (m *LibvirtNetworkManager) getNetworkInfo(conn *libvirt.Libvirt, network *libvirt.Network) (*NetworkInfo, error) {
	// Parse network definition from XML
	netDef, err := m.parseNetworkXML(conn, network)
	if err != nil {
		return nil, err
	}

	// Get status information
	statusInfo, err := m.getNetworkStatus(conn, network)
	if err != nil {
		return nil, err
	}

	// Build base network info
	info := m.buildBaseNetworkInfo(netDef, network, statusInfo)

	// Add IP configuration if present
	m.addIPConfiguration(info, netDef)

	// Add DHCP leases if network is active
	if err := m.addDHCPLeases(conn, network, info); err != nil {
		m.logger.Debug("Failed to get DHCP leases", logger.Error(err))
	}

	return info, nil
}

// parseNetworkXML gets and parses network XML definition.
func (m *LibvirtNetworkManager) parseNetworkXML(conn *libvirt.Libvirt, network *libvirt.Network) (*networkXMLDef, error) {
	xmlStr, err := conn.NetworkGetXMLDesc(*network, 0)
	if err != nil {
		return nil, fmt.Errorf("getting network XML: %w", err)
	}

	var netDef networkXMLDef
	if unmarshalErr := xml.Unmarshal([]byte(xmlStr), &netDef); unmarshalErr != nil {
		return nil, fmt.Errorf("parsing network XML: %w", unmarshalErr)
	}

	return &netDef, nil
}

// networkStatusInfo holds status information about a network.
type networkStatusInfo struct {
	active     bool
	persistent bool
	autostart  bool
}

// getNetworkStatus retrieves status information for a network.
func (m *LibvirtNetworkManager) getNetworkStatus(conn *libvirt.Libvirt, network *libvirt.Network) (*networkStatusInfo, error) {
	active, err := conn.NetworkIsActive(*network)
	if err != nil {
		return nil, fmt.Errorf("checking if network is active: %w", err)
	}

	persistent, err := conn.NetworkIsPersistent(*network)
	if err != nil {
		return nil, fmt.Errorf("checking if network is persistent: %w", err)
	}

	autostart, err := conn.NetworkGetAutostart(*network)
	if err != nil {
		m.logger.Warn("Failed to get network autostart", logger.Error(err))
		autostart = 0
	}

	return &networkStatusInfo{
		active:     active == 1,
		persistent: persistent == 1,
		autostart:  autostart == 1,
	}, nil
}

// buildBaseNetworkInfo creates the base NetworkInfo structure.
func (m *LibvirtNetworkManager) buildBaseNetworkInfo(netDef *networkXMLDef, network *libvirt.Network, status *networkStatusInfo) *NetworkInfo {
	uuidStr := formatUUID(network.UUID[:])

	info := &NetworkInfo{
		UUID:       uuidStr,
		Name:       netDef.Name,
		Active:     status.active,
		Persistent: status.persistent,
		Autostart:  status.autostart,
	}

	// Set bridge name if present
	if netDef.Bridge != nil {
		info.BridgeName = netDef.Bridge.Name
	}

	// Set forward mode if present
	if netDef.Forward != nil {
		info.Forward = NetworkForward{
			Mode: netDef.Forward.Mode,
			Dev:  netDef.Forward.Dev,
		}
	}

	return info
}

// addIPConfiguration adds IP configuration to NetworkInfo.
func (m *LibvirtNetworkManager) addIPConfiguration(info *NetworkInfo, netDef *networkXMLDef) {
	if netDef.IP == nil {
		return
	}

	info.IP = &NetworkIP{
		Address: netDef.IP.Address,
		Netmask: netDef.IP.Netmask,
	}

	// Parse DHCP configuration
	if netDef.IP.DHCP != nil {
		dhcpInfo := &NetworkDHCPInfo{
			Enabled: true,
		}

		// Set range if present
		if netDef.IP.DHCP.Range != nil {
			dhcpInfo.Start = netDef.IP.DHCP.Range.Start
			dhcpInfo.End = netDef.IP.DHCP.Range.End
		}

		// Add static hosts
		for _, host := range netDef.IP.DHCP.Hosts {
			dhcpInfo.Hosts = append(dhcpInfo.Hosts, NetworkDHCPStaticHost(host))
		}

		info.IP.DHCP = dhcpInfo
	}
}

// addDHCPLeases adds DHCP lease information to NetworkInfo.
func (m *LibvirtNetworkManager) addDHCPLeases(conn *libvirt.Libvirt, network *libvirt.Network, info *NetworkInfo) error {
	// Only get leases if network is active and has DHCP
	if !info.Active || info.IP == nil || info.IP.DHCP == nil {
		return nil
	}

	leases, _, err := conn.NetworkGetDhcpLeases(*network, nil, 0, 0)
	if err != nil {
		return fmt.Errorf("getting DHCP leases: %w", err)
	}

	for _, lease := range leases {
		dhcpLease := m.parseLeaseInfo(lease)
		info.DHCPLeases = append(info.DHCPLeases, dhcpLease)
	}

	return nil
}

// parseLeaseInfo extracts lease information from libvirt lease.
func (m *LibvirtNetworkManager) parseLeaseInfo(lease libvirt.NetworkDhcpLease) NetworkDHCPLease {
	// Extract values from OptString fields ([]string)
	macStr := ""
	if len(lease.Mac) > 0 {
		macStr = lease.Mac[0]
	}

	hostnameStr := ""
	if len(lease.Hostname) > 0 {
		hostnameStr = lease.Hostname[0]
	}

	clientIDStr := ""
	if len(lease.Clientid) > 0 {
		clientIDStr = lease.Clientid[0]
	}

	return NetworkDHCPLease{
		IPAddress:  lease.Ipaddr,
		MACAddress: macStr,
		Hostname:   hostnameStr,
		ClientID:   clientIDStr,
		ExpiryTime: lease.Expirytime,
	}
}

// buildNetworkXMLFromParams builds network XML from CreateNetworkParams.
func (m *LibvirtNetworkManager) buildNetworkXMLFromParams(params *CreateNetworkParams) (string, error) {
	// Use simple builder for basic configurations
	if m.shouldUseSimpleBuilder(params) {
		return m.buildSimpleNetworkXML(params)
	}

	// Build complex XML with detailed configuration
	netDef := m.buildNetworkDefinition(params)
	if err := m.addDHCPConfiguration(netDef, params); err != nil {
		return "", err
	}

	// Marshal to XML
	xmlData, err := xml.MarshalIndent(netDef, "", "  ")
	if err != nil {
		return "", fmt.Errorf("marshaling network XML: %w", err)
	}

	return string(xmlData), nil
}

// shouldUseSimpleBuilder determines if simple XML builder should be used.
func (m *LibvirtNetworkManager) shouldUseSimpleBuilder(params *CreateNetworkParams) bool {
	return params.IP == nil || params.IP.Address == ""
}

// buildSimpleNetworkXML builds XML using the simple builder for backward compatibility.
func (m *LibvirtNetworkManager) buildSimpleNetworkXML(params *CreateNetworkParams) (string, error) {
	dhcp := false
	if params.IP != nil && params.IP.DHCP != nil {
		dhcp = params.IP.DHCP.Enabled
	}
	return m.xmlBuilder.BuildNetworkXML(
		params.Name,
		params.BridgeName,
		"192.168.100.0/24", // Default CIDR
		dhcp,
	)
}

// buildNetworkDefinition creates the base network definition.
func (m *LibvirtNetworkManager) buildNetworkDefinition(params *CreateNetworkParams) *networkXMLDef {
	netDef := &networkXMLDef{
		Name: params.Name,
	}

	// Set bridge name
	if params.BridgeName != "" {
		netDef.Bridge = &networkBridge{Name: params.BridgeName}
	}

	// Set forward mode
	if params.Forward != nil {
		netDef.Forward = &networkForward{
			Mode: params.Forward.Mode,
			Dev:  params.Forward.Dev,
		}
	} else {
		// Default to NAT
		netDef.Forward = &networkForward{Mode: "nat"}
	}

	// Set IP configuration
	if params.IP != nil {
		netDef.IP = &networkIP{
			Address: params.IP.Address,
			Netmask: params.IP.Netmask,
		}
	}

	return netDef
}

// addDHCPConfiguration adds DHCP configuration to network definition.
func (m *LibvirtNetworkManager) addDHCPConfiguration(netDef *networkXMLDef, params *CreateNetworkParams) error {
	if params.IP == nil || params.IP.DHCP == nil || !params.IP.DHCP.Enabled {
		return nil
	}

	dhcp := &networkDHCP{}

	// Calculate or set DHCP range
	if params.IP.DHCP.Start != "" && params.IP.DHCP.End != "" {
		dhcp.Range = &dhcpRange{
			Start: params.IP.DHCP.Start,
			End:   params.IP.DHCP.End,
		}
	} else {
		// Auto-calculate range
		start, end, err := calculateDHCPRange(params.IP.Address, params.IP.Netmask)
		if err != nil {
			return fmt.Errorf("calculating DHCP range: %w", err)
		}
		dhcp.Range = &dhcpRange{Start: start, End: end}
	}

	// Add static hosts
	for _, host := range params.IP.DHCP.Hosts {
		dhcp.Hosts = append(dhcp.Hosts, dhcpHost(host))
	}

	netDef.IP.DHCP = dhcp
	return nil
}

// XML structures for parsing.
type networkXMLDef struct {
	Forward *networkForward `xml:"forward,omitempty"`
	Bridge  *networkBridge  `xml:"bridge,omitempty"`
	IP      *networkIP      `xml:"ip,omitempty"`
	XMLName xml.Name        `xml:"network"`
	Name    string          `xml:"name"`
	UUID    string          `xml:"uuid,omitempty"`
}

type networkForward struct {
	Mode string `xml:"mode,attr,omitempty"`
	Dev  string `xml:"dev,attr,omitempty"`
}

type networkBridge struct {
	Name string `xml:"name,attr"`
}

type networkIP struct {
	DHCP    *networkDHCP `xml:"dhcp,omitempty"`
	Address string       `xml:"address,attr"`
	Netmask string       `xml:"netmask,attr"`
}

type networkDHCP struct {
	Range *dhcpRange `xml:"range,omitempty"`
	Hosts []dhcpHost `xml:"host,omitempty"`
}

type dhcpRange struct {
	Start string `xml:"start,attr"`
	End   string `xml:"end,attr"`
}

type dhcpHost struct {
	MAC  string `xml:"mac,attr"`
	Name string `xml:"name,attr,omitempty"`
	IP   string `xml:"ip,attr"`
}

// Helper functions

func formatUUID(uuid []byte) string {
	if len(uuid) != 16 {
		return ""
	}
	return fmt.Sprintf("%02x%02x%02x%02x-%02x%02x-%02x%02x-%02x%02x-%02x%02x%02x%02x%02x%02x",
		uuid[0], uuid[1], uuid[2], uuid[3],
		uuid[4], uuid[5],
		uuid[6], uuid[7],
		uuid[8], uuid[9],
		uuid[10], uuid[11], uuid[12], uuid[13], uuid[14], uuid[15])
}

func calculateDHCPRange(ipAddr, netmask string) (start, end string, err error) {
	ip := net.ParseIP(ipAddr)
	if ip == nil {
		return "", "", fmt.Errorf("invalid IP address: %s", ipAddr)
	}

	mask := net.ParseIP(netmask)
	if mask == nil {
		return "", "", fmt.Errorf("invalid netmask: %s", netmask)
	}

	// Convert to IPv4
	ip = ip.To4()
	mask = mask.To4()

	// Calculate network address
	network := make(net.IP, 4)
	for i := 0; i < 4; i++ {
		network[i] = ip[i] & mask[i]
	}

	// Start at .100
	startIP := make(net.IP, 4)
	copy(startIP, network)
	startIP[3] = 100

	// End at .200
	endIP := make(net.IP, 4)
	copy(endIP, network)
	endIP[3] = 200

	return startIP.String(), endIP.String(), nil
}
