package network

import (
	"fmt"
	"net"
	"strings"
	"text/template"

	"github.com/wroersma/libgo/pkg/logger"
)

// TemplateXMLBuilder implements XMLBuilder using templates
type TemplateXMLBuilder struct {
	logger logger.Logger
}

// NewTemplateXMLBuilder creates a new TemplateXMLBuilder
func NewTemplateXMLBuilder(logger logger.Logger) *TemplateXMLBuilder {
	return &TemplateXMLBuilder{
		logger: logger,
	}
}

// networkXMLTemplate is the template for libvirt network XML
const networkXMLTemplate = `<network>
  <name>{{.Name}}</name>
  <bridge name="{{.Bridge}}"/>
  <forward mode="nat"/>
  <ip address="{{.Address}}" netmask="{{.Netmask}}">
    {{if .DHCP}}
    <dhcp>
      <range start="{{.RangeStart}}" end="{{.RangeEnd}}"/>
    </dhcp>
    {{end}}
  </ip>
</network>`

// networkTemplateData holds data for the network XML template
type networkTemplateData struct {
	Name       string
	Bridge     string
	Address    string
	Netmask    string
	DHCP       bool
	RangeStart string
	RangeEnd   string
}

// BuildNetworkXML implements XMLBuilder.BuildNetworkXML
func (b *TemplateXMLBuilder) BuildNetworkXML(name string, bridgeName string, cidr string, dhcp bool) (string, error) {
	// Parse CIDR
	ip, ipNet, err := net.ParseCIDR(cidr)
	if err != nil {
		return "", fmt.Errorf("invalid CIDR format %s: %w", cidr, err)
	}

	// Calculate network address
	networkIP := ip.Mask(ipNet.Mask)

	// Calculate netmask
	netmaskStr := strings.Join(strings.Split(fmt.Sprintf("%d.%d.%d.%d", ipNet.Mask[0], ipNet.Mask[1], ipNet.Mask[2], ipNet.Mask[3]), "."), ".")

	data := networkTemplateData{
		Name:    name,
		Bridge:  bridgeName,
		Address: networkIP.String(),
		Netmask: netmaskStr,
		DHCP:    dhcp,
	}

	// If DHCP is enabled, calculate range
	if dhcp {
		// For DHCP range, use from .100 to .200 in the subnet
		// This is a simple approach - for production, might want more flexibility
		rangeStart, rangeEnd := calculateDHCPRange(networkIP, ipNet.Mask)
		data.RangeStart = rangeStart
		data.RangeEnd = rangeEnd
	}

	// Parse the template
	tmpl, err := template.New("network").Parse(networkXMLTemplate)
	if err != nil {
		return "", fmt.Errorf("parsing network template: %w", err)
	}

	// Execute the template
	var xmlBuffer strings.Builder
	if err := tmpl.Execute(&xmlBuffer, data); err != nil {
		return "", fmt.Errorf("executing network template: %w", err)
	}

	return xmlBuffer.String(), nil
}

// calculateDHCPRange calculates DHCP range based on network address
func calculateDHCPRange(networkIP net.IP, mask net.IPMask) (string, string) {
	// Clone the IP to avoid modifying the original
	rangeStartIP := make(net.IP, len(networkIP))
	copy(rangeStartIP, networkIP)
	rangeEndIP := make(net.IP, len(networkIP))
	copy(rangeEndIP, networkIP)

	// Set the last octet to 100 for start and 200 for end
	// This assumes a /24 or wider subnet
	rangeStartIP[3] = 100
	rangeEndIP[3] = 200

	return rangeStartIP.String(), rangeEndIP.String()
}