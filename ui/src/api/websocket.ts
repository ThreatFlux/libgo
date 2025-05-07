import { VM, VMStatus } from '@/types/api';

// WebSocket message types
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

// WebSocket message interface
export interface WSMessage {
  type: MessageType;
  timestamp: string;
  data: any;
}

// VM metrics data
export interface VMMetrics {
  cpu: {
    utilization: number;
  };
  memory: {
    used: number;
    total: number;
  };
  network: {
    rxBytes: number;
    txBytes: number;
  };
  disk: {
    readBytes: number;
    writeBytes: number;
  };
}

// VM status data
export interface VMStatusData {
  status: VMStatus;
  lastStateChange: string;
  uptime: number;
}

// Console data
export interface ConsoleData {
  content: string;
  eof: boolean;
}

// Command data
export interface CommandData {
  action: 'start' | 'stop' | 'restart' | 'suspend' | 'resume';
  requestId?: string;
}

// Response data
export interface ResponseData {
  requestId: string;
  success: boolean;
  message: string;
}

// Error data
export interface ErrorData {
  code: string;
  message: string;
}

// Event handlers
export type MessageHandler = (message: WSMessage) => void;
export type StatusHandler = (data: VMStatusData) => void;
export type MetricsHandler = (data: VMMetrics) => void;
export type ConsoleHandler = (data: ConsoleData) => void;
export type ResponseHandler = (data: ResponseData) => void;
export type ErrorHandler = (data: ErrorData) => void;
export type ConnectionHandler = (connected: boolean) => void;

// WebSocket client
export class VMWebSocketClient {
  private socket: WebSocket | null = null;
  private connected = false;
  private reconnectTimer: ReturnType<typeof setTimeout> | null = null;
  private pingInterval: ReturnType<typeof setInterval> | null = null;
  private vmName: string;
  private isConsole: boolean;
  private messageHandlers: MessageHandler[] = [];
  private statusHandlers: StatusHandler[] = [];
  private metricsHandlers: MetricsHandler[] = [];
  private consoleHandlers: ConsoleHandler[] = [];
  private responseHandlers: ResponseHandler[] = [];
  private errorHandlers: ErrorHandler[] = [];
  private connectionHandlers: ConnectionHandler[] = [];

  constructor(vmName: string, isConsole = false) {
    this.vmName = vmName;
    this.isConsole = isConsole;
  }

  // Connect to WebSocket
  connect(): void {
    if (this.socket) {
      return;
    }

    // Get authentication token
    const token = localStorage.getItem('token');
    console.log('Attempting WebSocket connection with token available:', !!token);
    if (!token) {
      console.error('No authentication token found in localStorage');
      this.handleError({
        code: 'AUTH_ERROR',
        message: 'No authentication token available',
      });
      return;
    }
    
    // Check if the token looks valid (at least has proper format)
    if (token.split('.').length !== 3) {
      console.warn('Token does not appear to be a valid JWT');
    }

    // Determine WebSocket URL
    const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
    const path = this.isConsole ? `/ws/vms/${this.vmName}/console` : `/ws/vms/${this.vmName}`;
    
    // For development, use relative paths to ensure proxy works correctly
    const wsHost = window.location.host;
    const url = `${protocol}//${wsHost}${path}?token=${token}`;
    
    console.log(`Connecting to WebSocket URL: ${url.replace(/token=.*/, 'token=REDACTED')}`);

    try {
      // Create WebSocket connection
      this.socket = new WebSocket(url);

      // Set up event handlers
      this.socket.onopen = this.handleOpen.bind(this);
      this.socket.onclose = this.handleClose.bind(this);
      this.socket.onerror = this.handleSocketError.bind(this);
      this.socket.onmessage = this.handleMessage.bind(this);
    } catch (error) {
      console.error('Failed to create WebSocket connection:', error);
      this.scheduleReconnect();
    }
  }

  // Disconnect WebSocket
  disconnect(): void {
    if (this.reconnectTimer) {
      clearTimeout(this.reconnectTimer);
      this.reconnectTimer = null;
    }

    if (this.pingInterval) {
      clearInterval(this.pingInterval);
      this.pingInterval = null;
    }

    if (this.socket) {
      this.socket.close();
      this.socket = null;
    }

    this.connected = false;
  }

  // Send a command to the server
  sendCommand(action: CommandData['action']): void {
    const requestId = `cmd-${Date.now()}`;
    this.send({
      type: 'command',
      timestamp: new Date().toISOString(),
      data: {
        action,
        requestId,
      },
    });
  }

  // Send console input
  sendConsoleInput(content: string): void {
    if (!this.isConsole) {
      console.warn('Cannot send console input on non-console connection');
      return;
    }

    this.send({
      type: 'console_input',
      timestamp: new Date().toISOString(),
      data: {
        content,
      },
    });
  }

