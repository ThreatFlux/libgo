package cloudinit

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/threatflux/libgo/pkg/logger"
)

// CloudInitConfig holds cloud-init configuration data
type CloudInitConfig struct {
	UserData      string
	MetaData      string
	NetworkConfig string
}

// GenerateISO implements Manager.GenerateISO
// Creates a cloud-init ISO image with user-data, meta-data, and network-config
func (g *CloudInitGenerator) GenerateISO(ctx context.Context, config CloudInitConfig, outputPath string) error {
	g.logger.Debug("Generating cloud-init ISO",
		logger.String("outputPath", outputPath))

	// Create a temporary directory to store the files
	tmpDir, err := os.MkdirTemp("", "cloud-init-")
	if err != nil {
		return fmt.Errorf("creating temp directory: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	// Write config files to temp directory
	files := map[string]string{
		"user-data":      config.UserData,
		"meta-data":      config.MetaData,
		"network-config": config.NetworkConfig,
	}

	for filename, content := range files {
		filePath := filepath.Join(tmpDir, filename)
		err := os.WriteFile(filePath, []byte(content), 0644)
		if err != nil {
			return fmt.Errorf("writing %s: %w", filename, err)
		}
	}

	// Create directory for output if it doesn't exist
	outDir := filepath.Dir(outputPath)
	if err := os.MkdirAll(outDir, 0755); err != nil {
		return fmt.Errorf("creating output directory: %w", err)
	}

	// Create ISO using genisoimage or mkisofs
	// Prefer genisoimage if available, fall back to mkisofs
	var cmdName string
	var args []string

	// Try to find genisoimage or mkisofs
	if _, err := exec.LookPath("genisoimage"); err == nil {
		cmdName = "genisoimage"
		args = []string{
			"-output", outputPath,
			"-volid", "cidata",
			"-joliet",
			"-rock",
			tmpDir,
		}
	} else if _, err := exec.LookPath("mkisofs"); err == nil {
		cmdName = "mkisofs"
		args = []string{
			"-output", outputPath,
			"-volid", "cidata",
			"-joliet",
			"-rock",
			tmpDir,
		}
	} else {
		return fmt.Errorf("neither genisoimage nor mkisofs found")
	}

	cmd := exec.CommandContext(ctx, cmdName, args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("generating ISO: %w, output: %s", err, string(output))
	}

	g.logger.Info("Generated cloud-init ISO",
		logger.String("outputPath", outputPath))

	return nil
}
