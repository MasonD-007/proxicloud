# ProxiCloud Future Plans & Roadmap

> **Document Purpose:** This document outlines the strategic roadmap for ProxiCloud development, including user stories, implementation phases, and technical considerations. Based on AWS services research and homelab best practices.

**Last Updated:** December 9, 2025

---

## 📊 Project Status Summary

### ✅ Completed Features (MVP Progress: ~40%)

**Core Infrastructure:**
- ✅ **Container Management** - Full CRUD operations for LXC containers
- ✅ **Template System** - Browse, upload, and use LXC templates
- ✅ **Analytics & Monitoring** - Time-series metrics collection and visualization
- ✅ **Caching System** - Offline fallback for Proxmox data
- ✅ **Dashboard** - Real-time overview of containers and resource usage
- ✅ **Container Lifecycle** - Start, stop, reboot, and delete operations
- ✅ **Metrics API** - RESTful endpoints for metrics data
- ✅ **Web UI** - Modern Next.js frontend with TailwindCSS
- ✅ **Volume Management** - EBS-like persistent block storage with snapshots

**API Endpoints Implemented:**
- `GET /api/health` - Health check
- `GET /api/dashboard` - Dashboard statistics
- `GET /api/containers` - List all containers
- `POST /api/containers` - Create new container
- `GET /api/containers/{vmid}` - Get container details
- `DELETE /api/containers/{vmid}` - Delete container
- `POST /api/containers/{vmid}/start` - Start container
- `POST /api/containers/{vmid}/stop` - Stop container
- `POST /api/containers/{vmid}/reboot` - Reboot container
- `GET /api/templates` - List templates
- `POST /api/templates/upload` - Upload template
- `GET /api/analytics/stats` - Analytics statistics
- `GET /api/containers/{vmid}/metrics` - Container metrics
- `GET /api/containers/{vmid}/metrics/summary` - Metrics summary
- `GET /api/volumes` - List all volumes
- `POST /api/volumes` - Create volume
- `GET /api/volumes/{volid}` - Get volume details
- `DELETE /api/volumes/{volid}` - Delete volume
- `POST /api/volumes/{volid}/attach/{vmid}` - Attach volume
- `POST /api/volumes/{volid}/detach/{vmid}` - Detach volume
- `GET /api/volumes/{volid}/snapshots` - List snapshots
- `POST /api/volumes/{volid}/snapshots` - Create snapshot
- `POST /api/volumes/{volid}/snapshots/restore` - Restore snapshot
- `POST /api/volumes/{volid}/snapshots/clone` - Clone snapshot

### 🚧 In Progress
- 🚧 **One-Click App Deployment** - Template viewing complete, auto-configuration pending
- 🚧 **User Authentication** - IAM-like identity management (planned for MVP)

### ⏳ Planned
- ⏳ **Object Storage** - S3-like API (Mid-Tier release)
- ⏳ **Advanced Networking** - VPC-like networks and security groups (Mid-Tier release)
- ⏳ **Serverless Functions** - Lambda-like FaaS (Advanced release)
- ⏳ **Secrets Management** - Secure credential storage (Advanced release)

---

## Table of Contents

