package cloudinit

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"github.com/threatflux/libgo/internal/models/vm"
	"github.com/threatflux/libgo/pkg/logger"
)

// MetadataGenerator handles generation of cloud-init metadata.
type MetadataGenerator struct {
	logger logger.Logger
}

// NewMetadataGenerator creates a new MetadataGenerator.
func NewMetadataGenerator(logger logger.Logger) *MetadataGenerator {
	return &MetadataGenerator{
		logger: logger,
	}
}

// GenerateInstanceID generates a unique instance ID for cloud-init.
func (g *MetadataGenerator) GenerateInstanceID(vm *vm.VM) string {
	// Use VM UUID if available, otherwise use name
	if vm.UUID != "" {
		return vm.UUID
	}
	return fmt.Sprintf("iid-%s", vm.Name)
}

// GenerateHostname generates a hostname from VM name.
func (g *MetadataGenerator) GenerateHostname(vm *vm.VM) string {
	// Convert VM name to a valid hostname
	// Replace any character that's not alphanumeric or hyphen with hyphen
	reg := regexp.MustCompile(`[^a-zA-Z0-9-]`)
	hostname := reg.ReplaceAllString(vm.Name, "-")

	// Ensure hostname starts with a letter
	if !regexp.MustCompile(`^[a-zA-Z]`).MatchString(hostname) {
		hostname = "vm-" + hostname
	}

	// Ensure hostname ends with alphanumeric
	if !regexp.MustCompile(`[a-zA-Z0-9]$`).MatchString(hostname) {
		hostname += "0"
	}

	// Maximum length for a hostname
	if len(hostname) > 63 {
		hostname = hostname[:63]
	}

	// Convert to lowercase
	hostname = strings.ToLower(hostname)

	return hostname
}

// GenerateNetworkConfig generates network configuration for cloud-init.
func (g *MetadataGenerator) GenerateNetworkConfig(params vm.VMParams) (string, error) {
	// Create a simplified representation for cloud-init network config
	networkConfig := map[string]interface{}{
		"version": 2,
		"ethernets": map[string]interface{}{
			"ens3": map[string]interface{}{
				"dhcp4": true,
				"dhcp6": false,
			},
		},
	}

	// If network parameters have custom MAC, include that
	if params.Network.MacAddress != "" {
		if ethernets, ok := networkConfig["ethernets"].(map[string]interface{}); ok {
			if ens3, ok := ethernets["ens3"].(map[string]interface{}); ok {
				ens3["match"] = map[string]interface{}{
					"macaddress": params.Network.MacAddress,
				}
			}
		}
	}

	// Convert to YAML format
	networkConfigJSON, err := json.MarshalIndent(networkConfig, "", "  ")
	if err != nil {
		return "", fmt.Errorf("error marshaling network config: %w", err)
	}

	return string(networkConfigJSON), nil
}

// MetadataToJSON converts metadata key-value pairs to JSON format.
func (g *MetadataGenerator) MetadataToJSON(metadata map[string]string) (string, error) {
	// Convert to JSON
	metadataJSON, err := json.MarshalIndent(metadata, "", "  ")
	if err != nil {
		return "", fmt.Errorf("error marshaling metadata: %w", err)
	}

	return string(metadataJSON), nil
}

// ParseUserDataScript parses user data script and extracts key information.
func (g *MetadataGenerator) ParseUserDataScript(userData string) map[string]interface{} {
	result := make(map[string]interface{})

	// Look for hostname
	hostnameRE := regexp.MustCompile(`(?m)^hostname:\s*(.+)$`)
	if match := hostnameRE.FindStringSubmatch(userData); len(match) > 1 {
		result["hostname"] = strings.TrimSpace(match[1])
	}

	// Look for users
	userRE := regexp.MustCompile(`(?m)^users:`)
	if userRE.MatchString(userData) {
		result["has_users"] = true
	}

	// Look for packages
	packageRE := regexp.MustCompile(`(?m)^packages:`)
	if packageRE.MatchString(userData) {
		result["has_packages"] = true
	}

	// Look for SSH keys
	sshKeyRE := regexp.MustCompile(`(?m)^\s*ssh_authorized_keys:`)
	if sshKeyRE.MatchString(userData) {
		result["has_ssh_keys"] = true
	}

	return result
}
