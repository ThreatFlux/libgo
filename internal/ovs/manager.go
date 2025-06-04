package ovs

import (
	"context"
	"fmt"
	"strings"

	"github.com/threatflux/libgo/pkg/logger"
	"github.com/threatflux/libgo/pkg/utils/exec"
)

const (
	emptyListString = "[]\n"
)

// OVSManager implements Manager for Open vSwitch operations.
type OVSManager struct {
	executor exec.CommandExecutor
	logger   logger.Logger
}

// NewOVSManager creates a new OVSManager.
func NewOVSManager(executor exec.CommandExecutor, logger logger.Logger) *OVSManager {
	return &OVSManager{
		executor: executor,
		logger:   logger,
	}
}

// CreateBridge implements Manager.CreateBridge.
func (m *OVSManager) CreateBridge(ctx context.Context, name string) error {
	// Check if bridge already exists
	exists, err := m.bridgeExists(ctx, name)
	if err != nil {
		return fmt.Errorf("checking if bridge exists: %w", err)
	}
	if exists {
		return fmt.Errorf("bridge %s already exists", name)
	}

	// Create the bridge
	cmd := []string{"ovs-vsctl", "add-br", name}
	if _, err := m.executor.ExecuteContext(ctx, cmd[0], cmd[1:]...); err != nil {
		return fmt.Errorf("creating bridge: %w", err)
	}

	m.logger.Info("OVS bridge created", logger.String("bridge", name))
	return nil
}

// DeleteBridge implements Manager.DeleteBridge.
func (m *OVSManager) DeleteBridge(ctx context.Context, name string) error {
	// Check if bridge exists
	exists, err := m.bridgeExists(ctx, name)
	if err != nil {
		return fmt.Errorf("checking if bridge exists: %w", err)
	}
	if !exists {
		return nil // Bridge doesn't exist, nothing to delete
	}

	// Delete the bridge
	cmd := []string{"ovs-vsctl", "del-br", name}
	if _, err := m.executor.ExecuteContext(ctx, cmd[0], cmd[1:]...); err != nil {
		return fmt.Errorf("deleting bridge: %w", err)
	}

	m.logger.Info("OVS bridge deleted", logger.String("bridge", name))
	return nil
}

// ListBridges implements Manager.ListBridges.
func (m *OVSManager) ListBridges(ctx context.Context) ([]BridgeInfo, error) {
	// Get list of bridges
	cmd := []string{"ovs-vsctl", "list-br"}
	output, err := m.executor.ExecuteContext(ctx, cmd[0], cmd[1:]...)
	if err != nil {
		return nil, fmt.Errorf("listing bridges: %w", err)
	}

	bridgeNames := strings.Split(strings.TrimSpace(string(output)), "\n")
	if len(bridgeNames) == 1 && bridgeNames[0] == "" {
		return []BridgeInfo{}, nil
	}

	// Get details for each bridge
	bridges := make([]BridgeInfo, 0, len(bridgeNames))
	for _, name := range bridgeNames {
		if name == "" {
			continue
		}
		bridge, err := m.GetBridge(ctx, name)
		if err != nil {
			m.logger.Warn("Failed to get bridge info",
				logger.String("bridge", name),
				logger.Error(err))
			continue
		}
		bridges = append(bridges, *bridge)
	}

	return bridges, nil
}

