# API Reference

ProxiCloud exposes a RESTful API for programmatic access to container management and analytics.

---

## 📡 Base URL

When running on a Proxmox node:

```
http://YOUR-PROXMOX-IP:8080/api
```

For local development:
```
http://localhost:8080/api
```

---

## 🔐 Authentication

**Current Version (v1.0)**: No authentication required

The API is designed to run on a trusted local network. Authentication will be added in v2.0.

**Security Recommendations:**
- Only expose on trusted networks
- Use firewall rules to restrict access
- Consider a reverse proxy with authentication if internet-facing

---

## 📋 Response Format

All responses are in JSON format.

### Success Response

```json
{
  "success": true,
  "data": { ... }
}
```

### Error Response

```json
{
  "success": false,
  "error": "Error message description"
}
```

### HTTP Status Codes

| Code | Meaning |
|------|---------|
| 200 | Success |
| 201 | Created |
| 400 | Bad Request (invalid parameters) |
| 404 | Not Found |
| 500 | Internal Server Error |
| 503 | Service Unavailable (Proxmox unreachable) |

---

## 🏥 Health & Info

### Check API Health

```http
GET /api/health
```

Returns the health status of the API and connection to Proxmox.

**Response**:
```json
{
  "status": "healthy",
  "proxmox_connected": true,
  "database_connected": true,
  "timestamp": 1701619200
}
```

**Response Headers**:
- `X-Cache-Status`: `online` or `offline` (indicates if Proxmox is reachable)

---

### Get Node Information

```http
GET /api/info
```

Returns information about the Proxmox node.

**Response**:
```json
{
  "node_name": "pve",
  "version": "8.1.3",
  "uptime": 864000,
  "cpu_cores": 16,
  "total_memory": 67108864,
  "proxicloud_version": "1.0.0"
}
```

---

## 🗄️ Container Management

### List All Containers

```http
GET /api/containers
```

Returns a list of all LXC containers on the node.

**Response**:
```json
{
  "containers": [
    {
      "vmid": 100,
      "name": "proxicloud-001",
      "status": "running",
      "template": "ubuntu-24.04",
      "cpu": 2,
      "memory": 1024,
      "disk": 8589934592,
      "uptime": 3600,
      "cpu_usage": 23.5,
      "memory_usage": 524288000,
      "network_in": 1048576,
      "network_out": 2097152,
      "ip_address": "192.168.1.100"
    }
  ],
  "total": 1,
  "running": 1,
  "stopped": 0
}
```

**Query Parameters**:
| Parameter | Type | Description |
|-----------|------|-------------|
| `status` | string | Filter by status: `running`, `stopped`, `paused` |
| `sort` | string | Sort by: `vmid`, `name`, `status`, `cpu_usage` |
| `order` | string | Sort order: `asc`, `desc` |

**Example**:
```http
GET /api/containers?status=running&sort=cpu_usage&order=desc
```

---

### Get Container Details

```http
GET /api/containers/:id
```

Returns detailed information about a specific container.

**Parameters**:
- `:id` - Container VMID (integer)

**Response**:
```json
{
  "vmid": 100,
  "name": "proxicloud-001",
  "status": "running",
  "template": "ubuntu-24.04",
  "config": {
    "cores": 2,
    "memory": 1024,
    "swap": 512,
    "rootfs": "local-lvm:vm-100-disk-0,size=8G",
    "net0": "name=eth0,bridge=vmbr0,ip=dhcp",
    "onboot": true,
    "features": "nesting=1"
  },
  "current": {
    "status": "running",
    "cpu_usage": 23.5,
    "memory_usage": 524288000,
    "memory_total": 1073741824,
    "disk_usage": 2147483648,
    "disk_total": 8589934592,
    "uptime": 3600,
    "network": {
      "in": 1048576,
      "out": 2097152
    }
  },
  "ip_address": "192.168.1.100",
  "created_at": 1701519200
}
```

---

### Create Container

```http
POST /api/containers
```

Creates a new LXC container.

**Request Body**:
```json
{
  "hostname": "proxicloud-001",
  "template": "ubuntu-24.04-standard",
  "cores": 2,
  "memory": 1024,
  "swap": 512,
  "rootfs": 8,
  "storage": "local-lvm",
  "network": {
    "bridge": "vmbr0",
    "ip": "dhcp",
    "ip6": "dhcp"
  },
  "password": "secure-password",
  "ssh_keys": [
    "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAAB..."
  ],
  "start": true,
  "onboot": true,
  "unprivileged": true,
  "features": {
    "nesting": true,
    "keyctl": false
  }
}
```

