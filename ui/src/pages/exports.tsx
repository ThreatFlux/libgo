import React, { useState } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { getExports, cancelExport } from '@/api/export';
import { formatDate } from '@/lib/utils';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardHeader, CardTitle } from '@/lib/components';

// Icons
import { 
  LuFileDown, 
  LuRefreshCw, 
  LuCheck as LuCheckCircle, 
  LuTriangle as LuAlertCircle,
  LuCircle as LuXCircle,
  LuClock,
  LuChevronLeft,
  LuChevronRight,
  LuDownload
} from 'react-icons/lu';

export const ExportsPage: React.FC = () => {
  const [page, setPage] = useState(1);
  const [pageSize] = useState(10);
  const queryClient = useQueryClient();
  
  // Fetch exports
  const { data, isLoading, isError, refetch } = useQuery({
    queryKey: ['exports', page, pageSize],
    queryFn: () => getExports(page, pageSize),
    // Auto-refresh active exports
    refetchInterval: (queryData) => {
      const exportData = queryData as unknown as { jobs: Array<{ status: string }> };
      const hasActiveExports = exportData?.jobs?.some(
        job => job.status === 'pending' || job.status === 'running'
      );
      return hasActiveExports ? 5000 : false;
    }
  });

  // Cancel export mutation
  const cancelMutation = useMutation({
    mutationFn: cancelExport,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['exports'] });
    }
  });

  const handleRefresh = () => {
    refetch();
  };

  const handleCancel = (id: string) => {
    cancelMutation.mutate(id);
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

  // Get status badge styling
  const getStatusBadge = (status: string) => {
    switch (status) {
      case 'pending':
        return (
          <span className="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium bg-blue-100 text-blue-800 dark:bg-blue-900 dark:text-blue-200">
            <LuClock className="mr-1 h-3 w-3" />
            Pending
          </span>
        );
      case 'running':
        return (
          <span className="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium bg-yellow-100 text-yellow-800 dark:bg-yellow-900 dark:text-yellow-200">
            <svg className="animate-spin -ml-1 mr-2 h-3 w-3 text-yellow-800 dark:text-yellow-200" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24">
              <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4"></circle>
              <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
            </svg>
            Running
          </span>
        );
      case 'completed':
        return (
          <span className="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium bg-green-100 text-green-800 dark:bg-green-900 dark:text-green-200">
            <LuCheckCircle className="mr-1 h-3 w-3" />
            Completed
          </span>
        );
      case 'failed':
        return (
          <span className="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium bg-red-100 text-red-800 dark:bg-red-900 dark:text-red-200">
            <LuAlertCircle className="mr-1 h-3 w-3" />
            Failed
          </span>
        );
      case 'cancelled':
        return (
          <span className="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium bg-gray-100 text-gray-800 dark:bg-gray-700 dark:text-gray-200">
            <LuXCircle className="mr-1 h-3 w-3" />
            Cancelled
          </span>
        );
      default:
        return (
          <span className="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium bg-gray-100 text-gray-800 dark:bg-gray-700 dark:text-gray-200">
            {status}
          </span>
        );
    }
  };

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex justify-between items-center">
        <h1 className="text-2xl font-bold flex items-center">
          <LuFileDown className="mr-2 h-6 w-6" />
          VM Exports
        </h1>
        <Button variant="outline" onClick={handleRefresh}>
          <LuRefreshCw className="mr-2 h-4 w-4" />
          Refresh
        </Button>
      </div>

      {/* Export Jobs List */}
      <Card>
        <CardHeader className="pb-2">
          <CardTitle className="text-lg">Export Jobs</CardTitle>
        </CardHeader>
        <CardContent>
          {isLoading ? (
            <div className="flex justify-center items-center h-64">
              <div className="text-center">
                <svg className="animate-spin h-6 w-6 mx-auto text-primary" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24">
                  <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4"></circle>
                  <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
                </svg>
                <p className="mt-2">Loading export jobs...</p>
              </div>
            </div>
          ) : isError ? (
            <div className="flex justify-center items-center h-64">
              <div className="text-center text-red-500">
                <LuAlertCircle className="h-6 w-6 mx-auto" />
                <p className="mt-2">Error loading export jobs</p>
              </div>
            </div>
          ) : !data || data.jobs.length === 0 ? (
            <div className="flex flex-col justify-center items-center h-64 text-center text-muted-foreground">
              <LuFileDown className="h-12 w-12 mb-4 opacity-20" />
              <p>No export jobs found</p>
              <p className="text-sm mt-2">
                Export a VM from the VM details page to see export jobs here
              </p>
            </div>
          ) : (
            <div className="overflow-x-auto">
              <table className="w-full">
                <thead>
                  <tr className="border-b">
                    <th className="px-4 py-3 text-left text-sm font-medium text-muted-foreground">ID</th>
                    <th className="px-4 py-3 text-left text-sm font-medium text-muted-foreground">VM Name</th>
                    <th className="px-4 py-3 text-left text-sm font-medium text-muted-foreground">Format</th>
                    <th className="px-4 py-3 text-left text-sm font-medium text-muted-foreground">Status</th>
                    <th className="px-4 py-3 text-left text-sm font-medium text-muted-foreground">Progress</th>
                    <th className="px-4 py-3 text-left text-sm font-medium text-muted-foreground">Start Time</th>
                    <th className="px-4 py-3 text-left text-sm font-medium text-muted-foreground">Actions</th>
                  </tr>
                </thead>
                <tbody>
                  {data.jobs.map((job) => (
                    <tr key={job.id} className="border-b hover:bg-muted/50">
                      <td className="px-4 py-3 text-sm font-mono">{job.id.substring(0, 8)}...</td>
                      <td className="px-4 py-3 text-sm font-medium">{job.vmName}</td>
                      <td className="px-4 py-3 text-sm uppercase">{job.format}</td>
                      <td className="px-4 py-3 text-sm">{getStatusBadge(job.status)}</td>
                      <td className="px-4 py-3 text-sm">
                        {(job.status === 'pending' || job.status === 'running') ? (
                          <div className="w-full bg-muted rounded-full h-2.5">
                            <div 
                              className="bg-primary h-2.5 rounded-full" 
                              style={{ width: `${job.progress}%` }}
                            />
                          </div>
                        ) : job.status === 'completed' ? (
                          '100%'
                        ) : (
                          job.progress ? `${job.progress}%` : '-'
                        )}
                      </td>
                      <td className="px-4 py-3 text-sm">{formatDate(job.startTime)}</td>
                      <td className="px-4 py-3 text-sm">
                        {(job.status === 'pending' || job.status === 'running') && (
                          <Button
                            variant="outline"
                            size="sm"
                            onClick={() => handleCancel(job.id)}
                            disabled={cancelMutation.isPending}
                            className="text-red-500 border-red-500 hover:bg-red-500/10"
                          >
                            Cancel
                          </Button>
                        )}
                        {job.status === 'completed' && job.outputPath && (
                          <Button variant="outline" size="sm" className="text-green-500">
                            <LuDownload className="mr-1 h-3 w-3" />
                            Download
                          </Button>
                        )}
                        {job.status === 'failed' && job.error && (
                          <div className="text-red-500 text-xs max-w-xs truncate" title={job.error}>
                            {job.error}
                          </div>
                        )}
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          )}

          {/* Pagination */}
          {data && data.count > 0 && (
            <div className="px-4 py-3 border-t flex items-center justify-between mt-4">
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
        </CardContent>
      </Card>
    </div>
  );
};