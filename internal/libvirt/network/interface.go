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
}

// XMLBuilder defines interface for building network XML
type XMLBuilder interface {
	// BuildNetworkXML builds XML for network creation
	BuildNetworkXML(name string, bridgeName string, cidr string, dhcp bool) (string, error)
}
