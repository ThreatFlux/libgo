import apiClient from './client';

// Storage pool types
export interface StoragePoolInfo {
  uuid: string;
  name: string;
  type: string;
  state: 'inactive' | 'building' | 'running' | 'degraded' | 'inaccessible';
  autostart: boolean;
  persistent: boolean;
  capacity: number;
  allocation: number;
  available: number;
  path?: string;
  source?: {
    host?: string;
    dir?: string;
    device?: string;
    name?: string;
    format?: string;
  };
  target?: {
    path: string;
    permissions?: {
      mode?: string;
      owner?: string;
      group?: string;
    };
  };
  metadata?: Record<string, any>;
}

export interface CreatePoolParams {
  name: string;
  type: string;
  path?: string;
  source?: {
    host?: string;
    dir?: string;
    device?: string;
    name?: string;
    format?: string;
  };
  autostart: boolean;
  metadata?: Record<string, any>;
}

// Storage volume types
export interface StorageVolumeInfo {
  name: string;
  key: string;
  path: string;
  type: string;
  capacity: number;
  allocation: number;
  format: string;
  pool: string;
  backing_store?: {
    path: string;
    format: string;
  };
  metadata?: Record<string, any>;
}

export interface CreateVolumeParams {
  name: string;
  capacity_bytes: number;
  format?: string;
  backing_store?: string;
  metadata?: Record<string, any>;
}

// Storage pool API
export const storagePoolApi = {
  list: async () => {
    const response = await apiClient.get<{ pools: StoragePoolInfo[] }>('/api/v1/storage/pools');
    return response.data.pools;
  },

  get: async (name: string) => {
    const response = await apiClient.get<StoragePoolInfo>(`/api/v1/storage/pools/${name}`);
    return response.data;
  },

  create: async (params: CreatePoolParams) => {
    const response = await apiClient.post<StoragePoolInfo>('/api/v1/storage/pools', params);
    return response.data;
  },

  delete: async (name: string) => {
    await apiClient.delete(`/api/v1/storage/pools/${name}`);
  },

  start: async (name: string) => {
    const response = await apiClient.put<{ message: string }>(`/api/v1/storage/pools/${name}/start`);
    return response.data;
  },

  stop: async (name: string) => {
    const response = await apiClient.put<{ message: string }>(`/api/v1/storage/pools/${name}/stop`);
    return response.data;
  },
};

// Storage volume API
export const storageVolumeApi = {
  list: async (poolName: string) => {
    const response = await apiClient.get<{ volumes: StorageVolumeInfo[] }>(`/api/v1/storage/pools/${poolName}/volumes`);
    return response.data.volumes;
  },

  create: async (poolName: string, params: CreateVolumeParams) => {
    const response = await apiClient.post<StorageVolumeInfo>(`/api/v1/storage/pools/${poolName}/volumes`, params);
    return response.data;
  },

  delete: async (poolName: string, volumeName: string) => {
    await apiClient.delete(`/api/v1/storage/pools/${poolName}/volumes/${volumeName}`);
  },

  upload: async (poolName: string, volumeName: string, file: File, onProgress?: (progress: number) => void) => {
    const formData = new FormData();
    formData.append('file', file);

    const response = await apiClient.post(
      `/api/v1/storage/pools/${poolName}/volumes/${volumeName}/upload`,
      formData,
      {
        headers: {
          'Content-Type': 'multipart/form-data',
        },
        onUploadProgress: (progressEvent: any) => {
          if (onProgress && progressEvent.total) {
            const progress = Math.round((progressEvent.loaded * 100) / progressEvent.total);
            onProgress(progress);
          }
        },
      }
    );
    return response.data;
  },
};

// Helper functions
export const formatBytes = (bytes: number, decimals = 2) => {
  if (bytes === 0) return '0 Bytes';
  const k = 1024;
  const dm = decimals < 0 ? 0 : decimals;
  const sizes = ['Bytes', 'KB', 'MB', 'GB', 'TB', 'PB'];
  const i = Math.floor(Math.log(bytes) / Math.log(k));
  return parseFloat((bytes / Math.pow(k, i)).toFixed(dm)) + ' ' + sizes[i];
};

export const getPoolStateColor = (state: StoragePoolInfo['state']) => {
  switch (state) {
    case 'running':
      return 'success';
    case 'building':
      return 'warning';
    case 'inactive':
      return 'default';
    case 'degraded':
      return 'warning';
    case 'inaccessible':
      return 'destructive';
    default:
      return 'default';
  }
};