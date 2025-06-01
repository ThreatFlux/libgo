import { useEffect, useState } from 'react';
import { 
  VMWebSocketClient,
  VMMetrics,
  VMStatusData,
  ErrorData,
  ConsoleData
} from '@/api/websocket';

// Hook return type
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

/**
 * Hook for using WebSocket connection to monitor and control a VM
 * 
 * @param vmName The name of the VM to monitor
 * @param isConsole Whether to connect to the console endpoint
 * @returns Object with WebSocket state and methods
 */
export const useVMWebSocket = (
  vmName: string,
  isConsole = false
): UseVMWebSocketReturn => {
  // State
  const [client, setClient] = useState<VMWebSocketClient | null>(null);
  const [connected, setConnected] = useState(false);
  const [status, setStatus] = useState<VMStatusData | null>(null);
  const [metrics, setMetrics] = useState<VMMetrics | null>(null);
  const [error, setError] = useState<ErrorData | null>(null);
  const [consoleData, setConsoleData] = useState<ConsoleData[]>([]);

  // Create client and set up event handlers
  useEffect(() => {
    console.log(`useVMWebSocket: Creating client for VM ${vmName}, isConsole=${isConsole}`);
    const wsClient = new VMWebSocketClient(vmName, isConsole);
    
    // Connection handler
    const unsubConnection = wsClient.onConnection(connected => {
      console.log(`useVMWebSocket: Connection state changed to ${connected}`);
      setConnected(connected);
      if (!connected) {
        // Reset state on disconnect
        setStatus(null);
        setMetrics(null);
        setError(null);
      }
    });
    
    // Status handler
    const unsubStatus = wsClient.onStatus(data => {
      console.log(`useVMWebSocket: Received status update`, data);
      setStatus(data);
    });
    
    // Metrics handler
    const unsubMetrics = wsClient.onMetrics(data => {
      console.log(`useVMWebSocket: Received metrics update`, data);
      setMetrics(data);
    });
    
    // Error handler
    const unsubError = wsClient.onError(data => {
      console.log(`useVMWebSocket: Received error`, data);
      setError(data);
    });
    
    // Console handler
    const unsubConsole = wsClient.onConsole(data => {
      console.log(`useVMWebSocket: Received console data`, data);
      setConsoleData(prev => [...prev, data]);
    });
    
    // Connect to WebSocket
    console.log(`useVMWebSocket: Connecting to WebSocket`);
    wsClient.connect();
    setClient(wsClient);
    
    // Clean up on unmount
    return () => {
      unsubConnection();
      unsubStatus();
      unsubMetrics();
      unsubError();
      unsubConsole();
      wsClient.disconnect();
    };
  }, [vmName, isConsole]);
  
  // Send command
  const sendCommand = (action: 'start' | 'stop' | 'restart' | 'suspend' | 'resume') => {
    if (client) {
      client.sendCommand(action);
    }
  };
  
  // Send console input
  const sendConsoleInput = (content: string) => {
    if (client && isConsole) {
      client.sendConsoleInput(content);
    }
  };
  
  // Clear console
  const clearConsole = () => {
    setConsoleData([]);
  };
  
  return {
    connected,
    status,
    metrics,
    error,
    consoleData,
    sendCommand,
    sendConsoleInput,
    clearConsole,
  };
};

/**
 * Hook for VM monitoring only (no console)
 */
export const useVMMonitor = (vmName: string): Omit<UseVMWebSocketReturn, 'consoleData' | 'sendConsoleInput' | 'clearConsole'> => {
  const {
    connected,
    status,
    metrics,
    error,
    sendCommand,
  } = useVMWebSocket(vmName, false);
  
  return {
    connected,
    status,
    metrics,
    error,
    sendCommand,
  };
};

/**
 * Hook for VM console only
 */
export const useVMConsole = (vmName: string): Omit<UseVMWebSocketReturn, 'metrics'> => {
  const {
    connected,
    status,
    error,
    consoleData,
    sendCommand,
    sendConsoleInput,
    clearConsole,
  } = useVMWebSocket(vmName, true);
  
  return {
    connected,
    status,
    error,
    consoleData,
    sendCommand,
    sendConsoleInput,
    clearConsole,
  };
};