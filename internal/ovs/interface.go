package ovs

import (
	"context"
)

// Manager defines interface for managing Open vSwitch operations.
type Manager interface {
	// Bridge operations
	CreateBridge(ctx context.Context, name string) error
	DeleteBridge(ctx context.Context, name string) error
	ListBridges(ctx context.Context) ([]BridgeInfo, error)
	GetBridge(ctx context.Context, name string) (*BridgeInfo, error)

	// Port operations
	AddPort(ctx context.Context, bridge string, port string, options *PortOptions) error
	DeletePort(ctx context.Context, bridge string, port string) error
	ListPorts(ctx context.Context, bridge string) ([]PortInfo, error)
	GetPort(ctx context.Context, bridge string, port string) (*PortInfo, error)

	// VLAN operations
	SetPortVLAN(ctx context.Context, bridge string, port string, vlan int) error
	SetPortTrunk(ctx context.Context, bridge string, port string, vlans []int) error

	// Flow operations
	AddFlow(ctx context.Context, bridge string, flow *FlowRule) error
	DeleteFlow(ctx context.Context, bridge string, flowID string) error
	ListFlows(ctx context.Context, bridge string) ([]FlowRule, error)

	// Controller operations
	SetController(ctx context.Context, bridge string, controller string) error
	DeleteController(ctx context.Context, bridge string) error

	// Statistics
	GetBridgeStats(ctx context.Context, bridge string) (*BridgeStats, error)
	GetPortStats(ctx context.Context, bridge string, port string) (*PortStats, error)
}

// BridgeInfo represents information about an OVS bridge.
type BridgeInfo struct {
	ExternalIDs  map[string]string `json:"external_ids,omitempty"`
	OtherConfig  map[string]string `json:"other_config,omitempty"`
	Statistics   *BridgeStats      `json:"statistics,omitempty"`
	Name         string            `json:"name"`
	UUID         string            `json:"uuid"`
	Controller   string            `json:"controller,omitempty"`
	DatapathType string            `json:"datapath_type"`
	Ports        []string          `json:"ports"`
}

// PortInfo represents information about an OVS port.
type PortInfo struct {
	ExternalIDs map[string]string `json:"external_ids,omitempty"`
	OtherConfig map[string]string `json:"other_config,omitempty"`
	Statistics  *PortStats        `json:"statistics,omitempty"`
	Tag         *int              `json:"tag,omitempty"`
	Name        string            `json:"name"`
	UUID        string            `json:"uuid"`
	Bridge      string            `json:"bridge"`
	Type        string            `json:"type"`
	Interfaces  []string          `json:"interfaces"`
	Trunks      []int             `json:"trunks,omitempty"`
}

// PortOptions represents options for creating a port.
type PortOptions struct {
	ExternalIDs map[string]string `json:"external_ids,omitempty"`
	OtherConfig map[string]string `json:"other_config,omitempty"`
	Tag         *int              `json:"tag,omitempty"`
	Type        string            `json:"type,omitempty"`
	PeerPort    string            `json:"peer_port,omitempty"`
	RemoteIP    string            `json:"remote_ip,omitempty"`
	TunnelType  string            `json:"tunnel_type,omitempty"`
	Trunks      []int             `json:"trunks,omitempty"`
}

// FlowRule represents an OpenFlow rule.
type FlowRule struct {
	ID       string `json:"id"`
	Match    string `json:"match"`
	Actions  string `json:"actions"`
	Cookie   string `json:"cookie,omitempty"`
	Table    int    `json:"table"`
	Priority int    `json:"priority"`
}

// BridgeStats represents statistics for a bridge.
type BridgeStats struct {
	FlowCount    int64 `json:"flow_count"`
	PortCount    int64 `json:"port_count"`
	LookupCount  int64 `json:"lookup_count"`
	MatchedCount int64 `json:"matched_count"`
}

// PortStats represents statistics for a port.
type PortStats struct {
	RxPackets uint64 `json:"rx_packets"`
	RxBytes   uint64 `json:"rx_bytes"`
	RxDropped uint64 `json:"rx_dropped"`
	RxErrors  uint64 `json:"rx_errors"`
	TxPackets uint64 `json:"tx_packets"`
	TxBytes   uint64 `json:"tx_bytes"`
	TxDropped uint64 `json:"tx_dropped"`
	TxErrors  uint64 `json:"tx_errors"`
}

// CreateBridgeParams represents parameters for creating an OVS bridge.
type CreateBridgeParams struct {
	ExternalIDs  map[string]string `json:"external_ids,omitempty"`
	OtherConfig  map[string]string `json:"other_config,omitempty"`
	Name         string            `json:"name" binding:"required"`
	DatapathType string            `json:"datapath_type,omitempty"` // system or netdev
	Controller   string            `json:"controller,omitempty"`
}

// CreatePortParams represents parameters for creating an OVS port.
type CreatePortParams struct {
	ExternalIDs map[string]string `json:"external_ids,omitempty"`
	OtherConfig map[string]string `json:"other_config,omitempty"`
	Tag         *int              `json:"tag,omitempty"`
	Name        string            `json:"name" binding:"required"`
	Bridge      string            `json:"bridge" binding:"required"`
	Type        string            `json:"type,omitempty"`
	PeerPort    string            `json:"peer_port,omitempty"`
	RemoteIP    string            `json:"remote_ip,omitempty"`
	TunnelType  string            `json:"tunnel_type,omitempty"`
	Trunks      []int             `json:"trunks,omitempty"`
}

// UpdatePortParams represents parameters for updating an OVS port.
type UpdatePortParams struct {
	ExternalIDs map[string]string `json:"external_ids,omitempty"`
	OtherConfig map[string]string `json:"other_config,omitempty"`
	Tag         *int              `json:"tag,omitempty"`
	Trunks      []int             `json:"trunks,omitempty"`
}

// CreateFlowParams represents parameters for creating a flow rule.
type CreateFlowParams struct {
	Bridge   string `json:"bridge" binding:"required"`
	Match    string `json:"match" binding:"required"`
	Actions  string `json:"actions" binding:"required"`
	Cookie   string `json:"cookie,omitempty"`
	Table    int    `json:"table"`
	Priority int    `json:"priority"`
}
