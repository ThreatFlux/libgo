# LibGo Unified Compute Management API - Overview

The LibGo API provides a unified RESTful interface for managing both KVM virtual machines and Docker containers through a single, consistent API. This document provides an overview of the API's capabilities, design principles, and general usage information.

## API Features

### Unified Compute Management
- **Unified Instance API**: Single API for managing both VMs and containers
- **Backend Abstraction**: Seamless switching between KVM and Docker backends
- **Resource Management**: Consistent resource allocation and limits across backends
- **Mixed Workloads**: Support for running VMs and containers side-by-side

### KVM Virtual Machine Features
- **VM Lifecycle Management**: Create, start, stop, and delete virtual machines
- **VM Configuration**: Configure CPU, memory, storage, and networking
- **Cloud-Init Integration**: Customize VM deployments using cloud-init
- **VM Export**: Export VMs to various formats (QCOW2, VMDK, VDI, OVA, RAW)
- **Snapshot Management**: Create, revert, and manage VM snapshots
- **OVS Integration**: Advanced networking with OpenVSwitch

### Docker Container Features
- **Container Lifecycle**: Create, start, stop, pause, unpause, and remove containers
- **Image Management**: Pull, push, build, tag, and manage Docker images
- **Network Management**: Create and manage Docker networks
- **Volume Management**: Persistent storage with Docker volumes
- **Container Operations**: Execute commands, view logs, and monitor stats

### Common Features
- **Authentication**: JWT-based authentication with role-based access control
- **Monitoring**: Metrics endpoint for Prometheus integration
- **Health Checking**: API health status for monitoring
- **WebSocket Support**: Real-time monitoring and console access
- **Audit Logging**: Comprehensive operation logging

## API Design Principles

The API follows these key design principles:

1. **RESTful**: Uses standard HTTP methods and status codes
2. **JSON-based**: All requests and responses use JSON format
3. **Versioned**: API endpoints include version number (/api/v1/...)
4. **Secure**: Authentication required for all endpoints except health and login
5. **Well-documented**: OpenAPI/Swagger documentation
6. **Consistent**: Uniform error responses and pagination

## Authentication

The API uses JWT (JSON Web Token) for authentication. To authenticate:

1. Make a POST request to `/login` with username and password
2. Receive a JWT token in the response
3. Include the token in the `Authorization` header for subsequent requests:
   ```
   Authorization: Bearer {token}
   ```

## Content Types

- Requests should include `Content-Type: application/json` 
- Responses are always `application/json`

## Error Handling

All error responses follow a consistent structure:

```json
{
  "status": 400,
  "code": "INVALID_INPUT",
  "message": "Detailed error message"
}
```

Common error codes:
- `NOT_FOUND`: Requested resource not found
- `INVALID_INPUT`: Invalid input parameters
- `UNAUTHORIZED`: Authentication required or failed
- `FORBIDDEN`: Authenticated but insufficient permissions
- `RESOURCE_CONFLICT`: Resource already exists
- `INTERNAL_SERVER_ERROR`: Server error

## Pagination

List endpoints support pagination parameters:
- `page`: Page number (default: 1)
- `pageSize`: Number of items per page (default: 50, max: 100)

Paginated responses include:
- `count`: Total number of items
- `page`: Current page number
- `pageSize`: Number of items per page

## API Versioning

The API uses URL versioning with the format `/api/v1/...`. This allows introducing breaking changes in future versions while maintaining backward compatibility.

## Rate Limiting

The API implements rate limiting to prevent abuse. Clients exceeding rate limits will receive a `429 Too Many Requests` response.

## API Documentation

For detailed endpoint documentation, refer to [API Endpoints](endpoints.md) or view the OpenAPI documentation at `/swagger/index.html` when the server is running.
