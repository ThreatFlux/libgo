# Changelog

All notable changes to LibGo will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- **Unified Compute API**: New abstraction layer for managing both KVM VMs and Docker containers through a single API
- **Docker Integration**: Complete Docker support including:
  - Container lifecycle management (create, start, stop, restart, pause, unpause, remove)
  - Image management (pull, push, build, tag, remove, prune)
  - Network management (create, remove, connect/disconnect containers)
  - Volume management (create, remove, list, prune)
  - Container operations (exec, logs, stats, file operations)
- **Compute Endpoints**: New `/api/v1/compute/*` endpoints for unified management
- **Docker-specific API**: Direct Docker management via `/api/v1/docker/*` endpoints
- **Backend Selection**: Support for choosing between KVM and Docker backends per instance
- **Mixed Workloads**: Ability to run VMs and containers side-by-side
- **Resource Management**: Unified resource allocation and monitoring across backends
- **Configuration**: New Docker and Compute configuration sections
- **Documentation**: Comprehensive documentation for unified compute and Docker APIs

### Changed
- Renamed project from "KVM VM Management API" to "LibGo - Unified Compute Management API"
- Updated authentication to support both backend types
- Enhanced metrics collection to include Docker containers
- Improved error handling with backend-specific error codes

### Infrastructure
- Added Docker client v28.1.1 compatibility with Go 1.24.2
- Implemented modular service architecture for Docker operations
- Created comprehensive test structure for Docker services
- Added example configurations demonstrating Docker integration

## [Previous Versions]

### Security Fixes
- Resolved security scan issues and eliminated code duplication
- Fixed critical linting issues including unnecessary conversions and copy locks

### Code Quality
- Comprehensive linting improvements and code quality enhancements
- Added pre-commit hooks for automatic code formatting and linting
- Resolved major linting issues across the codebase