**Field Descriptions**:
| Field | Type | Required | Default | Description |
|-------|------|----------|---------|-------------|
| `hostname` | string | No | auto-generated | Container hostname |
| `template` | string | Yes | - | OS template ID |
| `cores` | integer | No | 2 | CPU cores |
| `memory` | integer | No | 1024 | RAM in MB |
| `swap` | integer | No | 512 | Swap in MB |
| `rootfs` | integer | No | 9 | Root disk size in GB |
| `storage` | string | No | `local-lvm` | Storage pool |
| `network.bridge` | string | No | `vmbr0` | Network bridge |
| `network.ip` | string | No | `dhcp` | IPv4 address or `dhcp` |
| `password` | string | No | - | Root password |
| `ssh_keys` | array | No | - | SSH public keys |
| `start` | boolean | No | `false` | Start after creation |
| `onboot` | boolean | No | `true` | Start on host boot |
| `unprivileged` | boolean | No | `true` | Use unprivileged container |

**Response** (201 Created):
```json
{
  "success": true,
  "vmid": 100,
  "name": "proxicloud-001",
  "task_id": "UPID:pve:00001234:00ABCDEF:...",
  "status": "creating"
}
```

**Error Response** (400 Bad Request):
```json
{
  "success": false,
  "error": "Template 'invalid-template' not found"
}
```

---

### Generate Container Name

```http
GET /api/containers/generate-name
```

Returns an auto-generated container name that's not in use.

**Response**:
```json
{
  "name": "proxicloud-003",
  "counter": 3
}
```

---

### Start Container

```http
POST /api/containers/:id/start
```

Starts a stopped container.

**Parameters**:
- `:id` - Container VMID

**Response**:
```json
{
  "success": true,
  "vmid": 100,
  "status": "starting",
  "task_id": "UPID:pve:..."
}
```

---

### Stop Container

```http
POST /api/containers/:id/stop
```

Stops a running container gracefully.

**Parameters**:
- `:id` - Container VMID

**Query Parameters**:
| Parameter | Type | Default | Description |
|-----------|------|---------|-------------|
| `force` | boolean | `false` | Force stop (kill) if true |
| `timeout` | integer | 60 | Seconds to wait before forcing |

**Example**:
```http
POST /api/containers/100/stop?timeout=30
```

**Response**:
```json
{
  "success": true,
  "vmid": 100,
  "status": "stopping",
  "task_id": "UPID:pve:..."
}
```

---

### Reboot Container

```http
POST /api/containers/:id/reboot
```

Reboots a running container.

**Parameters**:
- `:id` - Container VMID

**Response**:
```json
{
  "success": true,
  "vmid": 100,
  "status": "rebooting",
  "task_id": "UPID:pve:..."
}
```

---

### Delete Container

```http
DELETE /api/containers/:id
```

Deletes a container permanently.

**Parameters**:
- `:id` - Container VMID

**Query Parameters**:
| Parameter | Type | Default | Description |
|-----------|------|---------|-------------|
| `purge` | boolean | `true` | Remove from all backups/replications |

**Response**:
```json
{
  "success": true,
  "vmid": 100,
  "status": "deleted",
  "task_id": "UPID:pve:..."
}
```

---

### Get Container Status

```http
GET /api/containers/:id/status
```

Returns the current status of a container.

**Response**:
```json
{
  "vmid": 100,
  "status": "running",
  "uptime": 3600,
  "pid": 12345
}
```

---

## 📊 Analytics & Metrics

### Get CPU Metrics

```http
GET /api/analytics/:id/cpu
```

Returns historical CPU usage data for a container.

**Parameters**:
- `:id` - Container VMID

**Query Parameters**:
| Parameter | Type | Default | Description |
|-----------|------|---------|-------------|
| `range` | string | `1h` | Time range: `1h`, `24h`, `7d`, `30d` |
| `interval` | string | auto | Data point interval: `30s`, `1m`, `5m`, `1h` |

**Example**:
```http
GET /api/analytics/100/cpu?range=24h&interval=5m
```

**Response**:
```json
{
  "vmid": 100,
  "metric": "cpu",
  "range": "24h",
  "interval": "5m",
  "data_points": 288,
  "data": [
    {
      "timestamp": 1701619200,
      "value": 23.5
    },
    {
      "timestamp": 1701619500,
      "value": 25.1
    }
  ],
  "average": 24.3,
  "min": 12.1,
  "max": 45.8
}
```

---

### Get Memory Metrics

```http
GET /api/analytics/:id/memory
```

Returns historical memory usage data.

**Parameters & Response**: Same format as CPU metrics

