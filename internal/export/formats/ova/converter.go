package ova

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/threatflux/libgo/internal/models/vm"
	"github.com/threatflux/libgo/pkg/logger"
	"github.com/threatflux/libgo/pkg/utils/exec"
)

// OVAConverter implements Converter for OVA format.
type OVAConverter struct {
	templateGenerator *OVFTemplateGenerator
	logger            logger.Logger
}

// NewOVAConverter creates a new OVAConverter.
func NewOVAConverter(templateGenerator *OVFTemplateGenerator, logger logger.Logger) *OVAConverter {
	return &OVAConverter{
		templateGenerator: templateGenerator,
		logger:            logger,
	}
}

// Convert implements Converter.Convert.
// For OVA conversion, sourcePath is the path to the VM disk
// and destPath is the path where to create the OVA file.
func (c *OVAConverter) Convert(ctx context.Context, sourcePath string, destPath string, options map[string]string) error {
	// Create temporary directory for intermediate files
	tempDir, err := os.MkdirTemp("", "ova-export-")
	if err != nil {
		return fmt.Errorf("failed to create temporary directory: %w", err)
	}
	defer os.RemoveAll(tempDir)

	// Extract VM information from options
	vmInfo, err := c.getVMInfoFromOptions(options)
	if err != nil {
		return fmt.Errorf("invalid VM info in options: %w", err)
	}

	// Step 1: Convert disk to VMDK format (which is standard for OVA)
	vmdkPath := filepath.Join(tempDir, "disk.vmdk")
	if convertErr := c.convertToDisk(ctx, sourcePath, vmdkPath); convertErr != nil {
		return fmt.Errorf("failed to convert disk to VMDK: %w", convertErr)
	}

	// Get disk size
	diskInfo, err := os.Stat(vmdkPath)
	if err != nil {
		return fmt.Errorf("failed to get disk size: %w", err)
	}
	diskSize := uint64(diskInfo.Size())

	// Step 2: Generate OVF file
	ovfPath := filepath.Join(tempDir, "vm.ovf")
	ovfContent, err := c.templateGenerator.GenerateOVF(vmInfo, vmdkPath, diskSize)
	if err != nil {
		return fmt.Errorf("failed to generate OVF descriptor: %w", err)
	}

	if err := c.templateGenerator.WriteOVFToFile(ovfContent, ovfPath); err != nil {
		return fmt.Errorf("failed to write OVF descriptor: %w", err)
	}

	// Step 3: Package as OVA
	if err := c.packageOVA(ctx, vmdkPath, ovfPath, destPath); err != nil {
		return fmt.Errorf("failed to package as OVA: %w", err)
	}

	c.logger.Info("Successfully converted to OVA format",
		logger.String("destination", destPath))
	return nil
}

// convertToDisk converts source to VMDK format.
func (c *OVAConverter) convertToDisk(ctx context.Context, sourcePath string, destPath string) error {
	args := []string{
		"convert",
		"-f", "qcow2", // Source format, assume qcow2 as libvirt generally uses this
		"-O", "vmdk", // Output format
		"-o", "adapter_type=lsilogic,subformat=streamOptimized", // VMware compatible settings
		sourcePath,
		destPath,
	}

	c.logger.Info("Converting disk to VMDK format for OVA packaging",
		logger.String("source", sourcePath),
		logger.String("destination", destPath))

	// Execute qemu-img command
	execOpts := exec.CommandOptions{
		Timeout: 0, // No timeout, conversion might take a long time
	}

	output, err := exec.ExecuteCommand(ctx, "qemu-img", args, execOpts)
	if err != nil {
		c.logger.Error("Failed to convert disk to VMDK for OVA packaging",
			logger.String("error", err.Error()),
			logger.String("output", string(output)))
		return fmt.Errorf("disk conversion failed: %w", err)
	}

	return nil
}

// packageOVA packages VMDK and OVF into OVA.
func (c *OVAConverter) packageOVA(ctx context.Context, vmdkPath string, ovfPath string, ovaPath string) error {
	c.logger.Info("Packaging OVA",
		logger.String("ovf", ovfPath),
		logger.String("vmdk", vmdkPath),
		logger.String("destination", ovaPath))

	// Create a TAR archive
	// Change to the directory containing the files
	tempDir := filepath.Dir(ovfPath)

	// Construct tar command
	// OVF descriptor must be the first file in the archive
	args := []string{
		"-cf", // Create file
		ovaPath,
		"-C", tempDir,
		filepath.Base(ovfPath),  // OVF first
		filepath.Base(vmdkPath), // Then VMDK
	}

	execOpts := exec.CommandOptions{
		Timeout: 0, // No timeout, packaging might take a long time
	}

	output, err := exec.ExecuteCommand(ctx, "tar", args, execOpts)
	if err != nil {
		c.logger.Error("Failed to package OVA",
			logger.String("error", err.Error()),
			logger.String("output", string(output)))
		return fmt.Errorf("ova packaging failed: %w", err)
	}

	return nil
}

// getVMInfoFromOptions extracts VM information from options.
func (c *OVAConverter) getVMInfoFromOptions(options map[string]string) (*vm.VM, error) {
	// Minimal VM info required for OVF
	vmInfo := &vm.VM{
		Name: options["vm_name"],
	}

	if vmInfo.Name == "" {
		return nil, fmt.Errorf("vm_name is required in options")
	}

	// Add other VM properties if provided
	if uuid, ok := options["vm_uuid"]; ok {
		vmInfo.UUID = uuid
	}

	if cpuCount, ok := options["cpu_count"]; ok {
		count := 1
		if _, err := fmt.Sscanf(cpuCount, "%d", &count); err != nil {
			count = 1 // Default to 1 CPU if parsing fails
		}
		vmInfo.CPU.Count = count
	}

	if memoryMB, ok := options["memory_mb"]; ok {
		memory := uint64(1024)
		if _, err := fmt.Sscanf(memoryMB, "%d", &memory); err != nil {
			memory = 1024 // Default to 1GB if parsing fails
		}
		// Convert MB to bytes (1 MB = 1024 * 1024 bytes)
		vmInfo.Memory.SizeBytes = memory * 1024 * 1024
	}

	return vmInfo, nil
}

// GetFormatName implements Converter.GetFormatName.
func (c *OVAConverter) GetFormatName() string {
	return "ova"
}

// ValidateOptions implements Converter.ValidateOptions.
func (c *OVAConverter) ValidateOptions(options map[string]string) error {
	// Required options
	requiredOptions := []string{"vm_name"}
	for _, opt := range requiredOptions {
		if _, ok := options[opt]; !ok {
			return fmt.Errorf("missing required option: %s", opt)
		}
	}

	return nil
}
