package vm

import (
	"fmt"
	"time"
)

// VMStatus represents the status of a VM.
type VMStatus string

// Status constants.
const (
	VMStatusRunning  VMStatus = "running"
	VMStatusStopped  VMStatus = "stopped"
	VMStatusPaused   VMStatus = "paused"
	VMStatusShutdown VMStatus = "shutdown"
	VMStatusCrashed  VMStatus = "crashed"
	VMStatusUnknown  VMStatus = "unknown"
)

// Valid values for VM status.
var validStatuses = map[VMStatus]bool{
	VMStatusRunning:  true,
	VMStatusStopped:  true,
	VMStatusPaused:   true,
	VMStatusShutdown: true,
	VMStatusCrashed:  true,
	VMStatusUnknown:  true,
}

// IsValid checks if the status is valid.
func (s VMStatus) IsValid() bool {
	return validStatuses[s]
}

// String returns the string representation of the status.
func (s VMStatus) String() string {
	return string(s)
}

// IsActive returns true if the VM is in an active state.
func (s VMStatus) IsActive() bool {
	return s == VMStatusRunning || s == VMStatusPaused
}

// StatusInfo contains detailed status information.
type StatusInfo struct {
	LastStateChange time.Time    `json:"lastStateChange,omitempty"`
	Status          VMStatus     `json:"status"`
	NetworkUsage    NetworkUsage `json:"networkUsage,omitempty"`
	DiskUsage       DiskUsage    `json:"diskUsage,omitempty"`
	Uptime          int64        `json:"uptime,omitempty"`
	CPUUtilization  float64      `json:"cpuUtilization,omitempty"`
	MemoryUsage     uint64       `json:"memoryUsage,omitempty"`
}

// NetworkUsage contains network utilization information.
type NetworkUsage struct {
	RxBytes   uint64 `json:"rxBytes,omitempty"`   // Received bytes
	TxBytes   uint64 `json:"txBytes,omitempty"`   // Transmitted bytes
	RxPackets uint64 `json:"rxPackets,omitempty"` // Received packets
	TxPackets uint64 `json:"txPackets,omitempty"` // Transmitted packets
	RxDropped uint64 `json:"rxDropped,omitempty"` // Received packets dropped
	TxDropped uint64 `json:"txDropped,omitempty"` // Transmitted packets dropped
	RxErrors  uint64 `json:"rxErrors,omitempty"`  // Received errors
	TxErrors  uint64 `json:"txErrors,omitempty"`  // Transmitted errors
}

// DiskUsage contains disk utilization information.
type DiskUsage struct {
	ReadBytes  uint64 `json:"readBytes,omitempty"`  // Read bytes
	WriteBytes uint64 `json:"writeBytes,omitempty"` // Written bytes
	ReadOps    uint64 `json:"readOps,omitempty"`    // Read operations
	WriteOps   uint64 `json:"writeOps,omitempty"`   // Write operations
}

// StatusTransition represents a status transition.
type StatusTransition struct {
	From      VMStatus  `json:"from"`
	To        VMStatus  `json:"to"`
	Timestamp time.Time `json:"timestamp"`
	Initiator string    `json:"initiator,omitempty"` // Who/what initiated the change
	Reason    string    `json:"reason,omitempty"`    // Reason for the transition
}

// String returns a string representation of the status transition.
func (t StatusTransition) String() string {
	return fmt.Sprintf("%s → %s at %s", t.From, t.To, t.Timestamp.Format(time.RFC3339))
}
