package image

import (
	"context"
	"io"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/api/types/registry"
	"github.com/docker/docker/client"
	"github.com/threatflux/libgo/internal/docker"
	"github.com/threatflux/libgo/pkg/logger"
)

// Service provides Docker image management functionality
type Service interface {
	// Image operations
	Pull(ctx context.Context, refStr string, options image.PullOptions) (io.ReadCloser, error)
	Push(ctx context.Context, refStr string, options image.PushOptions) (io.ReadCloser, error)
	Build(ctx context.Context, buildContext io.Reader, options types.ImageBuildOptions) (types.ImageBuildResponse, error)
	Tag(ctx context.Context, source, target string) error
	Remove(ctx context.Context, imageID string, options image.RemoveOptions) ([]image.DeleteResponse, error)

	// Image inspection
	Inspect(ctx context.Context, imageID string) (*types.ImageInspect, error)
	List(ctx context.Context, options image.ListOptions) ([]image.Summary, error)
	History(ctx context.Context, imageID string) ([]image.HistoryResponseItem, error)

	// Image import/export
	Import(ctx context.Context, source image.ImportSource, ref string, options image.ImportOptions) (io.ReadCloser, error)
	Save(ctx context.Context, imageIDs []string) (io.ReadCloser, error)
	Load(ctx context.Context, input io.Reader, options ...client.ImageLoadOption) (image.LoadResponse, error)

	// Image search
	Search(ctx context.Context, term string, options registry.SearchOptions) ([]registry.SearchResult, error)

	// Image prune
	Prune(ctx context.Context, pruneFilter filters.Args) (image.PruneReport, error)

	// Utility methods
	Exists(ctx context.Context, imageID string) (bool, error)
	GetDigest(ctx context.Context, imageID string) (string, error)
}

// serviceImpl implements the Service interface
type serviceImpl struct {
	manager docker.Manager
	logger  logger.Logger
}

// NewService creates a new image service
func NewService(manager docker.Manager, logger logger.Logger) Service {
	return &serviceImpl{
		manager: manager,
		logger:  logger,
	}
}
