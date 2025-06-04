package config

import (
	"fmt"
	"os"
	"time"

	vmmodels "github.com/threatflux/libgo/internal/models/vm"
	yaml "gopkg.in/yaml.v3"
)

// TestConfig represents the structure of a test configuration file
type TestConfig struct {
	Test struct {
		Name        string `yaml:"name"`
		Description string `yaml:"description"`
		Timeout     string `yaml:"timeout"`
	} `yaml:"test"`
	Export struct {
		Options map[ // Duration string e.g. "60m"
		// Group map (8 bytes) before string (8 bytes) for better alignment
		string]string `yaml:"options"`
		Format string `yaml:"format"`
	} `yaml:"export"`
	VM struct {
		Network      vmmodels.NetParams `yaml:"network,omitempty"`
		Name         string             `yaml:"name"`
		Template     string             `yaml:"template"`
		Description  string             `yaml:"description"`
		Provisioning struct {
			Method        string `yaml:"method"`
			UnattendedXml string `yaml:"unattendedXml,omitempty"`
			Scripts       []struct {
				Name    string `yaml:"name"`
				Content string `yaml:"content"`
			} `yaml:"scripts,omitempty"`
		} `yaml:"provisioning"`
		Disk   vmmodels.DiskParams   `yaml:"disk,omitempty"`
		CPU    vmmodels.CPUParams    `yaml:"cpu,omitempty"`
		Memory vmmodels.MemoryParams `yaml:"memory,omitempty"`
	} `yaml:"vm"`
	Verification struct {
		Services []struct {
			Name            string `yaml:"name"`
			Protocol        string `yaml:"protocol"`
			ExpectedContent string `yaml:"expectedContent"`
			Port            int    `yaml:"port"`
			Timeout         int    `yaml:"timeout"`
		} `yaml:"services"`
	} `yaml:"verification"` // Group strings together
	// Group ints together (4 bytes each)
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
