import { get, post, del } from './client';

// OVS Bridge types
export interface OVSBridge {
  name: string;
  uuid: string;
  controller?: string;
  datapath_type: string;
  ports: string[];
  external_ids?: Record<string, string>;
  other_config?: Record<string, string>;
  statistics?: {
    flow_count: number;
    port_count: number;
    lookup_count: number;
    matched_count: number;
  };
}

export interface CreateBridgeRequest {
  name: string;
  datapath_type?: string;
  controller?: string;
  external_ids?: Record<string, string>;
  other_config?: Record<string, string>;
}

// OVS Port types
export interface OVSPort {
  name: string;
  uuid: string;
  bridge: string;
  type: string;
  tag?: number;
  trunks?: number[];
  interfaces: string[];
  external_ids?: Record<string, string>;
  other_config?: Record<string, string>;
  statistics?: {
    rx_packets: number;
    rx_bytes: number;
    rx_dropped: number;
    rx_errors: number;
    tx_packets: number;
    tx_bytes: number;
    tx_dropped: number;
    tx_errors: number;
  };
}

export interface CreatePortRequest {
  name: string;
  bridge: string;
  type?: string;
  tag?: number;
  trunks?: number[];
  peer_port?: string;
  remote_ip?: string;
  tunnel_type?: string;
  external_ids?: Record<string, string>;
  other_config?: Record<string, string>;
}

// OVS Flow types
export interface OVSFlow {
  id: string;
  table: number;
  priority: number;
  match: string;
  actions: string;
  cookie?: string;
}

export interface CreateFlowRequest {
  bridge: string;
  table: number;
  priority: number;
  match: string;
  actions: string;
  cookie?: string;
}

// API responses
export interface BridgeListResponse {
  bridges: OVSBridge[];
  count: number;
}

export interface BridgeResponse {
  bridge: OVSBridge;
}

export interface PortListResponse {
  ports: OVSPort[];
  count: number;
}

export interface PortResponse {
  port: OVSPort;
}

export interface FlowResponse {
  flow: OVSFlow;
}

export interface FlowListResponse {
  flows: OVSFlow[];
  count: number;
}

// Bridge API functions
export const listBridges = async (): Promise<BridgeListResponse> => {
  return get<BridgeListResponse>('/api/v1/ovs/bridges');
};

export const getBridge = async (name: string): Promise<BridgeResponse> => {
  return get<BridgeResponse>(`/api/v1/ovs/bridges/${name}`);
};

export const createBridge = async (data: CreateBridgeRequest): Promise<BridgeResponse> => {
  return post<BridgeResponse>('/api/v1/ovs/bridges', data);
};

export const deleteBridge = async (name: string): Promise<void> => {
  return del<void>(`/api/v1/ovs/bridges/${name}`);
};

// Port API functions
export const listPorts = async (bridge: string): Promise<PortListResponse> => {
  return get<PortListResponse>(`/api/v1/ovs/bridges/${bridge}/ports`);
};

export const createPort = async (data: CreatePortRequest): Promise<PortResponse> => {
  return post<PortResponse>(`/api/v1/ovs/bridges/${data.bridge}/ports`, data);
};

export const deletePort = async (bridge: string, port: string): Promise<void> => {
  return del<void>(`/api/v1/ovs/bridges/${bridge}/ports/${port}`);
};

// Flow API functions
export const createFlow = async (data: CreateFlowRequest): Promise<FlowResponse> => {
  return post<FlowResponse>(`/api/v1/ovs/bridges/${data.bridge}/flows`, data);
};

export const listFlows = async (bridge: string): Promise<FlowListResponse> => {
  return get<FlowListResponse>(`/api/v1/ovs/bridges/${bridge}/flows`);
};