  // Add event handlers
  onMessage(handler: MessageHandler): () => void {
    this.messageHandlers.push(handler);
    return () => {
      this.messageHandlers = this.messageHandlers.filter(h => h !== handler);
    };
  }

  onStatus(handler: StatusHandler): () => void {
    this.statusHandlers.push(handler);
    return () => {
      this.statusHandlers = this.statusHandlers.filter(h => h !== handler);
    };
  }

  onMetrics(handler: MetricsHandler): () => void {
    this.metricsHandlers.push(handler);
    return () => {
      this.metricsHandlers = this.metricsHandlers.filter(h => h !== handler);
    };
  }

  onConsole(handler: ConsoleHandler): () => void {
    this.consoleHandlers.push(handler);
    return () => {
      this.consoleHandlers = this.consoleHandlers.filter(h => h !== handler);
    };
  }

  onResponse(handler: ResponseHandler): () => void {
    this.responseHandlers.push(handler);
    return () => {
      this.responseHandlers = this.responseHandlers.filter(h => h !== handler);
    };
  }

  onError(handler: ErrorHandler): () => void {
    this.errorHandlers.push(handler);
    return () => {
      this.errorHandlers = this.errorHandlers.filter(h => h !== handler);
    };
  }

  onConnection(handler: ConnectionHandler): () => void {
    this.connectionHandlers.push(handler);
    return () => {
      this.connectionHandlers = this.connectionHandlers.filter(h => h !== handler);
    };
  }

  // Private methods
  private send(message: WSMessage): void {
    if (!this.socket || this.socket.readyState !== WebSocket.OPEN) {
      console.warn('Cannot send message: WebSocket not connected');
      return;
    }

    try {
      this.socket.send(JSON.stringify(message));
    } catch (error) {
      console.error('Failed to send WebSocket message:', error);
    }
  }

  private handleOpen(): void {
    console.log(`WebSocket connected to ${this.vmName}`);
    this.connected = true;

    // Start ping interval
    this.pingInterval = setInterval(() => {
      this.send({
        type: 'heartbeat',
        timestamp: new Date().toISOString(),
        data: {},
      });
    }, 30000);

    // Notify connection handlers
    this.connectionHandlers.forEach(handler => handler(true));
  }

  private handleClose(event: CloseEvent): void {
    console.log(`WebSocket disconnected from ${this.vmName}:`, event.code, event.reason);
    console.log('WebSocket close details:', {
      code: event.code,
      reason: event.reason,
      wasClean: event.wasClean
    });
    this.connected = false;

    // Clear ping interval
    if (this.pingInterval) {
      clearInterval(this.pingInterval);
      this.pingInterval = null;
    }

    // Notify connection handlers
    this.connectionHandlers.forEach(handler => handler(false));

    // Reconnect if not a clean close
    if (event.code !== 1000) {
      this.scheduleReconnect();
    }
  }

  private handleSocketError(event: Event): void {
    console.error(`WebSocket error for ${this.vmName}:`, event);
    console.error('WebSocket error details:', {
      type: event.type,
      target: event.target,
      timeStamp: event.timeStamp,
    });
    this.handleError({
      code: 'SOCKET_ERROR',
      message: 'WebSocket connection error',
    });
  }

  private handleMessage(event: MessageEvent): void {
    try {
      const message = JSON.parse(event.data) as WSMessage;
      
      // Call general message handlers
      this.messageHandlers.forEach(handler => handler(message));

      // Call type-specific handlers
      switch (message.type) {
        case 'status':
          this.statusHandlers.forEach(handler => handler(message.data as VMStatusData));
          break;
        case 'metrics':
          this.metricsHandlers.forEach(handler => handler(message.data as VMMetrics));
          break;
        case 'console':
          this.consoleHandlers.forEach(handler => handler(message.data as ConsoleData));
          break;
        case 'response':
          this.responseHandlers.forEach(handler => handler(message.data as ResponseData));
          break;
        case 'error':
          this.errorHandlers.forEach(handler => handler(message.data as ErrorData));
          break;
      }
    } catch (error) {
      console.error('Failed to parse WebSocket message:', error, event.data);
    }
  }

  private handleError(error: ErrorData): void {
    console.error(`WebSocket error for ${this.vmName}:`, error);
    this.errorHandlers.forEach(handler => handler(error));
  }

  private scheduleReconnect(): void {
    if (this.reconnectTimer) {
      return;
    }

    // Reconnect after 5 seconds
    this.reconnectTimer = setTimeout(() => {
      this.reconnectTimer = null;
      console.log(`Attempting to reconnect to ${this.vmName}...`);
      this.connect();
    }, 5000);
  }
}