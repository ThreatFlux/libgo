// Auth Types
export interface LoginRequest {
  username: string;
  password: string;
}

export interface LoginResponse {
  token: string;
  expiresAt: string;
  user: User;
}

export interface User {
  id: string;
  username: string;
  email: string;
  roles: string[];
  createdAt: string;
  updatedAt: string;
}

// VM Types
export interface VM {
  name: string;
  uuid: string;
  status: VMStatus;
  cpu: CPUConfig;
  memory: MemoryConfig;
  disks: Disk[];
  networks: Network[];
  createdAt: string;
  description: string;
}

export type VMStatus = 'running' | 'stopped' | 'paused' | 'shutdown' | 'crashed' | 'unknown';

export interface CPUConfig {
  count: number;
  model: string;
  sockets: number;
  cores: number;
  threads: number;
}

export interface MemoryConfig {
  sizeBytes: number;
}

export interface Disk {
  path: string;
  format: 'qcow2' | 'raw';
  sizeBytes: number;
  bus: 'virtio' | 'ide' | 'sata' | 'scsi';
  readOnly: boolean;
  bootable: boolean;
  shareable: boolean;
  serial: string;
  storagePool: string;
  device: string;
}

export interface Network {
  type: 'bridge' | 'network' | 'direct';
  source: string;
  model: string;
  macAddress: string;
  ipAddress: string;
  ipAddressV6: string;
}

export interface VMListResponse {
  vms: VM[];
  count: number;
  pageSize: number;
  page: number;
}

export interface VMCreateParams {
  name: string;
  description: string;
  cpu: {
    count: number;
    model: string;
    socket: number;
    cores: number;
    threads: number;
  };
  memory: {
    sizeBytes: number;
  };
  disk: {
    sizeBytes: number;
    format: 'qcow2' | 'raw';
    sourceImage: string;
    storagePool: string;
    bus: 'virtio' | 'ide' | 'sata' | 'scsi';
    cacheMode: 'none' | 'writeback' | 'writethrough' | 'directsync' | 'unsafe';
    shareable: boolean;
    readOnly: boolean;
  };
  network: {
    type: 'bridge' | 'network' | 'direct';
    source: string;
    model: 'virtio' | 'e1000' | 'rtl8139';
    macAddress: string;
  };
  cloudInit?: {
    userData: string;
    metaData: string;
    networkConfig: string;
    sshKeys: string[];
  };
  template?: string;
}

export interface VMCreateResponse {
  vm: VM;
}

// Export Types
export type ExportFormat = 'qcow2' | 'vmdk' | 'vdi' | 'ova' | 'raw';
export type ExportStatus = 'pending' | 'running' | 'completed' | 'failed' | 'cancelled';

export interface ExportJob {
  id: string;
  vmName: string;
  format: ExportFormat;
  status: ExportStatus;
  progress: number;
  startTime: string;
  endTime: string;
  error: string;
  outputPath: string;
  options: Record<string, string>;
}

export interface ExportRequest {
  format: ExportFormat;
  options?: Record<string, string>;
  fileName?: string;
}

export interface ExportResponse {
  job: ExportJob;
}

export interface ExportListResponse {
  jobs: ExportJob[];
  count: number;
  pageSize: number;
  page: number;
}

// System Metrics
export interface SystemMetrics {
  vmCount: number;
  cpuUtilization: number;
  memoryUtilization: number;
  storageUtilization: number;
  runningVMs: number;
  exportJobs: number;
  activeExportJobs: number;
}

export interface StoragePoolInfo {
  name: string;
  capacity: number;
  available: number;
  allocation: number;
}

export interface HealthStatus {
  status: 'UP' | 'DOWN';
  components: Record<string, ComponentHealth>;
  version: string;
  buildTime: string;
}

export interface ComponentHealth {
  status: 'UP' | 'DOWN';
  details?: Record<string, unknown>;
}