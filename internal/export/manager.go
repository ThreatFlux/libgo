package export

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/threatflux/libgo/internal/errors"
	"github.com/threatflux/libgo/internal/export/formats"
	"github.com/threatflux/libgo/internal/export/formats/ova"
	"github.com/threatflux/libgo/internal/export/formats/qcow2"
	"github.com/threatflux/libgo/internal/export/formats/raw"
	"github.com/threatflux/libgo/internal/export/formats/vdi"
	"github.com/threatflux/libgo/internal/export/formats/vmdk"
	"github.com/threatflux/libgo/internal/libvirt/domain"
	"github.com/threatflux/libgo/internal/libvirt/storage"
	"github.com/threatflux/libgo/pkg/logger"
)

// ExportManager implements Manager.
type ExportManager struct {
	jobStore       *jobStore
	formatManagers map[string]formats.Converter
	storageManager storage.VolumeManager
	domainManager  domain.Manager
	logger         logger.Logger
	baseExportDir  string
}

// NewExportManager creates a new ExportManager.
func NewExportManager(
	storageManager storage.VolumeManager,
	domainManager domain.Manager,
	baseExportDir string,
	logger logger.Logger,
) (*ExportManager, error) {
	// Create export directory if it doesn't exist
	if err := os.MkdirAll(baseExportDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create export directory: %w", err)
	}

	// Create OVF template generator for OVA exports
	ovfGenerator, err := ova.NewOVFTemplateGenerator(logger)
	if err != nil {
		return nil, fmt.Errorf("failed to create OVF template generator: %w", err)
	}

	// Initialize format converters
	formatConverters := map[string]formats.Converter{
		"qcow2": qcow2.NewQCOW2Converter(logger),
		"vmdk":  vmdk.NewVMDKConverter(logger),
		"vdi":   vdi.NewVDIConverter(logger),
		"ova":   ova.NewOVAConverter(ovfGenerator, logger),
		"raw":   raw.NewRAWConverter(logger),
	}

	manager := &ExportManager{
		jobStore:       newJobStore(),
		formatManagers: formatConverters,
		storageManager: storageManager,
		domainManager:  domainManager,
		baseExportDir:  baseExportDir,
		logger:         logger,
	}

	return manager, nil
}

// CreateExportJob implements Manager.CreateExportJob.
func (m *ExportManager) CreateExportJob(ctx context.Context, vmName string, params Params) (*Job, error) {
	// Check if VM exists
	_, err := m.domainManager.Get(ctx, vmName)
	if err != nil {
		return nil, fmt.Errorf("VM not found: %w", err)
	}

	// Validate format
	converter, ok := m.formatManagers[params.Format]
	if !ok {
		return nil, fmt.Errorf("%w: %s", errors.ErrUnsupportedFormat, params.Format)
	}

	// Validate options
	if err := converter.ValidateOptions(params.Options); err != nil {
		return nil, fmt.Errorf("invalid export options: %w", err)
	}

	// Generate a filename if not provided
	fileName := params.FileName
	if fileName == "" {
		timestamp := time.Now().Format("20060102-150405")
		fileName = fmt.Sprintf("%s-%s.%s", vmName, timestamp, params.Format)
	}

	// Create job
	job := m.jobStore.createJob(vmName, params.Format, params.Options)

	// Start processing job in background
	go m.processExportJob(job, fileName)

	return job, nil
}

// GetJob implements Manager.GetJob.
func (m *ExportManager) GetJob(ctx context.Context, jobID string) (*Job, error) {
	job, exists := m.jobStore.getJob(jobID)
	if !exists {
		return nil, fmt.Errorf("%w: %s", errors.ErrExportJobNotFound, jobID)
	}
	return job, nil
}

// CancelJob implements Manager.CancelJob.
func (m *ExportManager) CancelJob(ctx context.Context, jobID string) error {
	// Get job
	job, exists := m.jobStore.getJob(jobID)
	if !exists {
		return fmt.Errorf("%w: %s", errors.ErrExportJobNotFound, jobID)
	}

	// Only pending or running jobs can be canceled
	if job.Status != StatusPending && job.Status != StatusRunning {
		return fmt.Errorf("cannot cancel job in %s state", job.Status)
	}

	// Update job status
	if !m.jobStore.updateJobStatus(jobID, StatusCanceled, job.Progress, nil) {
		return fmt.Errorf("failed to update job status")
	}

	// If an output file was created, remove it
	if job.OutputPath != "" {
		if err := os.Remove(job.OutputPath); err != nil {
			m.logger.Warn("Failed to remove output file after job cancellation",
				logger.String("job_id", jobID),
				logger.String("file", job.OutputPath),
				logger.String("error", err.Error()))
		}
	}

	m.logger.Info("Export job canceled",
		logger.String("job_id", jobID),
		logger.String("vm_name", job.VMName))

	return nil
}

