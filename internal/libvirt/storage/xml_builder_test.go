package storage

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
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

func TestTemplateXMLBuilder_BuildStoragePoolXML(t *testing.T) {
	// Create a temporary directory for templates
	tmpDir := t.TempDir()

	// Create test storage pool template
	poolTemplate := `<pool type='dir'>
  <n>{{.Name}}</n>
  <target>
    <path>{{.Path}}</path>
    <permissions>
      <mode>0755</mode>
      <owner>0</owner>
      <group>0</group>
    </permissions>
  </target>
</pool>`

	// Write the template to the temporary directory
	templatePath := filepath.Join(tmpDir, "storage_pool.xml.tmpl")
	if err := os.WriteFile(templatePath, []byte(poolTemplate), 0644); err != nil {
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

	// Test pool XML generation
	poolName := "test-pool"
	poolPath := "/var/lib/libvirt/storage/test-pool"

	xml, err := builder.BuildStoragePoolXML(poolName, poolPath)
	if err != nil {
		t.Fatalf("BuildStoragePoolXML failed: %v", err)
	}

	// Verify the XML contains expected elements
	assert.Contains(t, xml, "<n>test-pool</n>")
	assert.Contains(t, xml, "<path>/var/lib/libvirt/storage/test-pool</path>")
	assert.Contains(t, xml, "<mode>0755</mode>")
}

func TestTemplateXMLBuilder_BuildStorageVolumeXML(t *testing.T) {
	// Create a temporary directory for templates
	tmpDir := t.TempDir()

	// Create test storage volume template
	volumeTemplate := `<volume>
  <n>{{.Name}}</n>
  <allocation>0</allocation>
  <capacity unit="bytes">{{.CapacityBytes}}</capacity>
  <target>
    <format type="{{.Format}}"/>
    <permissions>
      <mode>0644</mode>
      <owner>0</owner>
      <group>0</group>
    </permissions>
  </target>
</volume>`

	// Write the template to the temporary directory
	templatePath := filepath.Join(tmpDir, "storage_volume.xml.tmpl")
	if err := os.WriteFile(templatePath, []byte(volumeTemplate), 0644); err != nil {
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

	// Test cases
	testCases := []struct {
		name          string
		volName       string
		capacityBytes uint64
		format        string
		expectFormat  string
	}{
		{
			name:          "Basic QCOW2 Volume",
			volName:       "test-vol",
			capacityBytes: 10 * 1024 * 1024 * 1024, // 10GB
			format:        "qcow2",
			expectFormat:  "qcow2",
		},
		{
			name:          "Raw Format Volume",
			volName:       "test-raw-vol",
			capacityBytes: 5 * 1024 * 1024 * 1024, // 5GB
			format:        "raw",
			expectFormat:  "raw",
		},
		{
			name:          "Default Format",
			volName:       "default-format-vol",
			capacityBytes: 1 * 1024 * 1024 * 1024, // 1GB
			format:        "",
			expectFormat:  "qcow2", // Default format
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			xml, err := builder.BuildStorageVolumeXML(tc.volName, tc.capacityBytes, tc.format)
			if err != nil {
				t.Fatalf("BuildStorageVolumeXML failed: %v", err)
			}

			// Verify the XML contains expected elements
			assert.Contains(t, xml, "<n>"+tc.volName+"</n>")
			assert.Contains(t, xml, fmt.Sprintf("<capacity unit=\"bytes\">%d</capacity>", tc.capacityBytes))
			assert.Contains(t, xml, fmt.Sprintf("<format type=\"%s\"/>", tc.expectFormat))
		})
	}
}
