package network

import (
	"net"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	mocks_logger "github.com/threatflux/libgo/test/mocks/logger"
	"go.uber.org/mock/gomock"
)

func TestTemplateXMLBuilder_BuildNetworkXML(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLogger := mocks_logger.NewMockLogger(ctrl)
	builder := NewTemplateXMLBuilder(mockLogger)

	tests := []struct {
		name        string
		networkName string
		bridgeName  string
		cidr        string
		dhcp        bool
		wantErr     bool
	}{
		{
			name:        "Valid network with DHCP",
			networkName: "test-network",
			bridgeName:  "virbr0",
			cidr:        "192.168.100.0/24",
			dhcp:        true,
			wantErr:     false,
		},
		{
			name:        "Valid network without DHCP",
			networkName: "test-network",
			bridgeName:  "virbr0",
			cidr:        "192.168.100.0/24",
			dhcp:        false,
			wantErr:     false,
		},
		{
			name:        "Invalid CIDR",
			networkName: "test-network",
			bridgeName:  "virbr0",
			cidr:        "not-a-cidr",
			dhcp:        true,
			wantErr:     true,
		},
		{
			name:        "Different subnet",
			networkName: "test-network",
			bridgeName:  "virbr0",
			cidr:        "10.0.0.0/16",
			dhcp:        true,
			wantErr:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			xml, err := builder.BuildNetworkXML(tt.networkName, tt.bridgeName, tt.cidr, tt.dhcp)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.NotEmpty(t, xml)

			// Basic validation of XML content
			assert.Contains(t, xml, tt.networkName)
			assert.Contains(t, xml, tt.bridgeName)

			// Check DHCP section
			if tt.dhcp {
				assert.Contains(t, xml, "<dhcp>")
				assert.Contains(t, xml, "<range start=")
			} else {
				assert.NotContains(t, xml, "<dhcp>")
			}
		})
	}
}

func TestCalculateDHCPRange(t *testing.T) {
	tests := []struct {
		name      string
		cidr      string
		wantStart string
		wantEnd   string
	}{
		{
			name:      "Class C network",
			cidr:      "192.168.1.0/24",
			wantStart: "192.168.1.100",
			wantEnd:   "192.168.1.200",
		},
		{
			name:      "Class B network",
			cidr:      "172.16.0.0/16",
			wantStart: "172.16.0.100",
			wantEnd:   "172.16.0.200",
		},
		{
			name:      "Class A network",
			cidr:      "10.0.0.0/8",
			wantStart: "10.0.0.100",
			wantEnd:   "10.0.0.200",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Parse CIDR to get IP and mask
			ip, ipNet, err := net.ParseCIDR(tt.cidr)
			require.NoError(t, err)

			// Get network address
			networkIP := ip.Mask(ipNet.Mask)

			// Calculate DHCP range
			startIP, endIP := calculateDHCPRange(networkIP, ipNet.Mask)

			assert.Equal(t, tt.wantStart, startIP)
			assert.Equal(t, tt.wantEnd, endIP)
		})
	}
}
