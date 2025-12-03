export interface Container {
  vmid: number;
  name: string;
  status: 'running' | 'stopped';
  node: string;
  cpu: number;
  mem: number;
  maxmem: number;
  disk: number;
  maxdisk: number;
  uptime: number;
  template?: string;
  os?: string;
}

export interface CreateContainerRequest {
  hostname: string;
  cores: number;
  memory: number;
  disk: number;
  ostemplate: string;
  password?: string;
  ssh_keys?: string;
  start_on_boot?: boolean;
  unprivileged?: boolean;
}

export interface DashboardStats {
  total_containers: number;
  running_containers: number;
  stopped_containers: number;
  total_cpu: number;
  total_memory: number;
  used_memory: number;
  total_disk: number;
  used_disk: number;
}

export interface MetricsData {
  timestamp: number;
  cpu: number;
  memory: number;
  disk_read: number;
  disk_write: number;
  net_in: number;
  net_out: number;
}

export interface Template {
  volid: string;
  format: string;
  size: number;
  content: string;
}