**Response Example**:
```json
{
  "vmid": 100,
  "metric": "memory",
  "range": "1h",
  "data": [
    {
      "timestamp": 1701619200,
      "value": 524288000,
      "percent": 48.5
    }
  ],
  "average": 512000000,
  "min": 450000000,
  "max": 600000000
}
```

---

### Get Network Metrics

```http
GET /api/analytics/:id/network
```

Returns historical network throughput data.

**Response**:
```json
{
  "vmid": 100,
  "metric": "network",
  "range": "1h",
  "data": [
    {
      "timestamp": 1701619200,
      "in": 1048576,
      "out": 2097152,
      "in_rate": 34952,
      "out_rate": 69905
    }
  ],
  "totals": {
    "in": 3774873600,
    "out": 7549747200
  }
}
```

---

### Get Disk I/O Metrics

```http
GET /api/analytics/:id/disk
```

Returns historical disk I/O data.

**Response**:
```json
{
  "vmid": 100,
  "metric": "disk",
  "range": "1h",
  "data": [
    {
      "timestamp": 1701619200,
      "read": 1048576,
      "write": 2097152,
      "read_rate": 34952,
      "write_rate": 69905
    }
  ],
  "totals": {
    "read": 524288000,
    "write": 1048576000
  }
}
```

---

### Get All Metrics Summary

```http
GET /api/analytics/:id/summary
```

Returns a summary of all metrics for a container.

**Query Parameters**:
| Parameter | Type | Default | Description |
|-----------|------|---------|-------------|
| `range` | string | `1h` | Time range |

**Response**:
```json
{
  "vmid": 100,
  "range": "1h",
  "cpu": {
    "current": 23.5,
    "average": 24.3,
    "min": 12.1,
    "max": 45.8
  },
  "memory": {
    "current": 524288000,
    "average": 512000000,
    "total": 1073741824,
    "percent": 48.5
  },
  "network": {
    "in_rate": 34952,
    "out_rate": 69905,
    "total_in": 3774873600,
    "total_out": 7549747200
  },
  "disk": {
    "usage": 2147483648,
    "total": 8589934592,
    "percent": 25.0,
    "read_rate": 34952,
    "write_rate": 69905
  }
}
```

---

## 📦 Templates

### List Templates

```http
GET /api/templates
```

Returns a list of featured templates configured in ProxiCloud.

**Response**:
```json
{
  "templates": [
    {
      "id": "ubuntu-24.04-standard",
      "name": "Ubuntu 24.04 LTS",
      "description": "Ubuntu 24.04 Noble Numbat (LTS)",
      "distribution": "ubuntu",
      "version": "24.04",
      "featured": true,
      "size": 471859200,
      "available": true
    }
  ],
  "total": 5
}
```

---

### List All Templates

```http
GET /api/templates/all
```

Returns all templates available in Proxmox storage.

**Response**: Same format as `/api/templates`, but includes all templates, not just featured ones.

---

### Get Template Details

```http
GET /api/templates/:id/info
```

Returns detailed information about a specific template.

**Parameters**:
- `:id` - Template ID (e.g., `ubuntu-24.04-standard`)

**Response**:
```json
{
  "id": "ubuntu-24.04-standard",
  "name": "Ubuntu 24.04 LTS",
  "description": "Ubuntu 24.04 Noble Numbat (LTS)",
  "distribution": "ubuntu",
  "version": "24.04",
  "architecture": "amd64",
  "size": 471859200,
  "checksum": "sha512:abc123...",
  "url": "http://download.proxmox.com/...",
  "available": true,
  "storage": "local",
  "path": "/var/lib/vz/template/cache/ubuntu-24.04-standard_24.04-1_amd64.tar.zst"
}
```

---

## 📈 Dashboard

### Get Dashboard Summary

```http
GET /api/dashboard/summary
```

Returns overall statistics for the dashboard view.

**Response**:
```json
{
  "containers": {
    "total": 15,
    "running": 12,
    "stopped": 3,
    "paused": 0
  },
  "resources": {
    "cpu": {
      "used": 45.2,
      "total": 16
    },
    "memory": {
      "used": 25769803776,
      "total": 67108864000,
      "percent": 38.4
    },
    "storage": {
      "used": 107374182400,
      "total": 536870912000,
      "percent": 20.0
    }
  },
  "node": {
    "name": "pve",
    "version": "8.1.3",
    "uptime": 864000,
    "cpu_cores": 16,
    "load_average": [1.23, 1.45, 1.67]
  },
  "recent_activity": [
    {
      "timestamp": 1701619200,
      "action": "create",
      "vmid": 105,
      "name": "proxicloud-005",
      "details": "Created from ubuntu-24.04-standard"
    },
    {
      "timestamp": 1701619100,
      "action": "start",
      "vmid": 104,
      "name": "proxicloud-004"
    }
  ]
}
```

