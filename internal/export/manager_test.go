package export

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	customErrors "github.com/threatflux/libgo/internal/errors"
	"github.com/threatflux/libgo/internal/export/formats"
	"github.com/threatflux/libgo/internal/models/vm"
	mocks_domain "github.com/threatflux/libgo/test/mocks/libvirt/domain"
	mocks_storage "github.com/threatflux/libgo/test/mocks/libvirt/storage"
	mocks_logger "github.com/threatflux/libgo/test/mocks/logger"
)

// testConverter is a test implementation of formats.Converter
type testConverter struct {
	formatName  string
	convertErr  error
	validateErr error
}

func (tc *testConverter) Convert(ctx context.Context, sourcePath string, destPath string, options map[string]string) error {
	return tc.convertErr
}

func (tc *testConverter) GetFormatName() string {
	return tc.formatName
}

func (tc *testConverter) ValidateOptions(options map[string]string) error {
	return tc.validateErr
}

func TestExportManager_CreateExportJob(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping export manager test with background goroutines in short mode")
	}
	// Setup
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStorageManager := mocks_storage.NewMockVolumeManager(ctrl)
	mockDomainManager := mocks_domain.NewMockManager(ctrl)
	mockLogger := mocks_logger.NewMockLogger(ctrl)
	mockLogger.EXPECT().Info(gomock.Any(), gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Error(gomock.Any(), gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Warn(gomock.Any(), gomock.Any()).AnyTimes()

	// Create temporary directory for testing
	tmpDir, err := os.MkdirTemp("", "export-test-")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	// Create test converter
	testConverter := &testConverter{
		formatName:  "test-format",
		convertErr:  nil,
		validateErr: nil,
	}

	// Create manager with test converter
	manager := &ExportManager{
		jobStore: newJobStore(),
		formatManagers: map[string]formats.Converter{
			"test-format": testConverter,
		},
		storageManager: mockStorageManager,
		domainManager:  mockDomainManager,
		baseExportDir:  tmpDir,
		logger:         mockLogger,
	}

	// Create a test VM
	testVM := &vm.VM{
		Name:   "test-vm",
		UUID:   "test-uuid",
		Status: vm.VMStatusStopped,
		Disks: []vm.DiskInfo{
			{
				PoolName:   "default",
				VolumeName: "test-vm-disk",
				Path:       "/var/lib/libvirt/images/test-vm-disk.qcow2",
				SizeBytes:  1024 * 1024 * 1024,
				Format:     "qcow2",
			},
		},
	}

	// Test cases
	t.Run("VM does not exist", func(t *testing.T) {
		mockDomainManager.EXPECT().Get(gomock.Any(), "non-existent-vm").
			Return(nil, errors.New("VM not found"))

		job, err := manager.CreateExportJob(context.Background(), "non-existent-vm", Params{
			Format: "test-format",
		})

		assert.Error(t, err)
		assert.Nil(t, job)
		assert.Contains(t, err.Error(), "VM not found")
	})

	t.Run("Unsupported format", func(t *testing.T) {
		mockDomainManager.EXPECT().Get(gomock.Any(), "test-vm").
			Return(testVM, nil)

		job, err := manager.CreateExportJob(context.Background(), "test-vm", Params{
			Format: "unsupported-format",
		})

		assert.Error(t, err)
		assert.Nil(t, job)
		assert.True(t, errors.Is(err, customErrors.ErrUnsupportedFormat))
	})

	t.Run("Valid export job", func(t *testing.T) {
		mockDomainManager.EXPECT().Get(gomock.Any(), "test-vm").
			Return(testVM, nil).AnyTimes()

		job, err := manager.CreateExportJob(context.Background(), "test-vm", Params{
			Format:   "test-format",
			FileName: "test-export.test",
			Options: map[string]string{
				"option1": "value1",
			},
		})

		assert.NoError(t, err)
		assert.NotNil(t, job)
		assert.Equal(t, "test-vm", job.VMName)
		assert.Equal(t, "test-format", job.Format)
		assert.Equal(t, StatusPending, job.Status)
		assert.Equal(t, 0, job.Progress)
		assert.NotEmpty(t, job.ID)
		assert.Equal(t, "value1", job.Options["option1"])
	})
}

func TestExportManager_GetJob(t *testing.T) {
	// Setup
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStorageManager := mocks_storage.NewMockVolumeManager(ctrl)
	mockDomainManager := mocks_domain.NewMockManager(ctrl)
	mockLogger := mocks_logger.NewMockLogger(ctrl)

	// Create temporary directory for testing
	tmpDir, err := os.MkdirTemp("", "export-test-")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	// Create manager
	manager := &ExportManager{
		jobStore:       newJobStore(),
		formatManagers: map[string]formats.Converter{},
		storageManager: mockStorageManager,
		domainManager:  mockDomainManager,
		baseExportDir:  tmpDir,
		logger:         mockLogger,
	}

	// Create a test job
	job := manager.jobStore.createJob("test-vm", "qcow2", map[string]string{"key": "value"})

	t.Run("Get existing job", func(t *testing.T) {
		// Test get job
		retrievedJob, err := manager.GetJob(context.Background(), job.ID)
		assert.NoError(t, err)
		assert.Equal(t, job.ID, retrievedJob.ID)
		assert.Equal(t, "test-vm", retrievedJob.VMName)
		assert.Equal(t, "qcow2", retrievedJob.Format)
	})

	t.Run("Get non-existent job", func(t *testing.T) {
		// Test get non-existent job
		retrievedJob, err := manager.GetJob(context.Background(), "non-existent-job")
		assert.Error(t, err)
		assert.Nil(t, retrievedJob)
		assert.True(t, errors.Is(err, customErrors.ErrExportJobNotFound))
	})
}

func TestExportManager_ListJobs(t *testing.T) {
	// Setup
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStorageManager := mocks_storage.NewMockVolumeManager(ctrl)
	mockDomainManager := mocks_domain.NewMockManager(ctrl)
	mockLogger := mocks_logger.NewMockLogger(ctrl)

	// Create temporary directory for testing
	tmpDir, err := os.MkdirTemp("", "export-test-")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	// Create manager
	manager := &ExportManager{
		jobStore:       newJobStore(),
		formatManagers: map[string]formats.Converter{},
		storageManager: mockStorageManager,
		domainManager:  mockDomainManager,
		baseExportDir:  tmpDir,
		logger:         mockLogger,
	}

	// Create test jobs
	job1 := manager.jobStore.createJob("vm1", "qcow2", nil)
	job2 := manager.jobStore.createJob("vm2", "vmdk", nil)

	// Test list jobs
	jobs, err := manager.ListJobs(context.Background())
	assert.NoError(t, err)
	assert.Len(t, jobs, 2)

	// Verify all jobs are present (order not guaranteed)
	jobIDs := make(map[string]bool)
	for _, job := range jobs {
		jobIDs[job.ID] = true
	}
	assert.True(t, jobIDs[job1.ID])
	assert.True(t, jobIDs[job2.ID])
}

func TestExportManager_CancelJob(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping export manager test with background goroutines in short mode")
	}
	// Setup
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStorageManager := mocks_storage.NewMockVolumeManager(ctrl)
	mockDomainManager := mocks_domain.NewMockManager(ctrl)
	mockLogger := mocks_logger.NewMockLogger(ctrl)
	mockLogger.EXPECT().Info(gomock.Any(), gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Warn(gomock.Any(), gomock.Any()).AnyTimes()

	// Create temporary directory for testing
	tmpDir, err := os.MkdirTemp("", "export-test-")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	// Create manager
	manager := &ExportManager{
		jobStore:       newJobStore(),
		formatManagers: map[string]formats.Converter{},
		storageManager: mockStorageManager,
		domainManager:  mockDomainManager,
		baseExportDir:  tmpDir,
		logger:         mockLogger,
	}

	t.Run("Cancel pending job", func(t *testing.T) {
		// Create a test job
		job := manager.jobStore.createJob("test-vm", "qcow2", nil)

		// Cancel the job
		err := manager.CancelJob(context.Background(), job.ID)
		assert.NoError(t, err)

		// Verify job status
		updatedJob, exists := manager.jobStore.getJob(job.ID)
		assert.True(t, exists)
		assert.Equal(t, StatusCanceled, updatedJob.Status)
	})

	t.Run("Cancel running job", func(t *testing.T) {
		// Create a test job
		job := manager.jobStore.createJob("test-vm", "qcow2", nil)

		// Set job to running
		manager.jobStore.updateJobStatus(job.ID, StatusRunning, 50, nil)

		// Create output file
		outPath := filepath.Join(tmpDir, "test-output.qcow2")
		err := os.WriteFile(outPath, []byte("test"), 0644)
		require.NoError(t, err)
		manager.jobStore.setJobOutputPath(job.ID, outPath)

		// Cancel the job
		err = manager.CancelJob(context.Background(), job.ID)
		assert.NoError(t, err)

		// Verify job status
		updatedJob, exists := manager.jobStore.getJob(job.ID)
		assert.True(t, exists)
		assert.Equal(t, StatusCanceled, updatedJob.Status)
	})

	t.Run("Cancel completed job", func(t *testing.T) {
		// Create a test job
		job := manager.jobStore.createJob("test-vm", "qcow2", nil)

		// Set job to completed
		manager.jobStore.updateJobStatus(job.ID, StatusCompleted, 100, nil)

		// Try to cancel the job
		err := manager.CancelJob(context.Background(), job.ID)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "cannot cancel job in completed state")

		// Verify job status didn't change
		updatedJob, exists := manager.jobStore.getJob(job.ID)
		assert.True(t, exists)
		assert.Equal(t, StatusCompleted, updatedJob.Status)
	})

	t.Run("Cancel non-existent job", func(t *testing.T) {
		// Try to cancel a non-existent job
		err := manager.CancelJob(context.Background(), "non-existent-job")
		assert.Error(t, err)
		assert.True(t, errors.Is(err, customErrors.ErrExportJobNotFound))
	})
}

