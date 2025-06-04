package network

import (
	"context"

	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/network"
	"github.com/threatflux/libgo/pkg/logger"
)

// Create creates a new network
func (s *serviceImpl) Create(ctx context.Context, name string, options network.CreateOptions) (network.CreateResponse, error) {
	client, err := s.manager.GetWithContext(ctx)
	if err != nil {
		s.logger.Error("Failed to get Docker client", logger.Error(err))
		return network.CreateResponse{}, err
	}

	response, err := client.NetworkCreate(ctx, name, options)
	if err != nil {
		s.logger.Error("Failed to create network",
			logger.String("name", name),
			logger.Error(err))
		return network.CreateResponse{}, err
	}

	s.logger.Info("Created network",
		logger.String("name", name),
		logger.String("id", response.ID))
	return response, nil
}

// Remove removes a network
func (s *serviceImpl) Remove(ctx context.Context, networkID string) error {
	client, err := s.manager.GetWithContext(ctx)
	if err != nil {
		s.logger.Error("Failed to get Docker client", logger.Error(err))
		return err
	}

	err = client.NetworkRemove(ctx, networkID)
	if err != nil {
		s.logger.Error("Failed to remove network",
			logger.String("network_id", networkID),
			logger.Error(err))
		return err
	}

	s.logger.Info("Removed network", logger.String("network_id", networkID))
	return nil
}

// Connect connects a container to a network
func (s *serviceImpl) Connect(ctx context.Context, networkID, containerID string, config *network.EndpointSettings) error {
	client, err := s.manager.GetWithContext(ctx)
	if err != nil {
		s.logger.Error("Failed to get Docker client", logger.Error(err))
		return err
	}

	err = client.NetworkConnect(ctx, networkID, containerID, config)
	if err != nil {
		s.logger.Error("Failed to connect container to network",
			logger.String("network_id", networkID),
			logger.String("container_id", containerID),
			logger.Error(err))
		return err
	}

	s.logger.Info("Connected container to network",
		logger.String("network_id", networkID),
		logger.String("container_id", containerID))
	return nil
}

// Disconnect disconnects a container from a network
func (s *serviceImpl) Disconnect(ctx context.Context, networkID, containerID string, force bool) error {
	client, err := s.manager.GetWithContext(ctx)
	if err != nil {
		s.logger.Error("Failed to get Docker client", logger.Error(err))
		return err
	}

	err = client.NetworkDisconnect(ctx, networkID, containerID, force)
	if err != nil {
		s.logger.Error("Failed to disconnect container from network",
			logger.String("network_id", networkID),
			logger.String("container_id", containerID),
			logger.Bool("force", force),
			logger.Error(err))
		return err
	}

	s.logger.Info("Disconnected container from network",
		logger.String("network_id", networkID),
		logger.String("container_id", containerID))
	return nil
}

// Inspect returns detailed information about a network
func (s *serviceImpl) Inspect(ctx context.Context, networkID string, options network.InspectOptions) (*network.Inspect, error) {
	client, err := s.manager.GetWithContext(ctx)
	if err != nil {
		s.logger.Error("Failed to get Docker client", logger.Error(err))
		return nil, err
	}

	inspect, err := client.NetworkInspect(ctx, networkID, options)
	if err != nil {
		s.logger.Error("Failed to inspect network",
			logger.String("network_id", networkID),
			logger.Error(err))
		return nil, err
	}

	s.logger.Debug("Inspected network", logger.String("network_id", networkID))
	return &inspect, nil
}

// List returns a list of networks
func (s *serviceImpl) List(ctx context.Context, options network.ListOptions) ([]network.Summary, error) {
	client, err := s.manager.GetWithContext(ctx)
	if err != nil {
		s.logger.Error("Failed to get Docker client", logger.Error(err))
		return nil, err
	}

	networks, err := client.NetworkList(ctx, options)
	if err != nil {
		s.logger.Error("Failed to list networks", logger.Error(err))
		return nil, err
	}

	s.logger.Debug("Listed networks", logger.Int("count", len(networks)))
	return networks, nil
}

// Prune removes unused networks
func (s *serviceImpl) Prune(ctx context.Context, pruneFilter filters.Args) (network.PruneReport, error) {
	client, err := s.manager.GetWithContext(ctx)
	if err != nil {
		s.logger.Error("Failed to get Docker client", logger.Error(err))
		return network.PruneReport{}, err
	}

	report, err := client.NetworksPrune(ctx, pruneFilter)
	if err != nil {
		s.logger.Error("Failed to prune networks", logger.Error(err))
		return network.PruneReport{}, err
	}

	s.logger.Info("Pruned networks",
		logger.Int("deleted_count", len(report.NetworksDeleted)))
	return report, nil
}

// Exists checks if a network exists
func (s *serviceImpl) Exists(ctx context.Context, networkID string) (bool, error) {
	_, err := s.Inspect(ctx, networkID, network.InspectOptions{})
	if err != nil {
		return false, nil // Network doesn't exist
	}
	return true, nil
}

// GetByName returns a network by name
func (s *serviceImpl) GetByName(ctx context.Context, name string) (*network.Inspect, error) {
	return s.Inspect(ctx, name, network.InspectOptions{})
}