// ListJobs implements Manager.ListJobs.
func (m *ExportManager) ListJobs(ctx context.Context) ([]*Job, error) {
	return m.jobStore.listJobs(), nil
}

// processExportJob processes an export job.
func (m *ExportManager) processExportJob(job *Job, fileName string) {
	// Update job status to running
	m.jobStore.updateJobStatus(job.ID, StatusRunning, 5, nil)

	// Setup job environment
	_, cleanup, ctx, cancel, err := m.setupJobEnvironment(job)
	if err != nil {
		m.jobStore.updateJobStatus(job.ID, StatusFailed, 0, err)
		return
	}
	defer cleanup()
	defer cancel()

	// Get VM details
	vm, err := m.domainManager.Get(ctx, job.VMName)
	if err != nil {
		m.logger.Error("Failed to get VM information",
			logger.String("job_id", job.ID),
			logger.String("vm_name", job.VMName),
			logger.String("error", err.Error()))
		m.jobStore.updateJobStatus(job.ID, StatusFailed, 0, err)
		return
	}

	// Update progress
	m.jobStore.updateJobStatus(job.ID, StatusRunning, 10, nil)

	// Resolve disk path
	sourceDiskPath, err := m.resolveDiskPath(ctx, job, vm)
	if err != nil {
		m.jobStore.updateJobStatus(job.ID, StatusFailed, 10, err)
		return
	}

	// Update progress
	m.jobStore.updateJobStatus(job.ID, StatusRunning, 20, nil)

	// Prepare VM for export
	destPath := filepath.Join(m.baseExportDir, fileName)
	m.enrichJobOptions(job, vm)
	m.jobStore.updateJobStatus(job.ID, StatusRunning, 30, nil)

	// Handle VM state (stop if running)
	vmWasRunning, err := m.handleVMState(ctx, job, vm)
	if err != nil {
		m.jobStore.updateJobStatus(job.ID, StatusFailed, 30, err)
		return
	}

	// Perform the export conversion
	err = m.performConversion(ctx, job, sourceDiskPath, destPath)
	if err != nil {
		m.jobStore.updateJobStatus(job.ID, StatusFailed, 50, err)
		return
	}

	// Finalize the export
	m.finalizeExport(ctx, job, destPath, vmWasRunning)
}

// setupJobEnvironment creates the job directory and context.
func (m *ExportManager) setupJobEnvironment(job *Job) (string, func(), context.Context, context.CancelFunc, error) {
	// Create export directory for this job
	jobDir := filepath.Join(m.baseExportDir, job.ID)
	if err := os.MkdirAll(jobDir, 0755); err != nil {
		m.logger.Error("Failed to create export job directory",
			logger.String("job_id", job.ID),
			logger.String("dir", jobDir),
			logger.String("error", err.Error()))
		return "", nil, nil, nil, err
	}

	// Define cleanup function
	cleanup := func() {
		// In case of failure or after successful conversion, clean up temp directory
		// Leave the output file in place
		// In case of success, caller may choose to keep the final output file
		os.RemoveAll(jobDir)
	}

	// Setup context with cancellation
	ctx, cancel := context.WithCancel(context.Background())

	return jobDir, cleanup, ctx, cancel, nil
}

// resolveDiskPath resolves the source disk path for the VM.
func (m *ExportManager) resolveDiskPath(ctx context.Context, job *Job, vm interface{}) (string, error) {
	// Check if source_volume was specified in options
	if sourceVolume, ok := job.Options["source_volume"]; ok && sourceVolume != "" {
		return m.getSpecifiedVolumePath(ctx, job, sourceVolume)
	}

	// Use standard naming convention as fallback
	standardVolName := fmt.Sprintf("%s-disk-0", job.VMName)
	poolName := "default"

	// Try standard naming convention
	diskPath, err := m.storageManager.GetPath(ctx, poolName, standardVolName)
	if err == nil {
		m.logger.Info("Using standard disk naming convention",
			logger.String("job_id", job.ID),
			logger.String("vm_name", job.VMName),
			logger.String("volume", standardVolName))
		return diskPath, nil
	}

	return "", fmt.Errorf("could not determine valid disk volume name")
}

