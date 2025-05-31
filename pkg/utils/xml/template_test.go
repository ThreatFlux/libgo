package xml

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"testing"
)

func TestNewTemplateLoader(t *testing.T) {
	// Test with a valid directory
	tmpDir := t.TempDir()
	loader, err := NewTemplateLoader(tmpDir)
	if err != nil {
		t.Fatalf("NewTemplateLoader failed with valid directory: %v", err)
	}
	if loader.TemplateDir != tmpDir {
		t.Errorf("Expected template dir to be %s, got %s", tmpDir, loader.TemplateDir)
	}

	// Test with a non-existent directory
	_, err = NewTemplateLoader("/nonexistent/directory")
	if err == nil {
		t.Errorf("Expected error for non-existent directory, got nil")
	}
}

func TestLoadTemplate(t *testing.T) {
	// Create a temporary directory for templates
	tmpDir := t.TempDir()

	// Create a test template file
	templateName := "test.xml.tmpl"
	templateContent := `<test>{{ .Value }}</test>`
	templatePath := filepath.Join(tmpDir, templateName)
	if err := os.WriteFile(templatePath, []byte(templateContent), 0644); err != nil {
		t.Fatalf("Failed to create test template: %v", err)
	}

	// Create template loader
	loader, err := NewTemplateLoader(tmpDir)
	if err != nil {
		t.Fatalf("Failed to create template loader: %v", err)
	}

	// Load the template
	tmpl, err := loader.LoadTemplate(templateName)
	if err != nil {
		t.Fatalf("LoadTemplate failed: %v", err)
	}
	if tmpl == nil {
		t.Fatalf("LoadTemplate returned nil template")
	}

	// Test loading a non-existent template
	_, err = loader.LoadTemplate("nonexistent.tmpl")
	if err == nil {
		t.Errorf("Expected error for non-existent template, got nil")
	}
}

func TestRenderTemplate(t *testing.T) {
	// Create a temporary directory for templates
	tmpDir := t.TempDir()

	// Create a test template file
	templateName := "test.xml.tmpl"
	templateContent := `<test>{{ .Value }}</test>`
	templatePath := filepath.Join(tmpDir, templateName)
	if err := os.WriteFile(templatePath, []byte(templateContent), 0644); err != nil {
		t.Fatalf("Failed to create test template: %v", err)
	}

	// Create template loader
	loader, err := NewTemplateLoader(tmpDir)
	if err != nil {
		t.Fatalf("Failed to create template loader: %v", err)
	}

	// Render the template
	data := struct {
		Value string
	}{
		Value: "test-value",
	}

	result, err := loader.RenderTemplate(templateName, data)
	if err != nil {
		t.Fatalf("RenderTemplate failed: %v", err)
	}

	expectedResult := `<test>test-value</test>`
	if result != expectedResult {
		t.Errorf("Expected result to be '%s', got '%s'", expectedResult, result)
	}

	// Test rendering with a non-existent template
	_, err = loader.RenderTemplate("nonexistent.tmpl", data)
	if err == nil {
		t.Errorf("Expected error for non-existent template, got nil")
	}

	// Test rendering with invalid template data
	invalidTemplate := "invalid.xml.tmpl"
	invalidContent := `<test>{{ .MissingField.SubField }}</test>`
	invalidPath := filepath.Join(tmpDir, invalidTemplate)
	if err := os.WriteFile(invalidPath, []byte(invalidContent), 0644); err != nil {
		t.Fatalf("Failed to create invalid template: %v", err)
	}

	_, err = loader.RenderTemplate(invalidTemplate, data)
	if err == nil {
		t.Errorf("Expected error for invalid template data, got nil")
	}
}

func TestCacheAndClearCache(t *testing.T) {
	// Create a temporary directory for templates
	tmpDir := t.TempDir()

	// Create a test template file
	templateName := "test.xml.tmpl"
	templateContent := `<test>{{ .Value }}</test>`
	templatePath := filepath.Join(tmpDir, templateName)
	if err := os.WriteFile(templatePath, []byte(templateContent), 0644); err != nil {
		t.Fatalf("Failed to create test template: %v", err)
	}

	// Create template loader
	loader, err := NewTemplateLoader(tmpDir)
	if err != nil {
		t.Fatalf("Failed to create template loader: %v", err)
	}

	// Load the template to cache it
	_, err = loader.LoadTemplate(templateName)
	if err != nil {
		t.Fatalf("Failed to load template: %v", err)
	}

	// Modify the template file to verify caching
	newContent := `<test>modified-{{ .Value }}</test>`
	if err := os.WriteFile(templatePath, []byte(newContent), 0644); err != nil {
		t.Fatalf("Failed to modify template: %v", err)
	}

	// Render using cached template
	data := struct {
		Value string
	}{
		Value: "test-value",
	}

	result, err := loader.RenderTemplate(templateName, data)
	if err != nil {
		t.Fatalf("RenderTemplate failed: %v", err)
	}

	// Should use cached version (unmodified)
	expectedResult := `<test>test-value</test>`
	if result != expectedResult {
		t.Errorf("Expected cached result to be '%s', got '%s'", expectedResult, result)
	}

	// Clear cache
	loader.ClearCache()

	// Render again - should use the modified version
	result, err = loader.RenderTemplate(templateName, data)
	if err != nil {
		t.Fatalf("RenderTemplate failed after cache clear: %v", err)
	}

	expectedModifiedResult := `<test>modified-test-value</test>`
	if result != expectedModifiedResult {
		t.Errorf("Expected modified result to be '%s', got '%s'", expectedModifiedResult, result)
	}
}

func TestTemplateConcurrency(t *testing.T) {
	// Create a temporary directory for templates
	tmpDir := t.TempDir()

	// Create multiple test template files
	for i := 0; i < 5; i++ {
		templateName := fmt.Sprintf("test%d.xml.tmpl", i)
		templateContent := fmt.Sprintf(`<test id="%d">{{ .Value }}</test>`, i)
		templatePath := filepath.Join(tmpDir, templateName)
		if err := os.WriteFile(templatePath, []byte(templateContent), 0644); err != nil {
			t.Fatalf("Failed to create test template: %v", err)
		}
	}

	// Create template loader
	loader, err := NewTemplateLoader(tmpDir)
	if err != nil {
		t.Fatalf("Failed to create template loader: %v", err)
	}

	// Test concurrent access
	var wg sync.WaitGroup
	errors := make(chan error, 100)

	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < 5; j++ {
				// Load and render templates concurrently
				tmplIdx := j % 5
				templateName := fmt.Sprintf("test%d.xml.tmpl", tmplIdx)
				data := struct {
					Value string
				}{
					Value: fmt.Sprintf("value-%d-%d", id, j),
				}

				result, err := loader.RenderTemplate(templateName, data)
				if err != nil {
					errors <- fmt.Errorf("render error: %w", err)
					return
				}

				expected := fmt.Sprintf(`<test id="%d">value-%d-%d</test>`, tmplIdx, id, j)
				if result != expected {
					errors <- fmt.Errorf("expected '%s', got '%s'", expected, result)
					return
				}
			}
		}(i)
	}

	// Wait for all goroutines
	wg.Wait()
	close(errors)

	// Check for errors
	for err := range errors {
		t.Errorf("Concurrent test error: %v", err)
	}
}
