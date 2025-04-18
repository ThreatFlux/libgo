package domain

import (
	"fmt"
	"path/filepath"

	"github.com/google/uuid"
	"github.com/wroersma/libgo/internal/models/vm"
	"github.com/wroersma/libgo/pkg/logger"
	xmlutils "github.com/wroersma/libgo/pkg/utils/xml"
)

// TemplateXMLBuilder implements XMLBuilder using templates
type TemplateXMLBuilder struct {
	templateLoader *xmlutils.TemplateLoader
	logger         logger.Logger
}

// DomainTemplate contains data for domain XML template
type DomainTemplate struct {
	Name         string
	UUID         string
	Memory       MemoryTemplate
	CPU          CPUTemplate
	Disks        []DiskTemplate
	Networks     []NetworkTemplate
	CloudInitISO string
}

// MemoryTemplate contains memory data for the template
type MemoryTemplate struct {
	KiB uint64
}

// CPUTemplate contains CPU data for the template
type CPUTemplate struct {
	Count   int
	Model   string
	Cores   int
	Threads int
	Sockets int
}

// DiskTemplate contains disk data for the template
type DiskTemplate struct {
	Type       string
	Format     string
	Source     string
	SourceAttr string
	Device     string
	Bus        string
	Bootable   bool
	ReadOnly   bool
	Shareable  bool
}

// NetworkTemplate contains network data for the template
type NetworkTemplate struct {
	Type       string
	Source     string
	SourceAttr string
	MacAddress string
	Model      string
}

// NewTemplateXMLBuilder creates a new TemplateXMLBuilder
func NewTemplateXMLBuilder(templateLoader *xmlutils.TemplateLoader, logger logger.Logger) *TemplateXMLBuilder {
	return &TemplateXMLBuilder{
		templateLoader: templateLoader,
		logger:         logger,
	}
}

// BuildDomainXML implements XMLBuilder.BuildDomainXML
func (b *TemplateXMLBuilder) BuildDomainXML(params vm.VMParams) (string, error) {
	// Generate a UUID if not provided
	domainUUID := uuid.New().String()

	// Prepare memory (convert bytes to KiB)
	memoryKiB := params.Memory.SizeBytes / 1024

	// Prepare CPU info
	cpuTemplate := CPUTemplate{
		Count:   params.CPU.Count,
		Model:   params.CPU.Model,
		Cores:   params.CPU.Cores,
		Threads: params.CPU.Threads,
		Sockets: params.CPU.Socket,
	}

	// Prepare disk info
	disks := []DiskTemplate{}

	// Add the primary disk
	primaryDisk := DiskTemplate{
		Type:       string(vm.DiskTypeFile),
		Format:     string(params.Disk.Format),
		SourceAttr: "file",
		Device:     "vda", // Default device name for primary disk
		Bus:        string(params.Disk.GetBus()),
		Bootable:   true,
		ReadOnly:   params.Disk.ReadOnly,
		Shareable:  params.Disk.Shareable,
	}

	// If source image is provided, use it as the source
	if params.Disk.SourceImage != "" {
		primaryDisk.Source = params.Disk.SourceImage
	} else {
		// Otherwise create a new disk path
		storagePool := params.Disk.StoragePool
		if storagePool == "" {
			storagePool = vm.GetDefaultStoragePool()
		}
		primaryDisk.Source = fmt.Sprintf("/var/lib/libvirt/images/%s/%s.%s",
			storagePool,
			params.Name,
			params.Disk.Format)
	}

	disks = append(disks, primaryDisk)

	// Add any additional disks if they exist in the future
	// This would iterate through params.AdditionalDisks if implemented

	// Prepare network info
	networks := []NetworkTemplate{}
	if params.Network.Type != "" {
		networkTemplate := NetworkTemplate{
			Type:       string(params.Network.Type),
			Source:     params.Network.Source,
			MacAddress: params.Network.MacAddress,
			Model:      "virtio", // Default
		}

		// Set proper source attribute based on type
		switch string(params.Network.Type) {
		case "bridge":
			networkTemplate.SourceAttr = "bridge"
		case "network":
			networkTemplate.SourceAttr = "network"
		case "direct":
			networkTemplate.SourceAttr = "dev"
		}

		// Override model if specified
		if params.Network.Model != "" {
			networkTemplate.Model = params.Network.Model
		}

		networks = append(networks, networkTemplate)
	}

	// Cloud-init ISO path, if cloud-init config is provided
	var cloudInitISOPath string
	if params.CloudInit.UserData != "" || params.CloudInit.MetaData != "" {
		cloudInitISOPath = fmt.Sprintf("/home/vtriple/libgo-temp/cloudinit/%s-cloudinit.iso", params.Name)
	}

	// Prepare template data
	templateData := DomainTemplate{
		Name:         params.Name,
		UUID:         domainUUID,
		Memory:       MemoryTemplate{KiB: memoryKiB},
		CPU:          cpuTemplate,
		Disks:        disks,
		Networks:     networks,
		CloudInitISO: cloudInitISOPath,
	}

	// Render the template
	b.logger.Debug("Rendering domain XML template",
		logger.String("vm_name", params.Name),
		logger.Int("cpu_count", params.CPU.Count),
		logger.Uint64("memory_bytes", params.Memory.SizeBytes))

	domainXML, err := b.templateLoader.RenderTemplate("domain.xml.tmpl", templateData)
	if err != nil {
		return "", fmt.Errorf("failed to render domain XML template: %w", err)
	}

	return domainXML, nil
}

// GenerateCloudInitISOPath generates a path for cloud-init ISO
func (b *TemplateXMLBuilder) GenerateCloudInitISOPath(vmName string, isoDir string) string {
	// Create filename
	filename := fmt.Sprintf("%s-cloudinit.iso", vmName)

	// If directory is provided, join with filename
	if isoDir != "" {
		return filepath.Join(isoDir, filename)
	}

	// Default location in libvirt images directory
	return fmt.Sprintf("/var/lib/libvirt/images/cloudinit/%s", filename)
}
