package container

import (
	"context"
	"fmt"
	"io"

	"github.com/containerd/errdefs"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/network"
	"github.com/threatflux/libgo/internal/docker"
	"github.com/threatflux/libgo/pkg/logger"
)

// Ensure the docker package is used (this prevents import errors).
var _ docker.Manager

// Create creates a new container.
func (s *serviceImpl) Create(ctx context.Context, config *container.Config, hostConfig *container.HostConfig, name string) (string, error) {
	client, err := s.manager.GetWithContext(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to get Docker client: %w", err)
	}

	// Create container
	resp, err := client.ContainerCreate(ctx, config, hostConfig, &network.NetworkingConfig{}, nil, name)
	if err != nil {
		s.logger.Error("Failed to create container",
			logger.String("name", name),
			logger.Error(err))
		return "", fmt.Errorf("failed to create container: %w", err)
	}

	// Log any warnings
	for _, warning := range resp.Warnings {
		s.logger.Warn("Container creation warning",
			logger.String("container_id", resp.ID),
			logger.String("warning", warning))
	}

	s.logger.Info("Container created successfully",
		logger.String("container_id", resp.ID),
		logger.String("name", name))

	return resp.ID, nil
}

// Start starts a container.
func (s *serviceImpl) Start(ctx context.Context, containerID string) error {
	client, err := s.manager.GetWithContext(ctx)
	if err != nil {
		return fmt.Errorf("failed to get Docker client: %w", err)
	}

	if err := client.ContainerStart(ctx, containerID, container.StartOptions{}); err != nil {
		s.logger.Error("Failed to start container",
			logger.String("container_id", containerID),
			logger.Error(err))
		return fmt.Errorf("failed to start container: %w", err)
	}

	s.logger.Info("Container started successfully",
		logger.String("container_id", containerID))
	return nil
}

// Stop stops a container.
func (s *serviceImpl) Stop(ctx context.Context, containerID string, timeout *int) error {
	client, err := s.manager.GetWithContext(ctx)
	if err != nil {
		return fmt.Errorf("failed to get Docker client: %w", err)
	}

	var stopOptions container.StopOptions
	if timeout != nil {
		stopOptions.Timeout = timeout
	}

	if err := client.ContainerStop(ctx, containerID, stopOptions); err != nil {
		s.logger.Error("Failed to stop container",
			logger.String("container_id", containerID),
			logger.Error(err))
		return fmt.Errorf("failed to stop container: %w", err)
	}

	s.logger.Info("Container stopped successfully",
		logger.String("container_id", containerID))
	return nil
}

// Restart restarts a container.
func (s *serviceImpl) Restart(ctx context.Context, containerID string, timeout *int) error {
	client, err := s.manager.GetWithContext(ctx)
	if err != nil {
		return fmt.Errorf("failed to get Docker client: %w", err)
	}

	var stopOptions container.StopOptions
	if timeout != nil {
		stopOptions.Timeout = timeout
	}

	if err := client.ContainerRestart(ctx, containerID, stopOptions); err != nil {
		s.logger.Error("Failed to restart container",
			logger.String("container_id", containerID),
			logger.Error(err))
		return fmt.Errorf("failed to restart container: %w", err)
	}

	s.logger.Info("Container restarted successfully",
		logger.String("container_id", containerID))
	return nil
}

// Kill sends a signal to a container.
func (s *serviceImpl) Kill(ctx context.Context, containerID string, signal string) error {
	client, err := s.manager.GetWithContext(ctx)
	if err != nil {
		return fmt.Errorf("failed to get Docker client: %w", err)
	}

	if err := client.ContainerKill(ctx, containerID, signal); err != nil {
		s.logger.Error("Failed to kill container",
			logger.String("container_id", containerID),
			logger.String("signal", signal),
			logger.Error(err))
		return fmt.Errorf("failed to kill container: %w", err)
	}

	s.logger.Info("Container killed successfully",
		logger.String("container_id", containerID),
		logger.String("signal", signal))
	return nil
}