---

## 🔍 Search & Filter

### Search Containers

```http
GET /api/containers/search?q=web
```

Search containers by name, hostname, or IP address.

**Query Parameters**:
| Parameter | Type | Description |
|-----------|------|-------------|
| `q` | string | Search query |

**Response**:
```json
{
  "query": "web",
  "results": [
    {
      "vmid": 100,
      "name": "web-server-01",
      "status": "running",
      "ip_address": "192.168.1.100"
    }
  ],
  "total": 1
}
```

---

## 📡 WebSocket Events (Planned v2.0)

Real-time updates via WebSocket will be available in v2.0.

**Planned endpoint**:
```
ws://YOUR-PROXMOX-IP:8080/api/ws
```

**Events**:
- `container.created`
- `container.started`
- `container.stopped`
- `container.deleted`
- `metrics.update`

---

## 💡 Usage Examples

### cURL Examples

**List all containers**:
```bash
curl http://192.168.1.100:8080/api/containers
```

**Create a container**:
```bash
curl -X POST http://192.168.1.100:8080/api/containers \
  -H "Content-Type: application/json" \
  -d '{
    "hostname": "test-container",
    "template": "ubuntu-24.04-standard",
    "cores": 2,
    "memory": 1024,
    "start": true
  }'
```

**Get CPU metrics**:
```bash
curl "http://192.168.1.100:8080/api/analytics/100/cpu?range=24h"
```

---

### JavaScript/TypeScript Example

```typescript
// API Client wrapper
class ProxiCloudAPI {
  constructor(private baseUrl: string) {}

  async listContainers() {
    const response = await fetch(`${this.baseUrl}/containers`);
    return response.json();
  }

  async createContainer(config: ContainerConfig) {
    const response = await fetch(`${this.baseUrl}/containers`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(config)
    });
    return response.json();
  }

  async startContainer(vmid: number) {
    const response = await fetch(`${this.baseUrl}/containers/${vmid}/start`, {
      method: 'POST'
    });
    return response.json();
  }

  async getMetrics(vmid: number, metric: string, range: string = '1h') {
    const response = await fetch(
      `${this.baseUrl}/analytics/${vmid}/${metric}?range=${range}`
    );
    return response.json();
  }
}

// Usage
const api = new ProxiCloudAPI('http://192.168.1.100:8080/api');

// List containers
const containers = await api.listContainers();
console.log(`Total containers: ${containers.total}`);

// Create container
const result = await api.createContainer({
  hostname: 'web-server',
  template: 'ubuntu-24.04-standard',
  cores: 4,
  memory: 4096,
  start: true
});
console.log(`Created container: ${result.vmid}`);

// Get CPU metrics
const metrics = await api.getMetrics(100, 'cpu', '24h');
console.log(`Average CPU: ${metrics.average}%`);
```

---

### Python Example

```python
import requests

class ProxiCloudAPI:
    def __init__(self, base_url):
        self.base_url = base_url

    def list_containers(self):
        response = requests.get(f"{self.base_url}/containers")
        return response.json()

    def create_container(self, config):
        response = requests.post(
            f"{self.base_url}/containers",
            json=config
        )
        return response.json()

    def get_metrics(self, vmid, metric, range='1h'):
        response = requests.get(
            f"{self.base_url}/analytics/{vmid}/{metric}",
            params={'range': range}
        )
        return response.json()

# Usage
api = ProxiCloudAPI('http://192.168.1.100:8080/api')

# List containers
containers = api.list_containers()
print(f"Total containers: {containers['total']}")

# Create container
result = api.create_container({
    'hostname': 'web-server',
    'template': 'ubuntu-24.04-standard',
    'cores': 4,
    'memory': 4096,
    'start': True
})
print(f"Created container: {result['vmid']}")

# Get metrics
metrics = api.get_metrics(100, 'cpu', '24h')
print(f"Average CPU: {metrics['average']}%")
```

---

## 📚 Related Documentation

- [Installation Guide](INSTALLATION.md)
- [Configuration Reference](CONFIGURATION.md)
- [Development Guide](DEVELOPMENT.md)
- [Architecture Overview](ARCHITECTURE.md)

---

## 🤝 Contributing

API improvements and new endpoints are welcome! See [DEVELOPMENT.md](DEVELOPMENT.md) for contribution guidelines.
