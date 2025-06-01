import { get, post, del } from './client';
import { 
  ExportJob, 
  ExportListResponse, 
  ExportRequest, 
  ExportResponse 
} from '@/types/api';

// Export a VM
export const exportVM = async (
  vmName: string, 
  exportRequest: ExportRequest
): Promise<ExportResponse> => {
  return post<ExportResponse>(`/api/v1/vms/${vmName}/export`, exportRequest);
};

// Get list of export jobs
export const getExports = async (
  page: number = 1, 
  pageSize: number = 50
): Promise<ExportListResponse> => {
  const params = { page, pageSize };
  return get<ExportListResponse>('/api/v1/exports', { params });
};

// Get export job status
export const getExportStatus = async (id: string): Promise<ExportJob> => {
  return get<ExportJob>(`/api/v1/exports/${id}`);
};

// Cancel an export job
export const cancelExport = async (id: string): Promise<void> => {
  return del(`/api/v1/exports/${id}`);
};