// Remove removes a container.
func (s *serviceImpl) Remove(ctx context.Context, containerID string, force bool) error {
	client, err := s.manager.GetWithContext(ctx)
	if err != nil {
		return fmt.Errorf("failed to get Docker client: %w", err)
	}

	removeOptions := container.RemoveOptions{
		Force:         force,
		RemoveVolumes: true,
	}

	if err := client.ContainerRemove(ctx, containerID, removeOptions); err != nil {
		s.logger.Error("Failed to remove container",
			logger.String("container_id", containerID),
			logger.Bool("force", force),
			logger.Error(err))
		return fmt.Errorf("failed to remove container: %w", err)
	}

	s.logger.Info("Container removed successfully",
		logger.String("container_id", containerID))
	return nil
}

// Pause pauses a container.
func (s *serviceImpl) Pause(ctx context.Context, containerID string) error {
	client, err := s.manager.GetWithContext(ctx)
	if err != nil {
		return fmt.Errorf("failed to get Docker client: %w", err)
	}

	if err := client.ContainerPause(ctx, containerID); err != nil {
		s.logger.Error("Failed to pause container",
			logger.String("container_id", containerID),
			logger.Error(err))
		return fmt.Errorf("failed to pause container: %w", err)
	}

	s.logger.Info("Container paused successfully",
		logger.String("container_id", containerID))
	return nil
}

// Unpause unpauses a container.
func (s *serviceImpl) Unpause(ctx context.Context, containerID string) error {
	client, err := s.manager.GetWithContext(ctx)
	if err != nil {
		return fmt.Errorf("failed to get Docker client: %w", err)
	}

	if err := client.ContainerUnpause(ctx, containerID); err != nil {
		s.logger.Error("Failed to unpause container",
			logger.String("container_id", containerID),
			logger.Error(err))
		return fmt.Errorf("failed to unpause container: %w", err)
	}

	s.logger.Info("Container unpaused successfully",
		logger.String("container_id", containerID))
	return nil
}

// Inspect inspects a container.
func (s *serviceImpl) Inspect(ctx context.Context, containerID string) (*container.InspectResponse, error) {
	client, err := s.manager.GetWithContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get Docker client: %w", err)
	}

	containerJSON, err := client.ContainerInspect(ctx, containerID)
	if err != nil {
		s.logger.Error("Failed to inspect container",
			logger.String("container_id", containerID),
			logger.Error(err))
		return nil, fmt.Errorf("failed to inspect container: %w", err)
	}

	return &containerJSON, nil
}

// List lists containers.
func (s *serviceImpl) List(ctx context.Context, options container.ListOptions) ([]container.Summary, error) {
	client, err := s.manager.GetWithContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get Docker client: %w", err)
	}

	containers, err := client.ContainerList(ctx, options)
	if err != nil {
		s.logger.Error("Failed to list containers", logger.Error(err))
		return nil, fmt.Errorf("failed to list containers: %w", err)
	}

	return containers, nil
}

