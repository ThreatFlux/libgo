# KVM VM Management API

A comprehensive RESTful API for managing KVM virtual machines with features for VM lifecycle management, storage operations, network management, and VM export in multiple formats. The API is designed with security, scalability, and performance in mind.

## Features

- **Complete VM Lifecycle Management**: Create, start, stop, restart, and delete virtual machines
- **Storage Operations**: Create, resize, clone, and delete VM disks
- **Network Management**: Configure VM networking with DHCP support
- **VM Export**: Export VMs to multiple formats (QCOW2, VMDK, VDI, OVA)
- **Template Support**: Create VMs from templates
- **Cloud-Init Integration**: Configure VMs with cloud-init
- **Authentication**: JWT-based authentication with role-based access control
- **Monitoring**: Prometheus metrics and health checks

## Requirements

- Go 1.24.0 or later
- Libvirt 9.0.0 or later with development headers
- Docker 24.0.0 or later (for containerization)

## Installation

```bash
# Clone the repository
git clone https://github.com/wroersma/libgo.git
cd libgo

# Build the application
make build

# Run the application
./bin/libgo-server
```

## Configuration

Configuration is managed through a YAML file (`configs/config.yaml`) and can be overridden with environment variables. See the example configuration file for available options.

## API Documentation

API documentation is available at `/docs/api/` when the server is running, or see the [API documentation](docs/api/README.md) in this repository.

### Endpoints

- **Authentication**: `/api/v1/login`
- **VM Management**: 
  - List VMs: `GET /api/v1/vms`
  - Create VM: `POST /api/v1/vms`
  - Get VM details: `GET /api/v1/vms/:name`
  - Delete VM: `DELETE /api/v1/vms/:name`
  - Start VM: `PUT /api/v1/vms/:name/start`
  - Stop VM: `PUT /api/v1/vms/:name/stop`
  - Export VM: `POST /api/v1/vms/:name/export`
- **Export Jobs**: `GET /api/v1/exports/:id`
- **Health Check**: `GET /health`
- **Metrics**: `GET /metrics`

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
  - `export/`: VM export functionality
  - `libvirt/`: Libvirt integration
  - `models/`: Data models
  - `vm/`: VM management logic
- `pkg/`: Reusable packages
- `test/`: Test utilities and mocks

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.
