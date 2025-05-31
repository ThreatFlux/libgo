# Windows VM Setup with LibGo API

This directory contains examples and scripts for creating Windows VMs using the LibGo KVM Management API.

## Prerequisites

1. **LibGo Server**: Ensure the LibGo server is installed and running
2. **Windows ISO/Images**: 
   - Windows 11 ISO: `Win11_24H2_English_x64.iso`
   - Windows Server 2022 QCOW2: `windows-server-2022-base.qcow2`
3. **VirtIO Drivers**: Download from [Fedora Project](https://fedorapeople.org/groups/virt/virtio-win/direct-downloads/stable-virtio/virtio-win.iso)
4. **Storage**: At least 200GB free space for multiple VMs
5. **Python 3.6+** (for Python scripts)

## Quick Start

### 1. Start LibGo Server

```bash
# Start with example Windows configuration
cd /path/to/libgo
./bin/libgo-server -config examples/windows-vm-setup/windows-vm-config.yaml

# Or use the simpler Windows 11 config (auth disabled)
./bin/libgo-server -config configs/windows11-config.yaml
```

### 2. Create a Single Windows VM

Using the bash script:
```bash
# Make the script executable
chmod +x create-windows-vm.sh

# Create a Windows 11 VM
./create-windows-vm.sh create

# Create and start the VM
./create-windows-vm.sh full

# List all VMs
./create-windows-vm.sh list
```

Using Python:
```bash
# Make the script executable
chmod +x create-multiple-vms.py

# Create a basic Windows 11 VM
./create-multiple-vms.py create my-windows-vm --template windows-11-basic

# Create and start
./create-multiple-vms.py create my-windows-vm --template windows-11-developer --start
```

### 3. Create Multiple VMs

```bash
# Create a set of different Windows VMs
./create-multiple-vms.py create-multiple

# This creates:
# - win11-workstation (Basic Windows 11)
# - win11-dev (Developer Windows 11)
# - winserver-web (Web Server)
# - winserver-db (Database Server)
```

## Available Templates

| Template Name | OS | CPU | Memory | Disk | Use Case |
|--------------|-----|-----|---------|------|----------|
| windows-11-basic | Windows 11 | 4 | 8GB | 100GB | General use |
| windows-11-developer | Windows 11 | 8 | 16GB | 250GB | Development |
| windows-server-web | Server 2022 | 4 | 8GB | 80GB | Web hosting |
| windows-server-database | Server 2022 | 8 | 32GB | 500GB | SQL Server |

## API Examples

### Using cURL

```bash
# Create VM
curl -X POST http://localhost:8080/api/v1/vms \
  -H "Content-Type: application/json" \
  -d @- <<EOF
{
  "name": "win11-test",
  "template": "windows-11",
  "cpu": {"count": 4},
  "memory": {"sizeBytes": 8589934592},
  "disk": {"sizeBytes": 107374182400, "format": "qcow2", "bus": "sata"}
}
EOF

# Start VM
curl -X PUT http://localhost:8080/api/v1/vms/win11-test/start

# Create snapshot
curl -X POST http://localhost:8080/api/v1/vms/win11-test/snapshots \
  -H "Content-Type: application/json" \
  -d '{"name": "clean-install", "description": "Fresh Windows installation"}'
```

### Using Python Client

```python
from create_multiple_vms import LibGoClient

# Create client
client = LibGoClient("http://localhost:8080/api/v1")

# Create VM
vm_config = {
    "name": "python-windows-vm",
    "template": "windows-11",
    "cpu": {"count": 4},
    "memory": {"sizeBytes": 8589934592},
    "disk": {"sizeBytes": 107374182400, "format": "qcow2"}
}

success, result = client.create_vm(vm_config)
if success:
    print(f"VM created: {result['vm']['uuid']}")
```

## Configuration Files

### windows-vm-config.yaml

Complete configuration file with:
- Windows-specific settings
- Template definitions
- Performance tuning
- Security settings
- Feature flags

Key sections:
```yaml
storage:
  templates:
    windows-11: "/path/to/Win11.iso"
    windows-server-2022: "/path/to/WinServer2022.qcow2"

vm:
  defaults:
    windows:
      cpu:
        count: 4
      memory:
        sizeGB: 8
      disk:
        sizeGB: 100
```

## Workflow Examples

### 1. Development Environment Setup

```bash
# Create developer VM
./create-multiple-vms.py create win11-dev \
  --template windows-11-developer \
  --description "Development Environment" \
  --start

# Wait for Windows installation...

# Create snapshot after setup
./create-multiple-vms.py snapshot win11-dev dev-tools-installed \
  --description "VS Code, Git, Docker installed"
```

### 2. Web Server Farm

```bash
# Create multiple web servers
for i in {1..3}; do
  ./create-multiple-vms.py create "web-server-$i" \
    --template windows-server-web \
    --description "IIS Web Server $i"
done

# Start all servers
for i in {1..3}; do
  ./create-multiple-vms.py start "web-server-$i"
done
```

### 3. Test Environment with Snapshots

```bash
# Create test VM
./create-windows-vm.sh create

# Start and complete Windows setup
./create-windows-vm.sh start

# Create baseline snapshot
./create-windows-vm.sh snapshot windows11-vm baseline

# Install software, test, then revert if needed
curl -X PUT http://localhost:8080/api/v1/vms/windows11-vm/snapshots/baseline/revert
```

## Customization

### Custom VM Configuration

Create a JSON file with VM specifications:

```json
[
  {
    "name": "custom-vm-1",
    "template": "windows-11-basic",
    "description": "Custom configuration VM"
  },
  {
    "name": "custom-vm-2",
    "template": "windows-server-database",
    "description": "SQL Server VM"
  }
]
```

Then create VMs:
```bash
./create-multiple-vms.py create-multiple --config-file custom-vms.json
```

### Environment Variables

- `API_BASE_URL`: Override API URL (default: http://localhost:8080/api/v1)
- `VM_NAME`: Default VM name for bash script
- `VM_TEMPLATE`: Default template name

Example:
```bash
export API_BASE_URL=http://192.168.1.100:8080/api/v1
export VM_NAME=my-windows-workstation
./create-windows-vm.sh create
```

## Troubleshooting

### Common Issues

1. **VM Creation Fails**
   ```bash
   # Check if storage pool exists
   virsh pool-list --all
   
   # Check available space
   df -h /var/lib/libvirt/images
   ```

2. **Windows Won't Boot**
   - Ensure UEFI/BIOS settings are correct
   - Check if VirtIO drivers are needed
   - Verify ISO path is correct

3. **Performance Issues**
   - Install VirtIO drivers after Windows setup
   - Enable CPU host-passthrough
   - Increase memory allocation

### Debug Mode

Run server in debug mode:
```bash
./bin/libgo-server -config configs/windows11-config.yaml
```

Check logs:
```bash
tail -f ./logs/libgo.log
```

## Best Practices

1. **Initial Setup**
   - Use SATA bus for initial Windows installation
   - Switch to VirtIO after installing drivers
   - Create snapshot after clean install

2. **Resource Allocation**
   - Windows 11: Minimum 8GB RAM, 4 CPUs
   - Windows Server: Scale based on workload
   - Use thin provisioning for disks

3. **Network Configuration**
   - Start with e1000e for compatibility
   - Switch to VirtIO for performance
   - Configure static IPs for servers

4. **Security**
   - Enable authentication in production
   - Use strong passwords
   - Regular snapshots before changes

## Additional Resources

- [LibGo API Documentation](../../docs/api/)
- [Windows VM Templates](../../configs/templates/)
- [Snapshot API Guide](../../docs/api/snapshots.md)
- [VirtIO Driver Installation](https://docs.fedoraproject.org/en-US/quick-docs/creating-windows-virtual-machines-using-virtio-drivers/)

## Support

For issues or questions:
1. Check the troubleshooting section
2. Review server logs
3. Consult the API documentation
4. Open an issue on GitHub