# LibGo Quick Reference

## Common Operations

### Authentication
```bash
# Login
TOKEN=$(curl -s -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"password"}' | jq -r .token)

# Use token in requests
curl -H "Authorization: Bearer $TOKEN" ...
```

### Unified Compute Operations

#### Create Instance
```bash
# Docker container
curl -X POST http://localhost:8080/api/v1/compute/instances \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "nginx-web",
    "backend": "docker",
    "config": {"image": "nginx:latest", "ports": ["80:80"]},
    "resources": {"cpu": {"cores": 1}, "memory": {"limit": 536870912}}
  }'

# KVM VM
curl -X POST http://localhost:8080/api/v1/compute/instances \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "ubuntu-vm",
    "backend": "kvm",
    "config": {"image": "ubuntu-22.04"},
    "resources": {"cpu": {"cores": 2}, "memory": {"limit": 2147483648}}
  }'
```

#### List All Instances
```bash
curl -H "Authorization: Bearer $TOKEN" \
  http://localhost:8080/api/v1/compute/instances
```

#### Start/Stop Instance
```bash
# Start
curl -X PUT -H "Authorization: Bearer $TOKEN" \
  http://localhost:8080/api/v1/compute/instances/{id}/start

# Stop
curl -X PUT -H "Authorization: Bearer $TOKEN" \
  http://localhost:8080/api/v1/compute/instances/{id}/stop
```

#### Delete Instance
```bash
curl -X DELETE -H "Authorization: Bearer $TOKEN" \
  http://localhost:8080/api/v1/compute/instances/{id}?force=true
```

### Docker-Specific Operations

#### Pull Image
```bash
curl -X POST http://localhost:8080/api/v1/docker/images/pull \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"image": "nginx", "tag": "latest"}'
```

#### Create Network
```bash
curl -X POST http://localhost:8080/api/v1/docker/networks \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "app-network",
    "driver": "bridge",
    "ipam": {"config": [{"subnet": "172.20.0.0/16"}]}
  }'
```

#### Create Volume
```bash
curl -X POST http://localhost:8080/api/v1/docker/volumes \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"name": "app-data", "driver": "local"}'
```

#### Container Logs
```bash
curl -H "Authorization: Bearer $TOKEN" \
  "http://localhost:8080/api/v1/docker/containers/{id}/logs?tail=100"
```

### KVM-Specific Operations

#### Create VM with Cloud-Init
```bash
curl -X POST http://localhost:8080/api/v1/vms \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "web-server",
    "template": "ubuntu-22.04",
    "cpu": 2,
    "memory": 2048,
    "disk": 20,
    "cloudInit": {
      "userData": "#cloud-config\npackages:\n  - nginx\n"
    }
  }'
```

#### Create Snapshot
```bash
curl -X POST http://localhost:8080/api/v1/vms/{name}/snapshots \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"name": "before-upgrade", "description": "Snapshot before system upgrade"}'
```

#### Export VM
```bash
curl -X POST http://localhost:8080/api/v1/vms/{name}/export \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"format": "ova", "compression": true}'
```

### Monitoring

#### Health Check
```bash
curl http://localhost:8080/health
```

#### Metrics (Prometheus format)
```bash
curl http://localhost:8080/metrics
```

#### Resource Usage
```bash
# Single instance
curl -H "Authorization: Bearer $TOKEN" \
  http://localhost:8080/api/v1/compute/instances/{id}/usage

# Cluster overview
curl -H "Authorization: Bearer $TOKEN" \
  http://localhost:8080/api/v1/compute/cluster/status
```

### WebSocket Operations

#### VM Console
```javascript
const ws = new WebSocket('ws://localhost:8080/ws/vm/{name}');
ws.onmessage = (event) => console.log('VM Update:', event.data);
```

## Resource Specifications

### Memory Sizes (in bytes)
- 512 MB: `536870912`
- 1 GB: `1073741824`
- 2 GB: `2147483648`
- 4 GB: `4294967296`
- 8 GB: `8589934592`

### Storage Sizes (in bytes)
- 10 GB: `10737418240`
- 20 GB: `21474836480`
- 50 GB: `53687091200`
- 100 GB: `107374182400`

## Environment Variables

```bash
# Override config file settings
export LIBGO_SERVER_PORT=8080
export LIBGO_AUTH_ENABLED=true
export LIBGO_DOCKER_ENABLED=true
export LIBGO_DOCKER_HOST=unix:///var/run/docker.sock
export LIBGO_COMPUTE_DEFAULT_BACKEND=docker
```

## Troubleshooting

### Check Service Status
```bash
# LibGo logs
journalctl -u libgo -f

# Check backend status
curl -H "Authorization: Bearer $TOKEN" \
  http://localhost:8080/api/v1/compute/backends/docker/info

curl -H "Authorization: Bearer $TOKEN" \
  http://localhost:8080/api/v1/compute/backends/kvm/info
```

### Common Issues

1. **Connection Refused**
   - Check if LibGo is running: `systemctl status libgo`
   - Verify port in configuration

2. **Authentication Failed**
   - Token expired - login again
   - Check user credentials in config

3. **Backend Unavailable**
   - Docker: `systemctl status docker`
   - KVM: `systemctl status libvirtd`

4. **Resource Limits**
   - Check limits: `curl .../compute/cluster/status`
   - Adjust in configuration if needed