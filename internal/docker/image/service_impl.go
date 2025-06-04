package image

import (
	"context"
	"io"

	"github.com/docker/docker/api/types/build"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/api/types/registry"
	"github.com/docker/docker/client"
	"github.com/threatflux/libgo/internal/docker"
	"github.com/threatflux/libgo/pkg/logger"
)

// Ensure the docker package is used (this prevents import errors).
var _ docker.Manager

// Pull pulls an image from a registry.
func (s *serviceImpl) Pull(ctx context.Context, refStr string, options image.PullOptions) (io.ReadCloser, error) {
	client, err := s.manager.GetWithContext(ctx)
	if err != nil {
		s.logger.Error("Failed to get Docker client", logger.Error(err))
		return nil, err
	}

	reader, err := client.ImagePull(ctx, refStr, options)
	if err != nil {
		s.logger.Error("Failed to pull image",
			logger.String("image", refStr),
			logger.Error(err))
		return nil, err
	}

	s.logger.Info("Started image pull", logger.String("image", refStr))
	return reader, nil
}

// Push pushes an image to a registry.
func (s *serviceImpl) Push(ctx context.Context, refStr string, options image.PushOptions) (io.ReadCloser, error) {
	client, err := s.manager.GetWithContext(ctx)
	if err != nil {
		s.logger.Error("Failed to get Docker client", logger.Error(err))
		return nil, err
	}

	reader, err := client.ImagePush(ctx, refStr, options)
	if err != nil {
		s.logger.Error("Failed to push image",
			logger.String("image", refStr),
			logger.Error(err))
		return nil, err
	}

	s.logger.Info("Started image push", logger.String("image", refStr))
	return reader, nil
}

// Build builds an image from a Dockerfile.
func (s *serviceImpl) Build(ctx context.Context, buildContext io.Reader, options build.ImageBuildOptions) (build.ImageBuildResponse, error) {
	client, err := s.manager.GetWithContext(ctx)
	if err != nil {
		s.logger.Error("Failed to get Docker client", logger.Error(err))
		return build.ImageBuildResponse{}, err
	}

	response, err := client.ImageBuild(ctx, buildContext, options)
	if err != nil {
		s.logger.Error("Failed to build image", logger.Error(err))
		return build.ImageBuildResponse{}, err
	}

	s.logger.Info("Started image build")
	return response, nil
}

// Tag tags an image.
func (s *serviceImpl) Tag(ctx context.Context, source, target string) error {
	client, err := s.manager.GetWithContext(ctx)
	if err != nil {
		s.logger.Error("Failed to get Docker client", logger.Error(err))
		return err
	}

	err = client.ImageTag(ctx, source, target)
	if err != nil {
		s.logger.Error("Failed to tag image",
			logger.String("source", source),
			logger.String("target", target),
			logger.Error(err))
		return err
	}

	s.logger.Info("Tagged image",
		logger.String("source", source),
		logger.String("target", target))
	return nil
}

// Remove removes one or more images.
func (s *serviceImpl) Remove(ctx context.Context, imageID string, options image.RemoveOptions) ([]image.DeleteResponse, error) {
	client, err := s.manager.GetWithContext(ctx)
	if err != nil {
		s.logger.Error("Failed to get Docker client", logger.Error(err))
		return nil, err
	}

	deleteResponses, err := client.ImageRemove(ctx, imageID, options)
	if err != nil {
		s.logger.Error("Failed to remove image",
			logger.String("image_id", imageID),
			logger.Error(err))
		return nil, err
	}

	s.logger.Info("Removed image",
		logger.String("image_id", imageID),
		logger.Int("deleted_count", len(deleteResponses)))
	return deleteResponses, nil
}

// Inspect returns detailed information about an image.
func (s *serviceImpl) Inspect(ctx context.Context, imageID string) (*image.InspectResponse, error) {
	client, err := s.manager.GetWithContext(ctx)
	if err != nil {
		s.logger.Error("Failed to get Docker client", logger.Error(err))
		return nil, err
	}

	inspectResponse, err := client.ImageInspect(ctx, imageID)
	if err != nil {
		s.logger.Error("Failed to inspect image",
			logger.String("image_id", imageID),
			logger.Error(err))
		return nil, err
	}

	s.logger.Debug("Inspected image", logger.String("image_id", imageID))
	return &inspectResponse, nil
}

// List returns a list of images.
func (s *serviceImpl) List(ctx context.Context, options image.ListOptions) ([]image.Summary, error) {
	client, err := s.manager.GetWithContext(ctx)
	if err != nil {
		s.logger.Error("Failed to get Docker client", logger.Error(err))
		return nil, err
	}

	images, err := client.ImageList(ctx, options)
	if err != nil {
		s.logger.Error("Failed to list images", logger.Error(err))
		return nil, err
	}

	s.logger.Debug("Listed images", logger.Int("count", len(images)))
	return images, nil
}

