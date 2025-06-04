package docker

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/network"
	"github.com/threatflux/libgo/internal/compute"
	"github.com/threatflux/libgo/pkg/logger"
)

// BackendService implements the compute.BackendService interface for Docker
type BackendService struct {
	manager Manager
	logger  logger.Logger
}

// NewBackendService creates a new Docker backend service
func NewBackendService(manager Manager, logger logger.Logger) compute.BackendService {
	return &BackendService{
		manager: manager,
		logger:  logger,
	}
}

// Create creates a new Docker container
func (s *BackendService) Create(ctx context.Context, req compute.ComputeInstanceRequest) (*compute.ComputeInstance, error) {
	if req.Type != compute.InstanceTypeContainer {
		return nil, fmt.Errorf("unsupported instance type %s for Docker backend", req.Type)
	}

	client, err := s.manager.GetWithContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get Docker client: %w", err)
	}

	// Convert compute request to Docker container config
	containerConfig, hostConfig, networkConfig, err := s.convertToDockerConfig(req)
	if err != nil {
		return nil, fmt.Errorf("failed to convert config: %w", err)
	}

	// Create the container
	resp, err := client.ContainerCreate(ctx, containerConfig, hostConfig, networkConfig, nil, req.Name)
	if err != nil {
		return nil, fmt.Errorf("failed to create container: %w", err)
	}

	// Get the created container details
	containerJSON, err := client.ContainerInspect(ctx, resp.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to inspect created container: %w", err)
	}

	// Convert Docker container to compute instance
	instance := s.convertFromDockerContainer(containerJSON)

	// Auto-start if requested
	if req.AutoStart {
		if err := s.Start(ctx, instance.ID); err != nil {
			s.logger.Warn("Failed to auto-start container", logger.String("id", instance.ID), logger.Error(err))
		}
	}

	return instance, nil
}

// Get retrieves a Docker container by ID
func (s *BackendService) Get(ctx context.Context, id string) (*compute.ComputeInstance, error) {
	client, err := s.manager.GetWithContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get Docker client: %w", err)
	}

	containerJSON, err := client.ContainerInspect(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to inspect container: %w", err)
	}

	return s.convertFromDockerContainer(containerJSON), nil
}

// List retrieves Docker containers based on options
func (s *BackendService) List(ctx context.Context, opts compute.ComputeInstanceListOptions) ([]*compute.ComputeInstance, error) {
	client, err := s.manager.GetWithContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get Docker client: %w", err)
	}

	// Convert compute list options to Docker list options
	listOptions := s.convertListOptions(opts)

	containers, err := client.ContainerList(ctx, listOptions)
	if err != nil {
		return nil, fmt.Errorf("failed to list containers: %w", err)
	}

	instances := make([]*compute.ComputeInstance, 0, len(containers))
	for _, cont := range containers {
		// Get detailed information for each container
		containerJSON, err := client.ContainerInspect(ctx, cont.ID)
		if err != nil {
			s.logger.Warn("Failed to inspect container", logger.String("id", cont.ID), logger.Error(err))
			continue
		}

		instance := s.convertFromDockerContainer(containerJSON)
		instances = append(instances, instance)
	}

	return instances, nil
}

// Update updates a Docker container
func (s *BackendService) Update(ctx context.Context, id string, update compute.ComputeInstanceUpdate) (*compute.ComputeInstance, error) {
	client, err := s.manager.GetWithContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get Docker client: %w", err)
	}

	// Convert compute update to Docker update config
	updateConfig := s.convertUpdateConfig(update)

	// Update the container
	_, err = client.ContainerUpdate(ctx, id, updateConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to update container: %w", err)
	}

	// Return updated container
	return s.Get(ctx, id)
}

// Delete removes a Docker container
func (s *BackendService) Delete(ctx context.Context, id string, force bool) error {
	client, err := s.manager.GetWithContext(ctx)
	if err != nil {
		return fmt.Errorf("failed to get Docker client: %w", err)
	}

	removeOptions := container.RemoveOptions{
		Force:         force,
		RemoveVolumes: true,
	}

	return client.ContainerRemove(ctx, id, removeOptions)
}

