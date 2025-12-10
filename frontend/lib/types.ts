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
  project_id?: string;
}

export interface CreateContainerRequest {
  vmid?: number; // Optional: user can specify their own VMID
  hostname: string;
  cores: number;
  memory: number;
  disk: number;
  ostemplate: string;
  password?: string;
  ssh_keys?: string;
  start_on_boot?: boolean;
  unprivileged?: boolean;
  project_id?: string;
  // Network configuration
  ip_address?: string; // IP address with CIDR notation (e.g., "192.168.1.100/24")
  gateway?: string;    // Gateway IP address (e.g., "192.168.1.1")
  nameserver?: string; // DNS nameserver (e.g., "8.8.8.8")
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

export interface MetricsSummary {
  vmid: number;
  start_time: string;
  end_time: string;
  avg_cpu: number;
  max_cpu: number;
  avg_mem_usage: number;
  max_mem_usage: number;
  avg_disk_usage: number;
  total_net_in: number;
  total_net_out: number;
  data_points: number;
}

export interface Template {
  volid: string;
  format: string;
  size: number;
  content: string;
}

export interface Volume {
  volid: string;
  name: string;
  size: number; // Size in GB
  used: number; // Used space in GB
  node: string;
  storage: string;
  type: string; // ssd or hdd
  format: string;
  status: 'available' | 'in-use' | 'error';
  attached_to?: number; // VMID if attached
  mountpoint?: string;
  created_at?: number;
}

export interface CreateVolumeRequest {
  name: string;
  size: number; // Size in GB
  storage?: string; // Storage pool (default: local-lvm)
  type?: string; // ssd or hdd (default: ssd)
  node?: string;
}

export interface AttachVolumeRequest {
  vmid: number;
  mountpoint?: string; // Optional: mp0-mp9
}

export interface DetachVolumeRequest {
  vmid: number;
  force?: boolean;
}

export interface Snapshot {
  name: string;
  volid: string;
  description?: string;
  created_at: number;
  size: number;
  parent?: string;
}

export interface CreateSnapshotRequest {
  name: string;
  description?: string;
}

export interface RestoreSnapshotRequest {
  snapshot_name: string;
}

export interface CloneSnapshotRequest {
  snapshot_name: string;
  new_name: string;
  storage?: string;
}

export interface Project {
  id: string;
  name: string;
  description?: string;
  tags?: string[];
  created_at: number;
  updated_at: number;
}

export interface CreateProjectRequest {
  name: string;
  description?: string;
  tags?: string[];
}

export interface UpdateProjectRequest {
  name?: string;
  description?: string;
  tags?: string[];
}

export interface AssignProjectRequest {
  project_id: string; // Empty string means "No Project"
}

