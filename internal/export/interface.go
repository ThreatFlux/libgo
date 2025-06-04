package export

import (
	"context"
	"time"
)

// Params represents export parameters.
type Params struct {
	Format   string            `json:"format" binding:"required,oneof=qcow2 vmdk vdi ova raw"`
	Options  map[string]string `json:"options,omitempty"`
	FileName string            `json:"fileName,omitempty"`
}

// Status represents export job status.
type Status string

// Job status constants.
const (
	StatusPending   Status = "pending"
	StatusRunning   Status = "running"
	StatusCompleted Status = "completed"
	StatusFailed    Status = "failed"
	StatusCanceled  Status = "canceled"
)

// Job represents an export job.
type Job struct {
	Options    map[string]string `json:"options,omitempty"`
	StartTime  time.Time         `json:"startTime"`
	EndTime    time.Time         `json:"endTime,omitempty"`
	ID         string            `json:"id"`
	VMName     string            `json:"vmName"`
	Format     string            `json:"format"`
	Error      string            `json:"error,omitempty"`
	OutputPath string            `json:"outputPath,omitempty"`
	Status     Status            `json:"status"`
	Progress   int               `json:"progress"`
}

// Manager defines interface for export management.
type Manager interface {
	// CreateExportJob creates a new export job
	CreateExportJob(ctx context.Context, vmName string, params Params) (*Job, error)

	// GetJob gets an export job by ID
	GetJob(ctx context.Context, jobID string) (*Job, error)

	// CancelJob cancels an export job
	CancelJob(ctx context.Context, jobID string) error

	// ListJobs lists all export jobs
	ListJobs(ctx context.Context) ([]*Job, error)
}
