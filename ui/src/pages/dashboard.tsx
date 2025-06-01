import React from 'react';
import { useQuery } from '@tanstack/react-query';
import { getVMs } from '@/api/vm';
import { getSystemMetrics, getStoragePoolInfo } from '@/api/metrics';
import { StatusCard } from '@/components/dashboard/status-card';
import { StorageUsage } from '@/components/dashboard/storage-usage';
import { VMStatusChart } from '@/components/dashboard/vm-status-chart';
import { formatBytes } from '@/lib/utils';
import { Link } from '@tanstack/react-router';
import { Button } from '@/components/ui/button';
import { VM } from '@/types/api';

// Icons
import { 
  LuServer, 
  LuCpu, 
  LuDownload,
  LuPlus,
  LuActivity
} from 'react-icons/lu';

export const DashboardPage: React.FC = () => {
  // Fetch VMs
  const { data: vmListData } = useQuery({
    queryKey: ['vms'],
    queryFn: () => getVMs(),
  });

  // Fetch system metrics
  const { data: systemMetrics } = useQuery({
    queryKey: ['system-metrics'],
    queryFn: () => getSystemMetrics(),
  });

  // Fetch storage pool info
  const { data: storagePools } = useQuery({
    queryKey: ['storage-pools'],
    queryFn: () => getStoragePoolInfo(),
  });

  // Count VMs by status
  const statusCounts = React.useMemo(() => {
    const vms = vmListData?.vms || [];
    const counts = {
      running: 0,
      stopped: 0,
      paused: 0,
      other: 0
    };

    vms.forEach((vm: VM) => {
      if (vm.status === 'running') {
        counts.running++;
      } else if (vm.status === 'stopped') {
        counts.stopped++;
      } else if (vm.status === 'paused') {
        counts.paused++;
      } else {
        counts.other++;
      }
    });

    return counts;
  }, [vmListData]);

  // Get most recent VMs (last 5)
  const recentVMs = React.useMemo(() => {
    const vms = vmListData?.vms || [];
    
    // Sort by creation date (newest first)
    return [...vms]
      .sort((a, b) => new Date(b.createdAt).getTime() - new Date(a.createdAt).getTime())
      .slice(0, 5);
  }, [vmListData]);

  return (
    <div className="space-y-6">
      <div className="flex justify-between items-center">
        <h1 className="text-2xl font-bold">Dashboard</h1>
        <Link to="/vms/create">
          <Button>
            <LuPlus className="mr-2 h-4 w-4" />
            Create VM
          </Button>
        </Link>
      </div>

      {/* Status Cards */}
      <div className="grid grid-cols-1 sm:grid-cols-2 md:grid-cols-4 gap-4">
        <StatusCard
          title="Total VMs"
          value={(vmListData?.count || 0).toString()}
          icon={<LuServer className="h-4 w-4" />}
          variant="default"
        />
        
        <StatusCard
          title="Running VMs"
          value={statusCounts.running.toString()}
          icon={<LuActivity className="h-4 w-4" />}
          variant="success"
        />
        
        <StatusCard
          title="CPU Utilization"
          value={`${systemMetrics?.cpuUtilization?.toFixed(1) || '0'}%`}
          icon={<LuCpu className="h-4 w-4" />}
          variant={
            systemMetrics?.cpuUtilization && systemMetrics.cpuUtilization > 80 
              ? 'danger' 
              : systemMetrics?.cpuUtilization && systemMetrics.cpuUtilization > 60 
                ? 'warning' 
                : 'default'
          }
        />
        
        <StatusCard
          title="Active Exports"
          value={(systemMetrics?.activeExportJobs || 0).toString()}
          icon={<LuDownload className="h-4 w-4" />}
          variant="default"
        />
      </div>

      {/* Charts Section */}
      <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
        {/* VM Status Chart */}
        <VMStatusChart
          running={statusCounts.running}
          stopped={statusCounts.stopped}
          paused={statusCounts.paused}
          other={statusCounts.other}
        />
        
        {/* Storage Usage */}
        {storagePools && <StorageUsage pools={storagePools} />}
      </div>

      {/* Recent VMs */}
      <div>
        <h2 className="text-xl font-semibold mb-4">Recent VMs</h2>
        <div className="bg-card border rounded-lg overflow-hidden">
          <div className="overflow-x-auto">
            <table className="w-full">
              <thead>
                <tr className="border-b">
                  <th className="px-4 py-3 text-left text-sm font-medium text-muted-foreground">Name</th>
                  <th className="px-4 py-3 text-left text-sm font-medium text-muted-foreground">Status</th>
                  <th className="px-4 py-3 text-left text-sm font-medium text-muted-foreground">CPU</th>
                  <th className="px-4 py-3 text-left text-sm font-medium text-muted-foreground">Memory</th>
                  <th className="px-4 py-3 text-left text-sm font-medium text-muted-foreground">Created</th>
                </tr>
              </thead>
              <tbody>
                {recentVMs.length > 0 ? (
                  recentVMs.map((vm) => (
                    <tr key={vm.uuid} className="border-b hover:bg-muted/50">
                      <td className="px-4 py-3 text-sm">
                        <Link 
                          to="/vms/$name" 
                          params={{ name: vm.name }}
                          className="font-medium text-primary hover:underline"
                        >
                          {vm.name}
                        </Link>
                      </td>
                      <td className="px-4 py-3 text-sm">
                        <span 
                          className={`inline-block rounded-full w-2 h-2 mr-2 ${
                            vm.status === 'running' ? 'bg-green-500' : 
                            vm.status === 'stopped' ? 'bg-red-500' : 
                            vm.status === 'paused' ? 'bg-yellow-500' : 
                            'bg-gray-500'
                          }`}
                        />
                        {vm.status}
                      </td>
                      <td className="px-4 py-3 text-sm">{vm.cpu.count} vCPU</td>
                      <td className="px-4 py-3 text-sm">{formatBytes(vm.memory.sizeBytes)}</td>
                      <td className="px-4 py-3 text-sm">{new Date(vm.createdAt).toLocaleDateString()}</td>
                    </tr>
                  ))
                ) : (
                  <tr>
                    <td colSpan={5} className="px-4 py-3 text-sm text-center text-muted-foreground">
                      No VMs found
                    </td>
                  </tr>
                )}
              </tbody>
            </table>
          </div>
          {(vmListData?.count ?? 0) > 5 && (
            <div className="p-4 border-t">
              <Link to="/vms">
                <Button variant="outline" size="sm">
                  View all VMs
                </Button>
              </Link>
            </div>
          )}
        </div>
      </div>
    </div>
  );
};