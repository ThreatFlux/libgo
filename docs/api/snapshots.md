# VM Snapshot API Documentation

The VM Snapshot API provides endpoints for managing virtual machine snapshots, allowing you to create, list, retrieve, delete, and revert to snapshots.

## Endpoints

### Create Snapshot

Create a new snapshot of a virtual machine.

**Endpoint:** `POST /api/v1/vms/{name}/snapshots`

**Request Body:**
```json
{
  "name": "snapshot-name",
  "description": "Optional description",
  "include_memory": true,
  "quiesce": false
}
```

**Parameters:**
- `name` (required): Name of the snapshot
- `description` (optional): Description of the snapshot
- `include_memory` (optional): Whether to include memory state in the snapshot (default: false)
- `quiesce` (optional): Attempt to quiesce guest filesystems (requires guest agent) (default: false)

**Response:**
```json
{
  "snapshot": {
    "name": "snapshot-name",
    "description": "Optional description",
    "state": "running",
    "parent": "",
    "created_at": "2025-05-30T12:00:00Z",
    "is_current": true,
    "has_metadata": true,
    "has_memory": true,
    "has_disk": true
  }
}
```

### List Snapshots

List all snapshots for a virtual machine.

**Endpoint:** `GET /api/v1/vms/{name}/snapshots`

**Query Parameters:**
- `include_metadata` (optional): Include full metadata for each snapshot (default: false)
- `tree` (optional): Return snapshots in tree structure (default: false)

**Response:**
```json
{
  "snapshots": [
    {
      "name": "snapshot1",
      "description": "First snapshot",
      "state": "shutoff",
      "created_at": "2025-05-30T10:00:00Z",
      "is_current": false,
      "has_memory": true,
      "has_disk": true
    },
    {
      "name": "snapshot2",
      "description": "Second snapshot",
      "state": "running",
      "created_at": "2025-05-30T11:00:00Z",
      "is_current": true,
      "has_memory": false,
      "has_disk": true
    }
  ],
  "count": 2
}
```

### Get Snapshot

Get detailed information about a specific snapshot.

**Endpoint:** `GET /api/v1/vms/{name}/snapshots/{snapshot}`

**Response:**
```json
{
  "snapshot": {
    "name": "snapshot-name",
    "description": "Snapshot description",
    "state": "running",
    "parent": "parent-snapshot",
    "created_at": "2025-05-30T12:00:00Z",
    "is_current": true,
    "has_metadata": true,
    "has_memory": true,
    "has_disk": true
  }
}
```

### Delete Snapshot

Delete a snapshot from a virtual machine.

**Endpoint:** `DELETE /api/v1/vms/{name}/snapshots/{snapshot}`

**Response:**
```json
{
  "success": true,
  "message": "Snapshot deleted successfully"
}
```

### Revert to Snapshot

Revert a virtual machine to a specific snapshot state.

**Endpoint:** `PUT /api/v1/vms/{name}/snapshots/{snapshot}/revert`

**Response:**
```json
{
  "success": true,
  "message": "VM reverted to snapshot successfully"
}
```

## Error Responses

All endpoints may return the following error responses:

**400 Bad Request:**
```json
{
  "error": {
    "code": "INVALID_INPUT",
    "message": "Invalid input parameters"
  }
}
```

**404 Not Found:**
```json
{
  "error": {
    "code": "NOT_FOUND",
    "message": "VM or snapshot not found"
  }
}
```

**500 Internal Server Error:**
```json
{
  "error": {
    "code": "INTERNAL_ERROR",
    "message": "Failed to perform snapshot operation"
  }
}
```

## Usage Examples

### Create a snapshot with memory state
```bash
curl -X POST http://localhost:8080/api/v1/vms/my-vm/snapshots \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{
    "name": "before-update",
    "description": "Snapshot before system update",
    "include_memory": true
  }'
```

### List all snapshots
```bash
curl -X GET http://localhost:8080/api/v1/vms/my-vm/snapshots \
  -H "Authorization: Bearer $TOKEN"
```

### Revert to a snapshot
```bash
curl -X PUT http://localhost:8080/api/v1/vms/my-vm/snapshots/before-update/revert \
  -H "Authorization: Bearer $TOKEN"
```

### Delete a snapshot
```bash
curl -X DELETE http://localhost:8080/api/v1/vms/my-vm/snapshots/old-snapshot \
  -H "Authorization: Bearer $TOKEN"
```

## Notes

- Snapshots with memory state allow the VM to be restored to the exact running state
- Disk-only snapshots are faster to create but only preserve disk state
- The `quiesce` option requires the guest agent to be installed in the VM
- Reverting to a snapshot will discard all changes made after the snapshot was created