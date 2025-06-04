# Unified Compute API Documentation

## Overview

The Unified Compute API provides a single, consistent interface for managing both KVM virtual machines and Docker containers. This abstraction layer treats both VMs and containers as "compute instances", allowing for seamless management regardless of the underlying technology.

## Key Concepts

### Compute Instance
A compute instance is an abstraction that represents either a KVM VM or a Docker container. Each instance has:
- Unique ID
- Name
- Backend type (KVM or Docker)
- State (running, stopped, paused, etc.)
- Resource allocation (CPU, memory, storage)
- Configuration specific to the backend

### Backends
LibGo supports two backends:
- **KVM**: For full virtual machines with complete OS isolation
- **Docker**: For lightweight containers sharing the host kernel

## API Endpoints

### List Compute Instances
```
GET /api/v1/compute/instances
```

Query parameters:
- `backend`: Filter by backend type (`kvm` or `docker`)
- `state`: Filter by instance state
- `page`: Page number (default: 1)
- `pageSize`: Items per page (default: 50)

Response:
```json
{
  "instances": [
    {
      "id": "instance-123",
      "name": "web-server",
      "uuid": "550e8400-e29b-41d4-a716-446655440000",
      "type": "container",
      "backend": "docker",
      "state": "running",
      "status": "healthy",
      "config": {
        "image": "nginx:latest",
        "env": ["ENV=production"]
      },
      "resources": {
        "cpu": {"cores": 2},
        "memory": {"limit": 1073741824, "used": 524288000}
      },
      "createdAt": "2024-06-04T10:00:00Z",
      "updatedAt": "2024-06-04T10:05:00Z"
    }
  ],
  "total": 15,
  "page": 1,
  "pageSize": 50
}
```

### Create Compute Instance
```
POST /api/v1/compute/instances
```

Request body:
```json
{
  "name": "my-instance",
  "backend": "docker",  // or "kvm"
  "config": {
    "image": "ubuntu:22.04",
    "env": ["KEY=value"],
    "volumes": ["/data:/data"]
  },
  "resources": {
    "cpu": {"cores": 2},
    "memory": {"limit": 2147483648},
    "storage": {"totalSpace": 10737418240}
  },
  "metadata": {
    "labels": {"app": "web", "env": "prod"}
  }
}
```

### Get Compute Instance
```
GET /api/v1/compute/instances/:id
```

### Update Compute Instance
```
PUT /api/v1/compute/instances/:id
```

Request body:
```json
{
  "resources": {
    "cpu": {"cores": 4},
    "memory": {"limit": 4294967296}
  }
}
```

### Delete Compute Instance
```
DELETE /api/v1/compute/instances/:id?force=true
```

### Start Instance
```
PUT /api/v1/compute/instances/:id/start
```

### Stop Instance
```
PUT /api/v1/compute/instances/:id/stop
```

Request body (optional):
```json
{
  "force": false,
  "timeout": 30
}
```

### Restart Instance
```
PUT /api/v1/compute/instances/:id/restart
```

### Pause Instance (Docker only)
```
PUT /api/v1/compute/instances/:id/pause
```

### Unpause Instance (Docker only)
```
PUT /api/v1/compute/instances/:id/unpause
```

### Get Resource Usage
```
GET /api/v1/compute/instances/:id/usage
```

Response:
```json
{
  "cpu": {
    "usage": 45.2,
    "cores": 2
  },
  "memory": {
    "used": 1073741824,
    "limit": 2147483648,
    "percentage": 50.0
  },
  "storage": {
    "used": 5368709120,
    "total": 10737418240
  },
  "network": {
    "rxBytes": 1024000,
    "txBytes": 2048000
  }
}
```

### Get Instance by Name
```
GET /api/v1/compute/instances/name/:name
```

### Get Cluster Status
```
GET /api/v1/compute/cluster/status
```

