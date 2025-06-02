import React, { useState, useEffect } from 'react';
import { useForm, Controller } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import { z } from 'zod';
import { useMutation, useQueryClient } from '@tanstack/react-query';
import { useNavigate } from '@tanstack/react-router';
import { createVM } from '@/api/vm';
import { VMCreateParams } from '@/types/api';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardHeader, CardTitle } from '@/lib/components';
import { listBridgeNetworks, BridgeNetwork } from '@/api/bridge-network';

// Icons
import { 
  LuArrowLeft, 
  LuServer, 
  LuCheck, 
  LuCpu, 
  LuHardDrive, 
  LuNetwork,
  LuMemoryStick
} from 'react-icons/lu';

// Form schema
const createVMSchema = z.object({
  name: z.string().min(1, 'Name is required').max(50, 'Name is too long'),
  description: z.string().optional(),
  cpu: z.object({
    count: z.number().int().min(1, 'At least 1 vCPU required').max(32, 'Maximum 32 vCPUs'),
    model: z.string().min(1, 'CPU model is required'),
    socket: z.number().int().min(1, 'At least 1 socket required'),
    cores: z.number().int().min(1, 'At least 1 core required'),
    threads: z.number().int().min(1, 'At least 1 thread required')
  }),
  memory: z.object({
    sizeBytes: z.number().int().min(64 * 1024 * 1024, 'At least 64MB required')
  }),
  disk: z.object({
    sizeBytes: z.number().int().min(1 * 1024 * 1024 * 1024, 'At least 1GB required'),
    format: z.enum(['qcow2', 'raw']),
    sourceImage: z.string().optional(),
    storagePool: z.string().min(1, 'Storage pool is required'),
    bus: z.enum(['virtio', 'ide', 'sata', 'scsi']),
    cacheMode: z.enum(['none', 'writeback', 'writethrough', 'directsync', 'unsafe']),
    shareable: z.boolean().default(false),
    readOnly: z.boolean().default(false)
  }),
  network: z.object({
    type: z.enum(['bridge', 'network', 'direct']),
    source: z.string().min(1, 'Network source is required'),
    model: z.enum(['virtio', 'e1000', 'rtl8139']),
    macAddress: z.string().optional()
  }),
  cloudInit: z.object({
    userData: z.string().optional(),
    metaData: z.string().optional(),
    networkConfig: z.string().optional(),
    sshKeys: z.array(z.string()).optional()
  }).optional(),
  template: z.string().optional()
});

type CreateVMFormValues = z.infer<typeof createVMSchema>;

// Default values
const defaultValues: CreateVMFormValues = {
  name: '',
  description: '',
  cpu: {
    count: 1,
    model: 'host-model',
    socket: 1,
    cores: 1,
    threads: 1
  },
  memory: {
    sizeBytes: 1024 * 1024 * 1024 // 1GB
  },
  disk: {
    sizeBytes: 10 * 1024 * 1024 * 1024, // 10GB
    format: 'qcow2',
    storagePool: 'default',
    bus: 'virtio',
    cacheMode: 'none',
    shareable: false,
    readOnly: false
  },
  network: {
    type: 'network',
    source: 'default',
    model: 'virtio'
  },
  cloudInit: {
    sshKeys: []
  }
};

