package vm

// VMParams contains parameters for VM creation.
type VMParams struct {
	// String fields (16 bytes each)
	Name        string `json:"name" validate:"required,hostname_rfc1123"`
	Description string `json:"description,omitempty"`
	Template    string `json:"template,omitempty"` // Name of template to use
	// Structs (ordered by estimated size)
	CloudInit CloudInitConfig `json:"cloudInit,omitempty"`
	CPU       CPUParams       `json:"cpu" validate:"required"`
	Memory    MemoryParams    `json:"memory" validate:"required"`
	Disk      DiskParams      `json:"disk" validate:"required"`
	Network   NetParams       `json:"network"`
}

// CPUParams contains CPU parameters.
type CPUParams struct {
	Model   string `json:"model,omitempty"`
	Count   int    `json:"count" validate:"required,min=1,max=128"`
	Socket  int    `json:"socket,omitempty" validate:"omitempty,min=1"`
	Cores   int    `json:"cores,omitempty" validate:"omitempty,min=1"`
	Threads int    `json:"threads,omitempty" validate:"omitempty,min=1"`
}

// MemoryParams contains memory parameters.
type MemoryParams struct {
	SizeBytes uint64 `json:"sizeBytes" validate:"required,min=134217728"` // Minimum 128MB
	SizeMB    uint64 `json:"sizeMB,omitempty"`                            // Size in MB (optional, calculated from SizeBytes if not provided)
}

// Using DiskParams from disk.go.

// Using NetParams from network.go.

// CloudInitConfig contains cloud-init configuration.
type CloudInitConfig struct {
	// Slice fields (24 bytes)
	SSHKeys []string `json:"sshKeys,omitempty"`
	// String fields (16 bytes each)
	UserData      string `json:"userData,omitempty"`
	MetaData      string `json:"metaData,omitempty"`
	NetworkConfig string `json:"networkConfig,omitempty"`
	ISODir        string `json:"-"` // Internal use only - not exposed via API
}
