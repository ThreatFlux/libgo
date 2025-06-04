package network

import (
	"context"

	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/network"
	"github.com/threatflux/libgo/internal/docker"
	"github.com/threatflux/libgo/pkg/logger"
)

// Service provides Docker network management functionality
type Service interface {
	// Network operations
	Create(ctx context.Context, name string, options network.CreateOptions) (network.CreateResponse, error)
	Remove(ctx context.Context, networkID string) error
	Connect(ctx context.Context, networkID, containerID string, config *network.EndpointSettings) error
	Disconnect(ctx context.Context, networkID, containerID string, force bool) error

	// Network inspection
	Inspect(ctx context.Context, networkID string, options network.InspectOptions) (*network.Inspect, error)
	List(ctx context.Context, options network.ListOptions) ([]network.Summary, error)

	// Network prune
	Prune(ctx context.Context, pruneFilter filters.Args) (network.PruneReport, error)

	// Utility methods
	Exists(ctx context.Context, networkID string) (bool, error)
	GetByName(ctx context.Context, name string) (*network.Inspect, error)
}

// serviceImpl implements the Service interface
type serviceImpl struct {
	manager docker.Manager
	logger  logger.Logger
}

// NewService creates a new network service
func NewService(manager docker.Manager, logger logger.Logger) Service {
	return &serviceImpl{
		manager: manager,
		logger:  logger,
	}
}
