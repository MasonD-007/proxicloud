# Configuration Reference

This document describes all configuration options available in ProxiCloud.

---

## 📁 Configuration File Location

The main configuration file is located at:

```
/etc/proxicloud/config.yaml
```

### File Permissions

For security, the configuration file should have restricted permissions since it contains sensitive API tokens:

```bash
chmod 600 /etc/proxicloud/config.yaml
chown root:root /etc/proxicloud/config.yaml
```

---

## 📝 Configuration Format

ProxiCloud uses YAML format for configuration. Here's the complete structure:

```yaml
# ProxiCloud Configuration File
# Version: 1.0

proxmox:
  api_url: "https://localhost:8006/api2/json"
  api_token: "root@pam!proxicloud=xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx"
  node: ""
  insecure_skip_verify: true

server:
  backend_port: 8080
  frontend_port: 3000
  bind_address: "0.0.0.0"

analytics:
  interval: "30s"
  retention_days: 30
  database_path: "/var/lib/proxicloud/analytics.db"

storage:
  default_storage: "local-lvm"

templates:
  featured:
    - "ubuntu-22.04-standard"
    - "ubuntu-24.04-standard"
    - "debian-12-standard"
    - "alpine-3.19-default"
    - "rockylinux-9-default"

logging:
  level: "info"
  path: "/var/log/proxicloud/app.log"

defaults:
  cpu_cores: 2
  memory_mb: 1024
  disk_gb: 9
  network_bridge: "vmbr0"
  container_name_prefix: "proxicloud"
```

---

## 🔧 Configuration Sections

### `proxmox` - Proxmox Connection Settings

#### `api_url`
- **Type**: String
- **Default**: `"https://localhost:8006/api2/json"`
- **Description**: Proxmox API endpoint URL
- **Notes**: 
  - Use `localhost` when ProxiCloud runs on the same node
  - Always use HTTPS (Proxmox default)
  - Include `/api2/json` path

**Example**:
```yaml
proxmox:
  api_url: "https://192.168.1.100:8006/api2/json"  # Remote node
```

#### `api_token`
- **Type**: String
- **Required**: Yes
- **Format**: `user@realm!tokenid=secret`
- **Description**: Proxmox API authentication token
- **Notes**:
  - Create via: Datacenter → Permissions → API Tokens
  - Privilege separation must be **disabled**
  - Token must have required permissions (see Installation guide)

**Example**:
```yaml
proxmox:
  api_token: "root@pam!proxicloud=12345678-1234-1234-1234-123456789abc"
```

#### `node`
- **Type**: String
- **Default**: `""` (auto-detect)
- **Description**: Proxmox node name
- **Notes**:
  - Leave empty to auto-detect from hostname
  - Override if detection fails or using remote node

**Example**:
```yaml
proxmox:
  node: "pve01"  # Explicit node name
```

#### `insecure_skip_verify`
- **Type**: Boolean
- **Default**: `true`
- **Description**: Skip TLS certificate verification
- **Notes**:
  - Set to `true` for self-signed Proxmox certificates
  - Set to `false` if you have proper SSL certificates
  - Required for most Proxmox installations

**Example**:
```yaml
proxmox:
  insecure_skip_verify: false  # Use with valid SSL certs
```

---

### `server` - Server Configuration

#### `backend_port`
- **Type**: Integer
- **Default**: `8080`
- **Description**: Port for the Go API server
- **Notes**:
  - Must not conflict with other services
  - Firewall access not required (only accessed by frontend)

**Example**:
```yaml
server:
  backend_port: 9080  # Change if port 8080 is in use
```

#### `frontend_port`
- **Type**: Integer
- **Default**: `3000`
- **Description**: Port for the Next.js web interface
- **Notes**:
  - This is the port you'll access in your browser
  - Must be allowed through firewall
  - Standard ports: 80 (HTTP), 443 (HTTPS), 3000 (development)