// Start starts a Docker container
func (s *BackendService) Start(ctx context.Context, id string) error {
	client, err := s.manager.GetWithContext(ctx)
	if err != nil {
		return fmt.Errorf("failed to get Docker client: %w", err)
	}

	return client.ContainerStart(ctx, id, container.StartOptions{})
}

// Stop stops a Docker container
func (s *BackendService) Stop(ctx context.Context, id string, force bool) error {
	client, err := s.manager.GetWithContext(ctx)
	if err != nil {
		return fmt.Errorf("failed to get Docker client: %w", err)
	}

	var timeout *int
	if !force {
		t := 30 // 30 second graceful shutdown
		timeout = &t
	}

	return client.ContainerStop(ctx, id, container.StopOptions{Timeout: timeout})
}

// Restart restarts a Docker container
func (s *BackendService) Restart(ctx context.Context, id string, force bool) error {
	client, err := s.manager.GetWithContext(ctx)
	if err != nil {
		return fmt.Errorf("failed to get Docker client: %w", err)
	}

	var timeout *int
	if !force {
		t := 30
		timeout = &t
	}

	return client.ContainerRestart(ctx, id, container.StopOptions{Timeout: timeout})
}

// Pause pauses a Docker container
func (s *BackendService) Pause(ctx context.Context, id string) error {
	client, err := s.manager.GetWithContext(ctx)
	if err != nil {
		return fmt.Errorf("failed to get Docker client: %w", err)
	}

	return client.ContainerPause(ctx, id)
}

// Unpause unpauses a Docker container
func (s *BackendService) Unpause(ctx context.Context, id string) error {
	client, err := s.manager.GetWithContext(ctx)
	if err != nil {
		return fmt.Errorf("failed to get Docker client: %w", err)
	}

	return client.ContainerUnpause(ctx, id)
}

// GetResourceUsage gets current resource usage for a container
func (s *BackendService) GetResourceUsage(ctx context.Context, id string) (*compute.ResourceUsage, error) {
	client, err := s.manager.GetWithContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get Docker client: %w", err)
	}

	stats, err := client.ContainerStatsOneShot(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get container stats: %w", err)
	}
	defer stats.Body.Close()

	var dockerStats container.Stats
	if err := json.NewDecoder(stats.Body).Decode(&dockerStats); err != nil {
		return nil, fmt.Errorf("failed to decode stats: %w", err)
	}

	return s.convertResourceUsage(dockerStats), nil
}

// UpdateResourceLimits updates resource limits for a container
func (s *BackendService) UpdateResourceLimits(ctx context.Context, id string, resources compute.ComputeResources) error {
	client, err := s.manager.GetWithContext(ctx)
	if err != nil {
		return fmt.Errorf("failed to get Docker client: %w", err)
	}

	updateConfig := container.UpdateConfig{
		Resources: container.Resources{
			Memory:     resources.Memory.Limit,
			MemorySwap: resources.Memory.Swap,
			CPUShares:  resources.CPU.Shares,
			CPUQuota:   resources.CPU.Quota,
			CPUPeriod:  resources.CPU.Period,
			CpusetCpus: resources.CPU.SetCPUs,
			CpusetMems: resources.CPU.SetMems,
		},
	}

	_, err = client.ContainerUpdate(ctx, id, updateConfig)
	return err
}

// GetBackendInfo returns information about the Docker backend
func (s *BackendService) GetBackendInfo(ctx context.Context) (*compute.BackendInfo, error) {
	client, err := s.manager.GetWithContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get Docker client: %w", err)
	}

	info, err := client.Info(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get Docker info: %w", err)
	}

	version, err := client.ServerVersion(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get Docker version: %w", err)
	}

	// Ping to check health
	_, err = client.Ping(ctx)
	healthStatus := &compute.HealthStatus{
		Status:    "healthy",
		LastCheck: time.Now(),
	}
	if err != nil {
		healthStatus.Status = "unhealthy"
		healthStatus.Message = err.Error()
	}

	return &compute.BackendInfo{
		Type:           compute.BackendDocker,
		Version:        version.Version,
		APIVersion:     version.APIVersion,
		Status:         "running",
		Capabilities:   []string{"containers", "images", "networks", "volumes"},
		SupportedTypes: []compute.ComputeInstanceType{compute.InstanceTypeContainer},
		ResourceLimits: compute.ComputeResources{
			CPU: compute.CPUResources{
				Cores: float64(info.NCPU),
			},
			Memory: compute.MemoryResources{
				Limit: info.MemTotal,
			},
		},
		Configuration: map[string]interface{}{
			"driver":         info.Driver,
			"storage_driver": info.Driver, // Docker v28 uses Driver field for storage
			"kernel_version": info.KernelVersion,
			"os_type":        info.OSType,
			"architecture":   info.Architecture,
		},
		HealthCheck: healthStatus,
	}, nil
}

