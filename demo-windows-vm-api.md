# Windows VM API Demo - Expected Outputs

This document shows what the Windows VM API responses would look like when the server is properly running.

## 1. Health Check

**Request:**
```bash
curl http://localhost:8080/health
```

**Expected Response:**
```json
{
  "status": "up",
  "time": "2025-05-30T19:55:00Z"
}
```

## 2. Create Windows VM

**Request:**
```bash
curl -X POST http://localhost:8080/api/v1/vms \
  -H "Content-Type: application/json" \
  -d '{
    "name": "test-windows-11",
    "template": "windows-11",
    "description": "Test Windows 11 VM",
    "cpu": {"count": 4},
    "memory": {"sizeBytes": 8589934592},
    "disk": {"sizeBytes": 107374182400, "format": "qcow2", "bus": "sata"}
  }'
```

**Expected Response:**
```json
{
  "vm": {
    "name": "test-windows-11",
    "uuid": "550e8400-e29b-41d4-a716-446655440000",
    "description": "Test Windows 11 VM",
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
    "disks": [
      {
        "size_bytes": 107374182400,
        "format": "qcow2",
        "bus": "sata"
      }
    ],
    "networks": [
      {
        "type": "network",
        "mac_address": "52:54:00:12:34:56",
        "source": "default",
        "model": "virtio"
      }
    ],
    "status": "shutoff",
    "created_at": "2025-05-30T19:55:00Z"
  }
}
```

## 3. Start VM

**Request:**
```bash
curl -X PUT http://localhost:8080/api/v1/vms/test-windows-11/start
```

**Expected Response:**
```json
{
  "success": true,
  "message": "VM started successfully"
}
```

## 4. List VMs

**Request:**
```bash
curl http://localhost:8080/api/v1/vms
```

**Expected Response:**
```json
{
  "vms": [
    {
      "name": "test-windows-11",
      "uuid": "550e8400-e29b-41d4-a716-446655440000",
      "description": "Test Windows 11 VM",
      "cpu": {
        "count": 4
      },
      "memory": {
        "size_bytes": 8589934592
      },
      "status": "running",
      "created_at": "2025-05-30T19:55:00Z"
    }
  ],
  "count": 1
}
```

## 5. Create Snapshot

**Request:**
```bash
curl -X POST http://localhost:8080/api/v1/vms/test-windows-11/snapshots \
  -H "Content-Type: application/json" \
  -d '{
    "name": "fresh-install",
    "description": "Windows 11 fresh installation",
    "include_memory": false
  }'
```

**Expected Response:**
```json
{
  "snapshot": {
    "name": "fresh-install",
    "description": "Windows 11 fresh installation", 
    "state": "shutoff",
    "parent": "",
    "created_at": "2025-05-30T20:00:00Z",
    "is_current": true,
    "has_metadata": true,
    "has_memory": false,
    "has_disk": true
  }
}
```

## 6. List Snapshots

**Request:**
```bash
curl http://localhost:8080/api/v1/vms/test-windows-11/snapshots
```

**Expected Response:**
```json
{
  "snapshots": [
    {
      "name": "fresh-install",
      "description": "Windows 11 fresh installation",
      "state": "shutoff", 
      "created_at": "2025-05-30T20:00:00Z",
      "is_current": true,
      "has_memory": false,
      "has_disk": true
    }
  ],
  "count": 1
}
```

## 7. Get Snapshot Details

**Request:**
```bash
curl http://localhost:8080/api/v1/vms/test-windows-11/snapshots/fresh-install
```

**Expected Response:**
```json
{
  "snapshot": {
    "name": "fresh-install",
    "description": "Windows 11 fresh installation",
    "state": "shutoff",
    "parent": "",
    "created_at": "2025-05-30T20:00:00Z",
    "is_current": true,
    "has_metadata": true,
    "has_memory": false,
    "has_disk": true
  }
}
```

## 8. Revert to Snapshot

**Request:**
```bash
curl -X PUT http://localhost:8080/api/v1/vms/test-windows-11/snapshots/fresh-install/revert
```

**Expected Response:**
```json
{
  "success": true,
  "message": "VM reverted to snapshot successfully"
}
```

## 9. Delete Snapshot

**Request:**
```bash
curl -X DELETE http://localhost:8080/api/v1/vms/test-windows-11/snapshots/fresh-install
```

**Expected Response:**
```json
{
  "success": true,
  "message": "Snapshot deleted successfully"
}
```

## 10. Export VM

**Request:**
```bash
curl -X POST http://localhost:8080/api/v1/vms/test-windows-11/export \
  -H "Content-Type: application/json" \
  -d '{
    "format": "vmdk",
    "compress": true
  }'
```

**Expected Response:**
```json
{
  "job": {
    "id": "export-123456",
    "status": "running",
    "format": "vmdk",
    "created_at": "2025-05-30T20:05:00Z",
    "estimated_completion": "2025-05-30T20:15:00Z"
  }
}
```

## Complete Workflow Example

Here's how you would create a complete Windows VM setup:

```bash
# 1. Create the VM
curl -X POST http://localhost:8080/api/v1/vms \
  -H "Content-Type: application/json" \
  -d '{
    "name": "production-windows",
    "template": "windows-11",
    "cpu": {"count": 8},
    "memory": {"sizeBytes": 17179869184},
    "disk": {"sizeBytes": 214748364800}
  }'

# 2. Start the VM
curl -X PUT http://localhost:8080/api/v1/vms/production-windows/start

# 3. After Windows installation, create a snapshot
curl -X POST http://localhost:8080/api/v1/vms/production-windows/snapshots \
  -H "Content-Type: application/json" \
  -d '{
    "name": "post-install",
    "description": "Clean Windows installation with drivers"
  }'

# 4. Install applications, then create another snapshot
curl -X POST http://localhost:8080/api/v1/vms/production-windows/snapshots \
  -H "Content-Type: application/json" \
  -d '{
    "name": "with-applications", 
    "description": "Windows with Office, dev tools installed"
  }'

# 5. Export for backup
curl -X POST http://localhost:8080/api/v1/vms/production-windows/export \
  -H "Content-Type: application/json" \
  -d '{"format": "ova"}'
```

## Status Codes

The API returns standard HTTP status codes:

- `200 OK` - Successful operation
- `201 Created` - Resource created successfully
- `400 Bad Request` - Invalid input parameters
- `404 Not Found` - VM or snapshot not found
- `500 Internal Server Error` - Server error

## Error Response Format

When an error occurs, the API returns:

```json
{
  "error": {
    "code": "INVALID_INPUT",
    "message": "Invalid VM parameters: CPU count must be at least 1"
  }
}
```