// Exists checks if a container exists.
func (s *serviceImpl) Exists(ctx context.Context, containerID string) (bool, error) {
	_, err := s.Inspect(ctx, containerID)
	if err != nil {
		if errdefs.IsNotFound(err) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

// Logs gets container logs.
func (s *serviceImpl) Logs(ctx context.Context, containerID string, options container.LogsOptions) (io.ReadCloser, error) {
	client, err := s.manager.GetWithContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get Docker client: %w", err)
	}

	logs, err := client.ContainerLogs(ctx, containerID, options)
	if err != nil {
		s.logger.Error("Failed to get container logs",
			logger.String("container_id", containerID),
			logger.Error(err))
		return nil, fmt.Errorf("failed to get container logs: %w", err)
	}

	return logs, nil
}

// Stats gets container stats.
func (s *serviceImpl) Stats(ctx context.Context, containerID string, stream bool) (container.StatsResponseReader, error) {
	client, err := s.manager.GetWithContext(ctx)
	if err != nil {
		return container.StatsResponseReader{}, fmt.Errorf("failed to get Docker client: %w", err)
	}

	stats, err := client.ContainerStats(ctx, containerID, stream)
	if err != nil {
		s.logger.Error("Failed to get container stats",
			logger.String("container_id", containerID),
			logger.Bool("stream", stream),
			logger.Error(err))
		return container.StatsResponseReader{}, fmt.Errorf("failed to get container stats: %w", err)
	}

	return stats, nil
}

// ExecCreate creates an exec instance.
func (s *serviceImpl) ExecCreate(ctx context.Context, containerID string, config container.ExecOptions) (container.ExecCreateResponse, error) {
	client, err := s.manager.GetWithContext(ctx)
	if err != nil {
		return container.ExecCreateResponse{}, fmt.Errorf("failed to get Docker client: %w", err)
	}

	exec, err := client.ContainerExecCreate(ctx, containerID, config)
	if err != nil {
		s.logger.Error("Failed to create exec instance",
			logger.String("container_id", containerID),
			logger.Error(err))
		return container.ExecCreateResponse{}, fmt.Errorf("failed to create exec instance: %w", err)
	}

	return exec, nil
}

// ExecStart starts an exec instance.
func (s *serviceImpl) ExecStart(ctx context.Context, execID string, config container.ExecStartOptions) error {
	client, err := s.manager.GetWithContext(ctx)
	if err != nil {
		return fmt.Errorf("failed to get Docker client: %w", err)
	}

	if err := client.ContainerExecStart(ctx, execID, config); err != nil {
		s.logger.Error("Failed to start exec instance",
			logger.String("exec_id", execID),
			logger.Error(err))
		return fmt.Errorf("failed to start exec instance: %w", err)
	}

	return nil
}

// ExecInspect inspects an exec instance.
func (s *serviceImpl) ExecInspect(ctx context.Context, execID string) (container.ExecInspect, error) {
	client, err := s.manager.GetWithContext(ctx)
	if err != nil {
		return container.ExecInspect{}, fmt.Errorf("failed to get Docker client: %w", err)
	}

	inspect, err := client.ContainerExecInspect(ctx, execID)
	if err != nil {
		s.logger.Error("Failed to inspect exec instance",
			logger.String("exec_id", execID),
			logger.Error(err))
		return container.ExecInspect{}, fmt.Errorf("failed to inspect exec instance: %w", err)
	}

	return inspect, nil
}

// CopyToContainer copies content to a container.
func (s *serviceImpl) CopyToContainer(ctx context.Context, containerID, dstPath string, content io.Reader, options container.CopyToContainerOptions) error {
	client, err := s.manager.GetWithContext(ctx)
	if err != nil {
		return fmt.Errorf("failed to get Docker client: %w", err)
	}

	if err := client.CopyToContainer(ctx, containerID, dstPath, content, options); err != nil {
		s.logger.Error("Failed to copy to container",
			logger.String("container_id", containerID),
			logger.String("dst_path", dstPath),
			logger.Error(err))
		return fmt.Errorf("failed to copy to container: %w", err)
	}

	return nil
}

// CopyFromContainer copies content from a container.
func (s *serviceImpl) CopyFromContainer(ctx context.Context, containerID, srcPath string) (io.ReadCloser, container.PathStat, error) {
	client, err := s.manager.GetWithContext(ctx)
	if err != nil {
		return nil, container.PathStat{}, fmt.Errorf("failed to get Docker client: %w", err)
	}

	reader, stat, err := client.CopyFromContainer(ctx, containerID, srcPath)
	if err != nil {
		s.logger.Error("Failed to copy from container",
			logger.String("container_id", containerID),
			logger.String("src_path", srcPath),
			logger.Error(err))
		return nil, container.PathStat{}, fmt.Errorf("failed to copy from container: %w", err)
	}

	return reader, stat, nil
}

// Wait waits for a container.
func (s *serviceImpl) Wait(ctx context.Context, containerID string, condition container.WaitCondition) (<-chan container.WaitResponse, <-chan error) {
	client, err := s.manager.GetWithContext(ctx)
	if err != nil {
		errCh := make(chan error, 1)
		errCh <- fmt.Errorf("failed to get Docker client: %w", err)
		close(errCh)
		return nil, errCh
	}

	return client.ContainerWait(ctx, containerID, condition)
}

// Attach attaches to a container.
func (s *serviceImpl) Attach(ctx context.Context, containerID string, options container.AttachOptions) (types.HijackedResponse, error) {
	client, err := s.manager.GetWithContext(ctx)
	if err != nil {
		return types.HijackedResponse{}, fmt.Errorf("failed to get Docker client: %w", err)
	}

	resp, err := client.ContainerAttach(ctx, containerID, options)
	if err != nil {
		s.logger.Error("Failed to attach to container",
			logger.String("container_id", containerID),
			logger.Error(err))
		return types.HijackedResponse{}, fmt.Errorf("failed to attach to container: %w", err)
	}

	return resp, nil
}

// Update updates a container configuration.
func (s *serviceImpl) Update(ctx context.Context, containerID string, updateConfig container.UpdateConfig) (container.UpdateResponse, error) {
	client, err := s.manager.GetWithContext(ctx)
	if err != nil {
		return container.UpdateResponse{}, fmt.Errorf("failed to get Docker client: %w", err)
	}

	resp, err := client.ContainerUpdate(ctx, containerID, updateConfig)
	if err != nil {
		s.logger.Error("Failed to update container",
			logger.String("container_id", containerID),
			logger.Error(err))
		return container.UpdateResponse{}, fmt.Errorf("failed to update container: %w", err)
	}

	return resp, nil
}

// Rename renames a container.
func (s *serviceImpl) Rename(ctx context.Context, containerID string, newName string) error {
	client, err := s.manager.GetWithContext(ctx)
	if err != nil {
		return fmt.Errorf("failed to get Docker client: %w", err)
	}

	if err := client.ContainerRename(ctx, containerID, newName); err != nil {
		s.logger.Error("Failed to rename container",
			logger.String("container_id", containerID),
			logger.String("new_name", newName),
			logger.Error(err))
		return fmt.Errorf("failed to rename container: %w", err)
	}

	s.logger.Info("Container renamed successfully",
		logger.String("container_id", containerID),
		logger.String("new_name", newName))
	return nil
}

// Resize resizes a container TTY.
func (s *serviceImpl) Resize(ctx context.Context, containerID string, options container.ResizeOptions) error {
	client, err := s.manager.GetWithContext(ctx)
	if err != nil {
		return fmt.Errorf("failed to get Docker client: %w", err)
	}

	if err := client.ContainerResize(ctx, containerID, options); err != nil {
		s.logger.Error("Failed to resize container",
			logger.String("container_id", containerID),
			logger.Error(err))
		return fmt.Errorf("failed to resize container: %w", err)
	}

	return nil
}

// Prune removes unused containers.
func (s *serviceImpl) Prune(ctx context.Context, pruneFilters container.ListOptions) (container.PruneReport, error) {
	client, err := s.manager.GetWithContext(ctx)
	if err != nil {
		return container.PruneReport{}, fmt.Errorf("failed to get Docker client: %w", err)
	}

	report, err := client.ContainersPrune(ctx, filters.Args{})
	if err != nil {
		s.logger.Error("Failed to prune containers", logger.Error(err))
		return container.PruneReport{}, fmt.Errorf("failed to prune containers: %w", err)
	}

	s.logger.Info("Containers pruned successfully",
		logger.Int("containers_deleted", len(report.ContainersDeleted)),
		logger.Uint64("space_reclaimed", report.SpaceReclaimed))

	return report, nil
}
