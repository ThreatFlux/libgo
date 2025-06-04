package volume

import (
	"context"

	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/volume"
	"github.com/threatflux/libgo/internal/docker"
	"github.com/threatflux/libgo/pkg/logger"
)

// Service provides Docker volume management functionality
type Service interface {
	// Volume operations
	Create(ctx context.Context, options volume.CreateOptions) (*volume.Volume, error)
	Remove(ctx context.Context, volumeID string, force bool) error

	// Volume inspection
	Inspect(ctx context.Context, volumeID string) (*volume.Volume, error)
	List(ctx context.Context, options volume.ListOptions) (*volume.ListResponse, error)

	// Volume prune
	Prune(ctx context.Context, pruneFilter filters.Args) (volume.PruneReport, error)

	// Utility methods
	Exists(ctx context.Context, volumeID string) (bool, error)
	GetByName(ctx context.Context, name string) (*volume.Volume, error)
}

// serviceImpl implements the Service interface
type serviceImpl struct {
	manager docker.Manager
	logger  logger.Logger
}

// NewService creates a new volume service
func NewService(manager docker.Manager, logger logger.Logger) Service {
	return &serviceImpl{
		manager: manager,
		logger:  logger,
	}
}