// GetBridge implements Manager.GetBridge.
func (m *OVSManager) GetBridge(ctx context.Context, name string) (*BridgeInfo, error) {
	// Check if bridge exists
	exists, err := m.bridgeExists(ctx, name)
	if err != nil {
		return nil, fmt.Errorf("checking if bridge exists: %w", err)
	}
	if !exists {
		return nil, fmt.Errorf("bridge %s not found", name)
	}

	info := &BridgeInfo{
		Name: name,
	}

	// Get bridge UUID
	cmd := []string{"ovs-vsctl", "get", "Bridge", name, "_uuid"}
	output, err := m.executor.ExecuteContext(ctx, cmd[0], cmd[1:]...)
	if err == nil {
		info.UUID = strings.TrimSpace(string(output))
	}

	// Get controller
	cmd = []string{"ovs-vsctl", "get", "Bridge", name, "controller"}
	output, err = m.executor.ExecuteContext(ctx, cmd[0], cmd[1:]...)
	if err == nil && string(output) != emptyListString {
		info.Controller = strings.Trim(string(output), emptyListString)
	}

	// Get datapath type
	cmd = []string{"ovs-vsctl", "get", "Bridge", name, "datapath_type"}
	output, err = m.executor.ExecuteContext(ctx, cmd[0], cmd[1:]...)
	if err == nil {
		info.DatapathType = strings.Trim(string(output), "\"\n")
	}

	// Get ports
	cmd = []string{"ovs-vsctl", "list-ports", name}
	output, err = m.executor.ExecuteContext(ctx, cmd[0], cmd[1:]...)
	if err == nil {
		ports := strings.Split(strings.TrimSpace(string(output)), "\n")
		if len(ports) == 1 && ports[0] == "" {
			info.Ports = []string{}
		} else {
			info.Ports = ports
		}
	}

	// Get external IDs
	cmd = []string{"ovs-vsctl", "get", "Bridge", name, "external_ids"}
	output, err = m.executor.ExecuteContext(ctx, cmd[0], cmd[1:]...)
	if err == nil {
		info.ExternalIDs = m.parseOVSMap(string(output))
	}

	// Get other config
	cmd = []string{"ovs-vsctl", "get", "Bridge", name, "other_config"}
	output, err = m.executor.ExecuteContext(ctx, cmd[0], cmd[1:]...)
	if err == nil {
		info.OtherConfig = m.parseOVSMap(string(output))
	}

	// Get statistics
	stats, err := m.GetBridgeStats(ctx, name)
	if err == nil {
		info.Statistics = stats
	}

	return info, nil
}

// AddPort implements Manager.AddPort.
func (m *OVSManager) AddPort(ctx context.Context, bridge string, port string, options *PortOptions) error {
	if err := m.validateBridgeExists(ctx, bridge); err != nil {
		return err
	}

	cmd := m.buildAddPortCommand(bridge, port, options)
	if err := m.executeAddPortCommand(ctx, cmd); err != nil {
		return err
	}

	if err := m.configurePortOptions(ctx, bridge, port, options); err != nil {
		return err
	}

	m.logger.Info("Port added to OVS bridge",
		logger.String("bridge", bridge),
		logger.String("port", port))
	return nil
}

// validateBridgeExists checks if the bridge exists.
func (m *OVSManager) validateBridgeExists(ctx context.Context, bridge string) error {
	exists, err := m.bridgeExists(ctx, bridge)
	if err != nil {
		return fmt.Errorf("checking if bridge exists: %w", err)
	}
	if !exists {
		return fmt.Errorf("bridge %s not found", bridge)
	}
	return nil
}

// buildAddPortCommand builds the ovs-vsctl command.
func (m *OVSManager) buildAddPortCommand(bridge, port string, options *PortOptions) []string {
	cmd := []string{"ovs-vsctl", "add-port", bridge, port}

	if options == nil {
		return cmd
	}

	// Set port type
	if options.Type != "" {
		cmd = append(cmd, "--", "set", "Interface", port, fmt.Sprintf("type=%s", options.Type))
	}

	// Set tunnel options for tunnel ports
	if m.isTunnelPort(options.Type) && options.RemoteIP != "" {
		cmd = append(cmd, fmt.Sprintf("options:remote_ip=%s", options.RemoteIP))
	}

	// Set patch port peer
	if options.Type == "patch" && options.PeerPort != "" {
		cmd = append(cmd, fmt.Sprintf("options:peer=%s", options.PeerPort))
	}

	return cmd
}

// isTunnelPort checks if the port type is a tunnel port.
func (m *OVSManager) isTunnelPort(portType string) bool {
	return portType == "vxlan" || portType == "gre" || portType == "geneve"
}

// executeAddPortCommand executes the add port command.
func (m *OVSManager) executeAddPortCommand(ctx context.Context, cmd []string) error {
	if _, err := m.executor.ExecuteContext(ctx, cmd[0], cmd[1:]...); err != nil {
		return fmt.Errorf("adding port: %w", err)
	}
	return nil
}

