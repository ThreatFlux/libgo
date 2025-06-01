# Creating a Windows VM Using the LibGo API

This guide demonstrates how to create a Windows VM from scratch using the LibGo KVM Management API with YAML templates.

## Prerequisites

1. LibGo server running with Windows template configured
2. Windows 11 ISO or Windows Server 2022 base image available
3. VirtIO drivers ISO (for optimal performance)
4. API access (authentication token if enabled)

## Step 1: Start the LibGo Server

First, ensure your LibGo server is running with the Windows configuration:

```bash
# Using the Windows 11 configuration (auth disabled for testing)
./bin/libgo-server -config configs/windows11-config.yaml
```

## Step 2: Create a Windows VM via API

### Using cURL

```bash
# Create a Windows 11 VM
curl -X POST http://localhost:8080/api/v1/vms \
  -H "Content-Type: application/json" \
  -d '{
    "name": "win11-desktop",
    "template": "windows-11",
    "description": "Windows 11 Desktop VM",
    "cpu": {
      "count": 4,
      "cores": 2,
      "threads": 1,
      "socket": 2
    },
    "memory": {
      "sizeBytes": 8589934592
    },
    "disk": {
      "sizeBytes": 107374182400,
      "format": "qcow2",
      "bus": "sata",
      "storagePool": "default"
    },
    "network": {
      "type": "network",
      "source": "default",
      "model": "virtio"
    },
    "cloudInit": {
      "enabled": false
    }
  }'
```

### Using Python

```python
import requests
import json

# API endpoint
api_url = "http://localhost:8080/api/v1/vms"

# VM configuration
vm_config = {
    "name": "win11-desktop",
    "template": "windows-11",
    "description": "Windows 11 Desktop VM",
    "cpu": {
        "count": 4,
        "cores": 2,
        "threads": 1,
        "socket": 2
    },
    "memory": {
        "sizeBytes": 8589934592  # 8GB
    },
    "disk": {
        "sizeBytes": 107374182400,  # 100GB
        "format": "qcow2",
        "bus": "sata",
        "storagePool": "default"
    },
    "network": {
        "type": "network",
        "source": "default",
        "model": "virtio"
    },
    "cloudInit": {
        "enabled": False
    }
}

# Create the VM
response = requests.post(api_url, json=vm_config)

if response.status_code == 201:
    vm = response.json()['vm']
    print(f"VM created successfully!")
    print(f"Name: {vm['name']}")
    print(f"UUID: {vm['uuid']}")
    print(f"Status: {vm['status']}")
else:
    print(f"Error: {response.status_code}")
    print(response.json())
```

## Step 3: YAML Template Configuration

Create a custom YAML configuration file that includes Windows VM templates:

```yaml
# windows-vms-config.yaml
server:
  host: "0.0.0.0"
  port: 8080
  readTimeout: 30s
  writeTimeout: 30s

libvirt:
  uri: "qemu:///system"
  connectionTimeout: 10s
  maxConnections: 5
  poolName: "default"
  networkName: "default"

auth:
  enabled: false  # Set to true in production

logging:
  level: "info"
  format: "json"

storage:
  defaultPool: "default"
  poolPath: "/var/lib/libvirt/images"
  templates:
    # Windows 11 template
    windows-11: "/path/to/Win11_24H2_English_x64.iso"
    # Windows Server 2022 template
    windows-server-2022: "/path/to/windows-server-2022-base.qcow2"

features:
  cloudInit: true
  export: true
  metrics: true
  websocket: true

# VM creation parameters
vm:
  defaults:
    # Default values for Windows VMs
    windows:
      cpu_count: 4
      memory_gb: 8
      disk_gb: 100
      network_model: "virtio"
      disk_bus: "sata"  # Use SATA for initial install, switch to VirtIO later
```

## Step 4: Create Multiple Windows VMs with Different Configurations

### Windows 11 Development VM
```bash
curl -X POST http://localhost:8080/api/v1/vms \
  -H "Content-Type: application/json" \
  -d '{
    "name": "win11-dev",
    "template": "windows-11",
    "description": "Windows 11 Development Environment",
    "cpu": {"count": 8},
    "memory": {"sizeBytes": 17179869184},
    "disk": {"sizeBytes": 214748364800, "format": "qcow2", "bus": "virtio"}
  }'
```

### Windows Server 2022 with IIS
```bash
curl -X POST http://localhost:8080/api/v1/vms \
  -H "Content-Type: application/json" \
  -d '{
    "name": "winserver-web",
    "template": "windows-server-2022",
    "description": "Windows Server 2022 IIS Web Server",
    "cpu": {"count": 4},
    "memory": {"sizeBytes": 8589934592},
    "disk": {"sizeBytes": 85899345920, "format": "qcow2", "bus": "virtio"}
  }'
```

## Step 5: Start the VM

