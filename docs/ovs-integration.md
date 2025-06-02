# OpenVSwitch Integration

This document describes the OpenVSwitch (OVS) integration in libgo, which provides software-defined networking capabilities for VM management.

## Overview

The OVS integration allows libgo to:
- Create and manage OVS bridges
- Add and configure ports with VLAN support
- Manage OpenFlow rules
- Integrate with libvirt for VM networking
- Provide a complete SDN solution for virtualized environments

## Installation and Setup

### 1. Install OpenVSwitch

```bash
# Install OVS packages
make install-ovs

# Or install manually:
# Ubuntu/Debian:
sudo apt-get install openvswitch-switch openvswitch-common

# RHEL/CentOS:
sudo yum install openvswitch

# Fedora:
sudo dnf install openvswitch
```

### 2. Set Up Development Environment

```bash
# Install all development tools including OVS
make test-setup
```

### 3. Verify Installation

```bash
# Check OVS version
ovs-vsctl --version
ovs-ofctl --version

# Start OVS services (if not already running)
sudo systemctl start openvswitch-switch
sudo systemctl enable openvswitch-switch
```

## API Endpoints

### Bridge Management

- `POST /api/v1/ovs/bridges` - Create a new bridge
- `GET /api/v1/ovs/bridges` - List all bridges
- `GET /api/v1/ovs/bridges/{name}` - Get bridge details
- `DELETE /api/v1/ovs/bridges/{name}` - Delete a bridge

### Port Management

- `POST /api/v1/ovs/bridges/{bridge}/ports` - Add port to bridge
- `GET /api/v1/ovs/bridges/{bridge}/ports` - List bridge ports
- `DELETE /api/v1/ovs/bridges/{bridge}/ports/{port}` - Remove port

### Flow Management

- `POST /api/v1/ovs/bridges/{bridge}/flows` - Add OpenFlow rule

## Usage Examples

### Creating a Bridge

```bash
curl -X POST http://localhost:8700/api/v1/ovs/bridges \
  -H "Content-Type: application/json" \
  -d '{
    "name": "br-test",
    "datapath_type": "system"
  }'
```

### Adding a Port with VLAN

```bash
curl -X POST http://localhost:8700/api/v1/ovs/bridges/br-test/ports \
  -H "Content-Type: application/json" \
  -d '{
    "name": "veth0",
    "bridge": "br-test",
    "type": "internal",
    "tag": 100
  }'
```

### Creating an OpenFlow Rule

```bash
curl -X POST http://localhost:8700/api/v1/ovs/bridges/br-test/flows \
  -H "Content-Type: application/json" \
  -d '{
    "table": 0,
    "priority": 1000,
    "match": "in_port=1",
    "actions": "output:2"
  }'
```

## Frontend UI

The OVS management interface is available in the web UI:

1. Navigate to `http://localhost:3700/ovs`
2. Use the bridge management interface to:
   - View existing bridges
   - Create new bridges
   - Manage ports and VLANs
   - Configure OpenFlow rules

## Integration with VM Management

### Connecting VMs to OVS Bridges

When creating VMs, you can specify OVS bridges for networking:

```json
{
  "name": "test-vm",
  "networks": [
    {
      "bridge": "br-ovs",
      "type": "bridge",
      "vlan": 100
    }
  ]
}
```

### VXLAN Tunnels

Create VXLAN tunnels between hosts:

```bash
# On host 1 (192.168.1.10)
curl -X POST http://localhost:8700/api/v1/ovs/bridges/br-tunnel/ports \
  -d '{
    "name": "vxlan-host2",
    "type": "vxlan",
    "remote_ip": "192.168.1.20",
    "tunnel_type": "vxlan",
    "other_config": {"key": "1000"}
  }'

# On host 2 (192.168.1.20)
curl -X POST http://localhost:8700/api/v1/ovs/bridges/br-tunnel/ports \
  -d '{
    "name": "vxlan-host1", 
    "type": "vxlan",
    "remote_ip": "192.168.1.10",
    "tunnel_type": "vxlan",
    "other_config": {"key": "1000"}
  }'
```

## Testing

### Unit Tests

```bash
# Run OVS unit tests
go test -short ./internal/ovs/...
```

### Integration Tests

```bash
# Run OVS integration tests (requires root and OVS installation)
make test-ovs

# Or run manually:
sudo go test -tags=integration ./internal/ovs/...
```

### Manual Testing

```bash
# Start the backend
make start-backend

# Create a test bridge
curl -X POST http://localhost:8700/api/v1/ovs/bridges \
  -d '{"name": "test-bridge"}'

# List bridges
curl http://localhost:8700/api/v1/ovs/bridges

# Clean up
curl -X DELETE http://localhost:8700/api/v1/ovs/bridges/test-bridge
```

## Configuration

Add OVS configuration to your config file:

```yaml
ovs:
  enabled: true
  default_bridges:
    - name: "br-mgmt"
      datapath_type: "system"
  libvirt_integration: true
  command_timeout: "30s"
```

## Troubleshooting

### Common Issues

1. **Permission Denied**: OVS commands typically require root privileges
   ```bash
   sudo systemctl status openvswitch-switch
   ```

2. **OVS Not Running**: Ensure OVS services are started
   ```bash
   sudo systemctl start openvswitch-switch
   ```

3. **Command Not Found**: Ensure OVS is installed
   ```bash
   make install-ovs
   ```

### Logs

Check libgo logs for OVS-related errors:
```bash
tail -f backend.log | grep -i ovs
```

### Debug Commands

```bash
# List all bridges
ovs-vsctl list-br

# Show bridge details
ovs-vsctl show

# List flows
ovs-ofctl dump-flows br-test
```

## Security Considerations

- OVS operations require elevated privileges
- Use proper authentication for API access
- Validate all input parameters
- Monitor OpenFlow rules for security policies
- Implement proper VLAN isolation

## Performance Tips

- Use hardware-accelerated datapaths when available
- Configure appropriate flow table sizes
- Monitor bridge and port statistics
- Use OpenFlow groups for complex forwarding logic

## Further Reading

- [OpenVSwitch Documentation](http://docs.openvswitch.org/)
- [OpenVSwitch Integration Guide](http://docs.openvswitch.org/en/latest/topics/integration/)
- [libvirt Networking](https://libvirt.org/formatnetwork.html)