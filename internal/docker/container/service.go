package container

import (
	"context"
	"io"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/threatflux/libgo/internal/docker"
	"github.com/threatflux/libgo/pkg/logger"
)

// Service provides Docker container management functionality.
type Service interface {
	// Container lifecycle
	Create(ctx context.Context, config *container.Config, hostConfig *container.HostConfig, name string) (string, error)
	Start(ctx context.Context, containerID string) error
	Stop(ctx context.Context, containerID string, timeout *int) error
	Restart(ctx context.Context, containerID string, timeout *int) error
	Kill(ctx context.Context, containerID string, signal string) error
	Remove(ctx context.Context, containerID string, force bool) error
	Pause(ctx context.Context, containerID string) error
	Unpause(ctx context.Context, containerID string) error

	// Container inspection
	Inspect(ctx context.Context, containerID string) (*container.InspectResponse, error)
	List(ctx context.Context, options container.ListOptions) ([]container.Summary, error)
	Exists(ctx context.Context, containerID string) (bool, error)

	// Container logs
	Logs(ctx context.Context, containerID string, options container.LogsOptions) (io.ReadCloser, error)

	// Container stats
	Stats(ctx context.Context, containerID string, stream bool) (container.StatsResponseReader, error)

	// Container exec
	ExecCreate(ctx context.Context, containerID string, config container.ExecOptions) (container.ExecCreateResponse, error)
	ExecStart(ctx context.Context, execID string, config container.ExecStartOptions) error
	ExecInspect(ctx context.Context, execID string) (container.ExecInspect, error)

	// Container file operations
	CopyToContainer(ctx context.Context, containerID, dstPath string, content io.Reader, options container.CopyToContainerOptions) error
	CopyFromContainer(ctx context.Context, containerID, srcPath string) (io.ReadCloser, container.PathStat, error)

	// Container wait
	Wait(ctx context.Context, containerID string, condition container.WaitCondition) (<-chan container.WaitResponse, <-chan error)

	// Container attach
	Attach(ctx context.Context, containerID string, options container.AttachOptions) (types.HijackedResponse, error)

	// Container updates
	Update(ctx context.Context, containerID string, updateConfig container.UpdateConfig) (container.UpdateResponse, error)

	// Container rename
	Rename(ctx context.Context, containerID string, newName string) error

	// Container resize
	Resize(ctx context.Context, containerID string, options container.ResizeOptions) error

	// Container prune
	Prune(ctx context.Context, pruneFilters container.ListOptions) (container.PruneReport, error)
}

// serviceImpl implements the Service interface.
type serviceImpl struct {
	manager docker.Manager
	logger  logger.Logger
}

// NewService creates a new container service.
func NewService(manager docker.Manager, logger logger.Logger) Service {
	return &serviceImpl{
		manager: manager,
		logger:  logger,
	}
}