```bash
curl -X PUT http://localhost:8080/api/v1/vms/win11-desktop/start
```

## Step 6: Get VM Details

```bash
curl -X GET http://localhost:8080/api/v1/vms/win11-desktop
```

Response:
```json
{
  "vm": {
    "name": "win11-desktop",
    "uuid": "550e8400-e29b-41d4-a716-446655440000",
    "description": "Windows 11 Desktop VM",
    "cpu": {
      "count": 4,
      "model": "host-passthrough",
      "sockets": 2,
      "cores": 2,
      "threads": 1
    },
    "memory": {
      "size_bytes": 8589934592
    },
    "disks": [{
      "size_bytes": 107374182400,
      "format": "qcow2",
      "bus": "sata"
    }],
    "networks": [{
      "type": "network",
      "mac_address": "52:54:00:12:34:56",
      "source": "default",
      "model": "virtio"
    }],
    "status": "running",
    "created_at": "2025-05-30T12:00:00Z"
  }
}
```

## Step 7: Connect to the VM

### Via VNC
```bash
# Get VNC display number from VM details
vncviewer localhost:5900
```

### Via RDP (after Windows installation)
```bash
# Default RDP port is 3389
rdesktop localhost:3389 -u Administrator -p P@ssw0rd
```

## Advanced Configuration Options

### VM Parameters Reference

| Parameter | Type | Description | Windows Default |
|-----------|------|-------------|-----------------|
| `name` | string | VM name (required) | - |
| `template` | string | Template name | "windows-11" |
| `cpu.count` | int | Number of vCPUs | 4 |
| `cpu.cores` | int | Cores per socket | 2 |
| `cpu.threads` | int | Threads per core | 1 |
| `cpu.socket` | int | Number of sockets | 2 |
| `memory.sizeBytes` | int64 | Memory in bytes | 4294967296 (4GB) |
| `disk.sizeBytes` | int64 | Disk size in bytes | 64424509440 (60GB) |
| `disk.format` | string | Disk format | "qcow2" |
| `disk.bus` | string | Disk bus type | "sata" or "virtio" |
| `network.type` | string | Network type | "network" |
| `network.source` | string | Network source | "default" |
| `network.model` | string | NIC model | "virtio" |

### Recommended Specifications

**Windows 11 Desktop:**
- CPU: 4-8 vCPUs
- Memory: 8-16 GB
- Disk: 100-200 GB
- Network: VirtIO

**Windows Server 2022:**
- CPU: 4-16 vCPUs
- Memory: 8-32 GB
- Disk: 80-500 GB
- Network: VirtIO

## Automation Script

Create a script to automate Windows VM creation:

```bash
#!/bin/bash
# create-windows-vms.sh

API_URL="http://localhost:8080/api/v1/vms"

# Function to create VM
create_vm() {
    local name=$1
    local template=$2
    local cpu=$3
    local memory=$4
    local disk=$5
    
    echo "Creating VM: $name"
    
    curl -s -X POST $API_URL \
        -H "Content-Type: application/json" \
        -d @- <<EOF
{
    "name": "$name",
    "template": "$template",
    "cpu": {"count": $cpu},
    "memory": {"sizeBytes": $memory},
    "disk": {"sizeBytes": $disk, "format": "qcow2", "bus": "virtio"}
}
EOF
}

# Create multiple Windows VMs
create_vm "win11-office" "windows-11" 4 8589934592 107374182400
create_vm "win11-gaming" "windows-11" 8 17179869184 214748364800
create_vm "winserver-ad" "windows-server-2022" 4 8589934592 85899345920
create_vm "winserver-sql" "windows-server-2022" 8 17179869184 214748364800
```

## Troubleshooting

### Common Issues

1. **VM Creation Fails**
   - Check if the template ISO/image exists
   - Verify storage pool has enough space
   - Check libvirt connection

2. **Performance Issues**
   - Install VirtIO drivers in Windows
   - Enable CPU host-passthrough
   - Allocate sufficient memory

3. **Network Connectivity**
   - Verify network bridge configuration
   - Check Windows Firewall settings
   - Install VirtIO network drivers

### Useful Commands

```bash
# List all VMs
curl http://localhost:8080/api/v1/vms

# Stop a VM
curl -X PUT http://localhost:8080/api/v1/vms/win11-desktop/stop

# Delete a VM
curl -X DELETE http://localhost:8080/api/v1/vms/win11-desktop

# Export VM to different format
curl -X POST http://localhost:8080/api/v1/vms/win11-desktop/export \
  -H "Content-Type: application/json" \
  -d '{"format": "vmdk"}'
```

## Next Steps

1. Set up unattended Windows installation with AutoUnattend.xml
2. Configure VirtIO drivers for better performance
3. Create snapshots before major changes
4. Set up automated backups using the export API
5. Monitor VM performance using the metrics API