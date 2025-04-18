package cloudinit

import (
	"context"

	"github.com/wroersma/libgo/internal/models/vm"
)

// Manager defines the interface for cloud-init operations
type Manager interface {
	// GenerateUserData generates cloud-init user-data content
	GenerateUserData(params vm.VMParams) (string, error)

	// GenerateMetaData generates cloud-init meta-data content
	GenerateMetaData(params vm.VMParams) (string, error)

	// GenerateNetworkConfig generates cloud-init network-config content
	GenerateNetworkConfig(params vm.VMParams) (string, error)

	// GenerateISO creates a cloud-init ISO image with the provided configuration
	GenerateISO(ctx context.Context, config CloudInitConfig, outputPath string) error
}
