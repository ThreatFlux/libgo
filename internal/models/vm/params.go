package vm

// VMParams contains parameters for VM creation
type VMParams struct {
	Name        string          `json:"name" validate:"required,hostname_rfc1123"`
	Description string          `json:"description,omitempty"`
	CPU         CPUParams       `json:"cpu" validate:"required"`
	Memory      MemoryParams    `json:"memory" validate:"required"`
	Disk        DiskParams      `json:"disk" validate:"required"`
	Network     NetParams       `json:"network"`
	CloudInit   CloudInitConfig `json:"cloudInit,omitempty"`
	Template    string          `json:"template,omitempty"` // Name of template to use
}

// CPUParams contains CPU parameters
type CPUParams struct {
	Count   int    `json:"count" validate:"required,min=1,max=128"`
	Model   string `json:"model,omitempty"`
	Socket  int    `json:"socket,omitempty" validate:"omitempty,min=1"`
	Cores   int    `json:"cores,omitempty" validate:"omitempty,min=1"`
	Threads int    `json:"threads,omitempty" validate:"omitempty,min=1"`
}

// MemoryParams contains memory parameters
type MemoryParams struct {
	SizeBytes uint64 `json:"sizeBytes" validate:"required,min=134217728"` // Minimum 128MB
	SizeMB    uint64 `json:"sizeMB,omitempty"`                            // Size in MB (optional, calculated from SizeBytes if not provided)
}

// Using DiskParams from disk.go

// Using NetParams from network.go

// CloudInitConfig contains cloud-init configuration
type CloudInitConfig struct {
	UserData      string   `json:"userData,omitempty"`
	MetaData      string   `json:"metaData,omitempty"`
	NetworkConfig string   `json:"networkConfig,omitempty"`
	SSHKeys       []string `json:"sshKeys,omitempty"`
	ISODir        string   `json:"-"` // Internal use only - not exposed via API
}
