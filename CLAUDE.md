# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Development Commands

### Build and Run
- `make build` - Build the server binary to `bin/libgo-server`
- `make build-dev` - Build with race detection for development
- `./bin/libgo-server` - Run the server (use `./bin/libgo-server -config configs/config.yaml`)
- `make docker-build` - Build Docker image
- `make docker-run` - Run in Docker container

### Testing
- `make test` - Run all tests with coverage
- `make unit-test` - Run only unit tests (with `-short` flag)
- `make integration-test` - Run integration tests only
- `make test-ubuntu-docker` - Run the Ubuntu Docker deployment integration test
- `make test-ovs` - Run OpenVSwitch integration tests (requires OVS installation and root)
- `make coverage` - Generate HTML coverage report

### Code Quality
- `make lint` - Run golangci-lint on the codebase
- `make fmt` - Format code with go fmt and goimports
- `make vet` - Run go vet
- `make sec-scan` - Run security scan with gosec
- `make vuln-check` - Run vulnerability check with govulncheck

### Development Setup
- `make setup` - Install development tools (golangci-lint, gosec, govulncheck, mockgen)
- `make test-setup` - Install development tools and OpenVSwitch packages for testing
- `make install-ovs` - Install OpenVSwitch packages only
- `make mocks` - Generate mock implementations for interfaces

## Architecture Overview

### Core Components
The application is a unified compute management API for both KVM/libvirt and Docker, built with layered dependency injection:

**Dependency Flow:** `main.go` → **Managers** → **Handlers** → **API Routes**

1. **Unified Compute Layer** (`internal/compute/`) - Abstraction for both VMs and containers
2. **Backend Implementations**:
   - **KVM/Libvirt Backend**:
     - `internal/libvirt/connection/` - Connection pool management
     - `internal/libvirt/domain/` - VM lifecycle operations
     - `internal/libvirt/storage/` - Disk and storage pool management  
     - `internal/libvirt/network/` - Network configuration
   - **Docker Backend** (`internal/docker/`):
     - `container/` - Container lifecycle and operations
     - `image/` - Image management
     - `network/` - Docker network management
     - `volume/` - Volume management
3. **SDN Management** (`internal/ovs/`) - OpenVSwitch software-defined networking
4. **Business Logic** (`internal/vm/`) - Orchestrates resource managers for VM operations
5. **API Layer** (`internal/api/`) - HTTP handlers and routing for unified and backend-specific APIs
6. **Export System** (`internal/export/`) - VM export to multiple formats (QCOW2, VMDK, VDI, OVA)

### Key Patterns
- **XML Template System**: All libvirt XML is generated from templates in `configs/templates/`
- **Interface-Based Design**: All major components implement interfaces for testability
- **Dependency Injection**: Components are wired together in `cmd/server/main.go`
- **Job System**: Long-running operations (exports) use async job tracking

### Configuration System
- Primary config: `configs/config.yaml.example` 
- Environment variable overrides supported
- Config loading in `internal/config/`
- Template configs for different environments in `configs/`

### Authentication & Security
- JWT-based authentication (`internal/auth/jwt/`)
- Role-based access control (`internal/auth/user/`)
- Database-backed user management with GORM (`internal/auth/user/gorm_service.go`)
- Middleware for auth, logging, recovery (`internal/middleware/`)

### Template & Cloud-Init System
- VM templates in JSON format (`configs/templates/`)
- Cloud-init integration for VM configuration (`internal/vm/cloudinit/`)
- Template manager loads and validates VM definitions (`internal/vm/template/`)

### Testing Strategy
- Unit tests for all major components (use `*_test.go` pattern)
- Integration tests in `test/integration/` that require actual libvirt
- Mock generation with `go.uber.org/mock` (run `make mocks`)
- Test helpers in `test/integration/` for setting up test environments

### Frontend Integration  
- React UI in `ui/` directory with TypeScript
- WebSocket support for real-time VM monitoring (`internal/websocket/`)
- Separate Dockerfile and build process for UI

### Key Files to Understand
- `cmd/server/main.go` - Application bootstrap and dependency wiring
- `internal/compute/manager.go` - Unified compute instance management
- `internal/vm/manager.go` - Core VM management orchestration
- `internal/docker/manager.go` - Docker client management
- `internal/docker/backend.go` - Docker backend service implementation
- `internal/export/manager.go` - VM export job management  
- `internal/ovs/manager.go` - OpenVSwitch bridge, port, and flow management
- `internal/ovs/libvirt_integration.go` - OVS-libvirt integration for VM networking
- `internal/api/router.go` - API route definitions
- `internal/api/router_adapter.go` - Route configuration and middleware setup
- `internal/libvirt/connection/manager.go` - Libvirt connection pooling

### Required Tools
- Go 1.24.0+
- Libvirt 9.0.0+ with development headers (for KVM support)
- Docker 24.0.0+ (for container management functionality)
- OpenVSwitch 2.13.0+ (for SDN functionality) - install with `make install-ovs`
- Various linting tools (installed via `make setup` or `make test-setup`)

### Database
Uses GORM with SQLite (default) or PostgreSQL support for user management and persistent storage.