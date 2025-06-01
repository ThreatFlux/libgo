package domain

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/threatflux/libgo/internal/models/vm"
	"github.com/threatflux/libgo/pkg/logger"
	xmlutils "github.com/threatflux/libgo/pkg/utils/xml"
)

// Mock logger for testing
type mockLogger struct {
	mock.Mock
}

func (m *mockLogger) Debug(msg string, fields ...logger.Field) {
	m.Called(msg, fields)
}

func (m *mockLogger) Info(msg string, fields ...logger.Field) {
	m.Called(msg, fields)
}

func (m *mockLogger) Warn(msg string, fields ...logger.Field) {
	m.Called(msg, fields)
}

func (m *mockLogger) Error(msg string, fields ...logger.Field) {
	m.Called(msg, fields)
}

func (m *mockLogger) Fatal(msg string, fields ...logger.Field) {
	m.Called(msg, fields)
}

func (m *mockLogger) WithFields(fields ...logger.Field) logger.Logger {
	args := m.Called(fields)
	return args.Get(0).(logger.Logger)
}

func (m *mockLogger) WithError(err error) logger.Logger {
	args := m.Called(err)
	return args.Get(0).(logger.Logger)
}

func (m *mockLogger) Sync() error {
	args := m.Called()
	return args.Error(0)
}

func TestTemplateXMLBuilder_BuildDomainXML(t *testing.T) {
	// Create a temporary directory for templates
	tmpDir := t.TempDir()

	// Create test domain template
	domainTemplate := `<domain type='kvm'>
  <n>{{.Name}}</n>
  <uuid>{{.UUID}}</uuid>
  <memory unit='KiB'>{{.Memory.KiB}}</memory>
  <vcpu placement='static'>{{.CPU.Count}}</vcpu>
  <cpu mode='host-model'></cpu>
  <devices>
    {{range .Disks}}
    <disk type='{{.Type}}' device='disk'>
      <driver name='qemu' type='{{.Format}}'/>
      <source {{.SourceAttr}}='{{.Source}}'/>
      <target dev='{{.Device}}' bus='{{.Bus}}'/>
    </disk>
    {{end}}
    {{range .Networks}}
    <interface type='{{.Type}}'>
      <source {{.SourceAttr}}='{{.Source}}'/>
      {{if .MacAddress}}<mac address='{{.MacAddress}}'/>{{end}}
      <model type='{{.Model}}'/>
    </interface>
    {{end}}
  </devices>
</domain>`

	// Write the template to the temporary directory
	templatePath := filepath.Join(tmpDir, "domain.xml.tmpl")
	if err := os.WriteFile(templatePath, []byte(domainTemplate), 0644); err != nil {
		t.Fatalf("Failed to write test template: %v", err)
	}

	// Create template loader
	templateLoader, err := xmlutils.NewTemplateLoader(tmpDir)
	if err != nil {
		t.Fatalf("Failed to create template loader: %v", err)
	}

	// Create mock logger
	mockLog := new(mockLogger)
	mockLog.On("Debug", mock.Anything, mock.Anything).Return()

	// Create XML builder
	builder := NewTemplateXMLBuilder(templateLoader, mockLog)

	// Test VM params
	params := vm.VMParams{
		Name: "test-vm",
		CPU: vm.CPUParams{
			Count: 2,
		},
		Memory: vm.MemoryParams{
			SizeBytes: 2 * 1024 * 1024 * 1024, // 2GB
		},
		Disk: vm.DiskParams{
			Format:      "qcow2",
			SourceImage: "/var/lib/libvirt/images/test-vm.qcow2",
		},
		Network: vm.NetParams{
			Type:   "bridge",
			Source: "virbr0",
			Model:  "virtio",
		},
	}

	// Generate XML
	xml, err := builder.BuildDomainXML(params)
	if err != nil {
		t.Fatalf("BuildDomainXML failed: %v", err)
	}

	// Verify the XML contains expected elements
	assert.Contains(t, xml, "<n>test-vm</n>")
	assert.Contains(t, xml, "<memory unit='KiB'>2097152</memory>") // 2GB in KiB
	assert.Contains(t, xml, "<vcpu placement='static'>2</vcpu>")
	assert.Contains(t, xml, `<disk type='file' device='disk'>`)
	assert.Contains(t, xml, `<driver name='qemu' type='qcow2'/>`)
	assert.Contains(t, xml, `<source file='/var/lib/libvirt/images/test-vm.qcow2'/>`)
	assert.Contains(t, xml, `<interface type='bridge'>`)
	assert.Contains(t, xml, `<source bridge='virbr0'/>`)

	// UUID should be generated
	assert.Regexp(t, `<uuid>[a-f0-9]{8}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{12}</uuid>`, xml)
}

