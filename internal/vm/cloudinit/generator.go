package cloudinit

import (
	"fmt"
	"strings"
	"text/template"

	"github.com/wroersma/libgo/internal/models/vm"
	"github.com/wroersma/libgo/pkg/logger"
)

// CloudInitGenerator implements Manager for cloud-init
type CloudInitGenerator struct {
	templateDir string
	logger      logger.Logger
	templates   map[string]*template.Template
}

// NewCloudInitGenerator creates a new CloudInitGenerator
func NewCloudInitGenerator(templateDir string, logger logger.Logger) (*CloudInitGenerator, error) {
	g := &CloudInitGenerator{
		templateDir: templateDir,
		logger:      logger,
		templates:   make(map[string]*template.Template),
	}

	// Load templates
	err := g.loadTemplates()
	if err != nil {
		return nil, fmt.Errorf("loading cloud-init templates: %w", err)
	}

	return g, nil
}

// GenerateUserData implements Manager.GenerateUserData
func (g *CloudInitGenerator) GenerateUserData(params vm.VMParams) (string, error) {
	tmpl, ok := g.templates["user-data.tmpl"]
	if !ok {
		return "", fmt.Errorf("user-data template not found")
	}

	// If custom user-data is provided in params, use it directly
	if params.CloudInit.UserData != "" {
		return params.CloudInit.UserData, nil
	}

	// Build template data
	data := map[string]interface{}{
		"VM": params,
	}

	// Execute template
	var result strings.Builder
	if err := tmpl.Execute(&result, data); err != nil {
		return "", fmt.Errorf("executing user-data template: %w", err)
	}

	g.logger.Debug("Generated user-data", logger.String("name", params.Name))
	return result.String(), nil
}

// GenerateMetaData implements Manager.GenerateMetaData
func (g *CloudInitGenerator) GenerateMetaData(params vm.VMParams) (string, error) {
	tmpl, ok := g.templates["meta-data.tmpl"]
	if !ok {
		return "", fmt.Errorf("meta-data template not found")
	}

	// If custom meta-data is provided in params, use it directly
	if params.CloudInit.MetaData != "" {
		return params.CloudInit.MetaData, nil
	}

	// Generate a random UUID if not provided
	instanceID := params.Name

	// Build template data
	data := map[string]interface{}{
		"VM":         params,
		"InstanceID": instanceID,
		"Hostname":   params.Name,
	}

	// Execute template
	var result strings.Builder
	if err := tmpl.Execute(&result, data); err != nil {
		return "", fmt.Errorf("executing meta-data template: %w", err)
	}

	g.logger.Debug("Generated meta-data", logger.String("name", params.Name))
	return result.String(), nil
}

// GenerateNetworkConfig implements Manager.GenerateNetworkConfig
func (g *CloudInitGenerator) GenerateNetworkConfig(params vm.VMParams) (string, error) {
	tmpl, ok := g.templates["network-config.tmpl"]
	if !ok {
		return "", fmt.Errorf("network-config template not found")
	}

	// If custom network-config is provided in params, use it directly
	if params.CloudInit.NetworkConfig != "" {
		return params.CloudInit.NetworkConfig, nil
	}

	// Build template data
	data := map[string]interface{}{
		"VM": params,
	}

	// Execute template
	var result strings.Builder
	if err := tmpl.Execute(&result, data); err != nil {
		return "", fmt.Errorf("executing network-config template: %w", err)
	}

	g.logger.Debug("Generated network-config", logger.String("name", params.Name))
	return result.String(), nil
}

// loadTemplates loads cloud-init templates
func (g *CloudInitGenerator) loadTemplates() error {
	// List of template files to load
	templateFiles := []string{
		"user-data.tmpl",
		"meta-data.tmpl",
		"network-config.tmpl",
	}

	for _, filename := range templateFiles {
		templatePath := fmt.Sprintf("%s/%s", g.templateDir, filename)
		tmpl, err := template.ParseFiles(templatePath)
		if err != nil {
			g.logger.Warn("Failed to load template",
				logger.String("template", filename),
				logger.Error(err))
			
			// Create a default template if file is not found
			tmpl, err = g.createDefaultTemplate(filename)
			if err != nil {
				return fmt.Errorf("creating default template for %s: %w", filename, err)
			}
		}
		
		g.templates[filename] = tmpl
		g.logger.Debug("Loaded cloud-init template", logger.String("name", filename))
	}

	return nil
}

// createDefaultTemplate creates a default template if the template file is not found
func (g *CloudInitGenerator) createDefaultTemplate(filename string) (*template.Template, error) {
	var templateContent string
	
	switch filename {
	case "user-data.tmpl":
		templateContent = defaultUserDataTemplate
	case "meta-data.tmpl":
		templateContent = defaultMetaDataTemplate
	case "network-config.tmpl":
		templateContent = defaultNetworkConfigTemplate
	default:
		return nil, fmt.Errorf("no default template for %s", filename)
	}
	
	return template.New(filename).Parse(templateContent)
}

// Default template contents

const defaultUserDataTemplate = `#cloud-config
hostname: {{.VM.Name}}
users:
  - name: cloud-user
    groups: sudo
    shell: /bin/bash
    sudo: ['ALL=(ALL) NOPASSWD:ALL']
    ssh_authorized_keys:
      {{- if .VM.CloudInit.SSHKeys }}
      {{- range .VM.CloudInit.SSHKeys }}
      - {{.}}
      {{- end }}
      {{- end }}
packages:
  - qemu-guest-agent
  - cloud-utils
  - cloud-init
package_update: true
package_upgrade: true
runcmd:
  - systemctl enable qemu-guest-agent
  - systemctl start qemu-guest-agent
`

const defaultMetaDataTemplate = `instance-id: {{.InstanceID}}
local-hostname: {{.Hostname}}
`

const defaultNetworkConfigTemplate = `version: 2
ethernets:
  ens3:
    dhcp4: true
    dhcp6: false
`