# LibGo - Unified Compute Management API

A comprehensive RESTful API for managing both KVM virtual machines and Docker containers through a unified interface. LibGo provides a single-node hypervisor management solution that treats VMs and containers as compute instances, offering consistent APIs for lifecycle management, storage operations, network management, and more. The API is designed with security, scalability, and performance in mind.

## Features

### Unified Compute Management
- **Dual Backend Support**: Manage both KVM VMs and Docker containers through a single API
- **Unified Resource Management**: Consistent resource limits and monitoring across both backends
- **Backend Flexibility**: Configure default backend and allow mixed workloads

### KVM Virtual Machine Features
- **Complete VM Lifecycle Management**: Create, start, stop, restart, and delete virtual machines
- **Storage Operations**: Create, resize, clone, and delete VM disks
- **Network Management**: Configure VM networking with DHCP support
- **VM Export**: Export VMs to multiple formats (QCOW2, VMDK, VDI, OVA)
- **Template Support**: Create VMs from templates
- **Cloud-Init Integration**: Configure VMs with cloud-init
- **Snapshot Management**: Create, list, revert, and delete VM snapshots
- **OVS Integration**: OpenVSwitch support for advanced networking

### Docker Container Features
- **Container Lifecycle Management**: Create, start, stop, restart, pause, unpause, and remove containers
- **Image Management**: Pull, push, build, tag, and remove Docker images
- **Network Management**: Create and manage Docker networks, connect/disconnect containers
- **Volume Management**: Create, list, inspect, and remove Docker volumes
- **Container Operations**: Execute commands, view logs, monitor stats, copy files
- **Prune Operations**: Clean up unused containers, images, networks, and volumes

### Security & Operations
- **Authentication**: JWT-based authentication with role-based access control
- **Monitoring**: Prometheus metrics and health checks for both VMs and containers
- **WebSocket Support**: Real-time monitoring and console access
- **Audit Logging**: Comprehensive logging of all operations

## Requirements

- Go 1.24.0 or later
- Libvirt 9.0.0 or later with development headers (for KVM support)
- Docker 24.0.0 or later (for Docker container management)
- OpenVSwitch 2.13.0+ (optional, for advanced networking)
- SQLite3 or PostgreSQL (for user management and persistent storage)

## Installation

```bash
# Clone the repository
git clone https://github.com/threatflux/libgo.git
cd libgo

# Build the application
make build

# Run the application
./bin/libgo-server
```

## Quick Start

### Running with Docker Support

```bash
# Copy and customize the configuration
cp configs/config.yaml.example configs/config.yaml

# Edit the configuration to enable Docker
# Set docker.enabled: true in the config file

# Run the server
./bin/libgo-server -config configs/config.yaml
```

### Creating a Compute Instance

```bash
# Create a Docker container
curl -X POST http://localhost:8080/api/v1/compute/instances \
  -H "Content-Type: application/json" \
  -d '{
    "name": "my-container",
    "backend": "docker",
    "config": {
      "image": "nginx:latest"
    },
    "resources": {
      "cpu": {"cores": 1},
      "memory": {"limit": 536870912}
    }
  }'

# Create a KVM VM
curl -X POST http://localhost:8080/api/v1/compute/instances \
  -H "Content-Type: application/json" \
  -d '{
    "name": "my-vm",
    "backend": "kvm",
    "config": {
      "image": "ubuntu-22.04"
    },
    "resources": {
      "cpu": {"cores": 2},
      "memory": {"limit": 2147483648}
    }
  }'
```

## Configuration

Configuration is managed through a YAML file (`configs/config.yaml`) and can be overridden with environment variables. Key configuration sections include:

- **Server**: HTTP server settings (host, port, TLS)
- **Libvirt**: KVM/libvirt connection settings
- **Docker**: Docker daemon connection settings
- **Compute**: Unified compute resource limits and backend configuration
- **Auth**: JWT authentication settings
- **Storage**: Storage pools and paths
- **Network**: Network configuration for both KVM and Docker
- **OVS**: OpenVSwitch settings (if enabled)

