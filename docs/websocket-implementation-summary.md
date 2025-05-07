# WebSocket Implementation Summary

This document provides an overview of the WebSocket implementation for real-time VM monitoring, console access, and management in the LibGo application.

## Architecture Overview

The WebSocket implementation consists of these key components:

1. **Backend WebSocket Server**
   - Hub for managing client connections
   - Message broadcasting system
   - VM monitoring service
   - Console input/output handling

2. **Frontend WebSocket Client**
   - Connection management with auto-reconnect
   - Message handling and event system
   - React hooks for easy integration
   - UI components for metrics and console

3. **Integration Points**
   - Router configuration for WebSocket endpoints
   - VM manager extension for metrics collection
   - Authentication middleware integration

## Implementation Components

### Backend Components

1. **WebSocket Types (`/internal/websocket/types.go`)**
   - Message types and structures
   - Client connection management
   - Hub for broadcasting messages

2. **WebSocket Handler (`/internal/websocket/handler.go`)**
   - Connection handling and upgrading
   - Message processing
   - Command handling
   - Console input/output processing

3. **VM Monitor (`/internal/websocket/vm_monitor.go`)**
   - Real-time VM metrics collection
   - Status monitoring
   - Client registration and cleanup

4. **WebSocket Setup (`/internal/websocket/setup.go`)**
   - Route configuration
   - Integration with API server
   - Authentication middleware

### Frontend Components

1. **WebSocket Client (`/ui/src/api/websocket.ts`)**
   - Connection management
   - Message handling
   - Event subscription system
   - Command sending

2. **WebSocket Hooks (`/ui/src/hooks/useVMWebSocket.ts`)**
   - React hooks for WebSocket state
   - Specialized hooks for monitoring and console
   - Message handling and state updates

3. **VM Console Component (`/ui/src/components/vm/vm-console.tsx`)**
   - Terminal implementation using xterm.js
   - Console input/output handling
   - UI controls for fullscreen and clearing

4. **VM Metrics Component (`/ui/src/components/vm/vm-metrics.tsx`)**
   - Real-time charts for CPU, memory, disk, and network
   - Metrics calculation and formatting
   - User-friendly display with tooltips

## Configuration

WebSocket functionality can be configured in `config.yaml`:

```yaml
# WebSocket settings
websocket:
  # Enable WebSocket functionality (real-time VM monitoring and console access)
  enabled: true
  # Ping interval for keeping connections alive (seconds)
  pingInterval: 30
  # Maximum message size in bytes
  maxMessageSize: 8192
  # Write wait timeout in seconds
  writeWait: 10s
  # Read wait timeout in seconds (pong wait)
  pongWait: 60s
  # VM metrics sampling interval in seconds
  metricsInterval: 5s
```

## Fixes Implemented

During the implementation, several issues were fixed:

1. **Router Configuration**
   - Added missing WebSocket import in router.go
   - Fixed VM manager casting in router_adapter.go

2. **UI Dependencies**
   - Installed xterm.js dependencies (xterm, xterm-addon-fit, xterm-addon-web-links)
   - Added missing Tab components to lib/components.ts

3. **VM Manager Integration**
   - Added GetMetrics method to VM manager interface
   - Implemented temporary mock metrics generation for testing

4. **Documentation**
   - Created comprehensive WebSocket API documentation
   - Updated config.yaml.example with WebSocket settings
   - Created implementation summary

## Security Considerations

The WebSocket implementation includes several security measures:

1. **Authentication**: All WebSocket connections require a valid JWT token
2. **Authorization**: Different permissions for monitoring vs. console access
3. **Secure Communication**: Support for WSS (WebSocket Secure)
4. **Session Management**: Tracking of active connections and cleanup of stale ones
5. **Input Validation**: Validation of all messages received from clients

## Next Steps

Future improvements to consider:

1. **Real Metrics Collection**: Replace mock metrics with actual libvirt data
2. **More VM Controls**: Add support for snapshot, resize, and other operations
3. **Enhanced Console**: Add support for terminal resizing, copy/paste, and special keys
4. **Unit Tests**: Comprehensive test coverage for WebSocket components
5. **Metrics Storage**: Historical metrics collection and visualization

## Usage Example

With the implementation in place, users can now:

1. View real-time VM metrics in the UI
2. Access VM console directly in the browser
3. Control VMs (start, stop, restart) with immediate feedback
4. Monitor VM status changes in real-time

All of this without page refreshes or manual polling, providing a much more responsive and interactive experience.