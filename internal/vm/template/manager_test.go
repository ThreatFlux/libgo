package template

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/threatflux/libgo/internal/models/vm"
	mocks_logger "github.com/threatflux/libgo/test/mocks/logger"
	"go.uber.org/mock/gomock"
)

func TestTemplateManager_ListTemplates(t *testing.T) {
	// Create a temporary directory for test templates
	tempDir, err := os.MkdirTemp("", "vm-templates-*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Create test templates
	templates := map[string]vm.VMParams{
		"small": {
			Name: "small",
			CPU: vm.CPUParams{
				Count: 1,
			},
			Memory: vm.MemoryParams{
				SizeBytes: 1 * 1024 * 1024 * 1024, // 1 GB
			},
			Disk: vm.DiskParams{
				SizeBytes: 10 * 1024 * 1024 * 1024, // 10 GB
				Format:    "qcow2",
			},
		},
		"medium": {
			Name: "medium",
			CPU: vm.CPUParams{
				Count: 2,
			},
			Memory: vm.MemoryParams{
				SizeBytes: 2 * 1024 * 1024 * 1024, // 2 GB
			},
			Disk: vm.DiskParams{
				SizeBytes: 20 * 1024 * 1024 * 1024, // 20 GB
				Format:    "qcow2",
			},
		},
	}

	// Write test templates to files
	for name, template := range templates {
		templateData, err := json.Marshal(template)
		require.NoError(t, err)

		err = os.WriteFile(filepath.Join(tempDir, name+".json"), templateData, 0644)
		require.NoError(t, err)
	}

	// Create template manager
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLogger := mocks_logger.NewMockLogger(ctrl)
	mockLogger.EXPECT().Debug(gomock.Any(), gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Info(gomock.Any(), gomock.Any()).AnyTimes()

	manager, err := NewTemplateManager(tempDir, mockLogger)
	require.NoError(t, err)

	// Test ListTemplates
	templateNames, err := manager.ListTemplates()
	require.NoError(t, err)
	assert.ElementsMatch(t, []string{"small", "medium"}, templateNames)
}

func TestTemplateManager_GetTemplate(t *testing.T) {
	// Create a temporary directory for test templates
	tempDir, err := os.MkdirTemp("", "vm-templates-*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Create test template
	template := vm.VMParams{
		Name: "test-template",
		CPU: vm.CPUParams{
			Count: 2,
		},
		Memory: vm.MemoryParams{
			SizeBytes: 2 * 1024 * 1024 * 1024, // 2 GB
		},
		Disk: vm.DiskParams{
			SizeBytes: 20 * 1024 * 1024 * 1024, // 20 GB
			Format:    "qcow2",
		},
	}

	// Write test template to file
	templateData, err := json.Marshal(template)
	require.NoError(t, err)

	err = os.WriteFile(filepath.Join(tempDir, "test-template.json"), templateData, 0644)
	require.NoError(t, err)

	// Create template manager
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLogger := mocks_logger.NewMockLogger(ctrl)
	mockLogger.EXPECT().Debug(gomock.Any(), gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Info(gomock.Any(), gomock.Any()).AnyTimes()

	manager, err := NewTemplateManager(tempDir, mockLogger)
	require.NoError(t, err)

	// Test GetTemplate
	retrievedTemplate, err := manager.GetTemplate("test-template")
	require.NoError(t, err)
	assert.Equal(t, template.Name, retrievedTemplate.Name)
	assert.Equal(t, template.CPU.Count, retrievedTemplate.CPU.Count)
	assert.Equal(t, template.Memory.SizeBytes, retrievedTemplate.Memory.SizeBytes)
	assert.Equal(t, template.Disk.SizeBytes, retrievedTemplate.Disk.SizeBytes)
	assert.Equal(t, template.Disk.Format, retrievedTemplate.Disk.Format)

	// Test GetTemplate with non-existent template
	_, err = manager.GetTemplate("non-existent")
	assert.Error(t, err)
}

func TestTemplateManager_ApplyTemplate(t *testing.T) {
	// Create a temporary directory for test templates
	tempDir, err := os.MkdirTemp("", "vm-templates-*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Create test template
	template := vm.VMParams{
		Name: "test-template",
		CPU: vm.CPUParams{
			Count: 2,
		},
		Memory: vm.MemoryParams{
			SizeBytes: 2 * 1024 * 1024 * 1024, // 2 GB
		},
		Disk: vm.DiskParams{
			SizeBytes: 20 * 1024 * 1024 * 1024, // 20 GB
			Format:    "qcow2",
		},
		Network: vm.NetParams{
			Type:   "network",
			Source: "default",
		},
	}

	// Write test template to file
	templateData, err := json.Marshal(template)
	require.NoError(t, err)

	err = os.WriteFile(filepath.Join(tempDir, "test-template.json"), templateData, 0644)
	require.NoError(t, err)

	// Create template manager
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLogger := mocks_logger.NewMockLogger(ctrl)
	mockLogger.EXPECT().Debug(gomock.Any(), gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Info(gomock.Any(), gomock.Any()).AnyTimes()

	manager, err := NewTemplateManager(tempDir, mockLogger)
	require.NoError(t, err)

	// Test applying template to empty params
	emptyParams := &vm.VMParams{
		Name: "new-vm",
	}
	err = manager.ApplyTemplate("test-template", emptyParams)
	require.NoError(t, err)
	assert.Equal(t, "new-vm", emptyParams.Name)
	assert.Equal(t, 2, emptyParams.CPU.Count)
	assert.Equal(t, uint64(2*1024*1024*1024), emptyParams.Memory.SizeBytes)
	assert.Equal(t, uint64(20*1024*1024*1024), emptyParams.Disk.SizeBytes)
	assert.Equal(t, "qcow2", string(emptyParams.Disk.Format))
	assert.Equal(t, "network", string(emptyParams.Network.Type))
	assert.Equal(t, "default", emptyParams.Network.Source)

	// Test applying template to partial params
	partialParams := &vm.VMParams{
		Name: "partial-vm",
		CPU: vm.CPUParams{
			Count: 4, // Different from template
		},
		Disk: vm.DiskParams{
			SizeBytes: 40 * 1024 * 1024 * 1024, // Different from template
		},
	}
	err = manager.ApplyTemplate("test-template", partialParams)
	require.NoError(t, err)
	assert.Equal(t, "partial-vm", partialParams.Name)
	assert.Equal(t, 4, partialParams.CPU.Count)                               // Should keep the original value
	assert.Equal(t, uint64(2*1024*1024*1024), partialParams.Memory.SizeBytes) // Should use template value
	assert.Equal(t, uint64(40*1024*1024*1024), partialParams.Disk.SizeBytes)  // Should keep the original value
	assert.Equal(t, "qcow2", string(partialParams.Disk.Format))               // Should use template value
	assert.Equal(t, "network", string(partialParams.Network.Type))            // Should use template value
	assert.Equal(t, "default", partialParams.Network.Source)                  // Should use template value
}