// configurePortOptions configures VLAN and trunk options.
func (m *OVSManager) configurePortOptions(ctx context.Context, bridge, port string, options *PortOptions) error {
	if options == nil {
		return nil
	}

	// Set VLAN tag if specified
	if options.Tag != nil {
		if err := m.SetPortVLAN(ctx, bridge, port, *options.Tag); err != nil {
			m.cleanupPortOnError(ctx, bridge, port)
			return fmt.Errorf("setting VLAN tag: %w", err)
		}
	}

	// Set trunk VLANs if specified
	if len(options.Trunks) > 0 {
		if err := m.SetPortTrunk(ctx, bridge, port, options.Trunks); err != nil {
			m.cleanupPortOnError(ctx, bridge, port)
			return fmt.Errorf("setting trunk VLANs: %w", err)
		}
	}

	return nil
}

// cleanupPortOnError attempts to cleanup a port during error recovery.
func (m *OVSManager) cleanupPortOnError(ctx context.Context, bridge, port string) {
	if deleteErr := m.DeletePort(ctx, bridge, port); deleteErr != nil {
		m.logger.Debug("Failed to cleanup port during error recovery", logger.Error(deleteErr))
	}
}

// DeletePort implements Manager.DeletePort.
func (m *OVSManager) DeletePort(ctx context.Context, bridge string, port string) error {
	// Check if port exists
	exists, err := m.portExists(ctx, bridge, port)
	if err != nil {
		return fmt.Errorf("checking if port exists: %w", err)
	}
	if !exists {
		return nil // Port doesn't exist, nothing to delete
	}

	// Delete the port
	cmd := []string{"ovs-vsctl", "del-port", bridge, port}
	if _, err := m.executor.ExecuteContext(ctx, cmd[0], cmd[1:]...); err != nil {
		return fmt.Errorf("deleting port: %w", err)
	}

	m.logger.Info("Port deleted from OVS bridge",
		logger.String("bridge", bridge),
		logger.String("port", port))
	return nil
}

// ListPorts implements Manager.ListPorts.
func (m *OVSManager) ListPorts(ctx context.Context, bridge string) ([]PortInfo, error) {
	// Check if bridge exists
	exists, err := m.bridgeExists(ctx, bridge)
	if err != nil {
		return nil, fmt.Errorf("checking if bridge exists: %w", err)
	}
	if !exists {
		return nil, fmt.Errorf("bridge %s not found", bridge)
	}

	// Get list of ports
	cmd := []string{"ovs-vsctl", "list-ports", bridge}
	output, err := m.executor.ExecuteContext(ctx, cmd[0], cmd[1:]...)
	if err != nil {
		return nil, fmt.Errorf("listing ports: %w", err)
	}

	portNames := strings.Split(strings.TrimSpace(string(output)), "\n")
	if len(portNames) == 1 && portNames[0] == "" {
		return []PortInfo{}, nil
	}

	// Get details for each port
	ports := make([]PortInfo, 0, len(portNames))
	for _, name := range portNames {
		if name == "" {
			continue
		}
		port, err := m.GetPort(ctx, bridge, name)
		if err != nil {
			m.logger.Warn("Failed to get port info",
				logger.String("port", name),
				logger.Error(err))
			continue
		}
		ports = append(ports, *port)
	}

	return ports, nil
}

