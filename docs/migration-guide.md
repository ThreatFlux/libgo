# Migration Guide

## Overview

This guide helps you migrate from standalone KVM or Docker setups to LibGo's unified compute management platform. LibGo provides a single API to manage both VMs and containers, making it easier to operate mixed workloads.

## Why Migrate to LibGo?

- **Unified Management**: Single API for both VMs and containers
- **Consistent Operations**: Same commands for starting, stopping, and managing instances
- **Resource Management**: Unified resource allocation and monitoring
- **Enhanced Security**: Built-in authentication and RBAC
- **Better Observability**: Integrated metrics and logging

## Migrating from Standalone KVM/Libvirt

### Before Migration

With standalone libvirt:
```bash
virsh create domain.xml
virsh start my-vm
virsh shutdown my-vm
```

### After Migration with LibGo

Using the unified API:
```bash
# Create a VM
curl -X POST http://localhost:8080/api/v1/compute/instances \
  -H "Content-Type: application/json" \
  -d '{
    "name": "my-vm",
    "backend": "kvm",
    "config": {"image": "ubuntu-22.04"},
    "resources": {"cpu": {"cores": 2}, "memory": {"limit": 2147483648}}
  }'

# Or use the KVM-specific API
curl -X POST http://localhost:8080/api/v1/vms \
  -H "Content-Type: application/json" \
  -d '{
    "name": "my-vm",
    "template": "ubuntu-22.04",
    "cpu": 2,
    "memory": 2048
  }'
```

### Key Differences

1. **Template-Based Creation**: LibGo uses templates instead of raw XML
2. **API-First**: All operations through REST API instead of CLI
3. **Authentication**: JWT tokens required for all operations
4. **Resource Tracking**: Automatic resource usage monitoring

### Migration Steps

1. **Export Existing VMs** (if needed):
   ```bash
   virsh dumpxml my-vm > my-vm.xml
   ```

2. **Create VM Templates** in LibGo:
   - Place base images in the configured storage pool
   - Define templates in `configs/templates/`

3. **Configure LibGo**:
   ```yaml
   libvirt:
     uri: "qemu:///system"
     poolName: "default"
     networkName: "default"
   ```

4. **Import VMs** using the API:
   - Create new VMs based on existing configurations
   - Attach existing disks if needed

## Migrating from Standalone Docker

### Before Migration

With standalone Docker:
```bash
docker run -d --name nginx -p 80:80 nginx:latest
docker stop nginx
docker rm nginx
```

### After Migration with LibGo

Using the unified API:
```bash
# Create a container
curl -X POST http://localhost:8080/api/v1/compute/instances \
  -H "Content-Type: application/json" \
  -d '{
    "name": "nginx",
    "backend": "docker",
    "config": {
      "image": "nginx:latest",
      "ports": ["80:80"]
    },
    "resources": {"cpu": {"cores": 1}, "memory": {"limit": 536870912}}
  }'

# Or use the Docker-specific API
curl -X POST http://localhost:8080/api/v1/docker/containers \
  -H "Content-Type: application/json" \
  -d '{
    "name": "nginx",
    "image": "nginx:latest",
    "ports": ["80:80"]
  }'
```

### Key Differences

1. **API-Based Management**: REST API instead of Docker CLI
2. **Resource Limits**: Explicit resource allocation
3. **Unified Monitoring**: Integrated metrics for all containers
4. **Authentication**: JWT tokens for all operations

### Migration Steps

1. **Enable Docker in LibGo**:
   ```yaml
   docker:
     enabled: true
     host: "unix:///var/run/docker.sock"
   ```

2. **List Existing Containers**:
   ```bash
   docker ps -a --format "table {{.Names}}\t{{.Image}}\t{{.Status}}"
   ```

3. **Recreate Containers** via LibGo API:
   - Use the same images and configurations
   - Map volumes and networks appropriately

4. **Migrate Docker Compose** stacks:
   - Parse compose files
   - Create equivalent LibGo API calls
   - Maintain service dependencies

## Unified Operations

Once migrated, you can manage both VMs and containers uniformly:

### List All Instances
```bash
curl http://localhost:8080/api/v1/compute/instances \
  -H "Authorization: Bearer $TOKEN"
```

### Start Any Instance
```bash
# Works for both VMs and containers
curl -X PUT http://localhost:8080/api/v1/compute/instances/:id/start \
  -H "Authorization: Bearer $TOKEN"
```

### Monitor Resources
```bash
# Unified resource usage for both types
curl http://localhost:8080/api/v1/compute/instances/:id/usage \
  -H "Authorization: Bearer $TOKEN"
```

## Best Practices

1. **Use the Unified API** when possible for consistency
2. **Set Resource Limits** explicitly for better resource management
3. **Enable Authentication** in production environments
4. **Use Templates** for VM creation to ensure consistency
5. **Monitor Metrics** through the unified `/metrics` endpoint

## Troubleshooting

### Connection Issues

If LibGo can't connect to libvirt or Docker:

1. Check service status:
   ```bash
   systemctl status libvirtd
   systemctl status docker
   ```

2. Verify socket permissions:
   ```bash
   ls -la /var/run/libvirt/libvirt-sock
   ls -la /var/run/docker.sock
   ```

3. Check LibGo logs:
   ```bash
   journalctl -u libgo -f
   ```

### Permission Errors

Ensure the LibGo service user has appropriate permissions:

1. For libvirt: Add to `libvirt` group
2. For Docker: Add to `docker` group
3. For storage: Ensure write access to configured paths

### Resource Conflicts

If you see resource allocation errors:

1. Check current usage:
   ```bash
   curl http://localhost:8080/api/v1/compute/cluster/status
   ```

2. Adjust resource limits in configuration
3. Enable overcommit if appropriate

## Support

For additional help:
- Check the [API Documentation](api/README.md)
- Review [example configurations](../configs/)
- Submit issues on GitHub