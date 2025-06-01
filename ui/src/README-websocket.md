# WebSocket Implementation in LibGo UI

This document provides a summary of the WebSocket implementation in the LibGo UI, focusing on the architecture, components, and usage patterns.

## Architecture Overview

The WebSocket implementation follows a layered architecture:

1. **WebSocket Client Layer** (`websocket.ts`)
   - Handles low-level WebSocket communication
   - Manages connection lifecycle
   - Implements message serialization/deserialization
   - Provides event subscription mechanism

2. **React Hook Layer** (`useVMWebSocket.ts`)
   - Encapsulates WebSocket client in React hooks
   - Manages WebSocket state in React components
   - Provides specialized hooks for different use cases

3. **UI Component Layer** (`vm-console.tsx`, `vm-metrics.tsx`)
   - Implements user interface for WebSocket data
   - Provides interactive components for VM monitoring and console access

## WebSocket Client (`websocket.ts`)

The WebSocket client provides:

- Connection establishment with authentication
- Automatic reconnection on connection loss
- Message parsing and event dispatching
- Command sending for VM control
- Heartbeat mechanism for connection health monitoring

Key classes and interfaces:

```typescript
// Message types
export type MessageType =
  | 'status'
  | 'metrics'
  | 'command'
  | 'response'
  | 'console'
  | 'console_input'
  | 'error'
  | 'heartbeat'
  | 'connection';

// WebSocket client
export class VMWebSocketClient {
  // Establish connection
  connect(): void;
  
  // Disconnect WebSocket
  disconnect(): void;
  
  // Send a command to the server
  sendCommand(action: CommandData['action']): void;
  
  // Send console input
  sendConsoleInput(content: string): void;
  
  // Event handlers
  onMessage(handler: MessageHandler): () => void;
  onStatus(handler: StatusHandler): () => void;
  onMetrics(handler: MetricsHandler): () => void;
  onConsole(handler: ConsoleHandler): () => void;
  onResponse(handler: ResponseHandler): () => void;
  onError(handler: ErrorHandler): () => void;
  onConnection(handler: ConnectionHandler): () => void;
}
```

## React Hooks (`useVMWebSocket.ts`)

Three main hooks are provided:

1. **useVMWebSocket**: Base hook for WebSocket connections
   - Creates and manages WebSocket client
   - Handles connection state
   - Provides access to all WebSocket events

2. **useVMMonitor**: Specialized hook for VM monitoring
   - Focused on status and metrics data
   - Provides VM control commands

3. **useVMConsole**: Specialized hook for VM console
   - Provides console input/output
   - Includes console-specific commands

Example usage:

```typescript
// Hook return interface
interface UseVMWebSocketReturn {
  connected: boolean;
  status: VMStatusData | null;
  metrics: VMMetrics | null;
  error: ErrorData | null;
  consoleData: ConsoleData[];
  sendCommand: (action: 'start' | 'stop' | 'restart' | 'suspend' | 'resume') => void;
  sendConsoleInput: (content: string) => void;
  clearConsole: () => void;
}

// Using the hook
const {
  connected,
  status,
  metrics,
  error,
  sendCommand
} = useVMMonitor(vmName);
```

## UI Components

### VM Console Component (`vm-console.tsx`)

The VM console component provides an interactive terminal interface using xterm.js:

- Terminal emulation with keyboard input
- Console output display
- Fullscreen mode
- Console clearing
- Connection status indicator

```typescript
export const VMConsole: React.FC<VMConsoleProps> = ({ 
  vmName,
  className = '' 
}) => {
  // Terminal initialization
  // WebSocket connection for console data
  // User interface with controls
}
```

### VM Metrics Component (`vm-metrics.tsx`)

The VM metrics component displays real-time performance data using Chart.js:

- CPU utilization chart
- Memory usage chart
- Network activity chart
- Disk activity chart
- Current metrics display
- Rate calculation for network and disk metrics

```typescript
export const VMMetricsChart: React.FC<VMMetricsChartProps> = ({
  vmName,
  metrics,
  className = '',
}) => {
  // Chart data state
  // Metrics processing
  // Chart rendering
}
```

## Integration in VM Detail Page

The WebSocket components are integrated into the VM detail page in `vm-detail.tsx`:

- Tabs for different VM views (overview, monitoring, console)
- WebSocket status indicators
- Conditional rendering based on connection status
- VM control buttons that use WebSocket when available

Key integration points:

```typescript
// WebSocket monitoring
const {
  connected: wsConnected,
  status: wsStatus,
  metrics: wsMetrics,
  error: wsError,
  sendCommand
} = useVMMonitor(name);

// VM control integration
const handleStart = () => {
  if (vm) {
    // Use WebSocket command if connected, otherwise fallback to REST API
    if (wsConnected) {
      sendCommand('start');
    } else {
      startMutation.mutate(vm.name);
    }
  }
};
```

## Connection Management

WebSocket connections are managed intelligently:

1. **Authentication**: JWT token is included in the WebSocket URL
2. **Automatic Reconnection**: Attempts to reconnect after connection loss
3. **Heartbeat**: Regular ping messages to keep the connection alive
4. **Graceful Degradation**: Falls back to REST API when WebSocket is unavailable
5. **Error Handling**: Displays error messages to the user

## Security Considerations

1. **Authentication**: All WebSocket connections are authenticated using JWT tokens
2. **Authorization**: Different permissions for monitoring vs. console access
3. **Secure Websockets**: Support for WSS (WebSocket Secure)
4. **Input Validation**: Validation of all user input before sending to server