import apiClient from './client';

// Network types
export interface NetworkForward {
  mode: string; // nat, route, bridge, private, vepa, passthrough
  dev?: string;
}

export interface NetworkDHCPInfo {
  enabled: boolean;
  start?: string;
  end?: string;
  hosts?: NetworkDHCPStaticHost[];
}

export interface NetworkDHCPStaticHost {
  mac: string;
  name?: string;
  ip: string;
}

export interface NetworkIP {
  address: string;
  netmask: string;
  dhcp?: NetworkDHCPInfo;
}

export interface NetworkDHCPLease {
  ip_address: string;
  mac_address: string;
  hostname?: string;
  client_id?: string;
  expiry_time: number;
}

export interface NetworkInfo {
  uuid: string;
  name: string;
  bridge_name: string;
  active: boolean;
  persistent: boolean;
  autostart: boolean;
  forward: NetworkForward;
  ip?: NetworkIP;
  dhcp_leases?: NetworkDHCPLease[];
  metadata?: Record<string, any>;
}

export interface CreateNetworkParams {
  name: string;
  bridge_name?: string;
  forward?: NetworkForward;
  ip?: NetworkIP;
  autostart: boolean;
  metadata?: Record<string, any>;
}

export interface UpdateNetworkParams {
  forward?: NetworkForward;
  ip?: NetworkIP;
  autostart?: boolean;
  metadata?: Record<string, any>;
}

export interface NetworkListResponse {
  networks: NetworkInfo[];
  count: number;
}

// Network API client
export const networkAPI = {
  // List all networks
  list: async (): Promise<NetworkListResponse> => {
    const response = await apiClient.get<NetworkListResponse>('/api/v1/networks');
    return response.data;
  },

  // Get network details
  get: async (name: string): Promise<NetworkInfo> => {
    const response = await apiClient.get<NetworkInfo>(`/api/v1/networks/${name}`);
    return response.data;
  },

  // Create a new network
  create: async (params: CreateNetworkParams): Promise<NetworkInfo> => {
    const response = await apiClient.post<NetworkInfo>('/api/v1/networks', params);
    return response.data;
  },

  // Update network
  update: async (name: string, params: UpdateNetworkParams): Promise<NetworkInfo> => {
    const response = await apiClient.put<NetworkInfo>(`/api/v1/networks/${name}`, params);
    return response.data;
  },

  // Delete network
  delete: async (name: string): Promise<void> => {
    await apiClient.delete(`/api/v1/networks/${name}`);
  },

  // Start network
  start: async (name: string): Promise<NetworkInfo> => {
    const response = await apiClient.put<NetworkInfo>(`/api/v1/networks/${name}/start`);
    return response.data;
  },

  // Stop network
  stop: async (name: string): Promise<NetworkInfo> => {
    const response = await apiClient.put<NetworkInfo>(`/api/v1/networks/${name}/stop`);
    return response.data;
  },
};