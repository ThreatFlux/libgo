package vm

import (
	"time"
)

// SnapshotState represents the state of a snapshot.
type SnapshotState string

const (
	// SnapshotStateRunning - the domain is running.
	SnapshotStateRunning SnapshotState = "running"
	// SnapshotStateBlocked - the domain is blocked on resource.
	SnapshotStateBlocked SnapshotState = "blocked"
	// SnapshotStatePaused - the domain is paused by user.
	SnapshotStatePaused SnapshotState = "paused"
	// SnapshotStateShutdown - the domain is being shut down.
	SnapshotStateShutdown SnapshotState = "shutdown"
	// SnapshotStateShutoff - the domain is shut off.
	SnapshotStateShutoff SnapshotState = "shutoff"
	// SnapshotStateCrashed - the domain is crashed.
	SnapshotStateCrashed SnapshotState = "crashed"
)

// Snapshot represents a VM snapshot.
type Snapshot struct {
	CreatedAt   time.Time     `json:"created_at"`
	Name        string        `json:"name"`
	Description string        `json:"description,omitempty"`
	State       SnapshotState `json:"state"`
	Parent      string        `json:"parent,omitempty"`
	IsCurrent   bool          `json:"is_current"`
	HasMetadata bool          `json:"has_metadata"`
	HasMemory   bool          `json:"has_memory"`
	HasDisk     bool          `json:"has_disk"`
}

// SnapshotParams represents parameters for creating a snapshot.
type SnapshotParams struct {
	// Name is the snapshot name (required).
	Name string `json:"name" binding:"required"`

	// Description is an optional description.
	Description string `json:"description,omitempty"`

	// IncludeMemory determines if memory state should be saved.
	IncludeMemory bool `json:"include_memory"`

	// Quiesce attempts to quiesce guest filesystems (requires guest agent).
	Quiesce bool `json:"quiesce"`
}

// SnapshotListOptions represents options for listing snapshots.
type SnapshotListOptions struct {
	// IncludeMetadata includes full metadata for each snapshot.
	IncludeMetadata bool `json:"include_metadata"`

	// Tree returns snapshots in tree structure.
	Tree bool `json:"tree"`
}