// ValidateConfig validates a compute instance configuration for Docker
func (s *BackendService) ValidateConfig(ctx context.Context, config compute.ComputeInstanceConfig) error {
	if config.Image == "" {
		return fmt.Errorf("image is required for Docker containers")
	}

	// Validate resource constraints
	if config.SecurityContext != nil && config.SecurityContext.RunAsUser != nil && *config.SecurityContext.RunAsUser < 0 {
		return fmt.Errorf("invalid user ID: %d", *config.SecurityContext.RunAsUser)
	}

	return nil
}

// GetBackendType returns the backend type
func (s *BackendService) GetBackendType() compute.ComputeBackend {
	return compute.BackendDocker
}

// GetSupportedInstanceTypes returns supported instance types
func (s *BackendService) GetSupportedInstanceTypes() []compute.ComputeInstanceType {
	return []compute.ComputeInstanceType{compute.InstanceTypeContainer}
}

// Helper methods for conversion

func (s *BackendService) convertToDockerConfig(req compute.ComputeInstanceRequest) (*container.Config, *container.HostConfig, *network.NetworkingConfig, error) {
	// Container config
	config := &container.Config{
		Image:      req.Config.Image,
		Cmd:        req.Config.Command,
		Env:        s.convertEnvironment(req.Config.Environment),
		WorkingDir: req.Config.WorkingDir,
		User:       req.Config.User,
		Labels:     req.Labels,
	}

	// Host config
	hostConfig := &container.HostConfig{
		Privileged: req.Config.Privileged,
		Resources: container.Resources{
			Memory:     req.Resources.Memory.Limit,
			MemorySwap: req.Resources.Memory.Swap,
			CPUShares:  req.Resources.CPU.Shares,
			CPUQuota:   req.Resources.CPU.Quota,
			CPUPeriod:  req.Resources.CPU.Period,
			CpusetCpus: req.Resources.CPU.SetCPUs,
			CpusetMems: req.Resources.CPU.SetMems,
		},
		RestartPolicy: container.RestartPolicy{
			Name:              container.RestartPolicyMode(req.Config.RestartPolicy.Policy),
			MaximumRetryCount: req.Config.RestartPolicy.MaximumRetryCount,
		},
	}

	// Set security context if provided
	if req.Config.SecurityContext != nil {
		if req.Config.SecurityContext.ReadOnlyRootFS != nil {
			hostConfig.ReadonlyRootfs = *req.Config.SecurityContext.ReadOnlyRootFS
		}
		if req.Config.SecurityContext.Capabilities != nil {
			hostConfig.CapAdd = req.Config.SecurityContext.Capabilities.Add
			hostConfig.CapDrop = req.Config.SecurityContext.Capabilities.Drop
		}
	}

	// Capabilities
	if len(req.Config.Capabilities) > 0 {
		hostConfig.CapAdd = append(hostConfig.CapAdd, req.Config.Capabilities...)
	}

	// Network config
	networkConfig := &network.NetworkingConfig{}
	if len(req.Networks) > 0 {
		endpoints := make(map[string]*network.EndpointSettings)
		for _, net := range req.Networks {
			endpoints[net.Network] = &network.EndpointSettings{
				IPAMConfig: &network.EndpointIPAMConfig{
					IPv4Address: net.IPAddress,
					IPv6Address: net.IPv6Address,
				},
			}
		}
		networkConfig.EndpointsConfig = endpoints
	}

	return config, hostConfig, networkConfig, nil
}

