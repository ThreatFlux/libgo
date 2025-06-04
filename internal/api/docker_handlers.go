package api

import "github.com/threatflux/libgo/internal/api/handlers"

// DockerHandlers groups all Docker-related handlers.
type DockerHandlers struct {
	// Container handlers
	CreateContainer   *handlers.DockerContainerHandler
	ListContainers    *handlers.DockerContainerHandler
	GetContainer      *handlers.DockerContainerHandler
	StartContainer    *handlers.DockerContainerHandler
	StopContainer     *handlers.DockerContainerHandler
	RestartContainer  *handlers.DockerContainerHandler
	DeleteContainer   *handlers.DockerContainerHandler
	GetContainerLogs  *handlers.DockerContainerHandler
	GetContainerStats *handlers.DockerContainerHandler

	// Image handlers (future)
	// Network handlers (future)
	// Volume handlers (future)
}
