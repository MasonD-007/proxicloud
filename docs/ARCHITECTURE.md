# Architecture Overview

This document provides a detailed technical overview of ProxiCloud's architecture, design decisions, and implementation details.

---

## 📐 System Architecture

### High-Level Architecture

```
┌──────────────────────────────────────────────────────────────┐
│                     User Browser (LAN)                       │
│                   http://proxmox-ip:3000                     │
└─────────────────────────┬────────────────────────────────────┘
                          │ HTTP/WebSocket
                          │
┌─────────────────────────▼────────────────────────────────────┐
│               Next.js Frontend Server                        │
│                     (Port 3000)                              │
│                                                              │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐     │
│  │  App Router  │  │  Components  │  │  API Client  │     │
│  │   (Pages)    │  │   (React)    │  │  (lib/api)   │     │
│  └──────────────┘  └──────────────┘  └──────────────┘     │
│                                                              │
│  Server-Side Rendering + Client-Side Hydration              │
└─────────────────────────┬────────────────────────────────────┘
                          │ HTTP JSON API
                          │
┌─────────────────────────▼────────────────────────────────────┐
│                 Go Backend API Server                        │
│                     (Port 8080)                              │
│                                                              │
│  ┌─────────────────────────────────────────────────────┐   │
│  │                HTTP Server (Gorilla Mux)             │   │
│  │  ┌──────────┐  ┌──────────┐  ┌──────────┐         │   │
│  │  │  CORS    │  │  Logging │  │ Recovery │         │   │
│  │  └──────────┘  └──────────┘  └──────────┘         │   │
│  └─────────────────────────────────────────────────────┘   │
│                                                              │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐     │
│  │   Handlers   │  │   Proxmox    │  │  Analytics   │     │
│  │              │  │    Client    │  │  Collector   │     │
│  │ Containers   │  │              │  │              │     │
│  │ Analytics    │  │  LXC Ops     │  │  Goroutine   │     │
│  │ Templates    │  │  Metrics     │  │  (30s loop)  │     │
│  │ Dashboard    │  │  Auth        │  │              │     │
│  └──────────────┘  └──────┬───────┘  └──────┬───────┘     │
│                            │                  │              │
│                            │                  ▼              │
│  ┌──────────────┐  ┌──────▼───────┐  ┌──────────────┐     │
│  │    Cache     │  │   Proxmox    │  │   SQLite     │     │
│  │   (Offline   │  │  VE API      │  │   Database   │     │
│  │    Mode)     │  │              │  │              │     │
│  └──────────────┘  └──────────────┘  └──────────────┘     │
└─────────────────────────┬────────────────────────────────────┘
                          │ HTTPS (TLS skip verify)
                          │
┌─────────────────────────▼────────────────────────────────────┐
│                   Proxmox VE API                             │
│                     (Port 8006)                              │
│                                                              │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐     │
│  │  LXC API     │  │  RRD API     │  │ Storage API  │     │
│  │              │  │  (Metrics)   │  │              │     │
│  │ Create       │  │              │  │ Templates    │     │
│  │ Start/Stop   │  │ CPU, RAM     │  │ Volumes      │     │
│  │ Delete       │  │ Disk, Net    │  │ Usage        │     │
│  └──────────────┘  └──────────────┘  └──────────────┘     │
└──────────────────────────────────────────────────────────────┘
```

---

## 🏗️ Component Architecture

### Backend Components

#### 1. HTTP Server Layer

**Technology**: Gorilla Mux router

**Responsibilities**:
- Route HTTP requests to handlers
- Apply middleware (CORS, logging, recovery)
- Serve RESTful API endpoints

**Key Files**:
- `internal/server/server.go` - Server initialization
- `internal/server/routes.go` - Route registration
- `internal/middleware/*.go` - Middleware implementations

#### 2. Handlers Layer

**Responsibilities**:
- Parse HTTP requests
- Validate input
- Call business logic (Proxmox client, analytics)
- Format responses
- Handle errors

**Handler Categories**:
- `containers.go` - LXC CRUD operations
- `metrics.go` - Analytics endpoints
- `dashboard.go` - Summary statistics
- `templates.go` - Template management
- `health.go` - Health checks

#### 3. Proxmox Client Layer

**Responsibilities**:
- Authenticate with Proxmox API
- Make HTTP requests to Proxmox
- Parse Proxmox responses
- Handle API errors and retries