// GetPort implements Manager.GetPort.
func (m *OVSManager) GetPort(ctx context.Context, bridge string, port string) (*PortInfo, error) {
	// Check if port exists
	exists, err := m.portExists(ctx, bridge, port)
	if err != nil {
		return nil, fmt.Errorf("checking if port exists: %w", err)
	}
	if !exists {
		return nil, fmt.Errorf("port %s not found on bridge %s", port, bridge)
	}

	info := &PortInfo{
		Name:   port,
		Bridge: bridge,
	}

	// Get port UUID
	cmd := []string{"ovs-vsctl", "get", "Port", port, "_uuid"}
	output, err := m.executor.ExecuteContext(ctx, cmd[0], cmd[1:]...)
	if err == nil {
		info.UUID = strings.TrimSpace(string(output))
	}

	// Get interface type
	cmd = []string{"ovs-vsctl", "get", "Interface", port, "type"}
	output, err = m.executor.ExecuteContext(ctx, cmd[0], cmd[1:]...)
	if err == nil {
		info.Type = strings.Trim(string(output), "\"\n")
	}

	// Get VLAN tag
	cmd = []string{"ovs-vsctl", "get", "Port", port, "tag"}
	output, err = m.executor.ExecuteContext(ctx, cmd[0], cmd[1:]...)
	if err == nil && string(output) != emptyListString {
		var tag int
		if _, scanErr := fmt.Sscanf(string(output), "%d", &tag); scanErr == nil {
			info.Tag = &tag
		}
	}

	// Get trunk VLANs
	cmd = []string{"ovs-vsctl", "get", "Port", port, "trunks"}
	output, err = m.executor.ExecuteContext(ctx, cmd[0], cmd[1:]...)
	if err == nil && string(output) != emptyListString {
		info.Trunks = m.parseOVSIntArray(string(output))
	}

	// Get statistics
	stats, err := m.GetPortStats(ctx, bridge, port)
	if err == nil {
		info.Statistics = stats
	}

	return info, nil
}

// SetPortVLAN implements Manager.SetPortVLAN.
func (m *OVSManager) SetPortVLAN(ctx context.Context, bridge string, port string, vlan int) error {
	// Check if port exists
	exists, err := m.portExists(ctx, bridge, port)
	if err != nil {
		return fmt.Errorf("checking if port exists: %w", err)
	}
	if !exists {
		return fmt.Errorf("port %s not found on bridge %s", port, bridge)
	}

	// Set VLAN tag
	cmd := []string{"ovs-vsctl", "set", "Port", port, fmt.Sprintf("tag=%d", vlan)}
	if _, err := m.executor.ExecuteContext(ctx, cmd[0], cmd[1:]...); err != nil {
		return fmt.Errorf("setting VLAN tag: %w", err)
	}

	m.logger.Info("VLAN tag set on port",
		logger.String("port", port),
		logger.Int("vlan", vlan))
	return nil
}

// SetPortTrunk implements Manager.SetPortTrunk.
func (m *OVSManager) SetPortTrunk(ctx context.Context, bridge string, port string, vlans []int) error {
	// Check if port exists
	exists, err := m.portExists(ctx, bridge, port)
	if err != nil {
		return fmt.Errorf("checking if port exists: %w", err)
	}
	if !exists {
		return fmt.Errorf("port %s not found on bridge %s", port, bridge)
	}

	// Build VLAN list
	vlanStrs := make([]string, len(vlans))
	for i, vlan := range vlans {
		vlanStrs[i] = fmt.Sprintf("%d", vlan)
	}
	vlanList := strings.Join(vlanStrs, ",")

	// Set trunk VLANs
	cmd := []string{"ovs-vsctl", "set", "Port", port, fmt.Sprintf("trunks=[%s]", vlanList)}
	if _, err := m.executor.ExecuteContext(ctx, cmd[0], cmd[1:]...); err != nil {
		return fmt.Errorf("setting trunk VLANs: %w", err)
	}

	m.logger.Info("Trunk VLANs set on port",
		logger.String("port", port),
		logger.Any("vlans", vlans))
	return nil
}

// AddFlow implements Manager.AddFlow.
func (m *OVSManager) AddFlow(ctx context.Context, bridge string, flow *FlowRule) error {
	// Check if bridge exists
	exists, err := m.bridgeExists(ctx, bridge)
	if err != nil {
		return fmt.Errorf("checking if bridge exists: %w", err)
	}
	if !exists {
		return fmt.Errorf("bridge %s not found", bridge)
	}

	// Build flow specification
	flowSpec := fmt.Sprintf("table=%d,priority=%d,%s,actions=%s",
		flow.Table, flow.Priority, flow.Match, flow.Actions)

	if flow.Cookie != "" {
		flowSpec = fmt.Sprintf("cookie=%s,%s", flow.Cookie, flowSpec)
	}

	// Add flow
	cmd := []string{"ovs-ofctl", "add-flow", bridge, flowSpec}
	if _, err := m.executor.ExecuteContext(ctx, cmd[0], cmd[1:]...); err != nil {
		return fmt.Errorf("adding flow: %w", err)
	}

	m.logger.Info("Flow rule added",
		logger.String("bridge", bridge),
		logger.String("flow", flowSpec))
	return nil
}