func (s *BackendService) convertFromDockerContainer(containerJSON types.ContainerJSON) *compute.ComputeInstance {
	instance := &compute.ComputeInstance{
		ID:      containerJSON.ID,
		Name:    strings.TrimPrefix(containerJSON.Name, "/"),
		Type:    compute.InstanceTypeContainer,
		Backend: compute.BackendDocker,
		State:   s.convertContainerState(containerJSON.State),
		Status:  containerJSON.State.Status,
		Config: compute.ComputeInstanceConfig{
			Image:       containerJSON.Config.Image,
			Command:     containerJSON.Config.Cmd,
			Environment: s.convertEnvironmentFromSlice(containerJSON.Config.Env),
			WorkingDir:  containerJSON.Config.WorkingDir,
			User:        containerJSON.Config.User,
			RestartPolicy: compute.RestartPolicy{
				Policy:            string(containerJSON.HostConfig.RestartPolicy.Name),
				MaximumRetryCount: containerJSON.HostConfig.RestartPolicy.MaximumRetryCount,
			},
		},
		Resources: compute.ComputeResources{
			CPU: compute.CPUResources{
				Shares:  containerJSON.HostConfig.CPUShares,
				Quota:   containerJSON.HostConfig.CPUQuota,
				Period:  containerJSON.HostConfig.CPUPeriod,
				SetCPUs: containerJSON.HostConfig.CpusetCpus,
				SetMems: containerJSON.HostConfig.CpusetMems,
			},
			Memory: compute.MemoryResources{
				Limit: containerJSON.HostConfig.Memory,
				Swap:  containerJSON.HostConfig.MemorySwap,
			},
		},
		Labels:      containerJSON.Config.Labels,
		RuntimeInfo: s.convertRuntimeInfo(containerJSON),
		CreatedAt:   time.Time{}, // We'll parse the created time below
		BackendData: map[string]interface{}{
			"docker_id":     containerJSON.ID,
			"image_id":      containerJSON.Image,
			"platform":      containerJSON.Platform,
			"driver":        containerJSON.Driver,
			"mount_label":   containerJSON.MountLabel,
			"process_label": containerJSON.ProcessLabel,
		},
	}

	// Parse created time
	if containerJSON.Created != "" {
		if t, err := time.Parse(time.RFC3339Nano, containerJSON.Created); err == nil {
			instance.CreatedAt = t
		}
	}

	// Set started/finished times
	if containerJSON.State != nil {
		if containerJSON.State.StartedAt != "" {
			if t, err := time.Parse(time.RFC3339Nano, containerJSON.State.StartedAt); err == nil {
				instance.StartedAt = &t
			}
		}
		if containerJSON.State.FinishedAt != "" {
			if t, err := time.Parse(time.RFC3339Nano, containerJSON.State.FinishedAt); err == nil && !t.IsZero() {
				instance.FinishedAt = &t
			}
		}
	}

	return instance
}

func (s *BackendService) convertContainerState(state *types.ContainerState) compute.ComputeInstanceState {
	if state == nil {
		return compute.StateUnknown
	}

	switch state.Status {
	case "created":
		return compute.StateCreated
	case "running":
		return compute.StateRunning
	case "paused":
		return compute.StatePaused
	case "restarting":
		return compute.StateRestarting
	case "removing":
		return compute.StateStopping
	case "exited":
		return compute.StateStopped
	case "dead":
		return compute.StateError
	default:
		return compute.StateUnknown
	}
}

func (s *BackendService) convertEnvironment(env map[string]string) []string {
	if env == nil {
		return nil
	}

	result := make([]string, 0, len(env))
	for k, v := range env {
		result = append(result, fmt.Sprintf("%s=%s", k, v))
	}
	return result
}

func (s *BackendService) convertEnvironmentFromSlice(env []string) map[string]string {
	if env == nil {
		return nil
	}

	result := make(map[string]string)
	for _, e := range env {
		parts := strings.SplitN(e, "=", 2)
		if len(parts) == 2 {
			result[parts[0]] = parts[1]
		}
	}
	return result
}

func (s *BackendService) convertListOptions(opts compute.ComputeInstanceListOptions) container.ListOptions {
	listOpts := container.ListOptions{
		All: true, // Show all containers by default
	}

	if opts.State != "" {
		listOpts.Filters = filters.NewArgs()
		switch opts.State {
		case compute.StateRunning:
			listOpts.Filters.Add("status", "running")
		case compute.StateStopped:
			listOpts.Filters.Add("status", "exited")
		case compute.StatePaused:
			listOpts.Filters.Add("status", "paused")
		case compute.StateCreated:
			listOpts.Filters.Add("status", "created")
		}
	}

	if opts.Limit > 0 {
		listOpts.Limit = opts.Limit
	}

	return listOpts
}

