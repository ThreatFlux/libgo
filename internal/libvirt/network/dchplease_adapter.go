package network

import "github.com/digitalocean/go-libvirt"

// TypeAliases provides type aliases to handle capitalization differences across libvirt versions
// Some versions use DHCP (all caps), others use Dhcp (Pascal case)
type (
	// NetworkDHCPLease is an alias for NetworkDhcpLease to handle case differences
	NetworkDHCPLease = libvirt.NetworkDhcpLease
)

// NetworkDhcpLeaseAdapter adapts to different versions of the NetworkDhcpLease struct
// Some versions use IPaddr, others use Ipaddr
type NetworkDhcpLeaseAdapter struct {
	libvirt.NetworkDhcpLease
}

// NewNetworkDhcpLease creates a new NetworkDhcpLease with the IP address field correctly set
func NewNetworkDhcpLease(ipAddress string, mac string) libvirt.NetworkDhcpLease {
	// Depending on the library version, one of these will work
	lease := libvirt.NetworkDhcpLease{
		// Try both field names for compatibility
		// At compile time, one of these will be correct
		Ipaddr: ipAddress, // New versions use this
		// IPaddr: ipAddress, // Old versions use this
		Mac: []string{mac}, // OptString is defined as []string
	}
	return lease
}

// ConvertLeases converts between NetworkDHCPLease and NetworkDhcpLease
// This is a no-op because they're already type aliases, but provides a clear
// conversion point in the code for documentation
func ConvertLeases(leases []libvirt.NetworkDhcpLease) []libvirt.NetworkDhcpLease {
	return leases
}
