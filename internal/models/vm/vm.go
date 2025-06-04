package vm

import (
	"time"
)

// VM represents a virtual machine.
type VM struct {
	// Group strings together (8 bytes each on 64-bit)
	Name        string `json:"name"`
	UUID        string `json:"uuid"`
	Description string `json:"description,omitempty"`
	// Group slices together (8 bytes each)
	Disks    []DiskInfo `json:"disks"`
	Networks []NetInfo  `json:"networks"`
	// Group time.Time (8 bytes)
	CreatedAt time.Time `json:"createdAt"`
	// Group structs together
	Status VMStatus   `json:"status"`
	CPU    CPUInfo    `json:"cpu"`
	Memory MemoryInfo `json:"memory"`
}

// Using VMStatus from status.go, not redeclaring here

// CPUInfo contains CPU information.
type CPUInfo struct {
	Model   string `json:"model,omitempty"`
	Count   int    `json:"count"`
	Sockets int    `json:"sockets,omitempty"`
	Cores   int    `json:"cores,omitempty"`
	Threads int    `json:"threads,omitempty"`
}

// MemoryInfo contains memory information.
type MemoryInfo struct {
	SizeBytes uint64 `json:"sizeBytes"`
	SizeMB    uint64 `json:"sizeMB"`
}

// Using DiskInfo from disk.go
// Using NetInfo from network.go
