import React, { useState } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { getVM, startVM, stopVM, deleteVM } from '@/api/vm';
import { useNavigate, useParams } from '@tanstack/react-router';
import { Button } from '@/components/ui/button';
import { VMStatusBadge } from '@/components/vm/vm-status-badge';
import { formatBytes, formatDate, cn } from '@/lib/utils';
import { Card, CardContent, CardHeader, CardTitle } from '@/lib/components';
import { useVMMonitor } from '@/hooks/useVMWebSocket';
import { VMMetricsChart } from '@/components/vm/vm-metrics';
import { VMConsole } from '@/components/vm/vm-console';
import { LuActivity, LuTerminal } from 'react-icons/lu';

// Icons
import { 
  LuPlay, 
  LuSquare, 
  LuTrash2, 
  LuDownload, 
  LuArrowLeft,
  LuHardDrive,
  LuNetwork,
  LuCpu,
  LuTriangle as LuAlertTriangle
} from 'react-icons/lu';

export const VMDetailPage: React.FC = () => {
  // Access the name param from the route
  const params = useParams({ strict: false });
  const name = params?.name as string;
  
  const navigate = useNavigate();
  const queryClient = useQueryClient();
  const [showConfirmation, setShowConfirmation] = useState(false);
  const [activeTab, setActiveTab] = useState('overview');
  
  // Fetch VM details
  const { data: vm, isLoading, isError, error } = useQuery({
    queryKey: ['vm', name],
    queryFn: () => getVM(name),
    retry: 3,         // Retry failed requests 3 times
    retryDelay: 1000, // Wait 1 second between retries
  });

  // WebSocket monitoring
  const {
    connected: wsConnected,
    status: wsStatus,
    metrics: wsMetrics,
    error: wsError,
    sendCommand
  } = useVMMonitor(name);
  
  console.log('VM Detail Page - WebSocket state:', { 
    connected: wsConnected, 
    hasStatus: !!wsStatus,
    hasMetrics: !!wsMetrics,
    error: wsError
  });

  // Start VM mutation
  const startMutation = useMutation({
    mutationFn: startVM,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['vm', name] });
    },
  });

  // Stop VM mutation
  const stopMutation = useMutation({
    mutationFn: stopVM,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['vm', name] });
    },
  });

  // Delete VM mutation
  const deleteMutation = useMutation({
    mutationFn: deleteVM,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['vms'] });
      navigate({ to: '/vms' });
    },
  });

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

  const handleStop = () => {
    if (vm) {
      // Use WebSocket command if connected, otherwise fallback to REST API
      if (wsConnected) {
        sendCommand('stop');
      } else {
        stopMutation.mutate(vm.name);
      }
    }
  };
  
  const handleRestart = () => {
    if (vm) {
      // Use WebSocket command if connected, otherwise fallback to REST API
      if (wsConnected) {
        sendCommand('restart');
      } else {
        // For REST API, we need to stop then start
        stopMutation.mutate(vm.name, {
          onSuccess: () => {
            setTimeout(() => {
              startMutation.mutate(vm.name);
            }, 2000);
          }
        });
      }
    }
  };

  const handleDelete = () => {
    setShowConfirmation(true);
  };

  const confirmDelete = () => {
    if (vm) {
      deleteMutation.mutate(vm.name);
    }
  };

  const cancelDelete = () => {
    setShowConfirmation(false);
  };

  const handleExport = () => {
    if (vm) {
      navigate({ to: '/vms/$name/export', params: { name: vm.name } });
    }
  };

  // Handle loading, error, or invalid VM data
  if (isLoading) {
    return <div className="text-center py-8">Loading VM details...</div>;
  }

  if (isError || !vm || !vm.cpu || !vm.memory || !vm.disks || !vm.networks) {
    // Detailed debug information
    console.log("VM data received:", vm);
    console.log("Error state:", isError);
    console.log("Error object:", error);
    
    return (
      <div className="text-center py-8 text-red-500">
        <p>Error loading VM details. The VM might not exist or is still initializing.</p>
        
        {error instanceof Error && (
          <p className="mt-2 text-sm">{error.message}</p>
        )}
        
        {vm && (
          <div className="mt-4 p-4 bg-red-500/10 border border-red-500 rounded text-left text-sm max-w-2xl mx-auto">
            <p className="font-bold mb-2">Debug Information:</p>
            <ul className="list-disc pl-5 space-y-1">
              {!vm.cpu && <li>CPU information is missing</li>}
              {!vm.memory && <li>Memory information is missing</li>}
              {!vm.disks && <li>Disk information is missing</li>}
              {!vm.networks && <li>Network information is missing</li>}
              <li>Name: {vm.name || 'missing'}</li>
              <li>UUID: {vm.uuid || 'missing'}</li>
              <li>Status: {vm.status || 'missing'}</li>
            </ul>
            <pre className="mt-3 p-2 bg-gray-900 text-gray-100 rounded overflow-auto text-xs">
              {JSON.stringify(vm, null, 2)}
            </pre>
          </div>
        )}
        
        <div className="mt-4 flex space-x-4 justify-center">
          <Button 
            onClick={() => queryClient.invalidateQueries({ queryKey: ['vm', name] })}
          >
            Retry
          </Button>
          <Button 
            variant="outline"
            onClick={() => navigate({ to: '/vms' })}
          >
            Back to VMs
          </Button>
        </div>
      </div>
    );
  }

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex flex-col md:flex-row md:items-center justify-between gap-4">
        <div className="flex items-center space-x-4">
          <Button variant="outline" size="icon" onClick={() => navigate({ to: '/vms' })}>
            <LuArrowLeft className="h-4 w-4" />
          </Button>
          <div>
            <h1 className="text-2xl font-bold">{vm.name}</h1>
            <div className="flex items-center space-x-2 mt-1">
              {/* Use WebSocket status if available, otherwise use REST API status */}
              <VMStatusBadge status={wsStatus?.status || vm.status} />
              <span className="text-sm text-muted-foreground">
                Created {formatDate(vm.createdAt)}
              </span>
              {wsConnected && (
                <span className="text-xs px-2 py-1 bg-green-100 text-green-800 rounded-full">
                  Live
                </span>
              )}
            </div>
          </div>
        </div>
        
        <div className="flex flex-wrap gap-2">
          {(wsStatus?.status || vm.status) !== 'running' && (
            <Button 
              variant="outline"
              onClick={handleStart}
              disabled={startMutation.isPending}
              className="text-green-500 border-green-500 hover:bg-green-500/10"
            >
              <LuPlay className="mr-2 h-4 w-4" />
              {startMutation.isPending ? 'Starting...' : 'Start'}
            </Button>
          )}
          
          {(wsStatus?.status || vm.status) === 'running' && (
            <>
              <Button 
                variant="outline"
                onClick={handleStop}
                disabled={stopMutation.isPending}
                className="text-yellow-500 border-yellow-500 hover:bg-yellow-500/10"
              >
                <LuSquare className="mr-2 h-4 w-4" />
                {stopMutation.isPending ? 'Stopping...' : 'Stop'}
              </Button>
              
              <Button 
                variant="outline"
                onClick={handleRestart}
                className="text-orange-500 border-orange-500 hover:bg-orange-500/10"
              >
                <LuPlay className="mr-2 h-4 w-4" />
                Restart
              </Button>
            </>
          )}
          
          <Button 
            variant="outline"
            onClick={handleExport}
            className="text-blue-500 border-blue-500 hover:bg-blue-500/10"
          >
            <LuDownload className="mr-2 h-4 w-4" />
            Export
          </Button>
          
          <Button 
            variant="outline"
            onClick={handleDelete}
            disabled={deleteMutation.isPending}
            className="text-red-500 border-red-500 hover:bg-red-500/10"
          >
            <LuTrash2 className="mr-2 h-4 w-4" />
            {deleteMutation.isPending ? 'Deleting...' : 'Delete'}
          </Button>
        </div>
      </div>

      {/* Delete Confirmation Dialog */}
      {showConfirmation && (
        <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50">
          <div className="bg-card border rounded-lg p-6 max-w-md mx-auto">
            <div className="flex items-center text-red-500 mb-4">
              <LuAlertTriangle className="h-6 w-6 mr-2" />
              <h3 className="text-lg font-semibold">Confirm Deletion</h3>
            </div>
            <p className="mb-6">
              Are you sure you want to delete the VM <strong>{vm.name}</strong>? This action cannot be undone.
            </p>
            <div className="flex justify-end space-x-2">
              <Button variant="outline" onClick={cancelDelete}>
                Cancel
              </Button>
              <Button variant="destructive" onClick={confirmDelete}>
                Delete
              </Button>
            </div>
          </div>
        </div>
      )}
      
      {/* VM Description */}
      {vm.description && (
        <Card>
          <CardContent className="pt-6">
            <p>{vm.description}</p>
          </CardContent>
        </Card>
      )}
      
      {/* Tabs */}
      <div className="space-y-4">
        <div className="inline-flex h-10 items-center justify-center rounded-md bg-muted p-1 text-muted-foreground">
          <button
            onClick={() => setActiveTab('overview')}
            className={cn(
              "inline-flex items-center justify-center whitespace-nowrap rounded-sm px-3 py-1.5 text-sm font-medium ring-offset-background transition-all focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2",
              activeTab === 'overview' && "bg-background text-foreground shadow-sm"
            )}
          >
            Overview
          </button>
          <button
            onClick={() => setActiveTab('monitoring')}
            disabled={!wsConnected}
            className={cn(
              "inline-flex items-center justify-center whitespace-nowrap rounded-sm px-3 py-1.5 text-sm font-medium ring-offset-background transition-all focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 disabled:pointer-events-none disabled:opacity-50",
              activeTab === 'monitoring' && "bg-background text-foreground shadow-sm"
            )}
          >
            <div className="flex items-center">
              <LuActivity className="mr-2 h-4 w-4" />
              Monitoring
              {wsConnected && (
                <span className="ml-2 w-2 h-2 bg-green-500 rounded-full"></span>
              )}
            </div>
          </button>
          <button
            onClick={() => setActiveTab('console')}
            disabled={!wsConnected}
            className={cn(
              "inline-flex items-center justify-center whitespace-nowrap rounded-sm px-3 py-1.5 text-sm font-medium ring-offset-background transition-all focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 disabled:pointer-events-none disabled:opacity-50",
              activeTab === 'console' && "bg-background text-foreground shadow-sm"
            )}
          >
            <div className="flex items-center">
              <LuTerminal className="mr-2 h-4 w-4" />
              Console
              {wsConnected && (
                <span className="ml-2 w-2 h-2 bg-green-500 rounded-full"></span>
              )}
            </div>
          </button>
        </div>
        
        {/* Tab Content */}
        {activeTab === 'overview' && (
          <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
            {/* CPU and Memory */}
            <Card>
              <CardHeader className="pb-2">
                <CardTitle className="flex items-center text-lg">
                  <LuCpu className="mr-2 h-4 w-4" />
                  System Resources
                </CardTitle>
              </CardHeader>
              <CardContent>
                <div className="space-y-4">
                  <div>
                    <div className="text-sm font-medium text-muted-foreground mb-1">CPU</div>
                    <div className="flex flex-col gap-1">
                      <div className="flex justify-between">
                        <span>Virtual CPUs</span>
                        <span className="font-medium">{vm.cpu.count}</span>
                      </div>
                      <div className="flex justify-between">
                        <span>Model</span>
                        <span className="font-medium">{vm.cpu.model}</span>
                      </div>
                      <div className="flex justify-between">
                        <span>Topology</span>
                        <span className="font-medium">
                          {vm.cpu.sockets} socket{vm.cpu.sockets !== 1 ? 's' : ''}, 
                          {vm.cpu.cores} core{vm.cpu.cores !== 1 ? 's' : ''}, 
                          {vm.cpu.threads} thread{vm.cpu.threads !== 1 ? 's' : ''}
                        </span>
                      </div>
                    </div>
                  </div>
                  
                  <div>
                    <div className="text-sm font-medium text-muted-foreground mb-1">Memory</div>
                    <div className="flex justify-between">
                      <span>RAM</span>
                      <span className="font-medium">{formatBytes(vm.memory.sizeBytes)}</span>
                    </div>
                  </div>
                </div>
              </CardContent>
            </Card>
            
            {/* Storage */}
            <Card>
              <CardHeader className="pb-2">
                <CardTitle className="flex items-center text-lg">
                  <LuHardDrive className="mr-2 h-4 w-4" />
                  Storage
                </CardTitle>
              </CardHeader>
              <CardContent>
                <div className="space-y-4">
                  {vm.disks?.map((disk, index) => (
                    <div key={index} className="pb-3 border-b last:border-0 last:pb-0">
                      <div className="flex justify-between items-center mb-2">
                        <span className="font-medium">Disk {index + 1}</span>
                        <VMStatusBadge 
                          status={disk.readOnly ? 'paused' : 'running'} 
                          className="text-xs py-0" 
                        >
                          {disk.readOnly ? 'Read Only' : 'Read/Write'}
                        </VMStatusBadge>
                      </div>
                      <div className="grid grid-cols-2 gap-2 text-sm">
                        <div>Path:</div>
                        <div className="truncate">{disk.path}</div>
                        
                        <div>Size:</div>
                        <div>{formatBytes(disk.sizeBytes)}</div>
                        
                        <div>Format:</div>
                        <div>{disk.format}</div>
                        
                        <div>Bus:</div>
                        <div>{disk.bus}</div>
                        
                        <div>Storage Pool:</div>
                        <div>{disk.storagePool}</div>
                        
                        {disk.bootable && (
                          <>
                            <div>Bootable:</div>
                            <div>Yes</div>
                          </>
                        )}
                      </div>
                    </div>
                  ))}
                  
                  {(!vm.disks || vm.disks.length === 0) && (
                    <div className="text-muted-foreground text-sm text-center py-4">
                      No disks attached
                    </div>
                  )}
                </div>
              </CardContent>
            </Card>
            
            {/* Networks */}
            <Card className="md:col-span-2">
              <CardHeader className="pb-2">
                <CardTitle className="flex items-center text-lg">
                  <LuNetwork className="mr-2 h-4 w-4" />
                  Network Interfaces
                </CardTitle>
              </CardHeader>
              <CardContent>
                <div className="space-y-4">
                  {vm.networks?.map((network, index) => (
                    <div key={index} className="pb-3 border-b last:border-0 last:pb-0">
                      <div className="font-medium mb-2">Network Interface {index + 1}</div>
                      <div className="grid grid-cols-2 sm:grid-cols-4 gap-x-4 gap-y-2 text-sm">
                        <div>Type:</div>
                        <div>{network.type}</div>
                        
                        <div>Source:</div>
                        <div>{network.source}</div>
                        
                        <div>Model:</div>
                        <div>{network.model}</div>
                        
                        <div>MAC Address:</div>
                        <div>{network.macAddress}</div>
                        
                        {network.ipAddress && (
                          <>
                            <div>IP Address:</div>
                            <div>{network.ipAddress}</div>
                          </>
                        )}
                        
                        {network.ipAddressV6 && (
                          <>
                            <div>IPv6 Address:</div>
                            <div>{network.ipAddressV6}</div>
                          </>
                        )}
                      </div>
                    </div>
                  ))}
                  
                  {(!vm.networks || vm.networks.length === 0) && (
                    <div className="text-muted-foreground text-sm text-center py-4">
                      No network interfaces attached
                    </div>
                  )}
                </div>
              </CardContent>
            </Card>
          </div>
        )}
        
        {/* Monitoring Tab */}
        {activeTab === 'monitoring' && (
          <div>
            {wsConnected && wsMetrics ? (
              <VMMetricsChart vmName={vm.name} metrics={wsMetrics} />
            ) : (
              <div className="text-center py-8">
                <p className="text-muted-foreground">
                  {wsConnected ? 'Waiting for metrics data...' : 'WebSocket connection required for real-time monitoring'}
                </p>
              </div>
            )}
          </div>
        )}
        
        {/* Console Tab */}
        {activeTab === 'console' && (
          <div>
            {wsConnected ? (
              <VMConsole vmName={vm.name} className="mt-4" />
            ) : (
              <div className="text-center py-8">
                <p className="text-muted-foreground">
                  WebSocket connection required for console access
                </p>
              </div>
            )}
          </div>
        )}
      </div>
    </div>
  );
};