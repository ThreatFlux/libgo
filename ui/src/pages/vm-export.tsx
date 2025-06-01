import React from 'react';
import { useForm, Controller } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import { z } from 'zod';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { useNavigate, useParams } from '@tanstack/react-router';
import { getVM } from '@/api/vm';
import { exportVM } from '@/api/export';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardHeader, CardTitle } from '@/lib/components';
import { VMStatusBadge } from '@/components/vm/vm-status-badge';

// Icons
import {
  LuArrowLeft,
  LuCheck,
  LuServer,
  LuFileDown
} from 'react-icons/lu';

// Form schema
const exportVMSchema = z.object({
  format: z.enum(['qcow2', 'vmdk', 'vdi', 'ova', 'raw']),
  fileName: z.string().min(1, 'Filename is required'),
  options: z.record(z.string()).optional()
});

type ExportVMFormValues = z.infer<typeof exportVMSchema>;

export const VMExportPage: React.FC = () => {
  const { name } = useParams({ from: '/vms/$name/export' });
  const navigate = useNavigate();
  const queryClient = useQueryClient();
  
  // Fetch VM details
  const { data: vm, isLoading: vmLoading } = useQuery({
    queryKey: ['vm', name],
    queryFn: () => getVM(name),
  });

  // Form setup
  const {
    control,
    handleSubmit,
    watch,
    formState: { errors }
  } = useForm<ExportVMFormValues>({
    resolver: zodResolver(exportVMSchema),
    defaultValues: {
      format: 'qcow2',
      fileName: name ? `${name}.qcow2` : '',
    }
  });

  // Watch format to update filename extension
  const watchFormat = watch('format');

  // Export VM mutation
  const exportMutation = useMutation({
    mutationFn: ({vmName, data}: {vmName: string, data: ExportVMFormValues}) => 
      exportVM(vmName, data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['exports'] });
      navigate({ to: '/exports' });
    }
  });

  // Form submission
  const onSubmit = (data: ExportVMFormValues) => {
    if (vm) {
      exportMutation.mutate({vmName: vm.name, data});
    }
  };

  // Format descriptions
  const formatDescriptions: Record<string, { description: string, sizeRatio: string }> = {
    'qcow2': { 
      description: 'QEMU Copy-On-Write v2 format, efficient sparse file format for QEMU/KVM.',
      sizeRatio: '1x (efficient)'
    },
    'vmdk': { 
      description: 'VMware Virtual Machine Disk format. Good for VMware compatibility.',
      sizeRatio: '1-1.5x' 
    },
    'vdi': { 
      description: 'VirtualBox Disk Image format. Best for VirtualBox.',
      sizeRatio: '1-1.5x' 
    },
    'ova': { 
      description: 'Open Virtualization Archive, includes VM configuration and disk in OVF format.',
      sizeRatio: '1.5-2x' 
    },
    'raw': { 
      description: 'Raw disk image format. Maximum compatibility but no compression.',
      sizeRatio: '2-3x (full disk size)' 
    },
  };

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center space-x-4">
        <Button 
          variant="outline" 
          size="icon" 
          onClick={() => navigate({ to: '/vms/$name', params: { name } })}
        >
          <LuArrowLeft className="h-4 w-4" />
        </Button>
        <div>
          <h1 className="text-2xl font-bold">Export Virtual Machine</h1>
          {vm && (
            <div className="flex items-center mt-1">
              <VMStatusBadge status={vm.status} />
              <span className="ml-2 text-sm">{vm.name}</span>
            </div>
          )}
        </div>
      </div>

      {vmLoading ? (
        <div className="text-center py-8">Loading VM details...</div>
      ) : !vm ? (
        <div className="text-center py-8 text-red-500">VM not found</div>
      ) : (
        <div className="max-w-2xl mx-auto">
          <Card>
            <CardHeader>
              <CardTitle className="flex items-center">
                <LuFileDown className="mr-2 h-5 w-5" />
                Export {vm.name}
              </CardTitle>
            </CardHeader>
            <CardContent>
              <form onSubmit={handleSubmit(onSubmit)} className="space-y-6">
                {/* VM Status Warning */}
                {vm.status === 'running' && (
                  <div className="bg-yellow-500/10 border border-yellow-500 text-yellow-500 p-4 rounded">
                    <p className="font-medium">Warning: VM is currently running</p>
                    <p className="text-sm">
                      Exporting a running VM may result in an inconsistent disk state.
                      It's recommended to stop the VM before exporting.
                    </p>
                  </div>
                )}
                
                {/* Export Format */}
                <div className="space-y-2">
                  <label htmlFor="format" className="block text-sm font-medium">
                    Export Format *
                  </label>
                  <Controller
                    name="format"
                    control={control}
                    render={({ field }) => (
                      <select
                        {...field}
                        id="format"
                        className="w-full p-2 border rounded-md bg-background"
                      >
                        <option value="qcow2">QCOW2 (QEMU/KVM)</option>
                        <option value="vmdk">VMDK (VMware)</option>
                        <option value="vdi">VDI (VirtualBox)</option>
                        <option value="ova">OVA (Open Virtualization Format)</option>
                        <option value="raw">Raw</option>
                      </select>
                    )}
                  />
                  
                  {watchFormat && formatDescriptions[watchFormat] && (
                    <p className="text-sm text-muted-foreground mt-1">
                      {formatDescriptions[watchFormat].description}<br />
                      <span className="text-xs">
                        Approximate size ratio: {formatDescriptions[watchFormat].sizeRatio}
                      </span>
                    </p>
                  )}
                </div>
                
                {/* File Name */}
                <div className="space-y-2">
                  <label htmlFor="fileName" className="block text-sm font-medium">
                    Output Filename *
                  </label>
                  <Controller
                    name="fileName"
                    control={control}
                    render={({ field }) => (
                      <input
                        {...field}
                        id="fileName"
                        type="text"
                        className="w-full p-2 border rounded-md bg-background"
                        placeholder={`${vm.name}.${watchFormat}`}
                      />
                    )}
                  />
                  {errors.fileName && (
                    <p className="text-sm text-red-500">{errors.fileName.message}</p>
                  )}
                </div>
                
                {/* VM Information */}
                <div className="bg-muted/40 p-4 rounded border">
                  <h3 className="text-sm font-medium mb-2 flex items-center">
                    <LuServer className="mr-2 h-4 w-4" />
                    VM Details
                  </h3>
                  <div className="grid grid-cols-2 gap-x-4 gap-y-2 text-sm">
                    <div className="text-muted-foreground">Disk Count:</div>
                    <div>{vm.disks.length}</div>
                    
                    <div className="text-muted-foreground">Total Disk Size:</div>
                    <div>
                      {vm.disks.reduce((acc, disk) => acc + disk.sizeBytes, 0).toLocaleString()} bytes
                    </div>
                    
                    <div className="text-muted-foreground">Status:</div>
                    <div>{vm.status}</div>
                  </div>
                </div>
                
                <div className="flex justify-between pt-4">
                  <Button
                    type="button"
                    variant="outline"
                    onClick={() => navigate({ to: '/vms/$name', params: { name: vm.name } })}
                  >
                    Cancel
                  </Button>
                  
                  <Button
                    type="submit"
                    disabled={exportMutation.isPending}
                    className="flex items-center"
                  >
                    {exportMutation.isPending ? (
                      'Starting Export...'
                    ) : (
                      <>
                        <LuCheck className="mr-2 h-4 w-4" />
                        Export VM
                      </>
                    )}
                  </Button>
                </div>
                
                {exportMutation.isError && (
                  <div className="mt-4 p-3 bg-red-500/10 border border-red-500 rounded-md text-red-500">
                    {exportMutation.error instanceof Error 
                      ? exportMutation.error.message 
                      : 'Failed to start export. Please try again.'}
                  </div>
                )}
              </form>
            </CardContent>
          </Card>
        </div>
      )}
    </div>
  );
};