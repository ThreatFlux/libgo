import React, { useState } from 'react';
import { useQuery } from '@tanstack/react-query';
import { getVMs } from '@/api/vm';
import { Link } from '@tanstack/react-router';
import { Button } from '@/components/ui/button';
import { VMStatusBadge } from '@/components/vm/vm-status-badge';
import { formatBytes, formatDate } from '@/lib/utils';
import { VMStatus } from '@/types/api';

// Icons
import { 
  LuPlus, 
  LuSearch, 
  LuRefreshCw, 
  LuChevronLeft, 
  LuChevronRight,
  LuFilter
} from 'react-icons/lu';

export const VMListPage: React.FC = () => {
  const [page, setPage] = useState(1);
  const [pageSize] = useState(10);
  const [nameFilter, setNameFilter] = useState('');
  const [statusFilter, setStatusFilter] = useState<VMStatus | ''>('');
  
  // Fetch VMs with filters
  const { data, isLoading, isError, refetch } = useQuery({
    queryKey: ['vms', page, pageSize, nameFilter, statusFilter],
    queryFn: () => getVMs(page, pageSize, nameFilter || undefined, statusFilter || undefined),
  });

  const handleSearch = (e: React.FormEvent) => {
    e.preventDefault();
    refetch();
  };

  const handleRefresh = () => {
    refetch();
  };

  const handlePrevPage = () => {
    if (page > 1) {
      setPage(page - 1);
    }
  };

  const handleNextPage = () => {
    if (data && page < Math.ceil(data.count / pageSize)) {
      setPage(page + 1);
    }
  };

  return (
    <div className="space-y-6">
      <div className="flex justify-between items-center">
        <h1 className="text-2xl font-bold">Virtual Machines</h1>
        <Link to="/vms/create">
          <Button>
            <LuPlus className="mr-2 h-4 w-4" />
            Create VM
          </Button>
        </Link>
      </div>

      {/* Filters */}
      <div className="bg-card border rounded-lg p-4">
        <form onSubmit={handleSearch} className="flex flex-col md:flex-row gap-4">
          <div className="flex-grow">
            <div className="relative">
              <LuSearch className="absolute left-3 top-1/2 transform -translate-y-1/2 text-muted-foreground" />
              <input
                type="text"
                placeholder="Search by name..."
                value={nameFilter}
                onChange={(e) => setNameFilter(e.target.value)}
                className="w-full pl-10 pr-4 py-2 border rounded-md bg-background"
              />
            </div>
          </div>
          
          <div className="w-full md:w-64">
            <div className="relative">
              <LuFilter className="absolute left-3 top-1/2 transform -translate-y-1/2 text-muted-foreground" />
              <select
                value={statusFilter}
                onChange={(e) => setStatusFilter(e.target.value as VMStatus | '')}
                className="w-full pl-10 pr-4 py-2 border rounded-md appearance-none bg-background"
              >
                <option value="">All Status</option>
                <option value="running">Running</option>
                <option value="stopped">Stopped</option>
                <option value="paused">Paused</option>
                <option value="shutdown">Shutdown</option>
                <option value="crashed">Crashed</option>
              </select>
            </div>
          </div>
          
          <div className="flex gap-2">
            <Button type="submit">
              Search
            </Button>
            <Button type="button" variant="outline" onClick={handleRefresh}>
              <LuRefreshCw className="h-4 w-4" />
            </Button>
          </div>
        </form>
      </div>

      {/* VM List */}
      <div className="bg-card border rounded-lg overflow-hidden">
        <div className="overflow-x-auto">
          <table className="w-full">
            <thead>
              <tr className="border-b">
                <th className="px-4 py-3 text-left text-sm font-medium text-muted-foreground">Name</th>
                <th className="px-4 py-3 text-left text-sm font-medium text-muted-foreground">Status</th>
                <th className="px-4 py-3 text-left text-sm font-medium text-muted-foreground">CPU</th>
                <th className="px-4 py-3 text-left text-sm font-medium text-muted-foreground">Memory</th>
                <th className="px-4 py-3 text-left text-sm font-medium text-muted-foreground">Disks</th>
                <th className="px-4 py-3 text-left text-sm font-medium text-muted-foreground">Created</th>
                <th className="px-4 py-3 text-left text-sm font-medium text-muted-foreground">Actions</th>
              </tr>
            </thead>
            <tbody>
              {isLoading ? (
                <tr>
                  <td colSpan={7} className="px-4 py-3 text-center">
                    Loading...
                  </td>
                </tr>
              ) : isError ? (
                <tr>
                  <td colSpan={7} className="px-4 py-3 text-center text-red-500">
                    Error loading VMs
                  </td>
                </tr>
              ) : data?.vms.length === 0 ? (
                <tr>
                  <td colSpan={7} className="px-4 py-3 text-center text-muted-foreground">
                    No virtual machines found
                  </td>
                </tr>
              ) : (
                data?.vms.map((vm) => (
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
                      <VMStatusBadge status={vm.status} />
                    </td>
                    <td className="px-4 py-3 text-sm">
                      {vm.cpu.count} vCPU
                    </td>
                    <td className="px-4 py-3 text-sm">
                      {formatBytes(vm.memory.sizeBytes)}
                    </td>
                    <td className="px-4 py-3 text-sm">
                      {vm.disks.length} ({formatBytes(vm.disks.reduce((sum, disk) => sum + disk.sizeBytes, 0))})
                    </td>
                    <td className="px-4 py-3 text-sm">
                      {formatDate(vm.createdAt)}
                    </td>
                    <td className="px-4 py-3 text-sm">
                      <div className="flex space-x-2">
                        <Link 
                          to="/vms/$name" 
                          params={{ name: vm.name }}
                          className="text-blue-500 hover:text-blue-700"
                        >
                          View
                        </Link>
                      </div>
                    </td>
                  </tr>
                ))
              )}
            </tbody>
          </table>
        </div>
        
        {/* Pagination */}
        {data && data.count > 0 && (
          <div className="px-4 py-3 border-t flex items-center justify-between">
            <div className="text-sm text-muted-foreground">
              Showing {Math.min((page - 1) * pageSize + 1, data.count)} to {Math.min(page * pageSize, data.count)} of {data.count} results
            </div>
            <div className="flex items-center space-x-2">
              <Button
                variant="outline"
                size="sm"
                onClick={handlePrevPage}
                disabled={page <= 1}
              >
                <LuChevronLeft className="h-4 w-4" />
              </Button>
              <span className="text-sm">{page}</span>
              <Button
                variant="outline"
                size="sm"
                onClick={handleNextPage}
                disabled={page >= Math.ceil(data.count / pageSize)}
              >
                <LuChevronRight className="h-4 w-4" />
              </Button>
            </div>
          </div>
        )}
      </div>
    </div>
  );
};