**Key Operations**:
- `ListContainers()` - Get all containers
- `GetContainer(vmid)` - Get container details
- `CreateContainer(config)` - Create new container
- `StartContainer(vmid)` - Start container
- `StopContainer(vmid)` - Stop container
- `GetMetrics(vmid)` - Fetch RRD data

**API Authentication**:
```go
// Token format: user@realm!tokenid=secret
req.Header.Set("Authorization", "PVEAPIToken=" + token)
```

#### 4. Analytics Collector

**Responsibilities**:
- Run background goroutine
- Collect metrics from all containers
- Store in SQLite database
- Clean up old data

**Collection Flow**:
```
Every 30s:
  1. List all containers
  2. For each running container:
     a. Fetch RRD data (CPU, RAM, disk, network)
     b. Parse metrics
     c. Insert into SQLite
  3. Every hour: DELETE old data (>30 days)
```

**Metrics Collected**:
- CPU usage (percentage)
- Memory usage (bytes, percentage)
- Disk I/O (read/write bytes per second)
- Network throughput (in/out bytes per second)

#### 5. Cache Layer (Offline Mode)

**Responsibilities**:
- Cache container state
- Serve cached data when Proxmox unreachable
- Update cache on successful API calls

**Cache Strategy**:
```
On API success:
  - Update cache with fresh data
  - Set X-Cache-Status: online

On API failure:
  - Serve from cache (if available)
  - Set X-Cache-Status: offline
  - Frontend shows warning banner
```

---

### Frontend Components

#### 1. Layout Components

**TopBar** (`components/layout/TopBar.tsx`):
- Logo/branding
- Search (future)
- Node name display
- Help/settings menu

**Sidebar** (`components/layout/Sidebar.tsx`):
- Navigation menu
- Active route highlighting
- Collapsible (mobile)

**ConnectionBanner** (`components/layout/ConnectionBanner.tsx`):
- Displays when offline
- Retry button
- Polls `/api/health` every 5s

#### 2. Page Components

**Dashboard** (`app/page.tsx`):
- Summary cards (total, running, stopped)
- Resource usage (CPU, RAM, storage)
- Recent activity feed
- Quick actions

**Container List** (`app/containers/page.tsx`):
- Table/card view of containers
- Status badges
- Action buttons
- Search and filter

**Create Container** (`app/containers/create/page.tsx`):
- Form with validation
- Template selector
- Resource sliders
- Advanced options

**Analytics** (`app/analytics/page.tsx`):
- Container selector
- Time range picker
- Metric charts (Recharts)

#### 3. API Client

**Location**: `lib/api.ts`

**Features**:
- Fetch wrapper with error handling
- Automatic retry on failure
- Offline detection
- Type-safe responses (TypeScript)

**Example**:
```typescript
export async function listContainers(): Promise<Container[]> {
  const response = await fetch(`${API_URL}/containers`);
  if (!response.ok) throw new Error('Failed to fetch');
  const data = await response.json();
  return data.containers;
}
```

---

## 💾 Data Flow

### Container Creation Flow

```
User submits form
   │
   ▼
Frontend validates input
   │
   ▼
POST /api/containers
   │
   ▼
Backend handler parses request
   │
   ▼
Proxmox client formats API call
   │
   ▼
Proxmox API creates container
   │
   ▼
Backend receives task ID
   │
   ▼
Return 201 with VMID + task ID
   │
   ▼
Frontend polls container status
   │
   ▼
Show success message
```

### Metrics Collection Flow

```
Background goroutine (every 30s)
   │
   ▼
List all containers
   │
   ▼
For each running container:
   │
   ▼
Fetch RRD data (last 30s sample)
   │
   ▼
Parse metrics:
  - CPU: percentage
  - Memory: bytes used
  - Disk: read/write bytes
  - Network: in/out bytes
   │
   ▼
Insert into SQLite
   │
   ▼
Commit transaction
   │
   ▼
Wait 30s, repeat
```

### Analytics Query Flow

```
User views analytics page
   │
   ▼
GET /api/analytics/100/cpu?range=24h
   │
   ▼
Backend calculates time range
   │
   ▼
Query SQLite:
  SELECT timestamp, value
  FROM metrics
  WHERE vmid=100 AND type='cpu'
    AND timestamp > (now - 24h)
  ORDER BY timestamp ASC
   │
   ▼
Calculate statistics (avg, min, max)
   │
   ▼
Return JSON with data points
   │
   ▼
Frontend renders chart (Recharts)
```