// getSpecifiedVolumePath gets the path for an explicitly specified volume.
func (m *ExportManager) getSpecifiedVolumePath(ctx context.Context, job *Job, sourceVolume string) (string, error) {
	poolName := "default" // Default pool when explicitly specifying volume
	m.logger.Info("Using explicitly specified source volume",
		logger.String("job_id", job.ID),
		logger.String("vm_name", job.VMName),
		logger.String("volume", sourceVolume),
		logger.String("pool", poolName))

	diskPath, err := m.storageManager.GetPath(ctx, poolName, sourceVolume)
	if err != nil {
		m.logger.Error("Failed to get disk path for specified source volume",
			logger.String("job_id", job.ID),
			logger.String("vm_name", job.VMName),
			logger.String("pool", poolName),
			logger.String("volume", sourceVolume),
			logger.String("error", err.Error()))
		return "", err
	}
	return diskPath, nil
}

// enrichJobOptions adds VM information to job options.
func (m *ExportManager) enrichJobOptions(job *Job, vm interface{}) {
	if job.Options == nil {
		job.Options = make(map[string]string)
	}

	// For simplified implementation, add basic options
	job.Options["vm_name"] = job.VMName
	job.Options["vm_uuid"] = "unknown"
	job.Options["cpu_count"] = "1"
	job.Options["memory_mb"] = "1024"
}

// handleVMState handles VM state changes (stopping if requested).
func (m *ExportManager) handleVMState(ctx context.Context, job *Job, vm interface{}) (bool, error) {
	vmWasRunning := false

	// Check if VM should be stopped
	if shouldStop, ok := job.Options["stop_vm"]; ok && (shouldStop == "true" || shouldStop == "1") {
		m.logger.Info("Stopping VM for export",
			logger.String("job_id", job.ID),
			logger.String("vm_name", job.VMName))

		if err := m.domainManager.Stop(ctx, job.VMName); err != nil {
			m.logger.Error("Failed to stop VM",
				logger.String("job_id", job.ID),
				logger.String("vm_name", job.VMName),
				logger.String("error", err.Error()))
			return false, err
		}
		vmWasRunning = true
	}

	return vmWasRunning, nil
}

// performConversion performs the actual export conversion.
func (m *ExportManager) performConversion(ctx context.Context, job *Job, sourceDiskPath, destPath string) error {
	// Update progress
	m.jobStore.updateJobStatus(job.ID, StatusRunning, 40, nil)

	// Get converter
	converter := m.formatManagers[job.Format]

	// Perform conversion
	m.logger.Info("Starting export conversion",
		logger.String("job_id", job.ID),
		logger.String("vm_name", job.VMName),
		logger.String("format", job.Format),
		logger.String("source", sourceDiskPath),
		logger.String("destination", destPath))

	err := converter.Convert(ctx, sourceDiskPath, destPath, job.Options)
	if err != nil {
		m.logger.Error("Export conversion failed",
			logger.String("job_id", job.ID),
			logger.String("vm_name", job.VMName),
			logger.String("format", job.Format),
			logger.String("error", err.Error()))
		return err
	}

	return nil
}

// finalizeExport completes the export process.
func (m *ExportManager) finalizeExport(ctx context.Context, job *Job, destPath string, vmWasRunning bool) {
	// Update job status to completed
	m.jobStore.updateJobStatus(job.ID, StatusCompleted, 100, nil)
	m.jobStore.setJobOutputPath(job.ID, destPath)

	// Start VM again if it was running before
	if vmWasRunning {
		if err := m.domainManager.Start(ctx, job.VMName); err != nil {
			m.logger.Error("Failed to restart VM after export",
				logger.String("job_id", job.ID),
				logger.String("vm_name", job.VMName),
				logger.String("error", err.Error()))
		}
	}

	m.logger.Info("Export job completed successfully",
		logger.String("job_id", job.ID),
		logger.String("vm_name", job.VMName),
		logger.String("format", job.Format),
		logger.String("output", destPath))
}
