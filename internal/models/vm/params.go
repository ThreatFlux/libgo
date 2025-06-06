package vm

// VMParams contains parameters for VM creation.
type VMParams struct {
	// Largest struct first - DiskParams has ~10 fields (strings, uint64s, enums, bools)
	Disk DiskParams `json:"disk" validate:"required"`
	// CloudInit has slice + 4 strings
	CloudInit CloudInitConfig `json:"cloudInit,omitempty"`
	// CPU has 1 string + 4 ints
	CPU CPUParams `json:"cpu" validate:"required"`
	// Memory has 2 uint64s
	Memory MemoryParams `json:"memory" validate:"required"`
	// Network has 3 strings + 1 enum
	Network NetParams `json:"network"`
	// String fields (16 bytes each) - put at end for optimal alignment
	Name        string `json:"name" validate:"required,hostname_rfc1123"`
	Description string `json:"description,omitempty"`
	Template    string `json:"template,omitempty"` // Name of template to use
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
	// Slice fields (24 bytes) - largest first
	SSHKeys []string `json:"sshKeys,omitempty"`
	// String fields (16 bytes each)
	UserData      string `json:"userData,omitempty"`
	MetaData      string `json:"metaData,omitempty"`
	NetworkConfig string `json:"networkConfig,omitempty"`
	ISODir        string `json:"-"` // Internal use only - not exposed via API
}
