package ovs

import (
	"context"
	"fmt"
	"strings"

	"github.com/threatflux/libgo/internal/libvirt/connection"
	"github.com/threatflux/libgo/pkg/logger"
)

// LibvirtIntegration handles OVS-libvirt integration
type LibvirtIntegration struct {
	ovsManager  Manager
	connManager connection.Manager
	logger      logger.Logger
}

// NewLibvirtIntegration creates a new LibvirtIntegration
func NewLibvirtIntegration(ovsManager Manager, connManager connection.Manager, logger logger.Logger) *LibvirtIntegration {
	return &LibvirtIntegration{
		ovsManager:  ovsManager,
		connManager: connManager,
		logger:      logger,
	}
}

// CreateNetworkForBridge creates a libvirt network that uses an OVS bridge
func (l *LibvirtIntegration) CreateNetworkForBridge(ctx context.Context, networkName string, bridgeName string) error {
	// First ensure the OVS bridge exists
	_, err := l.ovsManager.GetBridge(ctx, bridgeName)
	if err != nil {
		return fmt.Errorf("OVS bridge %s not found: %w", bridgeName, err)
	}

	// Get libvirt connection
	conn, err := l.connManager.Connect(ctx)
	if err != nil {
		return fmt.Errorf("connecting to libvirt: %w", err)
	}
	defer func() {
		if releaseErr := l.connManager.Release(conn); releaseErr != nil {
			l.logger.Error("Failed to release libvirt connection", logger.Error(releaseErr))
		}
	}()

	libvirtConn := conn.GetLibvirtConnection()

	// Check if network already exists
	_, err = libvirtConn.NetworkLookupByName(networkName)
	if err == nil {
		return fmt.Errorf("libvirt network %s already exists", networkName)
	}

	// Create network XML that uses the OVS bridge
	networkXML := fmt.Sprintf(`
<network>
  <name>%s</name>
  <forward mode='bridge'/>
  <bridge name='%s'/>
  <virtualport type='openvswitch'/>
</network>`, networkName, bridgeName)

	// Define the network
	network, err := libvirtConn.NetworkDefineXML(networkXML)
	if err != nil {
		return fmt.Errorf("defining network: %w", err)
	}

	// Start the network
	if err := libvirtConn.NetworkCreate(network); err != nil {
		// Clean up on failure
		if undefineErr := libvirtConn.NetworkUndefine(network); undefineErr != nil {
			l.logger.Warn("Failed to undefine network during cleanup", logger.Error(undefineErr))
		}
		return fmt.Errorf("starting network: %w", err)
	}

	// Set autostart
	if err := libvirtConn.NetworkSetAutostart(network, 1); err != nil {
		l.logger.Warn("Failed to set network autostart",
			logger.String("network", networkName),
			logger.Error(err))
	}

	l.logger.Info("Created libvirt network for OVS bridge",
		logger.String("network", networkName),
		logger.String("bridge", bridgeName))

	return nil
}

// AttachVMToOVSBridge attaches a VM interface to an OVS bridge with optional VLAN
func (l *LibvirtIntegration) AttachVMToOVSBridge(ctx context.Context, vmName string, bridgeName string, vlan *int) (string, error) {
	// Ensure OVS bridge exists
	_, err := l.ovsManager.GetBridge(ctx, bridgeName)
	if err != nil {
		return "", fmt.Errorf("OVS bridge %s not found: %w", bridgeName, err)
	}

	// Generate interface XML for OVS
	interfaceXML := fmt.Sprintf(`
<interface type='bridge'>
  <source bridge='%s'/>
  <virtualport type='openvswitch'/>
  <model type='virtio'/>`, bridgeName)

	// Add VLAN tag if specified
	if vlan != nil {
		interfaceXML += fmt.Sprintf(`
  <vlan>
    <tag id='%d'/>
  </vlan>`, *vlan)
	}

	interfaceXML += `
</interface>`

	// Get libvirt connection
	conn, err := l.connManager.Connect(ctx)
	if err != nil {
		return "", fmt.Errorf("connecting to libvirt: %w", err)
	}
	defer func() {
		if releaseErr := l.connManager.Release(conn); releaseErr != nil {
			l.logger.Error("Failed to release libvirt connection", logger.Error(releaseErr))
		}
	}()

	libvirtConn := conn.GetLibvirtConnection()

	// Look up the domain
	domain, err := libvirtConn.DomainLookupByName(vmName)
	if err != nil {
		return "", fmt.Errorf("looking up domain: %w", err)
	}

	// Attach the interface
	if err := libvirtConn.DomainAttachDevice(domain, interfaceXML); err != nil {
		return "", fmt.Errorf("attaching interface: %w", err)
	}

	// Get the MAC address of the newly attached interface
	xmlDesc, err := libvirtConn.DomainGetXMLDesc(domain, 0)
	if err != nil {
		return "", fmt.Errorf("getting domain XML: %w", err)
	}

	// Parse XML to find the MAC address (simplified - would use proper XML parsing in production)
	macAddr := l.extractMACFromXML(xmlDesc, bridgeName)

	l.logger.Info("Attached VM to OVS bridge",
		logger.String("vm", vmName),
		logger.String("bridge", bridgeName),
		logger.String("mac", macAddr))

	return macAddr, nil
}

