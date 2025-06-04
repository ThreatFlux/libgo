package configtest

import (
	"fmt"
	"os"
	"time"

	vmmodels "github.com/threatflux/libgo/internal/models/vm"
	"gopkg.in/yaml.v3"
)

// TestConfig represents the structure of a test configuration file
type TestConfig struct {
	Test struct {
		Name        string `yaml:"name"`
		Description string `yaml:"description"`
		Timeout     string `yaml:"timeout"` // Duration string e.g. "60m"
	} `yaml:"test"`

	VM struct {
		Name        string                `yaml:"name"`
		Template    string                `yaml:"template"`
		Description string                `yaml:"description"`
		CPU         vmmodels.CPUParams    `yaml:"cpu,omitempty"`
		Memory      vmmodels.MemoryParams `yaml:"memory,omitempty"`
		Disk        vmmodels.DiskParams   `yaml:"disk,omitempty"`
		Network     vmmodels.NetParams    `yaml:"network,omitempty"`

		Provisioning struct {
			Method        string `yaml:"method"` // cloudinit, unattended, preconfigured
			UnattendedXml string `yaml:"unattendedXml,omitempty"`
			Scripts       []struct {
				Name    string `yaml:"name"`
				Content string `yaml:"content"`
			} `yaml:"scripts,omitempty"`
		} `yaml:"provisioning"`
	} `yaml:"vm"`

	Verification struct {
		Services []struct {
			Name            string `yaml:"name"`
			Protocol        string `yaml:"protocol"`
			ExpectedContent string `yaml:"expectedContent"`
			Port            int    `yaml:"port"`
			Timeout         int    `yaml:"timeout"`
		} `yaml:"services"`
	} `yaml:"verification"`

	Export struct {
		Options map[string]string `yaml:"options"`
		Format  string            `yaml:"format"`
	} `yaml:"export"`
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
	// Create a completely new set of parameters rather than depending on the template
	// This is more reliable than template+overrides for testing
	params := vmmodels.VMParams{
		Name:        c.VM.Name,
		Description: c.VM.Description,
		Template:    c.VM.Template,

		// Explicitly set all required parameters
		CPU: vmmodels.CPUParams{
			Count: c.VM.CPU.Count,
			Model: "host-model", // Default fallback
		},
		Memory: vmmodels.MemoryParams{
			SizeBytes: 2 * 1024 * 1024 * 1024, // Explicitly use 2GB
			SizeMB:    2 * 1024,               // 2GB
		},
		Disk: vmmodels.DiskParams{
			SizeBytes:   10 * 1024 * 1024 * 1024, // 10GB
			SizeMB:      10 * 1024,               // 10GB
			Format:      "qcow2",
			StoragePool: "default",
			Bus:         "virtio",
		},
		Network: vmmodels.NetParams{
			Type:   "network",
			Source: "default",
			Model:  "virtio",
		},
	}

	// Override with any user-specified values
	if c.VM.CPU.Count > 0 {
		params.CPU.Count = c.VM.CPU.Count
	}

	if c.VM.CPU.Model != "" {
		params.CPU.Model = c.VM.CPU.Model
	}

	if c.VM.Memory.SizeBytes > 0 {
		params.Memory.SizeBytes = c.VM.Memory.SizeBytes
	}

	if c.VM.Disk.SizeBytes > 0 {
		params.Disk.SizeBytes = c.VM.Disk.SizeBytes
	}

	if c.VM.Disk.Format != "" {
		params.Disk.Format = c.VM.Disk.Format
	}

	if c.VM.Disk.StoragePool != "" {
		params.Disk.StoragePool = c.VM.Disk.StoragePool
	}

	if c.VM.Disk.Bus != "" {
		params.Disk.Bus = c.VM.Disk.Bus
	}

	if c.VM.Network.Type != "" {
		params.Network.Type = c.VM.Network.Type
	}

	if c.VM.Network.Source != "" {
		params.Network.Source = c.VM.Network.Source
	}

	if c.VM.Network.Model != "" {
		params.Network.Model = c.VM.Network.Model
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
