# LibGo KVM UI

Modern React TypeScript UI for managing LibGo KVM virtual machines.

## Features

- Dark mode by default with light mode support
- Responsive collapsible sidebar navigation
- Dashboard with VM and system metrics
- Complete VM management (create, view, start, stop, delete)
- VM export functionality
- JWT authentication
- Real-time VM monitoring via WebSockets
- Interactive VM console access via WebSockets

## Getting Started

```bash
# Install dependencies
npm install

# Start development server
npm run dev

# Build for production
npm run build
```

## Technology Stack

- React 18+
- TypeScript
- Vite
- TanStack Query (React Query)
- TanStack Router
- Tailwind CSS
- shadcn/ui components
- Zustand for state management
- Axios for API requests
- WebSocket for real-time communication
- Chart.js for metrics visualization
- xterm.js for terminal emulation
- React Hook Form
- Zod for validation
- react-icons

## Project Structure

```
src/
├── api/            # API integration
├── components/     # Reusable UI components
├── contexts/       # React contexts
├── hooks/          # Custom hooks
├── lib/            # Utility functions
├── pages/          # Page components
├── routes/         # Route definitions
├── store/          # Global state
├── styles/         # Global styles
└── types/          # TypeScript types
```

## Configuration

To configure the API endpoint, edit the `.env` file:

```
VITE_API_BASE_URL=http://localhost:8080
```

## Docker

```bash
# Build Docker image
docker build -t libgo-ui .

# Run Docker container
docker run -p 3000:80 libgo-ui
```

## WebSocket Integration

### WebSocket Endpoints

- **VM Monitoring:** `/ws/vms/{name}`
- **VM Console:** `/ws/vms/{name}/console`

Both endpoints require authentication with a JWT token (passed as a query parameter).

### Components and Hooks

#### WebSocket Client

`src/api/websocket.ts` provides the WebSocket client implementation with:

- Connection management
- Message handling
- Event subscription system
- Command sending

#### React Hooks

`src/hooks/useVMWebSocket.ts` provides React hooks for using WebSockets:

- `useVMWebSocket`: Base hook for WebSocket connections
- `useVMMonitor`: Specialized hook for VM monitoring
- `useVMConsole`: Specialized hook for VM console access

#### UI Components

- `src/components/vm/vm-console.tsx`: Terminal implementation using xterm.js
- `src/components/vm/vm-metrics.tsx`: Real-time charts for VM metrics

### Usage Examples

#### VM Monitoring

```tsx
import { useVMMonitor } from '@/hooks/useVMWebSocket';
import { VMMetricsChart } from '@/components/vm/vm-metrics';

const VMMoitoringTab = ({ vmName }) => {
  const {
    connected,
    status,
    metrics,
    error,
    sendCommand
  } = useVMMonitor(vmName);

  if (!connected) {
    return <div>Connecting to VM...</div>;
  }

  if (!metrics) {
    return <div>Waiting for metrics data...</div>;
  }

  return (
    <div>
      <p>Status: {status?.status}</p>
      <VMMetricsChart vmName={vmName} metrics={metrics} />
      
      <div className="mt-4">
        <button onClick={() => sendCommand('restart')}>
          Restart VM
        </button>
      </div>
    </div>
  );
};
```

#### VM Console

```tsx
import { VMConsole } from '@/components/vm/vm-console';

const VMConsoleTab = ({ vmName }) => {
  return (
    <div className="h-[500px]">
      <VMConsole vmName={vmName} />
    </div>
  );
};
```

### Message Protocol

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

For more details on the WebSocket API, see the [WebSocket API Documentation](../docs/websocket-api.md).