// CreateVXLANTunnel creates a VXLAN tunnel port on an OVS bridge
func (l *LibvirtIntegration) CreateVXLANTunnel(ctx context.Context, bridgeName string, tunnelName string, remoteIP string, vni int) error {
	options := &PortOptions{
		Type:     "vxlan",
		RemoteIP: remoteIP,
		OtherConfig: map[string]string{
			"key": fmt.Sprintf("%d", vni),
		},
	}

	if err := l.ovsManager.AddPort(ctx, bridgeName, tunnelName, options); err != nil {
		return fmt.Errorf("adding VXLAN tunnel: %w", err)
	}

	l.logger.Info("Created VXLAN tunnel",
		logger.String("bridge", bridgeName),
		logger.String("tunnel", tunnelName),
		logger.String("remote", remoteIP),
		logger.Int("vni", vni))

	return nil
}

// CreatePatchPort creates a patch port between two OVS bridges
func (l *LibvirtIntegration) CreatePatchPort(ctx context.Context, bridge1 string, bridge2 string) error {
	// Create patch port names
	patch1 := fmt.Sprintf("patch-%s-to-%s", bridge1, bridge2)
	patch2 := fmt.Sprintf("patch-%s-to-%s", bridge2, bridge1)

	// Add patch port to bridge1
	options1 := &PortOptions{
		Type:     "patch",
		PeerPort: patch2,
	}
	if err := l.ovsManager.AddPort(ctx, bridge1, patch1, options1); err != nil {
		return fmt.Errorf("adding patch port to %s: %w", bridge1, err)
	}

	// Add patch port to bridge2
	options2 := &PortOptions{
		Type:     "patch",
		PeerPort: patch1,
	}
	if err := l.ovsManager.AddPort(ctx, bridge2, patch2, options2); err != nil {
		// Clean up the first port on failure
		if deleteErr := l.ovsManager.DeletePort(ctx, bridge1, patch1); deleteErr != nil {
			l.logger.Warn("Failed to delete patch port during cleanup", logger.Error(deleteErr))
		}
		return fmt.Errorf("adding patch port to %s: %w", bridge2, err)
	}

	l.logger.Info("Created patch connection between bridges",
		logger.String("bridge1", bridge1),
		logger.String("bridge2", bridge2))

	return nil
}

// SetupQoSForVM sets up QoS rules for a VM's port on an OVS bridge
func (l *LibvirtIntegration) SetupQoSForVM(ctx context.Context, vmMAC string, bridgeName string, ingressRate int64, egressRate int64) error {
	// Find the port associated with the VM MAC
	ports, err := l.ovsManager.ListPorts(ctx, bridgeName)
	if err != nil {
		return fmt.Errorf("listing ports: %w", err)
	}

	var vmPort string
	for _, port := range ports {
		// In a real implementation, we'd check the port's MAC address
		// For now, we'll assume the port name contains part of the MAC
		if strings.Contains(port.Name, strings.ReplaceAll(vmMAC, ":", "")) {
			vmPort = port.Name
			break
		}
	}

	if vmPort == "" {
		return fmt.Errorf("port not found for MAC %s", vmMAC)
	}

	// Add QoS flow rules for ingress limiting
	if ingressRate > 0 {
		ingressFlow := &FlowRule{
			Table:    0,
			Priority: 100,
			Match:    fmt.Sprintf("in_port=%s", vmPort),
			Actions:  fmt.Sprintf("meter:%d,normal", ingressRate),
		}
		if err := l.ovsManager.AddFlow(ctx, bridgeName, ingressFlow); err != nil {
			return fmt.Errorf("adding ingress QoS flow: %w", err)
		}
	}

	// Add QoS flow rules for egress limiting
	if egressRate > 0 {
		egressFlow := &FlowRule{
			Table:    0,
			Priority: 100,
			Match:    fmt.Sprintf("dl_src=%s", vmMAC),
			Actions:  fmt.Sprintf("meter:%d,normal", egressRate),
		}
		if err := l.ovsManager.AddFlow(ctx, bridgeName, egressFlow); err != nil {
			return fmt.Errorf("adding egress QoS flow: %w", err)
		}
	}

	l.logger.Info("Set up QoS for VM",
		logger.String("mac", vmMAC),
		logger.String("bridge", bridgeName),
		logger.Int64("ingress_rate", ingressRate),
		logger.Int64("egress_rate", egressRate))

	return nil
}

// Helper method to extract MAC address from domain XML
func (l *LibvirtIntegration) extractMACFromXML(xml string, bridgeName string) string {
	// This is a simplified implementation
	// In production, use proper XML parsing
	lines := strings.Split(xml, "\n")
	inInterface := false

	for i, line := range lines {
		if l.isInterfaceStart(line) {
			inInterface = true
			continue
		}

		if inInterface && strings.Contains(line, "</interface>") {
			inInterface = false
			continue
		}

		if inInterface && strings.Contains(line, bridgeName) {
			if mac := l.findMACInLines(lines, i); mac != "" {
				return mac
			}
		}
	}
	return ""
}

// isInterfaceStart checks if line is the start of a bridge interface
func (l *LibvirtIntegration) isInterfaceStart(line string) bool {
	return strings.Contains(line, "<interface") && strings.Contains(line, "bridge")
}

// findMACInLines searches for MAC address in the next few lines
func (l *LibvirtIntegration) findMACInLines(lines []string, startIndex int) string {
	maxLines := 5
	for j := startIndex; j < len(lines) && j < startIndex+maxLines; j++ {
		if strings.Contains(lines[j], "mac address=") {
			return l.extractMACFromLine(lines[j])
		}
	}
	return ""
}

// extractMACFromLine extracts MAC address from a line containing "address="
func (l *LibvirtIntegration) extractMACFromLine(line string) string {
	addressPrefix := "address='"
	start := strings.Index(line, addressPrefix)
	if start == -1 {
		return ""
	}
	start += len(addressPrefix)

	end := strings.Index(line[start:], "'")
	if end <= 0 {
		return ""
	}

	return line[start : start+end]
}