---

## 🗄️ Database Schema

### SQLite Database

**Location**: `/var/lib/proxicloud/analytics.db`

#### Tables

**1. metrics**
```sql
CREATE TABLE metrics (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    vmid INTEGER NOT NULL,
    type TEXT NOT NULL,            -- 'cpu', 'memory', 'disk', 'network'
    value REAL NOT NULL,
    timestamp INTEGER NOT NULL
);

CREATE INDEX idx_vmid_type_ts ON metrics(vmid, type, timestamp);
CREATE INDEX idx_timestamp ON metrics(timestamp);
```

**2. containers (cache)**
```sql
CREATE TABLE containers (
    vmid INTEGER PRIMARY KEY,
    hostname TEXT NOT NULL,
    status TEXT NOT NULL,
    template TEXT,
    cores INTEGER,
    memory INTEGER,
    disk INTEGER,
    ip_address TEXT,
    created_at INTEGER,
    last_seen INTEGER NOT NULL
);
```

**3. events (activity log)**
```sql
CREATE TABLE events (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    vmid INTEGER,
    action TEXT NOT NULL,         -- 'create', 'start', 'stop', 'delete', 'reboot'
    timestamp INTEGER NOT NULL,
    details TEXT,
    user TEXT
);

CREATE INDEX idx_timestamp ON events(timestamp DESC);
```

**4. meta (application metadata)**
```sql
CREATE TABLE meta (
    key TEXT PRIMARY KEY,
    value TEXT NOT NULL,
    updated_at INTEGER NOT NULL
);

-- Stores:
-- - container_counter: next auto-generated number
-- - last_cleanup: timestamp of last retention cleanup
-- - version: database schema version
```

---

## 🔐 Security Architecture

### API Token Security

**Storage**:
- Stored in `/etc/proxicloud/config.yaml`
- File permissions: `600` (root only)
- Never logged or exposed in responses

**Usage**:
```go
req.Header.Set("Authorization", "PVEAPIToken="+token)
```

### TLS Certificate Handling

**Self-Signed Certificates**:
```go
transport := &http.Transport{
    TLSClientConfig: &tls.Config{
        InsecureSkipVerify: true,  // For self-signed Proxmox certs
    },
}
```

**Future Enhancement**: Support custom CA certificates

### Network Security

**Binding**:
- Backend: Localhost only by default (frontend is proxy)
- Frontend: `0.0.0.0` for LAN access

**Firewall**:
- Only port 3000 needs to be open
- Backend port 8080 can be firewalled (internal only)

---

## ⚡ Performance Considerations

### Backend Performance

**Concurrent Request Handling**:
- Go's goroutines handle concurrent requests
- No global locks on read operations
- Write operations use SQLite transactions