Response:
```json
{
  "backends": {
    "kvm": {
      "status": "healthy",
      "instances": 5,
      "resources": {
        "cpu": {"used": 10, "total": 32},
        "memory": {"used": 21474836480, "total": 68719476736}
      }
    },
    "docker": {
      "status": "healthy",
      "instances": 12,
      "resources": {
        "cpu": {"used": 8, "total": 32},
        "memory": {"used": 8589934592, "total": 68719476736}
      }
    }
  },
  "totalInstances": 17,
  "health": "healthy"
}
```

### Get Backend Info
```
GET /api/v1/compute/backends/:backend/info
```

Response:
```json
{
  "type": "docker",
  "version": "24.0.7",
  "apiVersion": "1.43",
  "status": "running",
  "capabilities": ["containers", "images", "networks", "volumes"],
  "supportedTypes": ["container"],
  "healthCheck": {
    "status": "healthy",
    "lastCheck": "2024-06-04T10:00:00Z"
  }
}
```

## Backend-Specific Configuration

### KVM Configuration
When creating a KVM instance, the config object supports:
```json
{
  "image": "ubuntu-22.04",           // Template name
  "diskFormat": "qcow2",            // Disk format
  "networkType": "bridge",          // Network type
  "networkSource": "default",       // Network name
  "cloudInit": {                    // Optional cloud-init
    "userData": "...",
    "metaData": "..."
  }
}
```

### Docker Configuration
When creating a Docker instance, the config object supports:
```json
{
  "image": "nginx:latest",          // Docker image
  "command": ["/bin/bash"],         // Override command
  "env": ["KEY=value"],             // Environment variables
  "volumes": ["/host:/container"],  // Volume mounts
  "ports": ["80:80", "443:443"],   // Port mappings
  "networkMode": "bridge",          // Network mode
  "restartPolicy": "unless-stopped" // Restart policy
}
```

## Error Responses

All errors follow the standard format:
```json
{
  "error": {
    "code": "BACKEND_UNAVAILABLE",
    "message": "Docker backend is not available",
    "details": {
      "backend": "docker",
      "reason": "Docker daemon not responding"
    }
  }
}
```

Common error codes:
- `BACKEND_UNAVAILABLE`: Requested backend is not available
- `INVALID_BACKEND`: Invalid backend specified
- `RESOURCE_LIMIT_EXCEEDED`: Resource request exceeds limits
- `INSTANCE_NOT_FOUND`: Instance with given ID not found
- `OPERATION_NOT_SUPPORTED`: Operation not supported for backend
- `BACKEND_ERROR`: Backend-specific error occurred

## Examples

### Create a Docker Container
```bash
curl -X POST http://localhost:8080/api/v1/compute/instances \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{
    "name": "nginx-web",
    "backend": "docker",
    "config": {
      "image": "nginx:alpine",
      "ports": ["8080:80"],
      "env": ["NGINX_HOST=example.com"]
    },
    "resources": {
      "cpu": {"cores": 1},
      "memory": {"limit": 536870912}
    }
  }'
```

### Create a KVM Virtual Machine
```bash
curl -X POST http://localhost:8080/api/v1/compute/instances \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{
    "name": "ubuntu-vm",
    "backend": "kvm",
    "config": {
      "image": "ubuntu-22.04",
      "diskFormat": "qcow2",
      "cloudInit": {
        "userData": "#cloud-config\npackages:\n  - nginx\n"
      }
    },
    "resources": {
      "cpu": {"cores": 2},
      "memory": {"limit": 2147483648},
      "storage": {"totalSpace": 21474836480}
    }
  }'
```

### List All Running Instances
```bash
curl -X GET "http://localhost:8080/api/v1/compute/instances?state=running" \
  -H "Authorization: Bearer $TOKEN"
```

### Stop an Instance Gracefully
```bash
curl -X PUT http://localhost:8080/api/v1/compute/instances/instance-123/stop \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{
    "force": false,
    "timeout": 60
  }'
```