package template

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/wroersma/libgo/internal/models/vm"
	"github.com/wroersma/libgo/pkg/logger"
)

// TemplateManager implements Manager for VM templates
type TemplateManager struct {
	templates  map[string]vm.VMParams
	logger     logger.Logger
	templateDir string
}

// NewTemplateManager creates a new TemplateManager
func NewTemplateManager(templateDir string, logger logger.Logger) (*TemplateManager, error) {
	manager := &TemplateManager{
		templates:   make(map[string]vm.VMParams),
		logger:      logger,
		templateDir: templateDir,
	}

	// Load templates
	err := manager.loadTemplates()
	if err != nil {
		return nil, fmt.Errorf("loading templates: %w", err)
	}

	return manager, nil
}

// GetTemplate implements Manager.GetTemplate
func (m *TemplateManager) GetTemplate(name string) (*vm.VMParams, error) {
	template, exists := m.templates[name]
	if !exists {
		return nil, fmt.Errorf("template not found: %s", name)
	}

	// Return a copy of the template to prevent modification
	templateCopy := template
	return &templateCopy, nil
}

// ListTemplates implements Manager.ListTemplates
func (m *TemplateManager) ListTemplates() ([]string, error) {
	templateNames := make([]string, 0, len(m.templates))
	for name := range m.templates {
		templateNames = append(templateNames, name)
	}
	return templateNames, nil
}

// ApplyTemplate implements Manager.ApplyTemplate
func (m *TemplateManager) ApplyTemplate(templateName string, params *vm.VMParams) error {
	template, err := m.GetTemplate(templateName)
	if err != nil {
		return err
	}

	// Apply template values for fields that are not set in params
	if params.CPU.Count == 0 {
		params.CPU = template.CPU
	}
	
	if params.Memory.SizeBytes == 0 {
		params.Memory = template.Memory
	}
	
	if params.Disk.SizeBytes == 0 {
		params.Disk = template.Disk
	} else if params.Disk.Format == "" {
		params.Disk.Format = template.Disk.Format
	}
	
	// Only apply network settings if not already set
	if params.Network.Type == "" {
		params.Network = template.Network
	}
	
	// Apply cloud-init settings if not already provided
	if params.CloudInit.UserData == "" {
		params.CloudInit = template.CloudInit
	}

	return nil
}

// loadTemplates loads VM templates from JSON files in the template directory
func (m *TemplateManager) loadTemplates() error {
	if _, err := os.Stat(m.templateDir); os.IsNotExist(err) {
		m.logger.Warn("Template directory does not exist", logger.String("dir", m.templateDir))
		return nil
	}

	files, err := os.ReadDir(m.templateDir)
	if err != nil {
		return fmt.Errorf("reading template directory: %w", err)
	}

	for _, file := range files {
		if file.IsDir() || !strings.HasSuffix(file.Name(), ".json") {
			continue
		}

		templatePath := filepath.Join(m.templateDir, file.Name())
		templateName := strings.TrimSuffix(file.Name(), ".json")

		// Read template file
		templateData, err := os.ReadFile(templatePath)
		if err != nil {
			m.logger.Warn("Failed to read template file",
				logger.String("file", templatePath),
				logger.Error(err))
			continue
		}

		// Parse template JSON
		var template vm.VMParams
		if err := json.Unmarshal(templateData, &template); err != nil {
			m.logger.Warn("Failed to parse template JSON",
				logger.String("file", templatePath),
				logger.Error(err))
			continue
		}

		m.templates[templateName] = template
		m.logger.Debug("Loaded VM template", logger.String("name", templateName))
	}

	m.logger.Info("Loaded VM templates", logger.Int("count", len(m.templates)))
	return nil
}