export const VMCreatePage: React.FC = () => {
  const navigate = useNavigate();
  const queryClient = useQueryClient();
  const [step, setStep] = useState(1);
  const [bridgeNetworks, setBridgeNetworks] = useState<BridgeNetwork[]>([]);
  const [bridgeNetworksLoading, setBridgeNetworksLoading] = useState(false);
  
  // Form setup
  const { 
    control, 
    handleSubmit, 
    watch,
    formState: { errors, isValid }
  } = useForm<CreateVMFormValues>({
    resolver: zodResolver(createVMSchema),
    defaultValues,
    mode: 'onChange'
  });

  // Watch network type to load appropriate networks
  const networkType = watch('network.type');

  // Load bridge networks when network type is 'network'
  useEffect(() => {
    if (networkType === 'network') {
      setBridgeNetworksLoading(true);
      listBridgeNetworks()
        .then(response => {
          setBridgeNetworks(response.networks);
        })
        .catch(error => {
          console.error('Failed to load bridge networks:', error);
          setBridgeNetworks([]);
        })
        .finally(() => {
          setBridgeNetworksLoading(false);
        });
    }
  }, [networkType]);

  // VM creation mutation
  const createMutation = useMutation({
    mutationFn: createVM,
    onSuccess: (data) => {
      queryClient.invalidateQueries({ queryKey: ['vms'] });
      
      // Add a small delay before navigation to allow backend to fully initialize VM
      setTimeout(() => {
        navigate({ 
          to: '/vms/$name',
          params: { name: data.vm.name },
          replace: true
        });
      }, 1500);
    }
  });

  // Form submission
  const onSubmit = (data: CreateVMFormValues) => {
    // Ensure description is set
    const vmData = {
      ...data,
      description: data.description || ""
    };
    createMutation.mutate(vmData as VMCreateParams);
  };

  // Watch values for dynamic UI
  const watchCPU = watch('cpu');
  const watchMemory = watch('memory');
  
  // Step navigation
  const nextStep = () => {
    setStep(step + 1);
  };

  const prevStep = () => {
    setStep(step - 1);
  };

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center space-x-4">
        <Button variant="outline" size="icon" onClick={() => navigate({ to: '/vms' })}>
          <LuArrowLeft className="h-4 w-4" />
        </Button>
        <h1 className="text-2xl font-bold">Create Virtual Machine</h1>
      </div>

      {/* Stepper */}
      <div className="flex items-center justify-center mb-8">
        <div className="flex items-center w-full max-w-3xl">
          <div className={`flex items-center justify-center w-8 h-8 rounded-full ${step >= 1 ? 'bg-primary text-primary-foreground' : 'bg-muted text-muted-foreground'}`}>
            1
          </div>
          <div className={`flex-1 h-1 mx-2 ${step >= 2 ? 'bg-primary' : 'bg-muted'}`}></div>
          <div className={`flex items-center justify-center w-8 h-8 rounded-full ${step >= 2 ? 'bg-primary text-primary-foreground' : 'bg-muted text-muted-foreground'}`}>
            2
          </div>
          <div className={`flex-1 h-1 mx-2 ${step >= 3 ? 'bg-primary' : 'bg-muted'}`}></div>
          <div className={`flex items-center justify-center w-8 h-8 rounded-full ${step >= 3 ? 'bg-primary text-primary-foreground' : 'bg-muted text-muted-foreground'}`}>
            3
          </div>
          <div className={`flex-1 h-1 mx-2 ${step >= 4 ? 'bg-primary' : 'bg-muted'}`}></div>
          <div className={`flex items-center justify-center w-8 h-8 rounded-full ${step >= 4 ? 'bg-primary text-primary-foreground' : 'bg-muted text-muted-foreground'}`}>
            4
          </div>
        </div>
      </div>

      <form onSubmit={handleSubmit(onSubmit)}>
        {/* Step 1: Basic Information */}
        {step === 1 && (
          <Card>
            <CardHeader>
              <CardTitle className="flex items-center text-xl">
                <LuServer className="mr-2 h-5 w-5" />
                Basic Information
              </CardTitle>
            </CardHeader>
            <CardContent className="space-y-4">
              <div className="space-y-2">
                <label htmlFor="name" className="block text-sm font-medium">
                  Name *
                </label>
                <Controller
                  name="name"
                  control={control}
                  render={({ field }) => (
                    <input
                      {...field}
                      id="name"
                      type="text"
                      className="w-full p-2 border rounded-md bg-background"
                      placeholder="Enter VM name"
                    />
                  )}
                />
                {errors.name && (
                  <p className="text-sm text-red-500">{errors.name.message}</p>
                )}
              </div>
              
              <div className="space-y-2">
                <label htmlFor="description" className="block text-sm font-medium">
                  Description
                </label>
                <Controller
                  name="description"
                  control={control}
                  render={({ field }) => (
                    <textarea
                      {...field}
                      id="description"
                      rows={3}
                      className="w-full p-2 border rounded-md bg-background"
                      placeholder="Optional description"
                    />
                  )}
                />
              </div>
              
              <div className="space-y-2">
                <label htmlFor="template" className="block text-sm font-medium">
                  Template
                </label>
                <Controller
                  name="template"
                  control={control}
                  render={({ field }) => (
                    <select
                      {...field}
                      id="template"
                      className="w-full p-2 border rounded-md bg-background"
                    >
                      <option value="">None (Custom VM)</option>
                      <option value="ubuntu-2404">Ubuntu 24.04</option>
                      <option value="windows-server-2022">Windows Server 2022</option>
                    </select>
                  )}
                />
                <p className="text-xs text-muted-foreground">
                  Select a template to use predefined settings or choose "None" for custom configuration
                </p>
              </div>
            </CardContent>
          </Card>
        )}
        
        {/* Step 2: CPU and Memory */}
        {step === 2 && (
          <Card>
            <CardHeader>
              <CardTitle className="flex items-center text-xl">
                <LuCpu className="mr-2 h-5 w-5" />
                CPU & Memory
              </CardTitle>
            </CardHeader>
            <CardContent className="space-y-6">
              <div className="space-y-4">
                <h3 className="text-lg font-medium">CPU Configuration</h3>
                
                <div className="space-y-2">
                  <label htmlFor="cpu-count" className="block text-sm font-medium">
                    vCPU Count *
                  </label>
                  <Controller
                    name="cpu.count"
                    control={control}
                    render={({ field }) => (
                      <input
                        {...field}
                        id="cpu-count"
                        type="number"
                        min="1"
                        max="32"
                        onChange={(e) => field.onChange(parseInt(e.target.value))}
                        className="w-full p-2 border rounded-md bg-background"
                      />
                    )}
                  />
                  {errors.cpu?.count && (
                    <p className="text-sm text-red-500">{errors.cpu.count.message}</p>
                  )}
                </div>
                
                <div className="space-y-2">
                  <label htmlFor="cpu-model" className="block text-sm font-medium">
                    CPU Model *
                  </label>
                  <Controller
                    name="cpu.model"
                    control={control}
                    render={({ field }) => (
                      <select
                        {...field}
                        id="cpu-model"
                        className="w-full p-2 border rounded-md bg-background"
                      >
                        <option value="host-model">Host Model (recommended)</option>
                        <option value="host-passthrough">Host Passthrough</option>
                        <option value="qemu64">QEMU 64-bit</option>
                        <option value="core2duo">Intel Core2 Duo</option>
                        <option value="Nehalem">Intel Nehalem</option>
                        <option value="Skylake-Client">Intel Skylake Client</option>
                        <option value="Opteron_G5">AMD Opteron G5</option>
                      </select>
                    )}
                  />
                </div>
                
                <div className="grid grid-cols-3 gap-4">
                  <div className="space-y-2">
                    <label htmlFor="cpu-socket" className="block text-sm font-medium">
                      Sockets
                    </label>
                    <Controller
                      name="cpu.socket"
                      control={control}
                      render={({ field }) => (
                        <input
                          {...field}
                          id="cpu-socket"
                          type="number"
                          min="1"
                          onChange={(e) => field.onChange(parseInt(e.target.value))}
                          className="w-full p-2 border rounded-md bg-background"
                        />
                      )}
                    />
                  </div>
                  
                  <div className="space-y-2">
                    <label htmlFor="cpu-cores" className="block text-sm font-medium">
                      Cores per Socket
                    </label>
                    <Controller
                      name="cpu.cores"
                      control={control}
                      render={({ field }) => (
                        <input
                          {...field}
                          id="cpu-cores"
                          type="number"
                          min="1"
                          onChange={(e) => field.onChange(parseInt(e.target.value))}
                          className="w-full p-2 border rounded-md bg-background"
                        />
                      )}
                    />
                  </div>
                  
                  <div className="space-y-2">
                    <label htmlFor="cpu-threads" className="block text-sm font-medium">
                      Threads per Core
                    </label>
                    <Controller
                      name="cpu.threads"
                      control={control}
                      render={({ field }) => (
                        <input
                          {...field}
                          id="cpu-threads"
                          type="number"
                          min="1"
                          onChange={(e) => field.onChange(parseInt(e.target.value))}
                          className="w-full p-2 border rounded-md bg-background"
                        />
                      )}
                    />
                  </div>
                </div>
                
                <p className="text-xs text-muted-foreground">
                  Total vCPUs = Sockets × Cores × Threads = {watchCPU.socket * watchCPU.cores * watchCPU.threads}
                </p>
              </div>
              
              <hr className="border-border" />
              
              <div className="space-y-4">
                <h3 className="text-lg font-medium flex items-center">
                  <LuMemoryStick className="mr-2 h-4 w-4" />
                  Memory
                </h3>
                
                <div className="space-y-2">
                  <label htmlFor="memory-size" className="block text-sm font-medium">
                    Memory Size (MB) *
                  </label>
                  <Controller
                    name="memory.sizeBytes"
                    control={control}
                    render={({ field }) => (
                      <input
                        id="memory-size"
                        type="number"
                        min="64"
                        step="128"
                        value={Math.floor(field.value / (1024 * 1024))}
                        onChange={(e) => field.onChange(parseInt(e.target.value) * 1024 * 1024)}
                        className="w-full p-2 border rounded-md bg-background"
                      />
                    )}
                  />
                  {errors.memory?.sizeBytes && (
                    <p className="text-sm text-red-500">{errors.memory.sizeBytes.message}</p>
                  )}
                  <p className="text-xs text-muted-foreground">
                    Approximately {(watchMemory.sizeBytes / (1024 * 1024 * 1024)).toFixed(2)} GB
                  </p>
                </div>
              </div>
            </CardContent>
          </Card>
        )}
        
        {/* Step 3: Storage */}
        {step === 3 && (
          <Card>
            <CardHeader>
              <CardTitle className="flex items-center text-xl">
                <LuHardDrive className="mr-2 h-5 w-5" />
                Storage
              </CardTitle>
            </CardHeader>
            <CardContent className="space-y-4">
              <div className="space-y-2">
                <label htmlFor="disk-size" className="block text-sm font-medium">
                  Disk Size (GB) *
                </label>
                <Controller
                  name="disk.sizeBytes"
                  control={control}
                  render={({ field }) => (
                    <input
                      id="disk-size"
                      type="number"
                      min="1"
                      value={Math.floor(field.value / (1024 * 1024 * 1024))}
                      onChange={(e) => field.onChange(parseInt(e.target.value) * 1024 * 1024 * 1024)}
                      className="w-full p-2 border rounded-md bg-background"
                    />
                  )}
                />
                {errors.disk?.sizeBytes && (
                  <p className="text-sm text-red-500">{errors.disk.sizeBytes.message}</p>
                )}
              </div>
              
              <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                <div className="space-y-2">
                  <label htmlFor="disk-format" className="block text-sm font-medium">
                    Disk Format *
                  </label>
                  <Controller
                    name="disk.format"
                    control={control}
                    render={({ field }) => (
                      <select
                        {...field}
                        id="disk-format"
                        className="w-full p-2 border rounded-md bg-background"
                      >
                        <option value="qcow2">QCOW2 (recommended)</option>
                        <option value="raw">Raw</option>
                      </select>
                    )}
                  />
                </div>
                
                <div className="space-y-2">
                  <label htmlFor="disk-pool" className="block text-sm font-medium">
                    Storage Pool *
                  </label>
                  <Controller
                    name="disk.storagePool"
                    control={control}
                    render={({ field }) => (
                      <select
                        {...field}
                        id="disk-pool"
                        className="w-full p-2 border rounded-md bg-background"
                      >
                        <option value="default">default</option>
                      </select>
                    )}
                  />
                </div>
              </div>
              
              <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                <div className="space-y-2">
                  <label htmlFor="disk-bus" className="block text-sm font-medium">
                    Disk Bus *
                  </label>
                  <Controller
                    name="disk.bus"
                    control={control}
                    render={({ field }) => (
                      <select
                        {...field}
                        id="disk-bus"
                        className="w-full p-2 border rounded-md bg-background"
                      >
                        <option value="virtio">VirtIO (recommended)</option>
                        <option value="sata">SATA</option>
                        <option value="scsi">SCSI</option>
                        <option value="ide">IDE</option>
                      </select>
                    )}
                  />
                </div>
                
                <div className="space-y-2">
                  <label htmlFor="disk-cache" className="block text-sm font-medium">
                    Cache Mode *
                  </label>
                  <Controller
                    name="disk.cacheMode"
                    control={control}
                    render={({ field }) => (
                      <select
                        {...field}
                        id="disk-cache"
                        className="w-full p-2 border rounded-md bg-background"
                      >
                        <option value="none">None</option>
                        <option value="writeback">Writeback</option>
                        <option value="writethrough">Writethrough</option>
                        <option value="directsync">Direct Sync</option>
                        <option value="unsafe">Unsafe</option>
                      </select>
                    )}
                  />
                </div>
              </div>
              
              <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                <div className="space-y-2">
                  <label htmlFor="disk-source" className="block text-sm font-medium">
                    Source Image (Optional)
                  </label>
                  <Controller
                    name="disk.sourceImage"
                    control={control}
                    render={({ field }) => (
                      <input
                        {...field}
                        id="disk-source"
                        type="text"
                        placeholder="Path to source image"
                        className="w-full p-2 border rounded-md bg-background"
                      />
                    )}
                  />
                  <p className="text-xs text-muted-foreground">
                    Path to existing image to clone (leave empty for empty disk)
                  </p>
                </div>
              </div>
              
              <div className="flex items-center space-x-4 mt-4">
                <div className="flex items-center">
                  <Controller
                    name="disk.readOnly"
                    control={control}
                    render={({ field }) => (
                      <input
                        type="checkbox"
                        id="disk-readonly"
                        checked={field.value}
                        onChange={(e) => field.onChange(e.target.checked)}
                        className="mr-2"
                      />
                    )}
                  />
                  <label htmlFor="disk-readonly" className="text-sm">
                    Read Only
                  </label>
                </div>
                
                <div className="flex items-center">
                  <Controller
                    name="disk.shareable"
                    control={control}
                    render={({ field }) => (
                      <input
                        type="checkbox"
                        id="disk-shareable"
                        checked={field.value}
                        onChange={(e) => field.onChange(e.target.checked)}
                        className="mr-2"
                      />
                    )}
                  />
                  <label htmlFor="disk-shareable" className="text-sm">
                    Shareable
                  </label>
                </div>
              </div>
            </CardContent>
          </Card>
        )}
        
        {/* Step 4: Network */}
        {step === 4 && (
          <Card>
            <CardHeader>
              <CardTitle className="flex items-center text-xl">
                <LuNetwork className="mr-2 h-5 w-5" />
                Network
              </CardTitle>
            </CardHeader>
            <CardContent className="space-y-4">
              <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                <div className="space-y-2">
                  <label htmlFor="network-type" className="block text-sm font-medium">
                    Network Type *
                  </label>
                  <Controller
                    name="network.type"
                    control={control}
                    render={({ field }) => (
                      <select
                        {...field}
                        id="network-type"
                        className="w-full p-2 border rounded-md bg-background"
                      >
                        <option value="network">Virtual Network</option>
                        <option value="bridge">Bridge</option>
                        <option value="direct">Direct (Macvtap)</option>
                      </select>
                    )}
                  />
                </div>
                
                <div className="space-y-2">
                  <label htmlFor="network-source" className="block text-sm font-medium">
                    Network Source *
                  </label>
                  <Controller
                    name="network.source"
                    control={control}
                    render={({ field }) => {
                      if (networkType === 'network') {
                        return (
                          <select
                            {...field}
                            id="network-source"
                            className="w-full p-2 border rounded-md bg-background"
                            disabled={bridgeNetworksLoading}
                          >
                            <option value="">Select a network...</option>
                            <option value="default">default (NAT)</option>
                            {bridgeNetworks.map((network) => (
                              <option key={network.name} value={network.name}>
                                {network.name} (Bridge: {network.bridge_name})
                              </option>
                            ))}
                          </select>
                        );
                      }
                      
                      return (
                        <input
                          {...field}
                          id="network-source"
                          type="text"
                          className="w-full p-2 border rounded-md bg-background"
                          placeholder={
                            networkType === 'bridge' 
                              ? 'Bridge name (e.g., br0)' 
                              : 'Interface name'
                          }
                        />
                      );
                    }}
                  />
                  <p className="text-xs text-muted-foreground">
                    {networkType === 'network' 
                      ? 'Select a virtual network or bridge network' 
                      : networkType === 'bridge'
                      ? 'Name of the Linux bridge to connect to'
                      : 'Network device name depending on type'
                    }
                    {bridgeNetworksLoading && networkType === 'network' && ' (Loading networks...)'}
                  </p>
                </div>
              </div>
              
              <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                <div className="space-y-2">
                  <label htmlFor="network-model" className="block text-sm font-medium">
                    Network Model *
                  </label>
                  <Controller
                    name="network.model"
                    control={control}
                    render={({ field }) => (
                      <select
                        {...field}
                        id="network-model"
                        className="w-full p-2 border rounded-md bg-background"
                      >
                        <option value="virtio">VirtIO (recommended)</option>
                        <option value="e1000">Intel e1000</option>
                        <option value="rtl8139">Realtek RTL8139</option>
                      </select>
                    )}
                  />
                </div>
                
                <div className="space-y-2">
                  <label htmlFor="network-mac" className="block text-sm font-medium">
                    MAC Address (Optional)
                  </label>
                  <Controller
                    name="network.macAddress"
                    control={control}
                    render={({ field }) => (
                      <input
                        {...field}
                        id="network-mac"
                        type="text"
                        className="w-full p-2 border rounded-md bg-background"
                        placeholder="Autogenerated if empty"
                      />
                    )}
                  />
                  <p className="text-xs text-muted-foreground">
                    Leave empty to auto-generate
                  </p>
                </div>
              </div>
              
              <div className="space-y-2">
                <label htmlFor="ssh-keys" className="block text-sm font-medium">
                  SSH Keys (Optional, one per line)
                </label>
                <Controller
                  name="cloudInit.sshKeys"
                  control={control}
                  render={({ field }) => (
                    <textarea
                      id="ssh-keys"
                      rows={3}
                      value={field.value?.join('\n') || ''}
                      onChange={(e) => field.onChange(
                        e.target.value.split('\n').filter(line => line.trim() !== '')
                      )}
                      className="w-full p-2 border rounded-md bg-background"
                      placeholder="ssh-rsa AAAA..."
                    />
                  )}
                />
                <p className="text-xs text-muted-foreground">
                  For cloud-init enabled images only
                </p>
              </div>
            </CardContent>
          </Card>
        )}
        
        {/* Navigation Buttons */}
        <div className="flex justify-between mt-6">
          {step > 1 ? (
            <Button type="button" variant="outline" onClick={prevStep}>
              Previous
            </Button>
          ) : (
            <Button type="button" variant="outline" onClick={() => navigate({ to: '/vms' })}>
              Cancel
            </Button>
          )}
          
          {step < 4 ? (
            <Button type="button" onClick={nextStep}>
              Next
            </Button>
          ) : (
            <Button
              type="submit"
              disabled={createMutation.isPending || !isValid}
              className="flex items-center"
            >
              {createMutation.isPending ? (
                'Creating...'
              ) : (
                <>
                  <LuCheck className="mr-2 h-4 w-4" />
                  Create VM
                </>
              )}
            </Button>
          )}
        </div>
        
        {createMutation.isError && (
          <div className="mt-4 p-3 bg-red-500/10 border border-red-500 rounded-md text-red-500">
            {createMutation.error instanceof Error 
              ? createMutation.error.message 
              : 'Failed to create VM. Please try again.'}
          </div>
        )}
      </form>
    </div>
  );
};