// DeleteFlow implements Manager.DeleteFlow.
func (m *OVSManager) DeleteFlow(ctx context.Context, bridge string, flowID string) error {
	// For now, we'll delete by cookie
	cmd := []string{"ovs-ofctl", "del-flows", bridge, fmt.Sprintf("cookie=%s/-1", flowID)}
	if _, err := m.executor.ExecuteContext(ctx, cmd[0], cmd[1:]...); err != nil {
		return fmt.Errorf("deleting flow: %w", err)
	}

	return nil
}

// ListFlows implements Manager.ListFlows.
func (m *OVSManager) ListFlows(ctx context.Context, bridge string) ([]FlowRule, error) {
	// Check if bridge exists
	exists, err := m.bridgeExists(ctx, bridge)
	if err != nil {
		return nil, fmt.Errorf("checking if bridge exists: %w", err)
	}
	if !exists {
		return nil, fmt.Errorf("bridge %s not found", bridge)
	}

	// Get flows
	cmd := []string{"ovs-ofctl", "dump-flows", bridge}
	output, err := m.executor.ExecuteContext(ctx, cmd[0], cmd[1:]...)
	if err != nil {
		return nil, fmt.Errorf("listing flows: %w", err)
	}

	// Parse flows (simplified - would need more complex parsing in production)
	lines := strings.Split(string(output), "\n")
	flows := make([]FlowRule, 0)
	for _, line := range lines {
		if strings.Contains(line, "cookie=") {
			// Basic parsing - would need improvement
			flow := FlowRule{}
			flows = append(flows, flow)
		}
	}

	return flows, nil
}

// SetController implements Manager.SetController.
func (m *OVSManager) SetController(ctx context.Context, bridge string, controller string) error {
	// Check if bridge exists
	exists, err := m.bridgeExists(ctx, bridge)
	if err != nil {
		return fmt.Errorf("checking if bridge exists: %w", err)
	}
	if !exists {
		return fmt.Errorf("bridge %s not found", bridge)
	}

	// Set controller
	cmd := []string{"ovs-vsctl", "set-controller", bridge, controller}
	if _, err := m.executor.ExecuteContext(ctx, cmd[0], cmd[1:]...); err != nil {
		return fmt.Errorf("setting controller: %w", err)
	}

	m.logger.Info("Controller set on bridge",
		logger.String("bridge", bridge),
		logger.String("controller", controller))
	return nil
}

// DeleteController implements Manager.DeleteController.
func (m *OVSManager) DeleteController(ctx context.Context, bridge string) error {
	// Check if bridge exists
	exists, err := m.bridgeExists(ctx, bridge)
	if err != nil {
		return fmt.Errorf("checking if bridge exists: %w", err)
	}
	if !exists {
		return fmt.Errorf("bridge %s not found", bridge)
	}

	// Delete controller
	cmd := []string{"ovs-vsctl", "del-controller", bridge}
	if _, err := m.executor.ExecuteContext(ctx, cmd[0], cmd[1:]...); err != nil {
		return fmt.Errorf("deleting controller: %w", err)
	}

	m.logger.Info("Controller deleted from bridge", logger.String("bridge", bridge))
	return nil
}

// GetBridgeStats implements Manager.GetBridgeStats.
func (m *OVSManager) GetBridgeStats(ctx context.Context, bridge string) (*BridgeStats, error) {
	// Get flow count
	cmd := []string{"ovs-ofctl", "dump-flows", bridge}
	output, err := m.executor.ExecuteContext(ctx, cmd[0], cmd[1:]...)
	if err != nil {
		return nil, fmt.Errorf("getting flow count: %w", err)
	}

	flowCount := int64(len(strings.Split(string(output), "\n")) - 2) // Subtract header lines

	// Get port count
	cmd = []string{"ovs-vsctl", "list-ports", bridge}
	output, err = m.executor.ExecuteContext(ctx, cmd[0], cmd[1:]...)
	if err != nil {
		return nil, fmt.Errorf("getting port count: %w", err)
	}

	portCount := int64(0)
	if string(output) != "" {
		portCount = int64(len(strings.Split(strings.TrimSpace(string(output)), "\n")))
	}

	return &BridgeStats{
		FlowCount: flowCount,
		PortCount: portCount,
	}, nil
}