func (s *BackendService) convertUpdateConfig(update compute.ComputeInstanceUpdate) container.UpdateConfig {
	config := container.UpdateConfig{}

	if update.Resources != nil {
		config.Resources = container.Resources{
			Memory:     update.Resources.Memory.Limit,
			MemorySwap: update.Resources.Memory.Swap,
			CPUShares:  update.Resources.CPU.Shares,
			CPUQuota:   update.Resources.CPU.Quota,
			CPUPeriod:  update.Resources.CPU.Period,
			CpusetCpus: update.Resources.CPU.SetCPUs,
			CpusetMems: update.Resources.CPU.SetMems,
		}
	}

	return config
}

func (s *BackendService) convertRuntimeInfo(containerJSON types.ContainerJSON) compute.RuntimeInfo {
	runtimeInfo := compute.RuntimeInfo{
		ProcessID: containerJSON.State.Pid,
		ExitCode:  containerJSON.State.ExitCode,
		HostInfo: compute.HostInfo{
			NodeName:   containerJSON.Config.Hostname, // Use hostname from config
			Hypervisor: "docker",
		},
	}

	// Convert network settings
	if containerJSON.NetworkSettings != nil && containerJSON.NetworkSettings.Networks != nil {
		runtimeInfo.Networks = make(map[string]compute.NetworkRuntimeInfo)
		for name, network := range containerJSON.NetworkSettings.Networks {
			runtimeInfo.Networks[name] = compute.NetworkRuntimeInfo{
				InterfaceName: name,
				IPAddress:     network.IPAddress,
				IPv6Address:   network.GlobalIPv6Address,
				MacAddress:    network.MacAddress,
			}
		}
	}

	return runtimeInfo
}

func (s *BackendService) convertResourceUsage(stats container.Stats) *compute.ResourceUsage {
	usage := &compute.ResourceUsage{
		Timestamp: stats.Read,
		CPU: compute.CPUUsage{
			UsageNanos:  int64(stats.CPUStats.CPUUsage.TotalUsage),
			SystemUsage: int64(stats.CPUStats.SystemUsage),
			OnlineCPUs:  int(stats.CPUStats.OnlineCPUs),
		},
		Memory: compute.MemoryUsage{
			Usage:    int64(stats.MemoryStats.Usage),
			MaxUsage: int64(stats.MemoryStats.MaxUsage),
			Limit:    int64(stats.MemoryStats.Limit),
		},
	}

	// Calculate CPU percentage
	if stats.PreCPUStats.SystemUsage > 0 && stats.CPUStats.OnlineCPUs > 0 {
		cpuDelta := float64(stats.CPUStats.CPUUsage.TotalUsage - stats.PreCPUStats.CPUUsage.TotalUsage)
		systemDelta := float64(stats.CPUStats.SystemUsage - stats.PreCPUStats.SystemUsage)
		if systemDelta > 0 {
			usage.CPU.Usage = (cpuDelta / systemDelta) * float64(stats.CPUStats.OnlineCPUs) * 100.0
		}
	}

	// Calculate memory percentage
	if stats.MemoryStats.Limit > 0 {
		usage.Memory.UsagePercent = (float64(stats.MemoryStats.Usage) / float64(stats.MemoryStats.Limit)) * 100.0
	}

	// Network statistics
	for _, netStats := range stats.Networks {
		usage.Network.RxBytes += int64(netStats.RxBytes)
		usage.Network.TxBytes += int64(netStats.TxBytes)
		usage.Network.RxPackets += int64(netStats.RxPackets)
		usage.Network.TxPackets += int64(netStats.TxPackets)
	}

	// Storage statistics
	for _, bioStats := range stats.BlkioStats.IoServiceBytesRecursive {
		switch strings.ToLower(bioStats.Op) {
		case "read":
			usage.Storage.ReadBytes += int64(bioStats.Value)
		case "write":
			usage.Storage.WriteBytes += int64(bioStats.Value)
		}
	}

	return usage
}
