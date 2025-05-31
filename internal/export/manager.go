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

// ExportManager implements Manager
type ExportManager struct {
	jobStore       *jobStore
	formatManagers map[string]formats.Converter
	storageManager storage.VolumeManager
	domainManager  domain.Manager
	baseExportDir  string
	logger         logger.Logger
}

// NewExportManager creates a new ExportManager
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

// CreateExportJob implements Manager.CreateExportJob
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

// GetJob implements Manager.GetJob
func (m *ExportManager) GetJob(ctx context.Context, jobID string) (*Job, error) {
	job, exists := m.jobStore.getJob(jobID)
	if !exists {
		return nil, fmt.Errorf("%w: %s", errors.ErrExportJobNotFound, jobID)
	}
	return job, nil
}

// CancelJob implements Manager.CancelJob
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

// ListJobs implements Manager.ListJobs
func (m *ExportManager) ListJobs(ctx context.Context) ([]*Job, error) {
	return m.jobStore.listJobs(), nil
}

// processExportJob processes an export job
func (m *ExportManager) processExportJob(job *Job, fileName string) {
	// Update job status to running
	m.jobStore.updateJobStatus(job.ID, StatusRunning, 5, nil)

	// Create export directory for this job
	jobDir := filepath.Join(m.baseExportDir, job.ID)
	if err := os.MkdirAll(jobDir, 0755); err != nil {
		m.logger.Error("Failed to create export job directory",
			logger.String("job_id", job.ID),
			logger.String("dir", jobDir),
			logger.String("error", err.Error()))
		m.jobStore.updateJobStatus(job.ID, StatusFailed, 0, err)
		return
	}

	// Define cleanup function
	cleanup := func() {
		// In case of failure or after successful conversion, clean up temp directory
		// Leave the output file in place
		// In case of success, caller may choose to keep the final output file
		os.RemoveAll(jobDir)
	}
	defer cleanup()

	// Setup context with cancellation
	ctx, cancel := context.WithCancel(context.Background())
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

	// Get VM disk path
	var sourceDiskPath string
	var poolName string
	var volName string

	// Check if source_volume was specified in options
	if sourceVolume, ok := job.Options["source_volume"]; ok && sourceVolume != "" {
		// Use the specified source volume
		poolName = "default" // Default to the default pool when explicitly specifying volume
		volName = sourceVolume
		m.logger.Info("Using explicitly specified source volume",
			logger.String("job_id", job.ID),
			logger.String("vm_name", job.VMName),
			logger.String("volume", volName),
			logger.String("pool", poolName))

		// Get disk path for the explicitly specified volume
		diskPath, pathErr := m.storageManager.GetPath(ctx, poolName, volName)
		if pathErr != nil {
			m.logger.Error("Failed to get disk path for specified source volume",
				logger.String("job_id", job.ID),
				logger.String("vm_name", job.VMName),
				logger.String("pool", poolName),
				logger.String("volume", volName),
				logger.String("error", pathErr.Error()))
			m.jobStore.updateJobStatus(job.ID, StatusFailed, 10, pathErr)
			return
		}
		sourceDiskPath = diskPath
	} else if len(vm.Disks) > 0 {
		// Find the VM's primary disk if no source_volume specified
		disk := vm.Disks[0]
		poolName = disk.StoragePool
		if poolName == "" {
			poolName = "default" // Use default pool if not specified
			m.logger.Warn("No storage pool specified for disk, using default pool",
				logger.String("job_id", job.ID),
				logger.String("vm_name", job.VMName),
				logger.String("disk_path", disk.Path))
		}

		// First try the standard naming convention, as this is what's most likely to be correct
		standardVolName := fmt.Sprintf("%s-disk-0", vm.Name)

		// Try to get the disk path using the standard naming convention first
		diskPath, stdErr := m.storageManager.GetPath(ctx, poolName, standardVolName)
		if stdErr == nil {
			// Standard naming convention works
			volName = standardVolName
			sourceDiskPath = diskPath
			m.logger.Info("Using standard disk naming convention succeeded",
				logger.String("job_id", job.ID),
				logger.String("vm_name", job.VMName),
				logger.String("volume", volName),
				logger.String("path", diskPath))
		} else if disk.Path != "" {
			// If standard naming fails, try using the path from the VM configuration
			volName = filepath.Base(disk.Path)
			m.logger.Info("Using volume name from disk path",
				logger.String("job_id", job.ID),
				logger.String("vm_name", job.VMName),
				logger.String("volume", volName),
				logger.String("disk_path", disk.Path))

			// Get disk path using the determined volume name
			diskPath, err = m.storageManager.GetPath(ctx, poolName, volName)
			if err != nil {
				// If the original pool isn't found, try the default pool as fallback
				if poolName != "default" {
					m.logger.Warn("Failed to get disk path with original pool, trying default pool",
						logger.String("job_id", job.ID),
						logger.String("vm_name", job.VMName),
						logger.String("original_pool", poolName),
						logger.String("volume", volName),
						logger.String("error", err.Error()))

					diskPath, err = m.storageManager.GetPath(ctx, "default", volName)
					if err != nil {
						// If both approaches fail, report error
						m.logger.Error("Failed to get disk path with all attempted methods",
							logger.String("job_id", job.ID),
							logger.String("vm_name", job.VMName),
							logger.String("pool", poolName),
							logger.String("standard_volume", standardVolName),
							logger.String("path_volume", volName),
							logger.String("error", err.Error()))
						m.jobStore.updateJobStatus(job.ID, StatusFailed, 10, err)
						return
					}
				} else {
					m.logger.Error("Failed to get disk path",
						logger.String("job_id", job.ID),
						logger.String("vm_name", job.VMName),
						logger.String("pool", poolName),
						logger.String("volume", volName),
						logger.String("error", err.Error()))
					m.jobStore.updateJobStatus(job.ID, StatusFailed, 10, err)
					return
				}
			}
			sourceDiskPath = diskPath
		} else {
			// Neither method worked
			determineDiskErr := fmt.Errorf("could not determine valid disk volume name")
			m.logger.Error("Failed to determine disk volume",
				logger.String("job_id", job.ID),
				logger.String("vm_name", job.VMName),
				logger.String("error", determineDiskErr.Error()))
			m.jobStore.updateJobStatus(job.ID, StatusFailed, 10, determineDiskErr)
			return
		}
	} else {
		diskErr := fmt.Errorf("VM has no disks")
		m.logger.Error("Failed to find VM disks",
			logger.String("job_id", job.ID),
			logger.String("vm_name", job.VMName),
			logger.String("error", diskErr.Error()))
		m.jobStore.updateJobStatus(job.ID, StatusFailed, 10, diskErr)
		return
	}

	// Update progress
	m.jobStore.updateJobStatus(job.ID, StatusRunning, 20, nil)

	// Get the destination path
	destPath := filepath.Join(m.baseExportDir, fileName)

	// Add VM information to options
	if job.Options == nil {
		job.Options = make(map[string]string)
	}
	job.Options["vm_name"] = vm.Name
	job.Options["vm_uuid"] = vm.UUID
	job.Options["cpu_count"] = fmt.Sprintf("%d", vm.CPU.Count)
	job.Options["memory_mb"] = fmt.Sprintf("%d", vm.Memory.SizeBytes/1024/1024)

	// Update progress
	m.jobStore.updateJobStatus(job.ID, StatusRunning, 30, nil)

	// Check VM status and stop if running (if requested)
	vmWasRunning := false
	if vm.Status == "running" {
		if shouldStop, ok := job.Options["stop_vm"]; ok && (shouldStop == "true" || shouldStop == "1") {
			m.logger.Info("Stopping VM for export",
				logger.String("job_id", job.ID),
				logger.String("vm_name", job.VMName))

			if stopErr := m.domainManager.Stop(ctx, job.VMName); stopErr != nil {
				m.logger.Error("Failed to stop VM",
					logger.String("job_id", job.ID),
					logger.String("vm_name", job.VMName),
					logger.String("error", stopErr.Error()))
				m.jobStore.updateJobStatus(job.ID, StatusFailed, 30, stopErr)
				return
			}
			vmWasRunning = true
		} else {
			m.logger.Warn("Exporting running VM, snapshot may be inconsistent",
				logger.String("job_id", job.ID),
				logger.String("vm_name", job.VMName))
		}
	}

	// Update progress - conversion starting
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

	err = converter.Convert(ctx, sourceDiskPath, destPath, job.Options)
	if err != nil {
		m.logger.Error("Export conversion failed",
			logger.String("job_id", job.ID),
			logger.String("vm_name", job.VMName),
			logger.String("format", job.Format),
			logger.String("error", err.Error()))
		m.jobStore.updateJobStatus(job.ID, StatusFailed, 50, err)
		return
	}

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