func TestTemplateXMLBuilder_BuildDomainXML_AdvancedOptions(t *testing.T) {
	// Create a temporary directory for templates
	tmpDir := t.TempDir()

	// Create test domain template (more complex version)
	domainTemplate := `<domain type='kvm'>
  <n>{{.Name}}</n>
  <uuid>{{.UUID}}</uuid>
  <memory unit='KiB'>{{.Memory.KiB}}</memory>
  <vcpu placement='static'>{{.CPU.Count}}</vcpu>
  <cpu mode='{{if .CPU.Model}}custom{{else}}host-model{{end}}'{{if .CPU.Model}} match='exact'{{end}}>
    {{if .CPU.Model}}
    <model>{{.CPU.Model}}</model>
    {{end}}
    {{if and .CPU.Cores .CPU.Threads .CPU.Sockets}}
    <topology sockets='{{.CPU.Sockets}}' cores='{{.CPU.Cores}}' threads='{{.CPU.Threads}}'/>
    {{end}}
  </cpu>
  <devices>
    {{range .Disks}}
    <disk type='{{.Type}}' device='disk'>
      <driver name='qemu' type='{{.Format}}'/>
      <source {{.SourceAttr}}='{{.Source}}'/>
      <target dev='{{.Device}}' bus='{{.Bus}}'/>
      {{if .Bootable}}<boot order='1'/>{{end}}
      {{if .ReadOnly}}<readonly/>{{end}}
      {{if .Shareable}}<shareable/>{{end}}
    </disk>
    {{end}}
    {{if .CloudInitISO}}
    <disk type='file' device='cdrom'>
      <driver name='qemu' type='raw'/>
      <source file='{{.CloudInitISO}}'/>
      <target dev='sdb' bus='sata'/>
      <readonly/>
    </disk>
    {{end}}
    {{range .Networks}}
    <interface type='{{.Type}}'>
      <source {{.SourceAttr}}='{{.Source}}'/>
      {{if .MacAddress}}<mac address='{{.MacAddress}}'/>{{end}}
      <model type='{{.Model}}'/>
    </interface>
    {{end}}
  </devices>
</domain>`

	// Write the template to the temporary directory
	templatePath := filepath.Join(tmpDir, "domain.xml.tmpl")
	if err := os.WriteFile(templatePath, []byte(domainTemplate), 0644); err != nil {
		t.Fatalf("Failed to write test template: %v", err)
	}

	// Create template loader
	templateLoader, err := xmlutils.NewTemplateLoader(tmpDir)
	if err != nil {
		t.Fatalf("Failed to create template loader: %v", err)
	}

	// Create mock logger
	mockLog := new(mockLogger)
	mockLog.On("Debug", mock.Anything, mock.Anything).Return()

	// Create XML builder
	builder := NewTemplateXMLBuilder(templateLoader, mockLog)

	// Test VM params with advanced options
	params := vm.VMParams{
		Name: "advanced-vm",
		CPU: vm.CPUParams{
			Count:   4,
			Model:   "Haswell-noTSX",
			Socket:  1,
			Cores:   2,
			Threads: 2,
		},
		Memory: vm.MemoryParams{
			SizeBytes: 4 * 1024 * 1024 * 1024, // 4GB
		},
		Disk: vm.DiskParams{
			Format:      "qcow2",
			SourceImage: "/var/lib/libvirt/images/test-vm.qcow2",
		},
		Network: vm.NetParams{
			Type:       "network",
			Source:     "default",
			Model:      "e1000",
			MacAddress: "52:54:00:12:34:56",
		},
		CloudInit: vm.CloudInitConfig{
			UserData: "user data content",
			MetaData: "meta data content",
		},
	}

	// Generate XML
	xml, err := builder.BuildDomainXML(params)
	if err != nil {
		t.Fatalf("BuildDomainXML failed: %v", err)
	}

	// Verify advanced elements
	assert.Contains(t, xml, `<cpu mode='custom' match='exact'>`)
	assert.Contains(t, xml, `<model>Haswell-noTSX</model>`)
	assert.Contains(t, xml, `<topology sockets='1' cores='2' threads='2'/>`)
	assert.Contains(t, xml, `<mac address='52:54:00:12:34:56'/>`)
	assert.Contains(t, xml, `<model type='e1000'/>`)
	assert.Contains(t, xml, `<source network='default'/>`)
	assert.Contains(t, xml, `<source file='/tmp/libgo-cloudinit/advanced-vm-cloudinit.iso'/>`)
}

func TestTemplateXMLBuilder_GenerateCloudInitISOPath(t *testing.T) {
	// Create mock logger
	mockLog := new(mockLogger)

	// Create XML builder
	builder := &TemplateXMLBuilder{
		logger: mockLog,
	}

	// Test with default location
	path := builder.GenerateCloudInitISOPath("test-vm", "")
	expectedPath := "/tmp/libgo-cloudinit/test-vm-cloudinit.iso"
	assert.Equal(t, expectedPath, path)

	// Test with custom directory
	customDir := "/tmp/cloudinit"
	path = builder.GenerateCloudInitISOPath("test-vm", customDir)
	expectedPath = filepath.Join(customDir, "test-vm-cloudinit.iso")
	assert.Equal(t, expectedPath, path)
}