**Example**:
```yaml
server:
  frontend_port: 80  # Use standard HTTP port
```

#### `bind_address`
- **Type**: String
- **Default**: `"0.0.0.0"`
- **Description**: Network interface to bind to
- **Options**:
  - `"0.0.0.0"` - Listen on all interfaces (LAN accessible)
  - `"127.0.0.1"` - Listen only on localhost (local access only)
  - Specific IP - Listen on specific interface

**Example**:
```yaml
server:
  bind_address: "192.168.1.100"  # Specific interface
```

---

### `analytics` - Metrics Collection Settings

#### `interval`
- **Type**: Duration string
- **Default**: `"30s"`
- **Description**: How often to collect metrics from containers
- **Format**: `"Xs"` (seconds), `"Xm"` (minutes), `"Xh"` (hours)
- **Recommended**: 30s - 120s
- **Notes**:
  - Lower values = more frequent updates, more database writes
  - Higher values = less frequent updates, less resource usage

**Example**:
```yaml
analytics:
  interval: "60s"  # Collect metrics every minute
```

#### `retention_days`
- **Type**: Integer
- **Default**: `30`
- **Description**: Number of days to keep historical metrics
- **Notes**:
  - Older data is automatically deleted
  - Higher values = larger database size
  - Cleanup runs hourly

**Example**:
```yaml
analytics:
  retention_days: 90  # Keep 3 months of data
```

#### `database_path`
- **Type**: String
- **Default**: `"/var/lib/proxicloud/analytics.db"`
- **Description**: Path to SQLite database file
- **Notes**:
  - Directory must exist and be writable
  - Database is created automatically if missing
  - Use fast storage (SSD) for better performance

**Example**:
```yaml
analytics:
  database_path: "/mnt/ssd/proxicloud/analytics.db"
```

---

### `storage` - Storage Configuration

#### `default_storage`
- **Type**: String
- **Default**: `"local-lvm"`
- **Description**: Default storage for new containers
- **Notes**:
  - Must match a storage ID in Proxmox
  - Check available storage: `pvesm status`
  - Common values: `local`, `local-lvm`, `local-zfs`

**Example**:
```yaml
storage:
  default_storage: "local-zfs"  # Use ZFS storage
```

---

### `templates` - Container Template Settings

#### `featured`
- **Type**: Array of strings
- **Default**: See example below
- **Description**: List of featured templates to show in UI
- **Notes**:
  - Template IDs must match Proxmox template names
  - These appear first in the create container form
  - Users can still access all templates via "Show all"

**Example**:
```yaml
templates:
  featured:
    - "ubuntu-22.04-standard"
    - "ubuntu-24.04-standard"
    - "debian-12-standard"
    - "debian-11-standard"
    - "alpine-3.19-default"
    - "alpine-3.18-default"
    - "rockylinux-9-default"
    - "centos-stream-9-default"
```

**How to find template IDs**:
```bash
# List all available templates
pveam available

# List downloaded templates
pveam list local
```

---

### `logging` - Logging Configuration

#### `level`
- **Type**: String
- **Default**: `"info"`
- **Options**: `"debug"`, `"info"`, `"warn"`, `"error"`
- **Description**: Minimum log level to record
- **Notes**:
  - `debug` - Very verbose, for troubleshooting
  - `info` - Normal operation (recommended)
  - `warn` - Only warnings and errors
  - `error` - Only errors

**Example**:
```yaml
logging:
  level: "debug"  # Enable debug logging
```

#### `path`
- **Type**: String
- **Default**: `"/var/log/proxicloud/app.log"`
- **Description**: Path to application log file
- **Notes**:
  - Directory must exist and be writable
  - Logs are also sent to systemd journal
  - Use `journalctl -u proxicloud-api` to view

**Example**:
```yaml
logging:
  path: "/var/log/proxicloud/debug.log"
```

---

### `defaults` - Container Creation Defaults

