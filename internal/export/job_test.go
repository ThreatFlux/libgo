package export

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestJobStore_CreateJob(t *testing.T) {
	store := newJobStore()

	// Test job creation
	job := store.createJob("test-vm", "qcow2", map[string]string{"key": "value"})

	// Verify job properties
	assert.NotEmpty(t, job.ID)
	assert.Equal(t, "test-vm", job.VMName)
	assert.Equal(t, "qcow2", job.Format)
	assert.Equal(t, StatusPending, job.Status)
	assert.Equal(t, 0, job.Progress)
	assert.NotZero(t, job.StartTime)
	assert.True(t, job.StartTime.Before(time.Now().Add(time.Second)))
	assert.True(t, job.StartTime.After(time.Now().Add(-time.Minute)))
	assert.Equal(t, "value", job.Options["key"])
	assert.Zero(t, job.EndTime)
	assert.Empty(t, job.Error)
	assert.Empty(t, job.OutputPath)

	// Verify job was added to store
	storedJob, exists := store.jobs[job.ID]
	assert.True(t, exists)
	assert.Equal(t, job, storedJob)
}

func TestJobStore_GetJob(t *testing.T) {
	store := newJobStore()

	// Create test job
	job := store.createJob("test-vm", "qcow2", nil)

	// Test get existing job
	retrievedJob, exists := store.getJob(job.ID)
	assert.True(t, exists)
	assert.Equal(t, job, retrievedJob)

	// Test get non-existent job
	retrievedJob, exists = store.getJob("non-existent-id")
	assert.False(t, exists)
	assert.Nil(t, retrievedJob)
}

func TestJobStore_ListJobs(t *testing.T) {
	store := newJobStore()

	// Empty store should return empty list
	jobs := store.listJobs()
	assert.Empty(t, jobs)

	// Create test jobs
	job1 := store.createJob("vm1", "qcow2", nil)
	job2 := store.createJob("vm2", "vmdk", nil)
	job3 := store.createJob("vm3", "vdi", nil)

	// Test listing jobs
	jobs = store.listJobs()
	require.Len(t, jobs, 3)

	// Verify all jobs are in the list (order not guaranteed)
	jobMap := make(map[string]*Job)
	for _, job := range jobs {
		jobMap[job.ID] = job
	}
	assert.Equal(t, job1, jobMap[job1.ID])
	assert.Equal(t, job2, jobMap[job2.ID])
	assert.Equal(t, job3, jobMap[job3.ID])
}

func TestJobStore_UpdateJobStatus(t *testing.T) {
	store := newJobStore()

	// Create test job
	job := store.createJob("test-vm", "qcow2", nil)

	testCases := []struct {
		err      error  // 16 bytes (interface header)
		name     string // 16 bytes (string header)
		progress int    // 8 bytes (int64 on 64-bit systems)
		status   Status // 8 bytes (Status is likely int type)
	}{
		{nil, "Update to running", 25, StatusRunning},
		{nil, "Update progress", 50, StatusRunning},
		{errors.New("test error"), "Update with error", 75, StatusRunning},
		{nil, "Update to completed", 100, StatusCompleted},
		{errors.New("failure error"), "Update to failed", 60, StatusFailed},
		{nil, "Update to canceled", 30, StatusCanceled},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Update job status
			success := store.updateJobStatus(job.ID, tc.status, tc.progress, tc.err)
			assert.True(t, success)

			// Verify job was updated
			updatedJob, exists := store.getJob(job.ID)
			require.True(t, exists)
			assert.Equal(t, tc.status, updatedJob.Status)
			assert.Equal(t, tc.progress, updatedJob.Progress)

			if tc.err != nil {
				assert.Equal(t, tc.err.Error(), updatedJob.Error)
			}

			// Verify end time is set for terminal states
			if tc.status == StatusCompleted || tc.status == StatusFailed || tc.status == StatusCanceled {
				assert.NotZero(t, updatedJob.EndTime)
				assert.True(t, updatedJob.EndTime.After(updatedJob.StartTime))
			}
		})
	}

	// Test updating non-existent job
	success := store.updateJobStatus("non-existent-id", StatusRunning, 50, nil)
	assert.False(t, success)
}

func TestJobStore_SetJobOutputPath(t *testing.T) {
	store := newJobStore()

	// Create test job
	job := store.createJob("test-vm", "qcow2", nil)

	// Test setting output path
	success := store.setJobOutputPath(job.ID, "/path/to/output")
	assert.True(t, success)

	// Verify job was updated
	updatedJob, exists := store.getJob(job.ID)
	require.True(t, exists)
	assert.Equal(t, "/path/to/output", updatedJob.OutputPath)

	// Test setting path for non-existent job
	success = store.setJobOutputPath("non-existent-id", "/path/to/output")
	assert.False(t, success)
}

func TestJobStore_Concurrency(t *testing.T) {
	store := newJobStore()

	// Create concurrent access scenario
	const numJobs = 100
	jobCh := make(chan *Job, numJobs)

	// Create jobs concurrently
	for i := 0; i < numJobs; i++ {
		go func(i int) {
			job := store.createJob("vm", "format", nil)
			jobCh <- job
		}(i)
	}

	// Collect all created jobs
	jobs := make([]*Job, 0, numJobs)
	for i := 0; i < numJobs; i++ {
		jobs = append(jobs, <-jobCh)
	}

	// Verify all jobs were created with unique IDs
	jobMap := make(map[string]bool)
	for _, job := range jobs {
		if jobMap[job.ID] {
			t.Fatalf("Duplicate job ID: %s", job.ID)
		}
		jobMap[job.ID] = true
	}

	// Verify we can retrieve all jobs
	listedJobs := store.listJobs()
	assert.Len(t, listedJobs, numJobs)
}
