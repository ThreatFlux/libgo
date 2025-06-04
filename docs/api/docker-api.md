# Docker API Documentation

## Overview

The Docker API provides direct access to Docker-specific functionality for container, image, network, and volume management. While the Unified Compute API provides a abstracted interface, these endpoints give you full control over Docker resources.

## Container Management

### List Containers
```
GET /api/v1/docker/containers
```

Query parameters:
- `all`: Show all containers (default shows only running)
- `limit`: Limit number of containers returned
- `filters`: JSON-encoded filters (e.g., `{"status":["running"]}`)

### Create Container
```
POST /api/v1/docker/containers
```

Request body:
```json
{
  "name": "my-container",
  "image": "nginx:latest",
  "env": ["KEY=value"],
  "volumes": ["/host/path:/container/path"],
  "ports": ["80:80", "443:443"],
  "networkMode": "bridge",
  "restartPolicy": "unless-stopped",
  "command": ["/bin/sh", "-c", "nginx -g 'daemon off;'"],
  "workingDir": "/app",
  "labels": {
    "app": "web",
    "env": "production"
  }
}
```

### Get Container Details
```
GET /api/v1/docker/containers/:id
```

### Start Container
```
POST /api/v1/docker/containers/:id/start
```

### Stop Container
```
POST /api/v1/docker/containers/:id/stop
```

Request body (optional):
```json
{
  "timeout": 30
}
```

### Restart Container
```
POST /api/v1/docker/containers/:id/restart
```

### Delete Container
```
DELETE /api/v1/docker/containers/:id?force=true
```

### Get Container Logs
```
GET /api/v1/docker/containers/:id/logs
```

Query parameters:
- `stdout`: Include stdout logs (default: true)
- `stderr`: Include stderr logs (default: true)
- `since`: Show logs since timestamp
- `until`: Show logs before timestamp
- `timestamps`: Show timestamps (default: false)
- `tail`: Number of lines to show from end of logs

### Get Container Stats
```
GET /api/v1/docker/containers/:id/stats
```

Query parameters:
- `stream`: Stream stats (default: true)

## Image Management

### List Images
```
GET /api/v1/docker/images
```

Query parameters:
- `all`: Show all images including intermediate layers
- `filters`: JSON-encoded filters

### Pull Image
```
POST /api/v1/docker/images/pull
```

Request body:
```json
{
  "image": "nginx",
  "tag": "latest"
}
```

### Inspect Image
```
GET /api/v1/docker/images/:id
```

### Remove Image
```
DELETE /api/v1/docker/images/:id
```

Query parameters:
- `force`: Force removal
- `prune`: Remove untagged parents

### Prune Images
```
POST /api/v1/docker/images/prune
```

Removes all unused images.

## Network Management

### List Networks
```
GET /api/v1/docker/networks
```

### Create Network
```
POST /api/v1/docker/networks
```

Request body:
```json
{
  "name": "my-network",
  "driver": "bridge",
  "internal": false,
  "attachable": true,
  "enable_ipv6": false,
  "ipam": {
    "driver": "default",
    "config": [
      {
        "subnet": "172.20.0.0/16",
        "gateway": "172.20.0.1"
      }
    ]
  },
  "labels": {
    "app": "web"
  }
}
```

### Inspect Network
```
GET /api/v1/docker/networks/:id
```

### Remove Network
```
DELETE /api/v1/docker/networks/:id
```

### Connect Container to Network
```
POST /api/v1/docker/networks/:id/connect
```

Request body:
```json
{
  "container": "container-id",
  "endpoint_config": {
    "ipv4_address": "172.20.0.2"
  }
}
```

### Disconnect Container from Network
```
POST /api/v1/docker/networks/:id/disconnect
```

Request body:
```json
{
  "container": "container-id",
  "force": false
}
```

### Prune Networks
```
POST /api/v1/docker/networks/prune
```

## Volume Management

### List Volumes
```
GET /api/v1/docker/volumes
```

### Create Volume
```
POST /api/v1/docker/volumes
```

Request body:
```json
{
  "name": "my-volume",
  "driver": "local",
  "driver_opts": {
    "type": "nfs",
    "o": "addr=192.168.1.1,rw",
    "device": ":/path/to/dir"
  },
  "labels": {
    "app": "database"
  }
}
```

### Inspect Volume
```
GET /api/v1/docker/volumes/:name
```

### Remove Volume
```
DELETE /api/v1/docker/volumes/:name
```

Query parameters:
- `force`: Force removal even if in use

### Prune Volumes
```
POST /api/v1/docker/volumes/prune
```

## Examples

### Create and Start an Nginx Container
```bash
# Create the container
curl -X POST http://localhost:8080/api/v1/docker/containers \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{
    "name": "nginx-web",
    "image": "nginx:alpine",
    "ports": ["8080:80"],
    "volumes": ["/var/www:/usr/share/nginx/html:ro"],
    "restartPolicy": "unless-stopped"
  }'

# Start the container
curl -X POST http://localhost:8080/api/v1/docker/containers/nginx-web/start \
  -H "Authorization: Bearer $TOKEN"
```

### Create a Custom Network and Connect Container
```bash
# Create network
curl -X POST http://localhost:8080/api/v1/docker/networks \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{
    "name": "app-network",
    "driver": "bridge",
    "ipam": {
      "config": [{
        "subnet": "172.25.0.0/16"
      }]
    }
  }'

# Connect container to network
curl -X POST http://localhost:8080/api/v1/docker/networks/app-network/connect \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{
    "container": "nginx-web",
    "endpoint_config": {
      "ipv4_address": "172.25.0.10"
    }
  }'
```

### Stream Container Logs
```bash
curl -X GET "http://localhost:8080/api/v1/docker/containers/nginx-web/logs?tail=100&timestamps=true" \
  -H "Authorization: Bearer $TOKEN"
```

### Monitor Container Stats
```bash
curl -X GET "http://localhost:8080/api/v1/docker/containers/nginx-web/stats?stream=false" \
  -H "Authorization: Bearer $TOKEN"
```

## Error Handling

Docker-specific errors include:
- `IMAGE_NOT_FOUND`: Requested image not found
- `CONTAINER_NOT_FOUND`: Container with given ID not found
- `NETWORK_NOT_FOUND`: Network not found
- `VOLUME_NOT_FOUND`: Volume not found
- `CONTAINER_ALREADY_STARTED`: Container is already running
- `CONTAINER_NOT_RUNNING`: Container is not running
- `IMAGE_IN_USE`: Cannot remove image, it's being used
- `NETWORK_IN_USE`: Cannot remove network, containers are connected