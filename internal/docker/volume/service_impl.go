package volume

import (
	"context"

	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/volume"
	"github.com/threatflux/libgo/pkg/logger"
)

// Create creates a new volume
func (s *serviceImpl) Create(ctx context.Context, options volume.CreateOptions) (*volume.Volume, error) {
	client, err := s.manager.GetWithContext(ctx)
	if err != nil {
		s.logger.Error("Failed to get Docker client", logger.Error(err))
		return nil, err
	}

	vol, err := client.VolumeCreate(ctx, options)
	if err != nil {
		s.logger.Error("Failed to create volume",
			logger.String("name", options.Name),
			logger.Error(err))
		return nil, err
	}

	s.logger.Info("Created volume",
		logger.String("name", vol.Name),
		logger.String("driver", vol.Driver))
	return &vol, nil
}

// Remove removes a volume
func (s *serviceImpl) Remove(ctx context.Context, volumeID string, force bool) error {
	client, err := s.manager.GetWithContext(ctx)
	if err != nil {
		s.logger.Error("Failed to get Docker client", logger.Error(err))
		return err
	}

	err = client.VolumeRemove(ctx, volumeID, force)
	if err != nil {
		s.logger.Error("Failed to remove volume",
			logger.String("volume_id", volumeID),
			logger.Bool("force", force),
			logger.Error(err))
		return err
	}

	s.logger.Info("Removed volume",
		logger.String("volume_id", volumeID),
		logger.Bool("force", force))
	return nil
}

// Inspect returns detailed information about a volume
func (s *serviceImpl) Inspect(ctx context.Context, volumeID string) (*volume.Volume, error) {
	client, err := s.manager.GetWithContext(ctx)
	if err != nil {
		s.logger.Error("Failed to get Docker client", logger.Error(err))
		return nil, err
	}

	vol, err := client.VolumeInspect(ctx, volumeID)
	if err != nil {
		s.logger.Error("Failed to inspect volume",
			logger.String("volume_id", volumeID),
			logger.Error(err))
		return nil, err
	}

	s.logger.Debug("Inspected volume", logger.String("volume_id", volumeID))
	return &vol, nil
}

// List returns a list of volumes
func (s *serviceImpl) List(ctx context.Context, options volume.ListOptions) (*volume.ListResponse, error) {
	client, err := s.manager.GetWithContext(ctx)
	if err != nil {
		s.logger.Error("Failed to get Docker client", logger.Error(err))
		return nil, err
	}

	response, err := client.VolumeList(ctx, options)
	if err != nil {
		s.logger.Error("Failed to list volumes", logger.Error(err))
		return nil, err
	}

	s.logger.Debug("Listed volumes", logger.Int("count", len(response.Volumes)))
	return &response, nil
}

// Prune removes unused volumes
func (s *serviceImpl) Prune(ctx context.Context, pruneFilter filters.Args) (volume.PruneReport, error) {
	client, err := s.manager.GetWithContext(ctx)
	if err != nil {
		s.logger.Error("Failed to get Docker client", logger.Error(err))
		return volume.PruneReport{}, err
	}

	report, err := client.VolumesPrune(ctx, pruneFilter)
	if err != nil {
		s.logger.Error("Failed to prune volumes", logger.Error(err))
		return volume.PruneReport{}, err
	}

	s.logger.Info("Pruned volumes",
		logger.Int("deleted_count", len(report.VolumesDeleted)),
		logger.Uint64("space_reclaimed", report.SpaceReclaimed))
	return report, nil
}

// Exists checks if a volume exists
func (s *serviceImpl) Exists(ctx context.Context, volumeID string) (bool, error) {
	_, err := s.Inspect(ctx, volumeID)
	if err != nil {
		return false, nil // Volume doesn't exist
	}
	return true, nil
}

// GetByName returns a volume by name
func (s *serviceImpl) GetByName(ctx context.Context, name string) (*volume.Volume, error) {
	return s.Inspect(ctx, name)
}
