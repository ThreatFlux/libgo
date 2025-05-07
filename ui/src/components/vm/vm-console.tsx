import React, { useEffect, useRef, useState } from 'react';
import { Terminal } from 'xterm';
import { FitAddon } from 'xterm-addon-fit';
import { WebLinksAddon } from 'xterm-addon-web-links';
import { useVMConsole } from '@/hooks/useVMWebSocket';
import { Button } from '@/components/ui/button';
import 'xterm/css/xterm.css';

interface VMConsoleProps {
  vmName: string;
  className?: string;
}

export const VMConsole: React.FC<VMConsoleProps> = ({ 
  vmName,
  className = '' 
}) => {
  const terminalRef = useRef<HTMLDivElement>(null);
  const [terminal, setTerminal] = useState<Terminal | null>(null);
  const [fitAddon, setFitAddon] = useState<FitAddon | null>(null);
  const [isFullscreen, setIsFullscreen] = useState(false);
  
  // Connect to VM console via WebSocket
  const {
    connected,
    status,
    consoleData,
    error,
    sendConsoleInput,
    clearConsole
  } = useVMConsole(vmName);
  
  // Initialize xterm.js terminal
  useEffect(() => {
    if (!terminalRef.current) return;
    
    // Create terminal
    const term = new Terminal({
      cursorBlink: true,
      fontFamily: 'Menlo, Monaco, "Courier New", monospace',
      fontSize: 14,
      theme: {
        background: '#1a1b26',
        foreground: '#c0caf5',
        cursor: '#c0caf5',
        selection: 'rgba(128, 203, 196, 0.3)',
        black: '#1a1b26',
        red: '#f7768e',
        green: '#9ece6a',
        yellow: '#e0af68',
        blue: '#7aa2f7',
        magenta: '#bb9af7',
        cyan: '#7dcfff',
        white: '#c0caf5',
        brightBlack: '#414868',
        brightRed: '#f7768e',
        brightGreen: '#9ece6a',
        brightYellow: '#e0af68',
        brightBlue: '#7aa2f7',
        brightMagenta: '#bb9af7',
        brightCyan: '#7dcfff',
        brightWhite: '#c0caf5',
      },
    });
    
    // Create addons
    const fit = new FitAddon();
    const webLinks = new WebLinksAddon();
    
    // Load addons
    term.loadAddon(fit);
    term.loadAddon(webLinks);
    
    // Open terminal
    term.open(terminalRef.current);
    fit.fit();
    
    // Set up input handling
    term.onData(data => {
      sendConsoleInput(data);
    });
    
    // Store terminal and fit addon
    setTerminal(term);
    setFitAddon(fit);
    
    // Write welcome message
    term.writeln('Connecting to VM console...');
    term.writeln('');
    
    // Clean up on unmount
    return () => {
      term.dispose();
    };
  }, [terminalRef]);
  
  // Handle window resize
  useEffect(() => {
    const handleResize = () => {
      if (fitAddon) {
        try {
          fitAddon.fit();
        } catch (e) {
          console.error('Failed to fit terminal:', e);
        }
      }
    };
    
    window.addEventListener('resize', handleResize);
    return () => {
      window.removeEventListener('resize', handleResize);
    };
  }, [fitAddon]);
  
  // Handle connection status changes
  useEffect(() => {
    if (!terminal) return;
    
    if (connected) {
      terminal.writeln('\r\nConnected to VM console.');
      terminal.writeln('');
    } else {
      terminal.writeln('\r\nDisconnected from VM console. Reconnecting...');
    }
  }, [connected, terminal]);
  
  // Handle VM status changes
  useEffect(() => {
    if (!terminal || !status) return;
    
    terminal.writeln(`\r\nVM status: ${status.status}`);
  }, [status, terminal]);
  
  // Handle console data
  useEffect(() => {
    if (!terminal || consoleData.length === 0) return;
    
    // Process only the latest console data
    const lastData = consoleData[consoleData.length - 1];
    terminal.write(lastData.content);
  }, [consoleData, terminal]);
  
  // Handle errors
  useEffect(() => {
    if (!terminal || !error) return;
    
    terminal.writeln(`\r\nError: ${error.message} (${error.code})`);
  }, [error, terminal]);
  
  // Toggle fullscreen
  const toggleFullscreen = () => {
    setIsFullscreen(!isFullscreen);
    setTimeout(() => {
      if (fitAddon) {
        fitAddon.fit();
      }
    }, 100);
  };
  
  // Clear terminal
  const handleClear = () => {
    if (terminal) {
      terminal.clear();
    }
    clearConsole();
  };
  
  return (
    <div className={`flex flex-col ${isFullscreen ? 'fixed inset-0 z-50 bg-black p-4' : className}`}>
      <div className="flex justify-between mb-2">
        <div className="flex items-center">
          <div className={`w-3 h-3 rounded-full mr-2 ${connected ? 'bg-green-500' : 'bg-red-500'}`}></div>
          <span className="text-sm">
            {connected ? `Connected (${status?.status || 'unknown'})` : 'Disconnecting...'}
          </span>
        </div>
        <div className="flex space-x-2">
          <Button
            size="sm"
            variant="outline"
            onClick={handleClear}
          >
            Clear
          </Button>
          <Button
            size="sm"
            variant="outline"
            onClick={toggleFullscreen}
          >
            {isFullscreen ? 'Exit Fullscreen' : 'Fullscreen'}
          </Button>
        </div>
      </div>
      
      <div
        ref={terminalRef}
        className={`flex-1 rounded-md overflow-hidden border border-gray-600 ${isFullscreen ? 'h-[calc(100vh-80px)]' : 'h-[400px]'}`}
      />
    </div>
  );
};