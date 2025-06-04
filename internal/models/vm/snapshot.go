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
	// Name is the snapshot name.
	Name string `json:"name"`

	// Description is an optional description.
	Description string `json:"description,omitempty"`

	// State is the domain state at time of snapshot.
	State SnapshotState `json:"state"`

	// Parent is the parent snapshot name (if any).
	Parent string `json:"parent,omitempty"`

	// CreatedAt is when the snapshot was created.
	CreatedAt time.Time `json:"created_at"`

	// IsCurrent indicates if this is the current snapshot.
	IsCurrent bool `json:"is_current"`

	// HasMetadata indicates if the snapshot has metadata.
	HasMetadata bool `json:"has_metadata"`

	// HasMemory indicates if memory state is included.
	HasMemory bool `json:"has_memory"`

	// HasDisk indicates if disk state is included.
	HasDisk bool `json:"has_disk"`
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
