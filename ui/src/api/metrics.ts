import { get } from './client';
import { HealthStatus, SystemMetrics, StoragePoolInfo } from '@/types/api';

// Get health status
export const getHealthStatus = async (): Promise<HealthStatus> => {
  return get<HealthStatus>('/health');
};

// Get metrics for dashboard
export const getSystemMetrics = async (): Promise<SystemMetrics> => {
  // This is a mock implementation as the actual API endpoints for system metrics 
  // may need to be implemented on the backend
  
  // In a real implementation, you would call the actual API endpoint:
  // return get<SystemMetrics>('/api/v1/metrics/system');
  
  // For now, getting data from Prometheus metrics endpoint
  const prometheusMetrics = await get<string>('/metrics');
  
  // Parse Prometheus metrics (simplified)
  const vmCountMatch = prometheusMetrics.match(/libgo_vm_count (\d+)/);
  const activeExportJobsMatch = prometheusMetrics.match(/libgo_export_jobs_active (\d+)/);
  
  // Default dummy data
  return {
    vmCount: vmCountMatch ? parseInt(vmCountMatch[1]) : 0,
    cpuUtilization: Math.random() * 100,
    memoryUtilization: Math.random() * 100,
    storageUtilization: Math.random() * 100,
    runningVMs: Math.floor(Math.random() * 10),
    exportJobs: Math.floor(Math.random() * 5),
    activeExportJobs: activeExportJobsMatch ? parseInt(activeExportJobsMatch[1]) : 0
  };
};

// Get storage pool information
export const getStoragePoolInfo = async (): Promise<StoragePoolInfo[]> => {
  // This might need to be implemented on the backend
  // For now returning mock data
  return [
    {
      name: 'default',
      capacity: 1024 * 1024 * 1024 * 500, // 500 GB
      available: 1024 * 1024 * 1024 * 300, // 300 GB
      allocation: 1024 * 1024 * 1024 * 200, // 200 GB
    },
    {
      name: 'ssd-pool',
      capacity: 1024 * 1024 * 1024 * 1000, // 1 TB
      available: 1024 * 1024 * 1024 * 800,  // 800 GB
      allocation: 1024 * 1024 * 1024 * 200,  // 200 GB
    }
  ];
};