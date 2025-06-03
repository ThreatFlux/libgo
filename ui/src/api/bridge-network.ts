import client from './client';

export interface BridgeNetwork {
  name: string;
  bridge_name: string;
  active: boolean;
  auto_start: boolean;
  forward_mode: string;
}

export interface CreateBridgeNetworkRequest {
  name: string;
  bridge_name: string;
  auto_start?: boolean;
}

export interface BridgeNetworkListResponse {
  networks: BridgeNetwork[];
  count: number;
}

// List all bridge networks
export async function listBridgeNetworks(): Promise<BridgeNetworkListResponse> {
  const response = await client.get('/api/v1/bridge-networks');
  return response.data;
}

// Get a specific bridge network
export async function getBridgeNetwork(name: string): Promise<BridgeNetwork> {
  const response = await client.get(`/api/v1/bridge-networks/${name}`);
  return response.data;
}

// Create a new bridge network
export async function createBridgeNetwork(data: CreateBridgeNetworkRequest): Promise<{ message: string; name: string }> {
  const response = await client.post('/api/v1/bridge-networks', data);
  return response.data;
}

// Delete a bridge network
export async function deleteBridgeNetwork(name: string): Promise<{ message: string }> {
  const response = await client.delete(`/api/v1/bridge-networks/${name}`);
  return response.data;
}