**Metrics Collection**:
- Runs in separate goroutine
- Non-blocking (doesn't affect API responses)
- Batch inserts for efficiency

**Database Optimization**:
- Indexes on frequently queried columns
- Periodic VACUUM for database compaction
- Query optimization (EXPLAIN for slow queries)

### Frontend Performance

**Server-Side Rendering**:
- Next.js pre-renders pages on server
- Reduces time to first paint

**Code Splitting**:
- Next.js automatically splits code by route
- Only load JavaScript needed for current page

**Image Optimization**:
- Next.js Image component for automatic optimization
- WebP format support

**Caching**:
- Browser caches static assets
- API responses use appropriate cache headers

---

## 🔄 Offline Mode Design

### Detection

**Frontend**:
```typescript
const checkConnection = async () => {
  const res = await fetch('/api/health');
  const cacheStatus = res.headers.get('X-Cache-Status');
  setIsOffline(cacheStatus === 'offline');
};
```

**Backend**:
```go
containers, err := client.ListContainers()
if err != nil {
    // Proxmox unreachable, serve from cache
    cached := cache.GetContainers()
    w.Header().Set("X-Cache-Status", "offline")
    json.NewEncoder(w).Encode(cached)
    return
}
```

### Cache Strategy

**What's Cached**:
- Container list and status
- Last known resource usage
- Container configuration

**Cache Expiry**:
- Updated on every successful API call
- Served when Proxmox unreachable
- Stale after 5 minutes (visual indicator)

**User Experience**:
- Warning banner: "Connection lost. Showing cached data."
- Retry button
- Disabled action buttons (can't start/stop while offline)

---

## 🎨 UI Design Patterns

### AWS-Inspired Design

**Colors**:
- Background: Dark blue-gray (`#0f1419`)
- Surface: Slightly lighter (`#16181d`)
- Text: White with varying opacity
- Accent: Sky blue (`#0ea5e9`)

**Layout**:
- Fixed top bar (64px)
- Fixed left sidebar (240px)
- Scrollable main content area
- Responsive (collapsible sidebar on mobile)

**Components**:
- Cards for grouping related content
- Tables for data lists
- Forms with inline validation
- Toast notifications for feedback

---

## 📊 Metrics System

### RRD Data Format

Proxmox stores metrics in Round-Robin Database (RRD) format.

**API Endpoint**:
```
GET /nodes/{node}/lxc/{vmid}/rrddata?timeframe=hour
```

**Response**:
```json
{
  "data": [
    {
      "time": 1701619200,
      "cpu": 0.235,
      "mem": 524288000,
      "maxmem": 1073741824,
      "disk": 2147483648,
      "maxdisk": 8589934592,
      "netin": 1048576,
      "netout": 2097152
    }
  ]
}
```

### Storage Format

**Normalized Storage**:
- Store absolute values (not percentages)
- Calculate percentages at query time
- Allows for flexible aggregations

**Time Series**:
- Timestamp stored as Unix epoch (integer)
- Indexed for fast range queries
- Automatic cleanup of old data

---

## 🚀 Scalability Considerations

### Current Limitations

- Single-node deployment
- SQLite database (not clustered)
- No horizontal scaling

### Future Enhancements

**Multi-Node Support**:
- Connect to multiple Proxmox nodes
- Aggregate data across nodes
- Node selector in UI

**Database Migration**:
- PostgreSQL support for larger deployments
- Time-series database (TimescaleDB, InfluxDB)
- Distributed storage

**Caching Layer**:
- Redis for session data
- Distributed cache for multi-node setups

---

## 🔧 Configuration Management

### Config Loading Priority

1. Environment variables (highest priority)
2. Config file (`/etc/proxicloud/config.yaml`)
3. Defaults (lowest priority)

**Example**:
```go
// Default
port := 8080

// Override from config file
if cfg.Server.BackendPort != 0 {
    port = cfg.Server.BackendPort
}

// Override from env var
if envPort := os.Getenv("PROXICLOUD_SERVER_BACKEND_PORT"); envPort != "" {
    port, _ = strconv.Atoi(envPort)
}
```

---

## 📚 Technology Choices

### Why Go for Backend?

- **Performance**: Compiled, fast execution
- **Concurrency**: Built-in goroutines for background tasks
- **Single Binary**: Easy deployment
- **Standard Library**: Built-in HTTP server
- **Type Safety**: Compile-time error checking

### Why Next.js for Frontend?

- **Server-Side Rendering**: Better SEO and initial load
- **App Router**: Modern routing with React Server Components
- **TypeScript**: Type safety for large codebase
- **Developer Experience**: Hot reload, error overlays
- **Production Ready**: Built-in optimization

### Why SQLite for Database?

- **Embedded**: No separate database server
- **Simple Deployment**: Single file
- **Performance**: Fast for read-heavy workloads
- **Reliable**: ACID compliant, battle-tested
- **Zero Configuration**: No setup required

### Why Tailwind CSS?

- **Utility-First**: Rapid UI development
- **Consistency**: Standardized spacing and colors
- **Performance**: Purges unused CSS
- **Dark Mode**: Built-in dark mode support
- **Customizable**: Easy to match AWS theme

---

## 🎯 Design Principles

1. **Simplicity**: Easy to install, configure, and use
2. **Reliability**: Graceful degradation (offline mode)
3. **Performance**: Fast response times, efficient metrics collection
4. **Security**: Token-based auth, secure defaults
5. **Maintainability**: Clean code structure, comprehensive tests
6. **User-Centric**: AWS-familiar interface, helpful error messages

---

## 📖 Related Documentation

- [Installation Guide](INSTALLATION.md)
- [Configuration Reference](CONFIGURATION.md)
- [API Documentation](API.md)
- [Development Guide](DEVELOPMENT.md)

---

**Architecture Version**: 1.0.0  
**Last Updated**: 2024-12-03