#### `cpu_cores`
- **Type**: Integer
- **Default**: `2`
- **Description**: Default number of CPU cores for new containers
- **Notes**: Users can override when creating containers

**Example**:
```yaml
defaults:
  cpu_cores: 4  # Default to 4 cores
```

#### `memory_mb`
- **Type**: Integer
- **Default**: `1024`
- **Description**: Default RAM in megabytes
- **Notes**: 
  - 1024 = 1GB
  - Users can override via slider in UI

**Example**:
```yaml
defaults:
  memory_mb: 2048  # Default to 2GB
```

#### `disk_gb`
- **Type**: Integer
- **Default**: `9`
- **Description**: Default root filesystem size in gigabytes
- **Notes**: Users can override when creating containers

**Example**:
```yaml
defaults:
  disk_gb: 20  # Default to 20GB
```

#### `network_bridge`
- **Type**: String
- **Default**: `"vmbr0"`
- **Description**: Default network bridge for new containers
- **Notes**:
  - Must match a bridge in Proxmox
  - Check available bridges: `brctl show`

**Example**:
```yaml
defaults:
  network_bridge: "vmbr1"  # Use secondary bridge
```

#### `container_name_prefix`
- **Type**: String
- **Default**: `"proxicloud"`
- **Description**: Prefix for auto-generated container names
- **Format**: `{prefix}-{counter}` → `proxicloud-001`, `proxicloud-002`
- **Notes**: Use lowercase, no spaces

**Example**:
```yaml
defaults:
  container_name_prefix: "lxc"  # Generates: lxc-001, lxc-002, etc.
```

---

## 🔐 Environment Variable Overrides

Configuration values can be overridden using environment variables. This is useful for:
- Keeping secrets out of config files
- Docker deployments
- CI/CD pipelines

### Environment Variable Format

Format: `PROXICLOUD_{SECTION}_{KEY}`

All uppercase, sections and keys separated by underscores.

### Examples

```bash
# Override API token
export PROXICLOUD_PROXMOX_API_TOKEN="root@pam!token=secret"

# Override ports
export PROXICLOUD_SERVER_BACKEND_PORT=9080
export PROXICLOUD_SERVER_FRONTEND_PORT=8080

# Override log level
export PROXICLOUD_LOGGING_LEVEL="debug"

# Override retention
export PROXICLOUD_ANALYTICS_RETENTION_DAYS=90
```

### Systemd Service Override

To set environment variables for systemd services:

```bash
# Create override directory
mkdir -p /etc/systemd/system/proxicloud-api.service.d

# Create override file
cat > /etc/systemd/system/proxicloud-api.service.d/override.conf <<EOF
[Service]
Environment="PROXICLOUD_LOGGING_LEVEL=debug"
Environment="PROXICLOUD_ANALYTICS_INTERVAL=60s"
EOF

# Reload and restart
systemctl daemon-reload
systemctl restart proxicloud-api
```

---

## 📋 Configuration Examples

### Minimal Configuration

```yaml
proxmox:
  api_token: "root@pam!proxicloud=your-token-here"
```

All other values will use defaults.

---

### Production Configuration

```yaml
proxmox:
  api_url: "https://localhost:8006/api2/json"
  api_token: "root@pam!proxicloud=xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx"
  node: ""
  insecure_skip_verify: true

server:
  backend_port: 8080
  frontend_port: 80
  bind_address: "0.0.0.0"

analytics:
  interval: "60s"
  retention_days: 90
  database_path: "/var/lib/proxicloud/analytics.db"

storage:
  default_storage: "local-zfs"

templates:
  featured:
    - "ubuntu-24.04-standard"
    - "debian-12-standard"
    - "alpine-3.19-default"

logging:
  level: "info"
  path: "/var/log/proxicloud/app.log"

defaults:
  cpu_cores: 2
  memory_mb: 2048
  disk_gb: 20
  network_bridge: "vmbr0"
  container_name_prefix: "srv"
```

---

### Development Configuration