// History returns the history of an image.
func (s *serviceImpl) History(ctx context.Context, imageID string) ([]image.HistoryResponseItem, error) {
	client, err := s.manager.GetWithContext(ctx)
	if err != nil {
		s.logger.Error("Failed to get Docker client", logger.Error(err))
		return nil, err
	}

	history, err := client.ImageHistory(ctx, imageID)
	if err != nil {
		s.logger.Error("Failed to get image history",
			logger.String("image_id", imageID),
			logger.Error(err))
		return nil, err
	}

	s.logger.Debug("Retrieved image history",
		logger.String("image_id", imageID),
		logger.Int("layers", len(history)))
	return history, nil
}

// Import imports an image.
func (s *serviceImpl) Import(ctx context.Context, source image.ImportSource, ref string, options image.ImportOptions) (io.ReadCloser, error) {
	client, err := s.manager.GetWithContext(ctx)
	if err != nil {
		s.logger.Error("Failed to get Docker client", logger.Error(err))
		return nil, err
	}

	reader, err := client.ImageImport(ctx, source, ref, options)
	if err != nil {
		s.logger.Error("Failed to import image",
			logger.String("ref", ref),
			logger.Error(err))
		return nil, err
	}

	s.logger.Info("Started image import", logger.String("ref", ref))
	return reader, nil
}

// Save saves images to a tar archive.
func (s *serviceImpl) Save(ctx context.Context, imageIDs []string) (io.ReadCloser, error) {
	client, err := s.manager.GetWithContext(ctx)
	if err != nil {
		s.logger.Error("Failed to get Docker client", logger.Error(err))
		return nil, err
	}

	reader, err := client.ImageSave(ctx, imageIDs)
	if err != nil {
		s.logger.Error("Failed to save images",
			logger.Int("image_count", len(imageIDs)),
			logger.Error(err))
		return nil, err
	}

	s.logger.Info("Started image save", logger.Int("image_count", len(imageIDs)))
	return reader, nil
}

// Load loads images from a tar archive.
func (s *serviceImpl) Load(ctx context.Context, input io.Reader, options ...client.ImageLoadOption) (image.LoadResponse, error) {
	client, err := s.manager.GetWithContext(ctx)
	if err != nil {
		s.logger.Error("Failed to get Docker client", logger.Error(err))
		return image.LoadResponse{}, err
	}

	response, err := client.ImageLoad(ctx, input, options...)
	if err != nil {
		s.logger.Error("Failed to load images", logger.Error(err))
		return image.LoadResponse{}, err
	}

	s.logger.Info("Started image load")
	return response, nil
}

// Search searches for images in Docker Hub.
func (s *serviceImpl) Search(ctx context.Context, term string, options registry.SearchOptions) ([]registry.SearchResult, error) {
	client, err := s.manager.GetWithContext(ctx)
	if err != nil {
		s.logger.Error("Failed to get Docker client", logger.Error(err))
		return nil, err
	}

	results, err := client.ImageSearch(ctx, term, options)
	if err != nil {
		s.logger.Error("Failed to search images",
			logger.String("term", term),
			logger.Error(err))
		return nil, err
	}

	s.logger.Debug("Searched images",
		logger.String("term", term),
		logger.Int("results", len(results)))
	return results, nil
}

// Prune removes unused images.
func (s *serviceImpl) Prune(ctx context.Context, pruneFilter filters.Args) (image.PruneReport, error) {
	client, err := s.manager.GetWithContext(ctx)
	if err != nil {
		s.logger.Error("Failed to get Docker client", logger.Error(err))
		return image.PruneReport{}, err
	}

	report, err := client.ImagesPrune(ctx, pruneFilter)
	if err != nil {
		s.logger.Error("Failed to prune images", logger.Error(err))
		return image.PruneReport{}, err
	}

	s.logger.Info("Pruned images",
		logger.Int("deleted_count", len(report.ImagesDeleted)),
		logger.Uint64("space_reclaimed", report.SpaceReclaimed))
	return report, nil
}

// Exists checks if an image exists.
func (s *serviceImpl) Exists(ctx context.Context, imageID string) (bool, error) {
	_, err := s.Inspect(ctx, imageID)
	if err != nil {
		return false, nil // Image doesn't exist
	}
	return true, nil
}

// GetDigest returns the digest of an image.
func (s *serviceImpl) GetDigest(ctx context.Context, imageID string) (string, error) {
	inspect, err := s.Inspect(ctx, imageID)
	if err != nil {
		return "", err
	}

	if len(inspect.RepoDigests) > 0 {
		return inspect.RepoDigests[0], nil
	}

	return inspect.ID, nil // Return ID if no digest available
}
