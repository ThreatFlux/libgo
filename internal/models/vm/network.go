package vm

import (
	"crypto/rand"
	"fmt"
	"net"
)

// NetworkType represents the type of VM network
type NetworkType string

// Network type constants
const (
	NetworkTypeBridge  NetworkType = "bridge"
	NetworkTypeNetwork NetworkType = "network"
	NetworkTypeDirect  NetworkType = "direct"
)

// Valid network types
var validNetworkTypes = map[NetworkType]bool{
	NetworkTypeBridge:  true,
	NetworkTypeNetwork: true,
	NetworkTypeDirect:  true,
}

// IsValid checks if the network type is valid
func (t NetworkType) IsValid() bool {
	_, valid := validNetworkTypes[t]
	return valid
}

// String returns the string representation of the network type
func (t NetworkType) String() string {
	return string(t)
}

// NetParams contains network parameters for VM creation
type NetParams struct {
	Type         NetworkType `json:"type" validate:"required,oneof=bridge network direct"`
	Source       string      `json:"source" validate:"required"`
	Model        string      `json:"model,omitempty" validate:"omitempty,oneof=virtio e1000 rtl8139"`
	MacAddress   string      `json:"macAddress,omitempty" validate:"omitempty,mac"`
}

// NetInfo contains information about a VM's network interface
type NetInfo struct {
	Type         NetworkType `json:"type"`
	Source       string      `json:"source"`
	Model        string      `json:"model"`
	MacAddress   string      `json:"macAddress"`
	IPAddress    string      `json:"ipAddress,omitempty"`
	IPAddressV6  string      `json:"ipAddressV6,omitempty"`
}

// Validate validates the network parameters
func (p *NetParams) Validate() error {
	// Check network type
	if !NetworkType(p.Type).IsValid() {
		return fmt.Errorf("invalid network type: %s", p.Type)
	}

	// Check MAC address format if provided
	if p.MacAddress != "" {
		_, err := net.ParseMAC(p.MacAddress)
		if err != nil {
			return fmt.Errorf("invalid MAC address format: %w", err)
		}
	}

	// Validate model
	if p.Model != "" && p.Model != "virtio" && p.Model != "e1000" && p.Model != "rtl8139" {
		return fmt.Errorf("invalid network model: %s", p.Model)
	}

	return nil
}

// GenerateRandomMAC generates a random MAC address within KVM's private range
func GenerateRandomMAC() (string, error) {
	mac := make(net.HardwareAddr, 6)

	// Use KVM's OUI (Organizationally Unique Identifier) for the first 3 bytes
	// 52:54:00 is the OUI for QEMU/KVM
	mac[0] = 0x52
	mac[1] = 0x54
	mac[2] = 0x00

	// Generate random values for the last 3 bytes
	for i := 3; i < 6; i++ {
		b, err := randomByte()
		if err != nil {
			return "", fmt.Errorf("generating random byte: %w", err)
		}
		mac[i] = b
	}

	return mac.String(), nil
}

// randomByte generates a random byte (0-255)
func randomByte() (byte, error) {
	// For simplicity using crypto/rand
	var b [1]byte
	_, err := rand.Read(b[:])
	if err != nil {
		return 0, err
	}
	return b[0], nil
}

// NetworkDefinition represents a libvirt network definition
type NetworkDefinition struct {
	Name      string `json:"name" validate:"required"`
	Bridge    string `json:"bridge" validate:"required"`
	CIDR      string `json:"cidr" validate:"required,cidr"`
	DHCPEnabled bool  `json:"dhcpEnabled"`
}

// ParseCIDR parses the CIDR notation and returns the network address and mask
func (d *NetworkDefinition) ParseCIDR() (net.IP, *net.IPNet, error) {
	return net.ParseCIDR(d.CIDR)
}
