package network

import (
	"context"

	"github.com/digitalocean/go-libvirt"
)

// Manager defines interface for managing libvirt networks
type Manager interface {
	// EnsureExists ensures a network exists
	EnsureExists(ctx context.Context, name string, bridgeName string, cidr string, dhcp bool) error

	// Delete deletes a network
	Delete(ctx context.Context, name string) error

	// Get gets a network
	Get(ctx context.Context, name string) (*libvirt.Network, error)

	// GetDHCPLeases gets the DHCP leases for a network
	GetDHCPLeases(ctx context.Context, name string) ([]libvirt.NetworkDhcpLease, error)

	// FindIPByMAC finds the IP address of a MAC address in the network
	FindIPByMAC(ctx context.Context, networkName string, mac string) (string, error)

	// List lists all networks
	List(ctx context.Context) ([]*NetworkInfo, error)

	// Create creates a new network
	Create(ctx context.Context, params *CreateNetworkParams) (*NetworkInfo, error)

	// Update updates an existing network
	Update(ctx context.Context, name string, params *UpdateNetworkParams) (*NetworkInfo, error)

	// GetInfo gets detailed information about a network
	GetInfo(ctx context.Context, name string) (*NetworkInfo, error)

	// GetXML gets the XML configuration of a network
	GetXML(ctx context.Context, name string) (string, error)

	// IsActive checks if a network is active
	IsActive(ctx context.Context, name string) (bool, error)

	// Start starts an inactive network
	Start(ctx context.Context, name string) error

	// Stop stops an active network
	Stop(ctx context.Context, name string) error

	// SetAutostart sets the autostart flag for a network
	SetAutostart(ctx context.Context, name string, autostart bool) error
}

// XMLBuilder defines interface for building network XML
type XMLBuilder interface {
	// BuildNetworkXML builds XML for network creation
	BuildNetworkXML(name string, bridgeName string, cidr string, dhcp bool) (string, error)
}

// NetworkInfo represents detailed information about a network
type NetworkInfo struct {
	UUID       string                 `json:"uuid"`
	Name       string                 `json:"name"`
	BridgeName string                 `json:"bridge_name"`
	Active     bool                   `json:"active"`
	Persistent bool                   `json:"persistent"`
	Autostart  bool                   `json:"autostart"`
	Forward    NetworkForward         `json:"forward"`
	IP         *NetworkIP             `json:"ip,omitempty"`
	DHCPLeases []NetworkDHCPLease     `json:"dhcp_leases,omitempty"`
	Metadata   map[string]interface{} `json:"metadata,omitempty"`
}

// NetworkForward represents network forward configuration
type NetworkForward struct {
	Mode string `json:"mode"` // nat, route, bridge, private, vepa, passthrough
	Dev  string `json:"dev,omitempty"`
}

// NetworkIP represents network IP configuration
type NetworkIP struct {
	Address string           `json:"address"`
	Netmask string           `json:"netmask"`
	DHCP    *NetworkDHCPInfo `json:"dhcp,omitempty"`
}

// NetworkDHCPInfo represents DHCP configuration
type NetworkDHCPInfo struct {
	Enabled bool                    `json:"enabled"`
	Start   string                  `json:"start,omitempty"`
	End     string                  `json:"end,omitempty"`
	Hosts   []NetworkDHCPStaticHost `json:"hosts,omitempty"`
}

// NetworkDHCPStaticHost represents a static DHCP host entry
type NetworkDHCPStaticHost struct {
	MAC  string `json:"mac"`
	Name string `json:"name,omitempty"`
	IP   string `json:"ip"`
}

// NetworkDHCPLease represents an active DHCP lease
type NetworkDHCPLease struct {
	IPAddress  string `json:"ip_address"`
	MACAddress string `json:"mac_address"`
	Hostname   string `json:"hostname,omitempty"`
	ClientID   string `json:"client_id,omitempty"`
	ExpiryTime int64  `json:"expiry_time"`
}

// CreateNetworkParams represents parameters for creating a network
type CreateNetworkParams struct {
	Name       string                 `json:"name" binding:"required"`
	BridgeName string                 `json:"bridge_name,omitempty"`
	Forward    *NetworkForward        `json:"forward,omitempty"`
	IP         *NetworkIP             `json:"ip,omitempty"`
	Autostart  bool                   `json:"autostart"`
	Metadata   map[string]interface{} `json:"metadata,omitempty"`
}

// UpdateNetworkParams represents parameters for updating a network
type UpdateNetworkParams struct {
	Forward   *NetworkForward        `json:"forward,omitempty"`
	IP        *NetworkIP             `json:"ip,omitempty"`
	Autostart *bool                  `json:"autostart,omitempty"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
}
