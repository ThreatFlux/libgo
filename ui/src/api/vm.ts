import { get, post, put, del } from './client';
import { 
  VM, 
  VMListResponse, 
  VMCreateParams, 
  VMCreateResponse 
} from '@/types/api';

// Get list of VMs with optional pagination and filtering
export const getVMs = async (
  page: number = 1, 
  pageSize: number = 50, 
  name?: string, 
  status?: string
): Promise<VMListResponse> => {
  const params = { page, pageSize, name, status };
  const response = await get<VMListResponse>('/api/v1/vms', { params });
  console.log('VM list response:', response);
  return response;
};

// Get a specific VM by name
export const getVM = async (name: string): Promise<VM> => {
  try {
    // The backend returns { vm: VM } format, not directly VM
    const response = await get<{ vm: VM }>(`/api/v1/vms/${name}`);
    console.log('VM response structure:', response);
    
    // Return the VM from the response
    if (!response || !response.vm) {
      console.error('Invalid VM response structure:', response);
      throw new Error('Invalid VM response: VM data is missing');
    }
    
    return response.vm;
  } catch (error) {
    console.error(`Failed to get VM '${name}':`, error);
    // Re-throw the error so it can be handled by the component
    throw error;
  }
};

// Create a new VM
export const createVM = async (params: VMCreateParams): Promise<VMCreateResponse> => {
  return post<VMCreateResponse>('/api/v1/vms', params);
};

// Delete a VM
export const deleteVM = async (name: string): Promise<void> => {
  return del(`/api/v1/vms/${name}`);
};

// Start a VM
export const startVM = async (name: string): Promise<void> => {
  return put(`/api/v1/vms/${name}/start`);
};

// Stop a VM
export const stopVM = async (name: string): Promise<void> => {
  return put(`/api/v1/vms/${name}/stop`);
};