func TestJobStore(t *testing.T) {
	store := newJobStore()

	// Test creating a job
	job := store.createJob("test-vm", "qcow2", map[string]string{"key": "value"})
	assert.NotEmpty(t, job.ID)
	assert.Equal(t, "test-vm", job.VMName)
	assert.Equal(t, "qcow2", job.Format)
	assert.Equal(t, StatusPending, job.Status)
	assert.Equal(t, 0, job.Progress)
	assert.Equal(t, "value", job.Options["key"])

	// Test getting a job
	retrievedJob, exists := store.getJob(job.ID)
	assert.True(t, exists)
	assert.Equal(t, job.ID, retrievedJob.ID)

	// Test getting non-existent job
	_, exists = store.getJob("non-existent")
	assert.False(t, exists)

	// Test updating job status
	success := store.updateJobStatus(job.ID, StatusRunning, 50, nil)
	assert.True(t, success)

	retrievedJob, _ = store.getJob(job.ID)
	assert.Equal(t, StatusRunning, retrievedJob.Status)
	assert.Equal(t, 50, retrievedJob.Progress)

	// Test updating job with error
	testErr := errors.New("test error")
	success = store.updateJobStatus(job.ID, StatusFailed, 75, testErr)
	assert.True(t, success)

	retrievedJob, _ = store.getJob(job.ID)
	assert.Equal(t, StatusFailed, retrievedJob.Status)
	assert.Equal(t, 75, retrievedJob.Progress)
	assert.Equal(t, testErr.Error(), retrievedJob.Error)

	// Test setting output path
	success = store.setJobOutputPath(job.ID, "/path/to/output")
	assert.True(t, success)

	retrievedJob, _ = store.getJob(job.ID)
	assert.Equal(t, "/path/to/output", retrievedJob.OutputPath)

	// Test listing jobs
	job2 := store.createJob("another-vm", "vmdk", nil)
	jobs := store.listJobs()
	assert.Len(t, jobs, 2)

	// Verify jobs (order not guaranteed)
	jobIDs := make(map[string]bool)
	for _, j := range jobs {
		jobIDs[j.ID] = true
	}
	assert.True(t, jobIDs[job.ID])
	assert.True(t, jobIDs[job2.ID])
}
