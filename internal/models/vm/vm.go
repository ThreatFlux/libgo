package vm

import (
	"time"
)

// VM represents a virtual machine
type VM struct {
	Name        string     `json:"name"`
	UUID        string     `json:"uuid"`
	Status      VMStatus   `json:"status"`
	CPU         CPUInfo    `json:"cpu"`
	Memory      MemoryInfo `json:"memory"`
	Disks       []DiskInfo `json:"disks"`
	Networks    []NetInfo  `json:"networks"`
	CreatedAt   time.Time  `json:"createdAt"`
	Description string     `json:"description,omitempty"`
}

// Using VMStatus from status.go, not redeclaring here

// CPUInfo contains CPU information
type CPUInfo struct {
	Count   int    `json:"count"`
	Model   string `json:"model,omitempty"`
	Sockets int    `json:"sockets,omitempty"`
	Cores   int    `json:"cores,omitempty"`
	Threads int    `json:"threads,omitempty"`
}

// MemoryInfo contains memory information
type MemoryInfo struct {
	SizeBytes uint64 `json:"sizeBytes"`
	SizeMB    uint64 `json:"sizeMB"`
}

// Using DiskInfo from disk.go
// Using NetInfo from network.go