See the [example configuration](configs/config.yaml.example) or [test configuration with Docker](configs/test-docker-config.yaml) for available options.

## API Documentation

API documentation is available at `/docs/api/` when the server is running, or see the [API documentation](docs/api/README.md) in this repository.

### Core Endpoints

#### Unified Compute API
- **List Instances**: `GET /api/v1/compute/instances`
- **Create Instance**: `POST /api/v1/compute/instances`
- **Get Instance**: `GET /api/v1/compute/instances/:id`
- **Update Instance**: `PUT /api/v1/compute/instances/:id`
- **Delete Instance**: `DELETE /api/v1/compute/instances/:id`
- **Start Instance**: `PUT /api/v1/compute/instances/:id/start`
- **Stop Instance**: `PUT /api/v1/compute/instances/:id/stop`
- **Restart Instance**: `PUT /api/v1/compute/instances/:id/restart`
- **Get Instance by Name**: `GET /api/v1/compute/instances/name/:name`
- **Cluster Status**: `GET /api/v1/compute/cluster/status`
- **Backend Info**: `GET /api/v1/compute/backends/:backend/info`

#### KVM Virtual Machine API
- **List VMs**: `GET /api/v1/vms`
- **Create VM**: `POST /api/v1/vms`
- **Get VM Details**: `GET /api/v1/vms/:name`
- **Delete VM**: `DELETE /api/v1/vms/:name`
- **Start VM**: `PUT /api/v1/vms/:name/start`
- **Stop VM**: `PUT /api/v1/vms/:name/stop`
- **Export VM**: `POST /api/v1/vms/:name/export`
- **Snapshot Operations**: `/api/v1/vms/:name/snapshots/*`

#### Docker Container API
- **Container Management**: `/api/v1/docker/containers/*`
- **Image Management**: `/api/v1/docker/images/*`
- **Network Management**: `/api/v1/docker/networks/*`
- **Volume Management**: `/api/v1/docker/volumes/*`

#### Infrastructure APIs
- **Storage Pools**: `/api/v1/storage/pools/*`
- **Networks**: `/api/v1/networks/*`
- **OVS Bridges**: `/api/v1/ovs/bridges/*`
- **Authentication**: `/api/v1/auth/*`
- **Export Jobs**: `/api/v1/exports/:id`
- **Health Check**: `/health`
- **Metrics**: `/metrics`
- **WebSocket**: `/ws/vm/:name`

## Development

```bash
# Set up development environment
make setup

# Run tests
make test

# Run integration tests
make integration-test

# Generate test coverage report
make coverage

# Run linting
make lint

# Run security scanning
make security-check

# Build for development
make build-dev

# Build Docker image
make docker-build
```

## Project Structure

- `cmd/`: Application entry points
- `configs/`: Configuration files and templates
- `internal/`: Internal packages
  - `api/`: API handlers and server
  - `auth/`: Authentication and authorization
  - `compute/`: Unified compute management layer
  - `docker/`: Docker integration
    - `container/`: Container service implementation
    - `image/`: Image service implementation
    - `network/`: Network service implementation
    - `volume/`: Volume service implementation
  - `export/`: VM export functionality
  - `libvirt/`: Libvirt/KVM integration
    - `connection/`: Connection pool management
    - `domain/`: VM domain management
    - `network/`: Network management
    - `storage/`: Storage pool and volume management
  - `models/`: Data models
  - `ovs/`: OpenVSwitch integration
  - `vm/`: VM management logic
  - `websocket/`: WebSocket support for real-time operations
- `pkg/`: Reusable packages
  - `logger/`: Structured logging
  - `utils/`: Common utilities
- `test/`: Test utilities and mocks
- `ui/`: Web UI (React/TypeScript)

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.
