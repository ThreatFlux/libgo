# WebSocket API Documentation

The WebSocket API enables real-time VM monitoring, console access, and VM control. This document explains the available endpoints, message protocol, and usage examples.

## Endpoints

The WebSocket API is accessible through the following endpoints:

- **VM Monitoring**: `/ws/vms/{name}`
  - Provides real-time status updates and performance metrics for a specific VM
  - Requires authentication token and read permission

- **VM Console**: `/ws/vms/{name}/console`
  - Provides interactive terminal/console access to a specific VM
  - Requires authentication token and console permission

## Connection

All WebSocket connections require a valid JWT token, passed as a query parameter:

```
ws://localhost:8080/ws/vms/my-vm?token=eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
```

## Message Format

All WebSocket messages use JSON format with the following structure:

```json
{
  "type": "message_type",
  "timestamp": "2023-05-23T12:34:56Z",
  "data": {
    // Type-specific message data
  }
}
```

### Message Types

The `type` field indicates the message's purpose:

| Type | Direction | Description |
|------|-----------|-------------|
| `status` | Server → Client | VM status updates |
| `metrics` | Server → Client | VM performance metrics |
| `command` | Client → Server | VM control commands |
| `response` | Server → Client | Command response |
| `console` | Server → Client | Console output data |
| `console_input` | Client → Server | Console input data |
| `error` | Server → Client | Error messages |
| `heartbeat` | Both | Connection health check |
| `connection` | Server → Client | Connection status information |

## Message Data Formats

### Status Message

Provides the current status of a VM:

```json
{
  "type": "status",
  "timestamp": "2023-05-23T12:34:56Z",
  "data": {
    "status": "running",
    "lastStateChange": "2023-05-23T12:30:00Z",
    "uptime": 300
  }
}
```

Status values can be: `running`, `stopped`, `paused`, `suspended`, `error`.

### Metrics Message

Provides performance metrics for the VM:

```json
{
  "type": "metrics",
  "timestamp": "2023-05-23T12:34:56Z",
  "data": {
    "cpu": {
      "utilization": 23.5
    },
    "memory": {
      "used": 2147483648,
      "total": 4294967296
    },
    "network": {
      "rxBytes": 1024000,
      "txBytes": 512000
    },
    "disk": {
      "readBytes": 2048000,
      "writeBytes": 1024000
    }
  }
}
```

### Command Message

Sends a command to control the VM:

```json
{
  "type": "command",
  "timestamp": "2023-05-23T12:34:56Z",
  "data": {
    "action": "start",
    "requestId": "cmd-1621765234"
  }
}
```

Supported actions:
- `start`: Start the VM
- `stop`: Stop the VM
- `restart`: Restart the VM
- `suspend`: Suspend the VM
- `resume`: Resume the VM

The `requestId` is optional. If not provided, the server will generate one.

### Response Message

Response to a command:

```json
{
  "type": "response",
  "timestamp": "2023-05-23T12:34:56Z",
  "data": {
    "requestId": "cmd-1621765234",
    "success": true,
    "message": "Command 'start' completed successfully"
  }
}
```

### Console Messages

Console output from the VM:

```json
{
  "type": "console",
  "timestamp": "2023-05-23T12:34:56Z",
  "data": {
    "content": "Ubuntu 22.04 LTS\nmy-vm login: ",
    "eof": false
  }
}
```

Console input to the VM:

```json
{
  "type": "console_input",
  "timestamp": "2023-05-23T12:34:56Z",
  "data": {
    "content": "admin\n"
  }
}
```

### Error Message

Error information:

```json
{
  "type": "error",
  "timestamp": "2023-05-23T12:34:56Z",
  "data": {
    "code": "VM_NOT_RUNNING",
    "message": "Cannot connect to console: VM is not running"
  }
}
```

## Usage Examples

### Monitoring VM Performance

1. Connect to the VM monitoring WebSocket:
   ```
   ws://localhost:8080/ws/vms/my-vm?token=eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
   ```

2. Listen for status and metrics messages to display performance information.

3. Send commands to control the VM:
   ```json
   {
     "type": "command",
     "timestamp": "2023-05-23T12:34:56Z",
     "data": {
       "action": "restart",
       "requestId": "cmd-1621765234"
     }
   }
   ```

### Interactive Console Access

1. Connect to the VM console WebSocket:
   ```
   ws://localhost:8080/ws/vms/my-vm/console?token=eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
   ```

2. Listen for console output messages to display in a terminal emulator.

3. Send console input from the user:
   ```json
   {
     "type": "console_input",
     "timestamp": "2023-05-23T12:34:56Z",
     "data": {
       "content": "ls -la\n"
     }
   }
   ```

## Connection Management

- The server sends ping messages every 30 seconds to keep the connection alive.
- Clients should respond with pong messages.
- If a client doesn't receive a ping for 60 seconds, it should attempt to reconnect.
- On connection errors, clients should wait 5 seconds before attempting to reconnect.

## Error Handling

Common error codes:

| Code | Description |
|------|-------------|
| `AUTH_ERROR` | Authentication or authorization error |
| `VM_NOT_FOUND` | The requested VM does not exist |
| `VM_NOT_RUNNING` | The VM is not in running state |
| `INVALID_COMMAND` | Invalid command or parameters |
| `UNKNOWN_ERROR` | Unspecified server error |

## Security Considerations

- Always use HTTPS/WSS in production environments
- JWT tokens expire based on server configuration
- Console access requires special permissions
- All traffic is logged and monitored

## Frontend Integration

The WebSocket API is designed to work with the provided React components:

- `useVMWebSocket`: React hook for VM monitoring and control
- `useVMConsole`: React hook for VM console access
- `VMMetricsChart`: Component for displaying real-time metrics
- `VMConsole`: Component for interactive console access

See the UI documentation for more details on frontend integration.