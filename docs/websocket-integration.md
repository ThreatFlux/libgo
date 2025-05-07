# WebSocket Integration for VM Management

This document outlines the design and implementation plan for adding WebSocket support to enable real-time VM management including console/terminal access to VMs.

## Overview

The WebSocket integration will enable:
1. Real-time monitoring of VM metrics (CPU, memory, disk, network)
2. Interactive console/terminal access to VMs
3. Real-time status updates for VM state changes
4. Direct command execution for VM management operations

## Architecture

### Backend Components

1. **WebSocket Server**
   - Handles WebSocket connections and authentication
   - Routes messages to appropriate handlers
   - Manages client connections

2. **VM Monitor Service**
   - Collects VM metrics at regular intervals
   - Detects VM state changes
   - Publishes updates to connected clients

3. **VM Console Service**
   - Provides access to VM console via WebSocket
   - Handles console input/output
   - Manages terminal sessions

4. **Event System**
   - Broadcasts VM events to connected clients
   - Handles event subscription and filtering

### Frontend Components

1. **WebSocket Client**
   - Manages WebSocket connections
   - Handles reconnection logic
   - Processes incoming messages

2. **React Hooks**
   - `useVMWebSocket` - For VM monitoring and status updates
   - `useVMConsole` - For console/terminal access

3. **UI Components**
   - Real-time metrics dashboard
   - Interactive terminal/console
   - Status indicators and notifications

## WebSocket Protocol

### Connection Endpoints

- `/ws/vms/:name` - For VM monitoring and management
- `/ws/vms/:name/console` - For VM console/terminal access

### Message Types

#### VM Status Updates
```json
{
  "type": "status",
  "timestamp": "2023-05-06T13:13:42.637-0400",
  "data": {
    "status": "running|stopped|paused|shutdown|crashed|unknown",
    "lastStateChange": "2023-05-06T13:13:42.637-0400",
    "uptime": 3600
  }
}
```

#### VM Metrics Updates
```json
{
  "type": "metrics",
  "timestamp": "2023-05-06T13:13:42.637-0400",
  "data": {
    "cpu": {
      "utilization": 15.5
    },
    "memory": {
      "used": 1073741824,
      "total": 4294967296
    },
    "network": {
      "rxBytes": 1024000,
      "txBytes": 512000,
      "rxPackets": 1500,
      "txPackets": 750
    },
    "disk": {
      "readBytes": 51200000,
      "writeBytes": 25600000,
      "readOps": 1200,
      "writeOps": 600
    }
  }
}
```

#### VM Commands
```json
{
  "type": "command",
  "timestamp": "2023-05-06T13:13:42.637-0400",
  "data": {
    "action": "start|stop|restart|suspend|resume",
    "requestId": "550e8400-e29b-41d4-a716-446655440000"
  }
}
```

#### Command Responses
```json
{
  "type": "response",
  "timestamp": "2023-05-06T13:13:42.637-0400",
  "data": {
    "requestId": "550e8400-e29b-41d4-a716-446655440000",
    "success": true,
    "message": "VM started successfully"
  }
}
```

#### Console Messages
```json
{
  "type": "console",
  "timestamp": "2023-05-06T13:13:42.637-0400",
  "data": {
    "content": "base64-encoded-content",
    "eof": false
  }
}
```

#### Console Input
```json
{
  "type": "console_input",
  "timestamp": "2023-05-06T13:13:42.637-0400",
  "data": {
    "content": "base64-encoded-input"
  }
}
```

## Authentication and Security

1. **Authentication**
   - JWT-based authentication for WebSocket connections
   - Token passed as query parameter during connection

2. **Authorization**
   - Role-based access control for VM operations
   - Permission checks for sensitive operations

3. **Security Measures**
   - Rate limiting to prevent abuse
   - Message validation
   - Connection timeouts for inactive sessions
   - Sanitization of console input/output

## Implementation Plan

### Phase 1: Core WebSocket Infrastructure

1. Add WebSocket support to the Go backend
   - Implement connection handling
   - Add authentication middleware
   - Create basic message routing

2. Implement WebSocket client in the frontend
   - Create connection management
   - Add reconnection logic
   - Implement message processing

### Phase 2: VM Monitoring

1. Create VM monitoring service
   - Implement metric collection
   - Add state change detection
   - Create event broadcasting system

2. Add frontend components for monitoring
   - Create real-time metric charts
   - Add status indicators
   - Implement notification system

### Phase 3: VM Console/Terminal

1. Implement VM console service
   - Add console access via libvirt
   - Create console session management
   - Implement input/output handling

2. Add frontend terminal component
   - Create interactive terminal UI
   - Implement console input/output
   - Add terminal options (font size, colors, etc.)

### Phase 4: Enhanced VM Management

1. Add WebSocket-based VM commands
   - Implement command handlers
   - Add response processing
   - Create transaction management for commands

2. Enhance UI with WebSocket-based controls
   - Add real-time power controls
   - Implement progress indicators
   - Create command history

## Technical Considerations

### Performance

- Implement efficient message serialization
- Use goroutines for concurrent handling
- Consider binary protocols for console data
- Add message buffering for slow connections

### Scalability

- Design for multiple concurrent connections
- Implement connection pooling
- Consider distributed event broadcasting for multiple servers

### Error Handling

- Graceful connection termination
- Meaningful error messages
- Automatic reconnection with backoff
- Heartbeat mechanism for connection health

## Dependencies

- [gorilla/websocket](https://github.com/gorilla/websocket) - For WebSocket support in Go
- [xterm.js](https://xtermjs.org/) - For terminal emulation in the browser

## API Integration

The WebSocket endpoints will integrate with the existing API structure:

1. Authentication will use the same JWT mechanism
2. Authorization will use the same role-based system
3. VM operations will utilize the existing VM manager
4. Metrics will be collected using the existing libvirt integration

## Conclusion

This WebSocket integration will significantly enhance the user experience by providing real-time VM management capabilities and interactive console access. The implementation follows a phased approach to ensure stable, secure, and performant functionality.