```yaml
proxmox:
  api_url: "https://192.168.1.100:8006/api2/json"
  api_token: "root@pam!dev=test-token"
  node: "pve-test"
  insecure_skip_verify: true

server:
  backend_port: 8080
  frontend_port: 3000
  bind_address: "127.0.0.1"

analytics:
  interval: "10s"
  retention_days: 7
  database_path: "/tmp/proxicloud-dev.db"

logging:
  level: "debug"
  path: "/tmp/proxicloud.log"
```

---

### High-Frequency Monitoring

For infrastructure that needs frequent updates:

```yaml
analytics:
  interval: "10s"      # Collect every 10 seconds
  retention_days: 7     # Keep 1 week (reduces DB size)
```

---

### Low-Resource Mode

For resource-constrained environments:

```yaml
analytics:
  interval: "5m"       # Collect every 5 minutes
  retention_days: 14    # Keep 2 weeks only

defaults:
  cpu_cores: 1
  memory_mb: 512
  disk_gb: 8
```

---

## 🔄 Applying Configuration Changes

After modifying the configuration file:

### Restart Services

```bash
# Restart backend (API + metrics collector)
systemctl restart proxicloud-api

# Restart frontend
systemctl restart proxicloud-frontend
```

### Restart Only If Needed

Some changes only require restarting specific services:

| Configuration Section | Restart Required |
|----------------------|------------------|
| `proxmox.*` | Backend only |
| `server.backend_port` | Backend only |
| `server.frontend_port` | Frontend only |
| `server.bind_address` | Both |
| `analytics.*` | Backend only |
| `storage.*` | Backend only |
| `templates.*` | Backend only |
| `logging.*` | Backend only |
| `defaults.*` | Backend only |

### Verify Configuration

Check if configuration is valid:

```bash
# View current config
cat /etc/proxicloud/config.yaml

# Check if services started successfully
systemctl status proxicloud-api
systemctl status proxicloud-frontend

# Check logs for errors
journalctl -u proxicloud-api -n 50
```

---

## ⚠️ Common Configuration Mistakes

### 1. Invalid YAML Syntax

**Wrong**:
```yaml
proxmox:
api_token: "token"  # Missing indentation
```

**Correct**:
```yaml
proxmox:
  api_token: "token"  # Proper indentation (2 spaces)
```

### 2. Incorrect Token Format

**Wrong**:
```yaml
api_token: "root@pam=token"  # Missing !tokenid
```

**Correct**:
```yaml
api_token: "root@pam!proxicloud=token"
```

### 3. Port Conflicts

**Wrong**:
```yaml
server:
  backend_port: 8006  # Conflicts with Proxmox
  frontend_port: 22   # Conflicts with SSH
```

**Correct**:
```yaml
server:
  backend_port: 8080
  frontend_port: 3000
```

### 4. Invalid Storage

**Wrong**:
```yaml
storage:
  default_storage: "local-ssd"  # Storage doesn't exist
```

**Verify first**:
```bash
pvesm status
```

### 5. Invalid Duration Format

**Wrong**:
```yaml
analytics:
  interval: "30"      # Missing unit
  interval: "30 sec"  # Space not allowed
```

**Correct**:
```yaml
analytics:
  interval: "30s"
```

---

## 📞 Support

If you encounter configuration issues:

1. **Validate YAML syntax**: Use [YAML Lint](https://www.yamllint.com/)
2. **Check logs**: `journalctl -u proxicloud-api -n 100`
3. **Test API token**: See troubleshooting in [INSTALLATION.md](INSTALLATION.md)
4. **Ask for help**: [GitHub Discussions](https://github.com/yourusername/proxicloud/discussions)

---

## 📚 Related Documentation

- [Installation Guide](INSTALLATION.md) - Initial setup
- [API Reference](API.md) - REST API documentation
- [Development Guide](DEVELOPMENT.md) - Building from source
- [Architecture Overview](ARCHITECTURE.md) - Technical details
