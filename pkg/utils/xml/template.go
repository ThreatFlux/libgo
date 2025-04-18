package xml

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"text/template"
)

// TemplateLoader handles loading and rendering XML templates
type TemplateLoader struct {
	TemplateDir string
	templates   map[string]*template.Template
	mu          sync.RWMutex
}

// NewTemplateLoader creates a template loader
func NewTemplateLoader(templateDir string) (*TemplateLoader, error) {
	// Ensure the template directory exists
	info, err := os.Stat(templateDir)
	if err != nil {
		return nil, fmt.Errorf("template directory error: %w", err)
	}

	if !info.IsDir() {
		return nil, fmt.Errorf("template path is not a directory: %s", templateDir)
	}

	loader := &TemplateLoader{
		TemplateDir: templateDir,
		templates:   make(map[string]*template.Template),
	}

	// Pre-load templates if they exist
	err = loader.loadTemplates()
	if err != nil {
		return nil, fmt.Errorf("loading templates: %w", err)
	}

	return loader, nil
}

// loadTemplates loads all template files from the template directory
func (l *TemplateLoader) loadTemplates() error {
	l.mu.Lock()
	defer l.mu.Unlock()

	// Get all template files
	files, err := os.ReadDir(l.TemplateDir)
	if err != nil {
		return fmt.Errorf("reading template directory: %w", err)
	}

	// Process each template file
	for _, file := range files {
		if file.IsDir() {
			continue
		}

		// Only load files with .tmpl extension
		if filepath.Ext(file.Name()) != ".tmpl" {
			continue
		}

		// Parse the template
		tmplPath := filepath.Join(l.TemplateDir, file.Name())
		tmpl, err := template.ParseFiles(tmplPath)
		if err != nil {
			return fmt.Errorf("parsing template %s: %w", file.Name(), err)
		}

		// Store in map
		l.templates[file.Name()] = tmpl
	}

	return nil
}

// LoadTemplate loads a template from the template directory
func (l *TemplateLoader) LoadTemplate(templateName string) (*template.Template, error) {
	l.mu.RLock()
	tmpl, exists := l.templates[templateName]
	l.mu.RUnlock()

	if exists {
		return tmpl, nil
	}

	// Template is not cached, load it
	templatePath := filepath.Join(l.TemplateDir, templateName)
	if _, err := os.Stat(templatePath); os.IsNotExist(err) {
		return nil, fmt.Errorf("template does not exist: %s", templatePath)
	}

	// Load the template
	tmpl, err := template.ParseFiles(templatePath)
	if err != nil {
		return nil, fmt.Errorf("failed to parse template %s: %w", templateName, err)
	}

	// Cache the template
	l.mu.Lock()
	l.templates[templateName] = tmpl
	l.mu.Unlock()

	return tmpl, nil
}

// RenderTemplate renders a template with the given data
func (l *TemplateLoader) RenderTemplate(templateName string, data interface{}) (string, error) {
	tmpl, err := l.LoadTemplate(templateName)
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("failed to render template %s: %w", templateName, err)
	}

	return buf.String(), nil
}

// ClearCache clears the template cache
func (l *TemplateLoader) ClearCache() {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.templates = make(map[string]*template.Template)
}

// GetTemplatePath returns the path to the template directory
func (l *TemplateLoader) GetTemplatePath() string {
	return l.TemplateDir
}