1. [Vision & Goals](#vision--goals)
2. [Release Overview](#release-overview)
3. [User Stories by Release](#user-stories-by-release)
4. [Implementation Phases](#implementation-phases)
5. [Technical Considerations](#technical-considerations)
6. [Success Metrics](#success-metrics)

---

## Vision & Goals

ProxiCloud aims to bring AWS-like cloud services to self-hosted Proxmox environments, enabling homelab users to:

- **Simplify infrastructure management** with intuitive UI/UX
- **Leverage familiar AWS patterns** (EC2, S3, IAM, etc.) on local hardware
- **Reduce complexity** of Proxmox while maintaining power-user capabilities
- **Enable Infrastructure-as-Code** for homelab environments
- **Provide enterprise-grade features** (monitoring, backups, security) at homelab scale

---

## Release Overview

```
┌─────────────────────────────────────────────────────────────────┐
│ Release 1: MVP (Immediate)                                      │
│ Core compute, storage, and identity management                  │
│ Timeline: Weeks 1-4                                             │
└─────────────────────────────────────────────────────────────────┘
                            ↓
┌─────────────────────────────────────────────────────────────────┐
│ Release 2: Mid-Tier (1-2 months)                                │
│ Object storage, monitoring, templates, automation               │
│ Timeline: Weeks 5-12                                            │
└─────────────────────────────────────────────────────────────────┘
                            ↓
┌─────────────────────────────────────────────────────────────────┐
│ Release 3: Advanced (3-6 months)                                │
│ Serverless, SDN, enhanced observability                         │
│ Timeline: Weeks 13-24                                           │
└─────────────────────────────────────────────────────────────────┘
```

---

## User Stories by Release

### Release 1: MVP - Core Compute & Storage

#### Epic 1: EC2-like Compute Management ✅ (DONE)

**US-001: Launch Instances from Templates** ✅ (DONE)
> **As a homelab user**, I want to launch VMs and LXC containers from templates so that I can quickly deploy computing resources without manual Proxmox configuration.

**Acceptance Criteria:**
- ✅ Can browse available templates/ISOs in a catalog
- ✅ Can specify vCPU count and RAM allocation
- ✅ Can select network bridge and storage pool
- ✅ Instance launches successfully and appears in instance list
- ✅ Status updates in real-time (creating → running)

**Priority:** P0 (Critical)
**Status:** COMPLETED ✅

---

**US-002: View Instance Dashboard** ✅ (DONE)
> **As a homelab user**, I want to view all my running instances in a dashboard so that I can monitor my infrastructure at a glance.

**Acceptance Criteria:**
- ✅ Table displays: Name, Type (VM/LXC), Status, IP Address, CPU/RAM specs
- ✅ Can filter by status (running/stopped/all)
- ✅ Can search by name
- ✅ Shows current resource utilization
- ✅ Auto-refreshes every 5 seconds

**Priority:** P0 (Critical)
**Status:** COMPLETED ✅

---

**US-003: Manage Instance Lifecycle** ✅ (DONE)
> **As a homelab user**, I want to start/stop/terminate instances so that I can manage resource usage and costs.

**Acceptance Criteria:**
- ✅ Start button powers on stopped instances
- ✅ Stop button gracefully shuts down running instances
- ✅ Terminate button permanently deletes instances (with confirmation)
- ✅ Actions work reliably with proper error handling
- ✅ Status updates reflect changes immediately

**Priority:** P0 (Critical)
**Status:** COMPLETED ✅

---

**US-004: SSH Key Management**
> **As a homelab user**, I want to SSH into my instances using key pairs so that I can securely access them.

**Acceptance Criteria:**
- Can generate new SSH key pairs in UI
- Can upload existing public keys
- Keys automatically inject into new instances via cloud-init
- Can associate multiple keys with an instance
- Private keys download securely (one-time only)

**Priority:** P1 (High)

---

#### Epic 2: EBS-like Block Storage ✅ (DONE)

**US-005: Create Persistent Volumes** ✅ (DONE)
> **As a homelab user**, I want to create persistent storage volumes so that my data survives instance termination.

**Acceptance Criteria:**
- ✅ Can create ZFS zvol of specified size (1GB - 10TB)
- ✅ Can choose volume type (SSD/HDD pool)
- ✅ Volume appears in volumes list with status
- ✅ Can view volume details (size, type, attached instance)

**Priority:** P0 (Critical)
**Status:** COMPLETED ✅

---

**US-006: Attach/Detach Volumes** ✅ (DONE)
> **As a homelab user**, I want to attach/detach volumes to running instances so that I can move data between workloads.

**Acceptance Criteria:**
- ✅ Can attach unattached volume to running instance
- ✅ Volume appears as mount point in container (e.g., /mnt/data)
- ✅ Can detach volume from running container
- ✅ Detach operation is safe (requires manual unmount first)
- ✅ Attachment state persists across instance restarts

**Priority:** P0 (Critical)
**Status:** COMPLETED ✅

---

**US-007: Snapshot Volumes** ✅ (DONE)
> **As a homelab user**, I want to snapshot volumes for backup so that I can recover from mistakes.

**Acceptance Criteria:**
- ✅ Can create snapshot of any volume
- ✅ Can name snapshots with descriptions
- ✅ Can view snapshot list with timestamps
- ✅ Can restore volume from snapshot
- ✅ Can create new volume from snapshot (clone)

**Priority:** P1 (High)
**Status:** COMPLETED ✅

---

#### Epic 3: IAM-like Identity Management

**US-008: Multi-User Management**
> **As an administrator**, I want to create user accounts with different permissions so that I can safely share my homelab.

**Acceptance Criteria:**
- Can create users with username/password
- Can assign roles: Admin, Operator, ReadOnly
- Roles have predefined permissions (CRUD on resources)
- Can disable/delete user accounts
- Admin can reset user passwords

**Priority:** P1 (High)

---

**US-009: User Resource Isolation**
> **As a user**, I want to log in and see only my own resources so that my instances are isolated from other users.

**Acceptance Criteria:**
- JWT/session-based authentication
- Users see only resources they own
- Cannot view or modify other users' resources
- Admin role can view all resources
- Proper 401/403 error handling

**Priority:** P1 (High)

---

### Release 2: Mid-Tier - Enhanced Services

#### Epic 4: S3-like Object Storage

**US-010: Create Storage Buckets**
> **As a homelab user**, I want to create storage buckets so that I can store files via API like S3.

**Acceptance Criteria:**
- Can create named buckets (DNS-compliant names)
- Buckets map to ZFS datasets on backend
- Can list all owned buckets
- Can delete empty buckets
- Bucket names are globally unique per system

**Priority:** P2 (Medium)

---

**US-011: Object Upload/Download API**
> **As a developer**, I want to upload/download objects via REST API so that my applications can use object storage.

**Acceptance Criteria:**
- PUT /api/v1/buckets/{bucket}/objects/{key} uploads file
- GET /api/v1/buckets/{bucket}/objects/{key} downloads file
- Supports multipart uploads for large files (>100MB)
- Returns proper Content-Type headers
- Calculates and verifies MD5 checksums

**Priority:** P2 (Medium)

---

**US-012: Web-Based Bucket Browser**
> **As a homelab user**, I want to browse bucket contents in a web UI so that I can manage files visually.

**Acceptance Criteria:**
- AWS S3 console-like interface
- Can navigate folder structure (prefixes)
- Drag-and-drop file upload
- Download/delete actions per object
- Shows object metadata (size, modified date, content-type)

**Priority:** P2 (Medium)

---

#### Epic 5: VPC-lite Networking

**US-013: Isolated Networks**
> **As a homelab user**, I want to create isolated networks so that I can separate production and development workloads.

**Acceptance Criteria:**
- Can create named networks (bridges/VLANs)
- Can specify IP subnet (CIDR notation)
- Can assign instances to specific networks
- Instances on different networks cannot communicate by default
- Can attach instance to multiple networks

**Priority:** P2 (Medium)

---

**US-014: Security Groups & Firewall Rules**
> **As a security-conscious user**, I want to define firewall rules between networks so that I can control traffic flow.

**Acceptance Criteria:**
- Can create security groups with rules
- Rules specify: protocol, port range, source CIDR
- Can attach security groups to instances
- Rules enforce via iptables on Proxmox host
- Can allow inter-network communication selectively

**Priority:** P2 (Medium)

---

#### Epic 6: CloudWatch-like Monitoring ✅ (DONE)

**US-015: Instance Metrics Dashboard** ✅ (DONE)
> **As a homelab user**, I want to see CPU/memory/disk graphs for my instances so that I can identify performance issues.

**Acceptance Criteria:**
- ✅ Dashboard shows CPU%, memory%, disk I/O, network I/O
- ✅ Time-series graphs with selectable periods (5min, 1hr, 1day)
- ✅ Real-time updates (30-second refresh)
- ✅ Can view metrics for individual instances
- ✅ Host-level metrics also visible

**Priority:** P2 (Medium)
**Status:** COMPLETED ✅

---

**US-016: Audit Log Viewer**
> **As an administrator**, I want to view audit logs of all API actions so that I can track who did what.

**Acceptance Criteria:**
- Logs capture: timestamp, user, action, resource, result
- Searchable by user, action type, date range
- Filterable table with pagination
- Export logs as CSV
- Logs are append-only (tamper-evident)

**Priority:** P2 (Medium)

---

#### Epic 7: Template Library & Quick Deploy ✅ (PARTIALLY DONE)

**US-017: One-Click App Deployment** 🚧 (IN PROGRESS)
> **As a homelab user**, I want to deploy popular apps (Jellyfin, Nextcloud, Vaultwarden) with one click so that I don't need to manually configure them.

**Acceptance Criteria:**
- ✅ App catalog page lists available templates
- ✅ Each template has description, version, resource requirements
- ✅ "Launch" button creates instance with pre-configured settings
- ⏳ Apps auto-configure (database, ports, volumes)
- ⏳ Template library is extensible (can add custom templates)

**Priority:** P2 (Medium)
**Status:** IN PROGRESS 🚧 (Template viewing and uploading complete, one-click deployment pending)

**Suggested Apps:**
- Media: Jellyfin, Plex
- Storage: Nextcloud, Syncthing
- Security: Vaultwarden, Authentik
- Monitoring: Uptime Kuma, Netdata
- Development: GitLab, VS Code Server
- Home Automation: Home Assistant, Node-RED

---

**US-018: Infrastructure as Code (YAML)**
> **As a power user**, I want to define infrastructure as YAML so that I can version-control my homelab setup.

**Acceptance Criteria:**
- Can write YAML spec defining instances, volumes, networks
- `proxicloud apply -f infra.yaml` creates resources
- `proxicloud destroy -f infra.yaml` removes resources
- Validates YAML before applying
- Supports templating/variables
- Idempotent operations (re-applying is safe)

**Priority:** P2 (Medium)

**Example YAML:**
```yaml
instances:
  - name: web-server
    type: lxc
    template: ubuntu-22.04
    vcpu: 2
    memory: 2048
    volumes:
      - size: 20GB
        mount: /var/www
    networks:
      - production

networks:
  - name: production
    subnet: 10.0.1.0/24
```

---

#### Epic 8: Console Access

**US-019: Browser-Based Console**
> **As a homelab user**, I want to access instance consoles in my browser so that I don't need VNC clients.

**Acceptance Criteria:**
- Console button on instance details page
- Opens noVNC viewer in new tab/modal
- Works for VMs (VNC/SPICE) and LXC (terminal)
- Keyboard input works correctly
- Can copy/paste text into console

**Priority:** P2 (Medium)

---

#### Epic 9: Automated Backups

**US-020: Scheduled Backup Policies**
> **As a homelab user**, I want to schedule automatic backups with retention policies so that I don't lose data.

**Acceptance Criteria:**
- Can create backup schedule (daily/weekly/monthly)
- Can specify retention (keep last N backups)
- Backups run automatically via cron
- Old backups auto-delete per retention policy
- Can manually trigger backup anytime
- Email/notification on backup failure

**Priority:** P2 (Medium)

---

### Release 3: Advanced - Cloud Platform Features

#### Epic 10: Lambda-like Serverless Functions

**US-021: Deploy Code Functions**
> **As a developer**, I want to deploy code functions that run on-demand so that I can build event-driven applications.

**Acceptance Criteria:**
- Can upload code ZIP (Python, Node.js, Go)
- Can specify runtime, handler, memory, timeout
- Functions execute in isolated containers
- Logs available per invocation
- Can update function code/config

**Priority:** P3 (Low)

---

**US-022: Function Triggers**
> **As a developer**, I want functions to be triggered by HTTP requests or schedules so that I can automate workflows.

**Acceptance Criteria:**
- Each function gets unique HTTP endpoint
- Can configure cron schedule (e.g., `0 2 * * *`)
- Can pass JSON payload to invocations
- Can view invocation history and logs
- Support for environment variables

**Priority:** P3 (Low)

---

#### Epic 11: Advanced Networking (SDN)

**US-023: Custom Route Tables**
> **As a network engineer**, I want to define custom route tables and NAT rules so that I can build complex network topologies.

**Acceptance Criteria:**
- Can create route tables with custom routes
- Can associate route table with networks
- Can deploy NAT gateway VM
- NAT VM forwards traffic from private to public subnet
- Route changes take effect immediately

**Priority:** P3 (Low)

---

**US-024: VPC Peering**
> **As a homelab user**, I want VPC peering so that isolated networks can communicate securely.

**Acceptance Criteria:**
- Can peer two isolated networks
- Traffic flows bidirectionally
- Can specify allowed CIDR ranges
- Peering shows in network topology view
- Can delete peering connection

**Priority:** P3 (Low)

---

#### Epic 12: Enhanced Observability

**US-025: Prometheus/Grafana Integration**
> **As a DevOps user**, I want Prometheus/Grafana integration so that I can create custom dashboards and alerts.

**Acceptance Criteria:**
- Metrics exported at /metrics endpoint
- Prometheus can scrape ProxiCloud metrics
- Pre-built Grafana dashboards available
- Can create custom alerts (CPU > 80%, disk full, etc.)
- Alert notifications (email, webhook)

**Priority:** P3 (Low)

---

**US-026: Health Map Dashboard**
> **As a homelab user**, I want a single-page health map so that I can see all services status instantly.

**Acceptance Criteria:**
- Visual topology showing host + all instances
- Color-coded status: green (healthy), yellow (warning), red (critical)
- Shows CPU/memory usage per instance
- Click instance to view details
- Auto-refreshes every 10 seconds

**Priority:** P3 (Low)

---

#### Epic 13: Secrets Management

**US-027: Secure Vault for Credentials**
> **As a developer**, I want a secure vault for API keys and passwords so that I don't hardcode credentials.

**Acceptance Criteria:**
- Can store key-value secrets
- Secrets encrypted at rest
- REST API to create/read/delete secrets
- Can inject secrets as environment variables into instances
- Role-based access to secrets
- Audit log for secret access

**Priority:** P3 (Low)

---

## Implementation Phases

### Phase 1: MVP Foundation (Weeks 1-4) 🚧 IN PROGRESS

**Goal:** Launch core compute and storage functionality with basic IAM.

**Overall Progress:** ~50% Complete ✅

#### Backend Tasks

**1. Database Setup** ⏳ PENDING
- Create PostgreSQL/SQLite schema for core entities
- Tables: `instances`, `volumes`, `users`, `roles`, `user_roles`, `keypairs`, `security_groups`
- Implement migration system (golang-migrate or similar)
- Seed initial admin user

**2. Proxmox API Integration** ✅ DONE
- ✅ Build Go client wrapper for Proxmox REST API
- ✅ Authentication: Store API tokens securely
- ✅ Implement VM lifecycle methods:
  - ✅ `CreateVM(config VMConfig) (vmid int, error)`
  - ✅ `StartVM(vmid int) error`
  - ✅ `StopVM(vmid int) error`
  - ✅ `DeleteVM(vmid int) error`
- ✅ Implement LXC lifecycle methods (similar)
- ✅ Handle ZFS volume operations:
  - ✅ `CreateVolume(size int, pool string) (path string, error)`
  - ✅ `AttachVolume(vmid int, volumePath string) error`
  - ✅ `DetachVolume(vmid int, volumePath string) error`
  - ✅ `SnapshotVolume(volumePath string, name string) error`

**3. Core REST API (Go)** ✅ DONE
- ✅ Set up HTTP server with router (Gorilla Mux)
- ✅ Implement endpoints for instances, containers, templates
- ✅ Implement analytics endpoints
- ✅ Implement volume management endpoints (10 endpoints)
- ⏳ Implement authentication endpoints (pending)
- ⏳ Implement user/role management endpoints (pending)

**4. Authentication & Authorization** ⏳ PENDING
- JWT token generation with 24hr expiry
- Bcrypt password hashing
- Role-based middleware
- Per-user resource isolation in database queries

#### Frontend Tasks

**5. Core UI Pages (Next.js/React)** ✅ DONE
- ✅ Set up Next.js 14 project with App Router
- ✅ Install dependencies: TailwindCSS, Shadcn/ui components
- ✅ Create pages:
  - ✅ `/` - Dashboard with stats cards
  - ✅ `/containers` - Container table with actions
  - ✅ `/containers/create` - Container launch wizard
  - ✅ `/containers/[id]` - Container details page with volume management
  - ✅ `/templates` - Template catalog with upload
  - ✅ `/analytics` - Analytics dashboard
  - ✅ `/volumes` - Volume list with filtering
  - ✅ `/volumes/create` - Volume creation wizard
  - ✅ `/volumes/[volid]` - Volume details with snapshot management
- ✅ Implement API client (fetch wrapper)
- ⏳ Set up authentication context (pending)

**6. Instance Launch Wizard** ✅ DONE
- ✅ Multi-step form for creating containers
- ✅ Template selection
- ✅ Resource configuration (CPU, RAM, storage)
- ✅ Real-time validation
- ✅ Redirect to instance details on success

#### Infrastructure

**7. Deployment Scripts** ✅ PARTIALLY DONE
- ✅ Update `deploy/install.sh`
- ✅ Create systemd service files
- ✅ Development scripts (`deploy/scripts/dev.sh`)
- ⏳ Database migration scripts (pending)
- ⏳ Seed data scripts (pending)

**Phase 1 Deliverables:**
- ✅ Functional API server responding to requests
- ✅ Working Next.js frontend
- ✅ Users can create/start/stop/terminate instances
- ✅ Template browsing and uploading
- ✅ Analytics and metrics collection
- ✅ Users can create/attach/detach volumes
- ✅ Volume snapshot management (create, restore, clone)
- ✅ Full volume lifecycle management UI
- ⏳ Basic role-based access control (pending)

---

### Phase 2: Mid-Tier Enhancements (Weeks 5-12)

**Goal:** Add object storage, monitoring, templates, and automation features.

#### Backend Tasks

**8. Object Storage Service**
- Create `/api/v1/buckets` and `/api/v1/buckets/{bucket}/objects` endpoints
- Implement multipart upload support (RFC 7233)
- ZFS dataset creation for buckets:
  - `zfs create pool/buckets/{bucket-name}`
- Store object metadata in database:
  - Table: `objects (id, bucket_id, key, size_bytes, content_type, md5sum, file_path)`
- Generate pre-signed URLs for temporary access
- Implement S3-compatible headers (ETag, Content-MD5)

**9. Monitoring Service**
- Poll Proxmox RRD stats every 30 seconds
- Store time-series data (use PostgreSQL with TimescaleDB extension, or InfluxDB)
- Collect metrics: CPU%, memory%, disk I/O, network I/O
- Implement `/api/v1/metrics/instances/{id}?period=1h` endpoint
- Return JSON: `{timestamps: [], cpu: [], memory: []}`

**10. Audit Logging**
- Middleware to log all API calls:
  - Capture: timestamp, user_id, method, path, status_code, response_time
- Store in `audit_logs` table
- Implement `/api/v1/logs/audit?user=X&action=Y&from=Z` endpoint
- Add log retention policy (delete logs older than 90 days)

**11. Networking Service**
- Implement `/api/v1/networks` endpoints
- Create Proxmox bridges via API:
  - `POST /nodes/{node}/network` with bridge config
- VLAN support: create VLAN-aware bridges
- Security groups:
  - Store rules in `security_groups` table
  - Apply iptables rules on instance start
  - Example rule: `ACCEPT tcp 0.0.0.0/0 port 22`

#### Frontend Tasks

**12. Storage UI**
- Buckets page: `/storage/buckets`
  - List buckets with size, object count
  - "Create Bucket" button
- Bucket detail page: `/storage/buckets/{name}`
  - Object table with columns: Key, Size, Modified
  - Folder navigation (prefix-based)
  - Drag-and-drop upload zone
  - Multi-select delete
- Upload component with progress bar

**13. Monitoring Dashboard**
- Dashboard page: `/monitoring`
  - Grid layout with metric cards
  - CPU/Memory/Disk graphs using Chart.js or Recharts
  - Period selector (5min, 1hr, 1day)
  - Real-time updates (WebSocket or polling)
- Audit logs page: `/monitoring/logs`
  - Filterable table with search
  - Export CSV button

**14. Template Library**
- Templates page: `/templates`
  - Card grid showing available apps
  - Each card: logo, name, description, version
  - "Deploy" button opens config modal
- Backend: Store templates in `templates` table or JSON config file
- Template format:
```json
{
  "id": "jellyfin",
  "name": "Jellyfin Media Server",
  "type": "lxc",
  "base_template": "debian-12",
  "vcpu": 2,
  "memory": 2048,
  "scripts": ["setup-jellyfin.sh"],
  "ports": [8096, 8920],
  "volumes": [
    {"mount": "/media", "size": 100}
  ]
}
```

**15. Console Viewer**
- Add "Console" button to instance details page
- Open modal/new tab with noVNC viewer
- Backend: Proxy noVNC requests to Proxmox:
  - `GET /api/v1/instances/{id}/console` returns noVNC URL
- Use Proxmox's noVNC endpoint: `/nodes/{node}/qemu/{vmid}/vncproxy`

#### DevOps

**16. YAML Deployer**
- Create CLI tool: `proxicloud` (Go cobra CLI)
- Commands:
  - `proxicloud apply -f infra.yaml` - Create resources
  - `proxicloud destroy -f infra.yaml` - Delete resources
  - `proxicloud validate -f infra.yaml` - Check syntax
- YAML parser (use `gopkg.in/yaml.v3`)
- Implement dependency resolution (create networks before instances)
- Store deployment state (track what was created)

**17. Backup Automation**
- Implement backup scheduler (cron-based)
- Backend service polls database for backup policies
- Create policy model:
```go
type BackupPolicy struct {
  ID         string
  ResourceID string
  Schedule   string // cron expression
  Retention  int    // keep last N backups
  Enabled    bool
}
```
- Execute ZFS snapshots at scheduled times
- Delete old snapshots per retention policy
- UI: Backup policies page with create/edit/delete

**Deliverables:**
- S3-like object storage with web UI
- Real-time monitoring dashboards
- Audit log viewer
- Template library with 5+ pre-built apps
- Browser-based console access
- YAML-based infrastructure deployment
- Automated backup scheduling

---

### Phase 3: Advanced Features (Weeks 13-24)

**Goal:** Implement serverless functions, advanced networking, and enhanced observability.

#### Backend Tasks

**18. Serverless Functions (FaaS)**
- Design function execution architecture:
  - Use ephemeral LXC containers for isolation
  - Pre-built runtime containers (Python, Node.js, Go)
- Implement `/api/v1/functions` CRUD endpoints
- Function invocation:
  - `POST /api/v1/functions/{id}/invoke` with JSON payload
  - Spawn container, mount code, execute handler
  - Capture stdout/stderr as logs
  - Return JSON response
- Resource limits: CPU quota, memory limit, timeout
- Event triggers:
  - HTTP: Create API Gateway-like routing
  - Cron: Schedule function invocations

**19. Advanced Networking**
- Route table management:
  - Store routes in `route_tables` and `routes` tables
  - Apply routes using `ip route` commands
- NAT gateway:
  - Deploy special VM with IP forwarding enabled
  - Configure iptables MASQUERADE rules
- VPC peering:
  - Create bridge connections between isolated networks
  - Store peering relationships in database

**20. Secrets Manager**
- Implement encryption at rest (AES-256)
- Store encrypted secrets in `secrets` table
- Master key management (store in /etc/proxicloud/master.key)
- API endpoints:
```
GET    /api/v1/secrets           # List secret names (not values)
POST   /api/v1/secrets           # Create secret
GET    /api/v1/secrets/{name}    # Get secret value
DELETE /api/v1/secrets/{name}    # Delete secret
```
- Instance integration: Inject secrets as env vars on start

#### Frontend Tasks

**21. Functions UI**
- Functions page: `/functions`
  - List functions with name, runtime, last invoked
  - Create function form (upload ZIP, select runtime)
- Function editor: `/functions/{id}/edit`
  - Monaco editor for inline code editing
  - "Test Invoke" button with JSON payload input
  - Logs viewer below editor
- Metrics: Invocation count, success rate, avg duration

**22. Health Map Dashboard**
- Visual topology page: `/monitoring/health`
- Use D3.js or React Flow for graph visualization
- Nodes: host, instances, volumes, networks
- Edges: attachments, network connections
- Color coding:
  - Green: healthy (CPU < 70%, reachable)
  - Yellow: warning (CPU 70-90%, high memory)
  - Red: critical (CPU > 90%, unreachable)
- Click node for details panel

**23. Network Designer**
- Visual network builder: `/networks/designer`
- Drag-drop interface to create networks and subnets
- Route table editor with visual connections
- Firewall rule builder with port/protocol selectors
- Export as YAML for IaC

#### Integration

**24. Prometheus/Grafana**
- Implement Prometheus exporter:
  - Endpoint: `/metrics` (Prometheus text format)
  - Export instance metrics, API request counts, error rates
- Create Grafana provisioning config:
  - Auto-provision ProxiCloud dashboard
  - Pre-built panels: CPU, memory, API latency
- Alert rules in Prometheus:
  - `instance_cpu_high` - CPU > 80% for 5 minutes
  - `instance_unreachable` - No metrics for 2 minutes

**25. Local Package Registry (Optional)**
- Deploy Docker Registry v2
- UI for browsing images: `/registry`
- Integration with instance creation (use private registry images)
- Garbage collection for unused images

**Deliverables:**
- Serverless function platform with HTTP/cron triggers
- Advanced networking with custom routes and NAT
- Secrets management service
- Interactive health map dashboard
- Prometheus/Grafana integration
- Network designer UI

---

## Technical Considerations

### Required Technologies

**Backend Stack:**
- **Language:** Go 1.21+
- **Web Framework:** Chi router or Gorilla Mux
- **Database:** PostgreSQL 14+ (production) or SQLite (development)
- **ORM:** GORM or sqlc for type-safe queries
- **Authentication:** JWT with `golang-jwt/jwt`
- **Testing:** Standard library `testing` + `testify` assertions

**Frontend Stack:**
- **Framework:** Next.js 14+ (App Router)
- **Language:** TypeScript 5+
- **UI Library:** React 18+
- **Styling:** TailwindCSS 3+
- **Components:** Shadcn/ui or Headless UI
- **Charts:** Chart.js or Recharts
- **State Management:** React Context or Zustand

**Infrastructure:**
- **Proxmox:** VE 8.0+
- **Storage:** ZFS on Proxmox host
- **Networking:** Linux bridges, VLANs
- **Monitoring:** Prometheus + Grafana (Phase 3)
- **Container Runtime:** LXC (for instances and functions)

**Development Tools:**
- **Version Control:** Git
- **CI/CD:** GitHub Actions (build, test, lint)
- **Database Migrations:** golang-migrate
- **API Documentation:** OpenAPI/Swagger
- **Linting:** golangci-lint, ESLint

---

### Architecture Decisions

**1. Single-Node Design**
- ProxiCloud is designed for single Proxmox nodes
- No distributed consensus (no etcd, Consul, etc.)
- Simplifies deployment and reduces complexity
- Trade-off: No high availability or automatic failover

**2. Database Choice**
- PostgreSQL for production (robust, feature-rich)
- SQLite for development/testing (zero-config)
- Schema designed to support both via GORM

**3. API Design Philosophy**
- RESTful conventions (resources, HTTP verbs)
- Consistent error responses (JSON with error codes)
- Pagination for list endpoints (limit, offset)
- Filtering via query parameters

**4. Security Model**
- JWT tokens with short expiry (24hr)
- Bcrypt for password hashing (cost factor 12)
- Role-based access control (RBAC)
- Per-user resource isolation at database level
- Audit logging for compliance

**5. Storage Strategy**
- ZFS for all storage (instances, volumes, backups)
- Leverage ZFS features: snapshots, compression, dedup
- Object storage uses filesystem (ZFS datasets)
- No external storage systems (Ceph, GlusterFS)

---

### Key Risks & Mitigations

**Risk 1: Proxmox API Limitations**
- Some features may require shell access (not available via API)
- **Mitigation:** Use SSH connection for complex operations; wrap in Go functions

**Risk 2: Single-Node Constraints**
- No true HA or multi-AZ support
- **Mitigation:** Focus on backup/restore; clear documentation of limitations

**Risk 3: Performance with Large Files**
- S3-lite uploads may be slow for files >1GB
- **Mitigation:** Implement multipart uploads; use chunked transfer encoding

**Risk 4: Security of Serverless Functions**
- User code execution is inherently risky
- **Mitigation:** 
  - Run in unprivileged LXC containers
  - Apply strict resource limits (CPU, memory, timeout)
  - Network isolation (no internet access by default)
  - Audit all function code uploads

**Risk 5: Database Growth**
- Metrics and logs can grow indefinitely
- **Mitigation:** Implement retention policies; use partitioning; archive old data

**Risk 6: Proxmox Version Compatibility**
- API changes between Proxmox versions
- **Mitigation:** Document supported versions; version detection in code

---

### Testing Strategy

**Unit Tests**
- All Go handlers and services
- Target: 80%+ code coverage
- Mock Proxmox API client for tests
- Use table-driven tests for edge cases

**Integration Tests**
- Test API endpoints against real database
- Use testcontainers for PostgreSQL
- Verify request/response cycles
- Test authentication flows

**E2E Tests**
- Critical user workflows:
  1. Launch instance → attach volume → start instance
  2. Create bucket → upload object → download object
  3. Create user → login → view resources (isolation)
- Use Playwright or Cypress for frontend
- Run against staging environment

**Load Tests**
- Simulate multiple concurrent users
- Test API rate limits
- Identify bottlenecks (database queries, Proxmox API calls)
- Tools: k6 or Apache Bench

**Security Tests**
- Penetration testing for auth bypass
- SQL injection attempts
- CSRF protection verification
- Secrets leakage in logs/errors

---

### Performance Targets

**API Response Times:**
- List endpoints: < 200ms (p95)
- Create/update operations: < 1s (p95)
- Delete operations: < 500ms (p95)
- Metrics queries: < 300ms (p95)

**UI Responsiveness:**
- Time to Interactive (TTI): < 2s
- First Contentful Paint (FCP): < 1s
- Page transitions: < 100ms

**Throughput:**
- API: 1000 req/sec (single instance)
- Object storage uploads: 100 MB/s (limited by disk)
- Concurrent VM launches: 10 (limited by Proxmox)

**Resource Utilization:**
- Backend: < 512MB RAM, < 5% CPU (idle)
- Frontend: < 100MB RAM, < 2% CPU (idle)
- Database: < 1GB storage for 100 instances

---

## Success Metrics

### MVP Success Criteria
- [ ] 10 beta users successfully deploy ProxiCloud
- [x] ✅ Users can launch 5+ instance types without errors
- [ ] Zero critical security vulnerabilities
- [x] ✅ Documentation covers core features
- [x] ✅ Average API response time < 500ms
- [x] ✅ Analytics system collecting metrics (30-second intervals)
- [x] ✅ Caching system operational for offline fallback
- [x] ✅ Template upload and management working
- [x] ✅ Volume management fully operational (create, attach, detach, snapshot)
- [x] ✅ All volume endpoints documented in API.md
- [x] ✅ VOLUME_MANAGEMENT.md user guide created

### Mid-Tier Success Criteria
- [ ] 50+ active users (monthly)
- [ ] 10+ templates in library
- [ ] Users deploy 100+ instances via templates
- [ ] Object storage handles 1TB+ total data
- [ ] Monitoring captures 99%+ uptime

### Advanced Success Criteria
- [ ] 100+ active users (monthly)
- [ ] Users deploy 10+ serverless functions
- [ ] Prometheus integration used by 50% of users
- [ ] Community contributes 5+ custom templates
- [ ] Project reaches 500+ GitHub stars

### User Satisfaction Metrics
- **Net Promoter Score (NPS):** > 50
- **Feature Adoption Rate:** > 70% use core features
- **Support Ticket Volume:** < 10/month
- **Documentation Quality:** > 4.5/5 user rating

---

## Appendix

### AWS to ProxiCloud Feature Mapping

| AWS Service | ProxiCloud Equivalent | Status | Release |
|-------------|----------------------|--------|---------|
| EC2 | Instances (VM/LXC) | ✅ DONE | MVP |
| EBS | Volumes (ZFS zvols) | ✅ DONE | MVP |
| S3 | Object Storage (ZFS datasets) | ⏳ Planned | Mid-Tier |
| IAM | Users & Roles | ⏳ Planned | MVP |
| VPC | Networks (Bridges/VLANs) | ⏳ Planned | Mid-Tier |
| Lambda | Serverless Functions | ⏳ Planned | Advanced |
| CloudWatch | Monitoring Dashboard | ✅ DONE | Mid-Tier |
| CloudTrail | Audit Logs | ⏳ Planned | Mid-Tier |
| Secrets Manager | Secrets Vault | ⏳ Planned | Advanced |
| CloudFormation | YAML Deployer | ⏳ Planned | Mid-Tier |
| Route 53 | DNS Management | ❌ Not Planned | - |
| RDS | Managed Databases | ❌ Not Planned | - |
| EKS | Kubernetes | ❌ Not Planned | - |

### Database Schema Overview

**Core Entities:**
```sql
-- Users and authentication
users (id, username, password_hash, full_name, created_at)
roles (id, name, permissions)
user_roles (user_id, role_id)

-- Compute resources
instances (id, name, type, vcpu, memory_mb, status, user_id, created_at)
volumes (id, size_gb, type, attached_to, user_id, created_at)
snapshots (id, volume_id, name, description, created_at)

-- Storage
buckets (id, name, user_id, created_at)
objects (id, bucket_id, key, size_bytes, content_type, md5sum, file_path)

-- Networking
networks (id, name, subnet_cidr, user_id, created_at)
security_groups (id, name, rules, user_id)
instance_networks (instance_id, network_id, ip_address)

-- Monitoring & Logging
metrics (timestamp, instance_id, metric_type, value)
audit_logs (id, timestamp, user_id, action, resource_type, resource_id, details)

-- Advanced features
functions (id, name, runtime, handler, code_path, user_id, created_at)
function_invocations (id, function_id, timestamp, duration_ms, success, logs)
secrets (id, name, encrypted_value, user_id, created_at)
backup_policies (id, resource_id, schedule, retention, enabled)
```

### API Endpoint Reference

See full API documentation at `/docs/API.md` (to be created).

Quick reference:
- **Auth:** `/api/v1/auth/*`
- **Instances:** `/api/v1/instances/*`
- **Volumes:** `/api/v1/volumes/*`
- **Storage:** `/api/v1/buckets/*`
- **Monitoring:** `/api/v1/metrics/*`, `/api/v1/logs/*`
- **Functions:** `/api/v1/functions/*`
- **Secrets:** `/api/v1/secrets/*`

---

## Document History

| Version | Date | Author | Changes |
|---------|------|--------|---------|
| 1.0 | 2025-12-08 | System | Initial comprehensive roadmap |
| 1.1 | 2025-12-09 | System | Updated with completion status for implemented features:<br>- ✅ Container management (CRUD operations)<br>- ✅ Template system (browse, upload)<br>- ✅ Analytics & monitoring dashboard<br>- ✅ Caching system<br>- ✅ Core REST API endpoints<br>- ✅ Modern Next.js frontend<br>- 🚧 Phase 1 MVP ~40% complete |
| 1.2 | 2025-12-09 | System | ✅ **Epic 2: EBS-like Block Storage - COMPLETED**<br>- Implemented full volume management system<br>- 17/17 tasks complete (100%)<br>- Backend: Types, Proxmox client, handlers, cache, routes<br>- Frontend: Types, API client, pages (list, create, details)<br>- Features: Create, attach, detach, snapshot, restore, clone volumes<br>- Documentation: API.md updated, VOLUME_MANAGEMENT.md created<br>- 🎉 Phase 1 MVP now ~50% complete |

---

**Questions or Feedback?**

This roadmap is a living document. If you have suggestions, questions, or want to contribute, please:
- Open an issue on GitHub
- Join our community discussions
- Submit a pull request with improvements

**Next Steps:**
1. Review and approve roadmap with stakeholders
2. Set up project tracking (GitHub Projects or Jira)
3. Begin Phase 1 implementation
4. Schedule regular roadmap reviews (monthly)
