package config

import (
	"fmt"
	"os"
	"time"

	vmmodels "github.com/threatflux/libgo/internal/models/vm"
	yaml "gopkg.in/yaml.v3"
)

// TestConfig represents the structure of a test configuration file.
// Field alignment optimized for memory efficiency.
type TestConfig struct {
	// Largest struct fields first
	VM struct {
		// Slice fields (24 bytes) - largest first
		Provisioning struct {
			Scripts []struct {
				Name    string `yaml:"name"`    // 16 bytes each
				Content string `yaml:"content"` // 16 bytes each
			} `yaml:"scripts,omitempty"`
			Method        string `yaml:"method"`                  // 16 bytes
			UnattendedXml string `yaml:"unattendedXml,omitempty"` // 16 bytes
		} `yaml:"provisioning"`
		// Struct fields (group by size)
		Network vmmodels.NetParams    `yaml:"network,omitempty"`
		Disk    vmmodels.DiskParams   `yaml:"disk,omitempty"`
		CPU     vmmodels.CPUParams    `yaml:"cpu,omitempty"`
		Memory  vmmodels.MemoryParams `yaml:"memory,omitempty"`
		// String fields (16 bytes each) - group together
		Name        string `yaml:"name"`
		Template    string `yaml:"template"`
		Description string `yaml:"description"`
	} `yaml:"vm"`
	Verification struct {
		// Slice fields (24 bytes)
		Services []struct {
			// String fields (16 bytes each) - group together
			Name            string `yaml:"name"`
			Protocol        string `yaml:"protocol"`
			ExpectedContent string `yaml:"expectedContent"`
			// Int fields (8 bytes on 64-bit) - group together
			Port    int `yaml:"port"`
			Timeout int `yaml:"timeout"`
		} `yaml:"services"`
	} `yaml:"verification"`
	Export struct {
		// Map fields (24 bytes) - largest first
		Options map[string]string `yaml:"options"`
		// String fields (16 bytes)
		Format string `yaml:"format"`
	} `yaml:"export"`
	Test struct {
		// String fields (16 bytes each) - group together
		Name        string `yaml:"name"`
		Description string `yaml:"description"`
		Timeout     string `yaml:"timeout"`
	} `yaml:"test"`
}

// GetTimeout returns the test timeout duration with a default fallback
func (c *TestConfig) GetTimeout() time.Duration {
	timeout, err := time.ParseDuration(c.Test.Timeout)
	if err != nil {
		return 60 * time.Minute // Default timeout
	}
	return timeout
}

// CreateVMParams creates VM parameters from test config
func (c *TestConfig) CreateVMParams() vmmodels.VMParams {
	params := vmmodels.VMParams{
		Name:        c.VM.Name,
		Description: c.VM.Description,
		Template:    c.VM.Template,
	}

	// Only override template settings if specified in config with non-zero values
	if c.VM.CPU.Count > 0 {
		params.CPU = c.VM.CPU
	}

	if c.VM.Memory.SizeBytes > 0 {
		params.Memory = c.VM.Memory
	}

	if c.VM.Disk.SizeBytes > 0 {
		params.Disk = c.VM.Disk
	}

	if c.VM.Network.Type != "" {
		params.Network = c.VM.Network
	}

	// Handle provisioning method
	switch c.VM.Provisioning.Method {
	case "cloudinit":
		// For Linux VMs, extract cloud-init data from config
		userData := ""
		for _, script := range c.VM.Provisioning.Scripts {
			userData += script.Content + "\n"
		}
		params.CloudInit = vmmodels.CloudInitConfig{
			UserData: userData,
		}

	case "unattended":
		// For Windows VMs, we'll store unattended XML in CloudInit.UserData as a placeholder
		// This will need special handling in the VM manager
		params.CloudInit = vmmodels.CloudInitConfig{
			UserData: c.VM.Provisioning.UnattendedXml,
		}
	}

	return params
}

// LoadTestConfig loads a test configuration from a YAML file
func LoadTestConfig(filename string) (*TestConfig, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("reading config file: %w", err)
	}

	var config TestConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("parsing config YAML: %w", err)
	}

	return &config, nil
}