// GetPortStats implements Manager.GetPortStats.
func (m *OVSManager) GetPortStats(ctx context.Context, bridge string, port string) (*PortStats, error) {
	// Get port statistics
	cmd := []string{"ovs-vsctl", "get", "Interface", port, "statistics"}
	output, err := m.executor.ExecuteContext(ctx, cmd[0], cmd[1:]...)
	if err != nil {
		return nil, fmt.Errorf("getting port statistics: %w", err)
	}

	// Parse statistics map
	statsMap := m.parseOVSMap(string(output))
	stats := &PortStats{}

	// Extract values using helper function
	m.parseStatValue(statsMap, "rx_packets", &stats.RxPackets)
	m.parseStatValue(statsMap, "rx_bytes", &stats.RxBytes)
	m.parseStatValue(statsMap, "rx_dropped", &stats.RxDropped)
	m.parseStatValue(statsMap, "rx_errors", &stats.RxErrors)
	m.parseStatValue(statsMap, "tx_packets", &stats.TxPackets)
	m.parseStatValue(statsMap, "tx_bytes", &stats.TxBytes)
	m.parseStatValue(statsMap, "tx_dropped", &stats.TxDropped)
	m.parseStatValue(statsMap, "tx_errors", &stats.TxErrors)

	return stats, nil
}

// parseStatValue parses a single statistic value from the map.
func (m *OVSManager) parseStatValue(statsMap map[string]string, key string, dest *uint64) {
	if val, ok := statsMap[key]; ok {
		if _, scanErr := fmt.Sscanf(val, "%d", dest); scanErr != nil {
			m.logger.Debug(fmt.Sprintf("Failed to parse %s", key), logger.Error(scanErr))
		}
	}
}

// Helper methods.

func (m *OVSManager) bridgeExists(ctx context.Context, name string) (bool, error) {
	cmd := []string{"ovs-vsctl", "br-exists", name}
	_, err := m.executor.ExecuteContext(ctx, cmd[0], cmd[1:]...)
	if err != nil {
		// ovs-vsctl returns exit code 2 if bridge doesn't exist
		if strings.Contains(err.Error(), "exit status 2") {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func (m *OVSManager) portExists(ctx context.Context, bridge string, port string) (bool, error) {
	cmd := []string{"ovs-vsctl", "list-ports", bridge}
	output, err := m.executor.ExecuteContext(ctx, cmd[0], cmd[1:]...)
	if err != nil {
		return false, err
	}

	ports := strings.Split(strings.TrimSpace(string(output)), "\n")
	for _, p := range ports {
		if p == port {
			return true, nil
		}
	}
	return false, nil
}

func (m *OVSManager) parseOVSMap(input string) map[string]string {
	result := make(map[string]string)
	input = strings.Trim(input, "{}\n")
	if input == "" {
		return result
	}

	// Simple parser for OVS map format: {key=value, key2=value2}
	pairs := strings.Split(input, ", ")
	for _, pair := range pairs {
		parts := strings.SplitN(pair, "=", 2)
		if len(parts) == 2 {
			key := strings.Trim(parts[0], "\"")
			value := strings.Trim(parts[1], "\"")
			result[key] = value
		}
	}
	return result
}

func (m *OVSManager) parseOVSIntArray(input string) []int {
	result := []int{}
	input = strings.Trim(input, emptyListString)
	if input == "" {
		return result
	}

	parts := strings.Split(input, ",")
	for _, part := range parts {
		part = strings.TrimSpace(part)
		var val int
		if _, err := fmt.Sscanf(part, "%d", &val); err == nil {
			result = append(result, val)
		}
	}
	return result
}
