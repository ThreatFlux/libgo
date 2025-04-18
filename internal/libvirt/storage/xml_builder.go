package storage

import (
	"fmt"

	"github.com/wroersma/libgo/pkg/logger"
	xmlutils "github.com/wroersma/libgo/pkg/utils/xml"
)

// TemplateXMLBuilder implements XMLBuilder using templates
type TemplateXMLBuilder struct {
	templateLoader *xmlutils.TemplateLoader
	logger         logger.Logger
}

// PoolTemplate contains data for storage pool XML template
type PoolTemplate struct {
	Name string
	Path string
}

// VolumeTemplate contains data for storage volume XML template
type VolumeTemplate struct {
	Name          string
	CapacityBytes uint64
	Format        string
}

// NewTemplateXMLBuilder creates a new TemplateXMLBuilder
func NewTemplateXMLBuilder(templateLoader *xmlutils.TemplateLoader, logger logger.Logger) *TemplateXMLBuilder {
	return &TemplateXMLBuilder{
		templateLoader: templateLoader,
		logger:         logger,
	}
}

// BuildStoragePoolXML implements XMLBuilder.BuildStoragePoolXML
func (b *TemplateXMLBuilder) BuildStoragePoolXML(name string, path string) (string, error) {
	// Prepare template data
	templateData := PoolTemplate{
		Name: name,
		Path: path,
	}

	// Render the template
	b.logger.Debug("Rendering storage pool XML template",
		logger.String("pool_name", name),
		logger.String("path", path))

	poolXML, err := b.templateLoader.RenderTemplate("storage_pool.xml.tmpl", templateData)
	if err != nil {
		return "", fmt.Errorf("failed to render storage pool XML template: %w", err)
	}

	return poolXML, nil
}

// BuildStorageVolumeXML implements XMLBuilder.BuildStorageVolumeXML
func (b *TemplateXMLBuilder) BuildStorageVolumeXML(volName string, capacityBytes uint64, format string) (string, error) {
	// Default format if not specified
	if format == "" {
		format = "qcow2"
	}

	// Prepare template data
	templateData := VolumeTemplate{
		Name:          volName,
		CapacityBytes: capacityBytes,
		Format:        format,
	}

	// Render the template
	b.logger.Debug("Rendering storage volume XML template",
		logger.String("volume_name", volName),
		logger.Uint64("capacity_bytes", capacityBytes),
		logger.String("format", format))

	volumeXML, err := b.templateLoader.RenderTemplate("storage_volume.xml.tmpl", templateData)
	if err != nil {
		return "", fmt.Errorf("failed to render storage volume XML template: %w", err)
	}

	return volumeXML, nil
}
