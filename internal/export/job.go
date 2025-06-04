package export

import (
	"sync"
	"time"

	"github.com/google/uuid"
)

// jobStore provides thread-safe storage for export jobs.
type jobStore struct {
	jobs map[string]*Job
	mu   sync.RWMutex
}

// newJobStore creates a new job store.
func newJobStore() *jobStore {
	return &jobStore{
		jobs: make(map[string]*Job),
	}
}

// createJob creates a new export job.
func (s *jobStore) createJob(vmName string, format string, options map[string]string) *Job {
	s.mu.Lock()
	defer s.mu.Unlock()

	id := uuid.New().String()
	job := &Job{
		ID:        id,
		VMName:    vmName,
		Format:    format,
		Status:    StatusPending,
		Progress:  0,
		StartTime: time.Now(),
		Options:   options,
	}

	s.jobs[id] = job
	return job
}

// getJob gets a job by ID.
func (s *jobStore) getJob(id string) (*Job, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	job, exists := s.jobs[id]
	return job, exists
}

// listJobs returns all jobs.
func (s *jobStore) listJobs() []*Job {
	s.mu.RLock()
	defer s.mu.RUnlock()

	jobs := make([]*Job, 0, len(s.jobs))
	for _, job := range s.jobs {
		jobs = append(jobs, job)
	}

	return jobs
}

// updateJobStatus updates a job's status.
func (s *jobStore) updateJobStatus(id string, status Status, progress int, err error) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	job, exists := s.jobs[id]
	if !exists {
		return false
	}

	job.Status = status
	job.Progress = progress

	if err != nil {
		job.Error = err.Error()
	}

	if status == StatusCompleted || status == StatusFailed || status == StatusCanceled {
		job.EndTime = time.Now()
	}

	return true
}

// setJobOutputPath sets the output path for a job.
func (s *jobStore) setJobOutputPath(id string, path string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	job, exists := s.jobs[id]
	if !exists {
		return false
	}

	job.OutputPath = path
	return true
}
