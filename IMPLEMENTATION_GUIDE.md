# ProxiCloud Implementation Guide

This guide provides code templates and step-by-step instructions for implementing ProxiCloud.

---

## 📋 Table of Contents

1. [Backend Implementation](#backend-implementation)
2. [Frontend Implementation](#frontend-implementation)
3. [Testing](#testing)
4. [Deployment](#deployment)

---

## 🔧 Backend Implementation

### Step 1: Configuration System

**File**: `backend/internal/config/types.go`

```go
package config

type Config struct {
    Proxmox   ProxmoxConfig   `yaml:"proxmox"`
    Server    ServerConfig    `yaml:"server"`
    Analytics AnalyticsConfig `yaml:"analytics"`
    Storage   StorageConfig   `yaml:"storage"`
    Templates TemplatesConfig `yaml:"templates"`
    Logging   LoggingConfig   `yaml:"logging"`
    Defaults  DefaultsConfig  `yaml:"defaults"`
}

type ProxmoxConfig struct {
    APIURL             string `yaml:"api_url"`
    APIToken           string `yaml:"api_token"`
    Node               string `yaml:"node"`
    InsecureSkipVerify bool   `yaml:"insecure_skip_verify"`
}

type ServerConfig struct {
    BackendPort  int    `yaml:"backend_port"`
    FrontendPort int    `yaml:"frontend_port"`
    BindAddress  string `yaml:"bind_address"`
}

type AnalyticsConfig struct {
    Interval       string `yaml:"interval"`
    RetentionDays  int    `yaml:"retention_days"`
    DatabasePath   string `yaml:"database_path"`
}

type StorageConfig struct {
    DefaultStorage string `yaml:"default_storage"`
}

type TemplatesConfig struct {
    Featured []string `yaml:"featured"`
}

type LoggingConfig struct {
    Level string `yaml:"level"`
    Path  string `yaml:"path"`
}

type DefaultsConfig struct {
    CPUCores             int    `yaml:"cpu_cores"`
    MemoryMB             int    `yaml:"memory_mb"`
    DiskGB               int    `yaml:"disk_gb"`
    NetworkBridge        string `yaml:"network_bridge"`
    ContainerNamePrefix  string `yaml:"container_name_prefix"`
}
```

**File**: `backend/internal/config/config.go`

```go
package config

import (
    "fmt"
    "os"
    "strconv"
    "gopkg.in/yaml.v3"
)

func Load(path string) (*Config, error) {
    // Read config file
    data, err := os.ReadFile(path)
    if err != nil {
        return nil, fmt.Errorf("failed to read config: %w", err)
    }

    // Parse YAML
    var cfg Config
    if err := yaml.Unmarshal(data, &cfg); err != nil {
        return nil, fmt.Errorf("failed to parse config: %w", err)
    }

    // Apply environment variable overrides
    applyEnvOverrides(&cfg)

    // Auto-detect node name if not set
    if cfg.Proxmox.Node == "" {
        hostname, err := os.Hostname()
        if err == nil {
            cfg.Proxmox.Node = hostname
        }
    }

    // Validate config
    if err := validate(&cfg); err != nil {
        return nil, fmt.Errorf("invalid config: %w", err)
    }

    return &cfg, nil
}

func applyEnvOverrides(cfg *Config) {
    if token := os.Getenv("PROXICLOUD_PROXMOX_API_TOKEN"); token != "" {
        cfg.Proxmox.APIToken = token
    }
    if port := os.Getenv("PROXICLOUD_SERVER_BACKEND_PORT"); port != "" {
        if p, err := strconv.Atoi(port); err == nil {
            cfg.Server.BackendPort = p
        }
    }
    // Add more as needed...
}

func validate(cfg *Config) error {
    if cfg.Proxmox.APIToken == "" {
        return fmt.Errorf("proxmox.api_token is required")
    }
    if cfg.Server.BackendPort == 0 {
        cfg.Server.BackendPort = 8080
    }
    return nil
}
```

---

### Step 2: Proxmox API Client

**File**: `backend/internal/proxmox/client.go`

```go
package proxmox

import (
    "bytes"
    "crypto/tls"
    "encoding/json"
    "fmt"
    "io"
    "net/http"
    "time"
)

type Client struct {
    baseURL    string
    token      string
    node       string
    httpClient *http.Client
}

func NewClient(baseURL, token, node string, insecureSkipVerify bool) *Client {
    return &Client{
        baseURL: baseURL,
        token:   token,
        node:    node,
        httpClient: &http.Client{
            Timeout: 30 * time.Second,
            Transport: &http.Transport{
                TLSClientConfig: &tls.Config{
                    InsecureSkipVerify: insecureSkipVerify,
                },
            },
        },
    }
}

func (c *Client) request(method, path string, body interface{}) ([]byte, error) {
    var reqBody io.Reader
    if body != nil {
        jsonData, err := json.Marshal(body)
        if err != nil {
            return nil, fmt.Errorf("failed to marshal body: %w", err)
        }
        reqBody = bytes.NewReader(jsonData)
    }

    req, err := http.NewRequest(method, c.baseURL+path, reqBody)
    if err != nil {
        return nil, fmt.Errorf("failed to create request: %w", err)
    }

    req.Header.Set("Authorization", "PVEAPIToken="+c.token)
    req.Header.Set("Content-Type", "application/json")

    resp, err := c.httpClient.Do(req)
    if err != nil {
        return nil, fmt.Errorf("request failed: %w", err)
    }
    defer resp.Body.Close()

    respBody, err := io.ReadAll(resp.Body)
    if err != nil {
        return nil, fmt.Errorf("failed to read response: %w", err)
    }

    if resp.StatusCode >= 400 {
        return nil, fmt.Errorf("API error %d: %s", resp.StatusCode, string(respBody))
    }

    return respBody, nil
}

// HealthCheck tests the connection to Proxmox
func (c *Client) HealthCheck() error {
    _, err := c.request("GET", "/version", nil)
    return err
}
```

**File**: `backend/internal/proxmox/lxc.go`

```go
package proxmox

import (
    "encoding/json"
    "fmt"
)

type Container struct {
    VMID     int     `json:"vmid"`
    Name     string  `json:"name"`
    Status   string  `json:"status"`
    CPU      float64 `json:"cpu"`
    Memory   int64   `json:"mem"`
    MaxMem   int64   `json:"maxmem"`
    Disk     int64   `json:"disk"`
    MaxDisk  int64   `json:"maxdisk"`
    Uptime   int64   `json:"uptime"`
    NetIn    int64   `json:"netin"`
    NetOut   int64   `json:"netout"`
}

type ContainerConfig struct {
    Hostname    string            `json:"hostname"`
    Template    string            `json:"ostemplate"`
    Cores       int               `json:"cores"`
    Memory      int               `json:"memory"`
    Swap        int               `json:"swap"`
    RootFS      string            `json:"rootfs"`
    Net0        string            `json:"net0"`
    Password    string            `json:"password,omitempty"`
    SSHKeys     string            `json:"ssh-public-keys,omitempty"`
    Start       bool              `json:"start"`
    OnBoot      bool              `json:"onboot"`
    Unprivileged bool             `json:"unprivileged"`
}

func (c *Client) ListContainers() ([]Container, error) {
    path := fmt.Sprintf("/nodes/%s/lxc", c.node)
    data, err := c.request("GET", path, nil)
    if err != nil {
        return nil, err
    }

    var response struct {
        Data []Container `json:"data"`
    }
    if err := json.Unmarshal(data, &response); err != nil {
        return nil, fmt.Errorf("failed to parse response: %w", err)
    }

    return response.Data, nil
}

func (c *Client) GetContainer(vmid int) (*Container, error) {
    path := fmt.Sprintf("/nodes/%s/lxc/%d/status/current", c.node, vmid)
    data, err := c.request("GET", path, nil)
    if err != nil {
        return nil, err
    }

    var response struct {
        Data Container `json:"data"`
    }
    if err := json.Unmarshal(data, &response); err != nil {
        return nil, fmt.Errorf("failed to parse response: %w", err)
    }

    return &response.Data, nil
}

func (c *Client) CreateContainer(config ContainerConfig) (int, string, error) {
    // Find next available VMID
    vmid, err := c.getNextVMID()
    if err != nil {
        return 0, "", err
    }

    path := fmt.Sprintf("/nodes/%s/lxc", c.node)
    
    // Add VMID to config
    configMap := map[string]interface{}{
        "vmid":     vmid,
        "hostname": config.Hostname,
        "ostemplate": config.Template,
        "cores":    config.Cores,
        "memory":   config.Memory,
        "rootfs":   config.RootFS,
        "net0":     config.Net0,
    }
    
    if config.Password != "" {
        configMap["password"] = config.Password
    }
    
    data, err := c.request("POST", path, configMap)
    if err != nil {
        return 0, "", err
    }

    var response struct {
        Data string `json:"data"` // Task ID (UPID)
    }
    if err := json.Unmarshal(data, &response); err != nil {
        return 0, "", fmt.Errorf("failed to parse response: %w", err)
    }

    return vmid, response.Data, nil
}

func (c *Client) StartContainer(vmid int) error {
    path := fmt.Sprintf("/nodes/%s/lxc/%d/status/start", c.node, vmid)
    _, err := c.request("POST", path, nil)
    return err
}

func (c *Client) StopContainer(vmid int) error {
    path := fmt.Sprintf("/nodes/%s/lxc/%d/status/stop", c.node, vmid)
    _, err := c.request("POST", path, nil)
    return err
}

func (c *Client) RebootContainer(vmid int) error {
    path := fmt.Sprintf("/nodes/%s/lxc/%d/status/reboot", c.node, vmid)
    _, err := c.request("POST", path, nil)
    return err
}

func (c *Client) DeleteContainer(vmid int) error {
    path := fmt.Sprintf("/nodes/%s/lxc/%d", c.node, vmid)
    _, err := c.request("DELETE", path, nil)
    return err
}

func (c *Client) getNextVMID() (int, error) {
    data, err := c.request("GET", "/cluster/nextid", nil)
    if err != nil {
        return 0, err
    }

    var response struct {
        Data int `json:"data"`
    }
    if err := json.Unmarshal(data, &response); err != nil {
        return 0, fmt.Errorf("failed to parse response: %w", err)
    }

    return response.Data, nil
}
```

---

### Step 3: Main Application Entry Point

**File**: `backend/cmd/api/main.go`

```go
package main

import (
    "context"
    "log"
    "os"
    "os/signal"
    "syscall"
    "time"

    "github.com/MasonD-007/proxicloud/backend/internal/config"
    "github.com/MasonD-007/proxicloud/backend/internal/proxmox"
    "github.com/MasonD-007/proxicloud/backend/internal/server"
)

var version = "1.0.0"

func main() {
    // Load configuration
    configPath := os.Getenv("PROXICLOUD_CONFIG")
    if configPath == "" {
        configPath = "/etc/proxicloud/config.yaml"
    }

    cfg, err := config.Load(configPath)
    if err != nil {
        log.Fatalf("Failed to load config: %v", err)
    }

    log.Printf("ProxiCloud v%s starting...", version)
    log.Printf("Node: %s", cfg.Proxmox.Node)

    // Create Proxmox client
    client := proxmox.NewClient(
        cfg.Proxmox.APIURL,
        cfg.Proxmox.APIToken,
        cfg.Proxmox.Node,
        cfg.Proxmox.InsecureSkipVerify,
    )

    // Test connection
    if err := client.HealthCheck(); err != nil {
        log.Fatalf("Failed to connect to Proxmox: %v", err)
    }
    log.Println("Connected to Proxmox successfully")

    // Create server
    srv := server.New(cfg, client)

    // Start server
    go func() {
        if err := srv.Start(); err != nil {
            log.Fatalf("Server error: %v", err)
        }
    }()

    log.Printf("API server listening on %s:%d", cfg.Server.BindAddress, cfg.Server.BackendPort)

    // Wait for interrupt signal
    quit := make(chan os.Signal, 1)
    signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
    <-quit

    log.Println("Shutting down server...")

    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()

    if err := srv.Shutdown(ctx); err != nil {
        log.Fatalf("Server forced to shutdown: %v", err)
    }

    log.Println("Server exited")
}
```

---

## 🎨 Frontend Implementation

### Step 1: API Client

**File**: `frontend/lib/api.ts`

```typescript
const API_URL = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080/api';

export interface Container {
  vmid: number;
  name: string;
  status: 'running' | 'stopped' | 'paused';
  cpu: number;
  memory: number;
  maxmem: number;
  disk: number;
  maxdisk: number;
  uptime: number;
  netin: number;
  netout: number;
}

export interface CreateContainerRequest {
  hostname?: string;
  template: string;
  cores?: number;
  memory?: number;
  rootfs?: number;
  storage?: string;
  network?: {
    bridge: string;
    ip: string;
  };
  password?: string;
  ssh_keys?: string[];
  start?: boolean;
}

async function fetchAPI<T>(endpoint: string, options?: RequestInit): Promise<T> {
  const response = await fetch(`${API_URL}${endpoint}`, {
    ...options,
    headers: {
      'Content-Type': 'application/json',
      ...options?.headers,
    },
  });

  if (!response.ok) {
    throw new Error(`API error: ${response.statusText}`);
  }

  return response.json();
}

export async function healthCheck() {
  return fetchAPI<{ status: string }>('/health');
}

export async function listContainers() {
  const data = await fetchAPI<{ containers: Container[] }>('/containers');
  return data.containers;
}

export async function getContainer(vmid: number) {
  return fetchAPI<Container>(`/containers/${vmid}`);
}

export async function createContainer(config: CreateContainerRequest) {
  return fetchAPI<{ vmid: number; task_id: string }>('/containers', {
    method: 'POST',
    body: JSON.stringify(config),
  });
}

export async function startContainer(vmid: number) {
  return fetchAPI(`/containers/${vmid}/start`, { method: 'POST' });
}

export async function stopContainer(vmid: number) {
  return fetchAPI(`/containers/${vmid}/stop`, { method: 'POST' });
}

export async function deleteContainer(vmid: number) {
  return fetchAPI(`/containers/${vmid}`, { method: 'DELETE' });
}
```

---

### Step 2: Layout Components

**File**: `frontend/components/layout/Layout.tsx`

```typescript
import { TopBar } from './TopBar';
import { Sidebar } from './Sidebar';
import { ConnectionBanner } from './ConnectionBanner';

export function Layout({ children }: { children: React.ReactNode }) {
  return (
    <div className="min-h-screen bg-[rgb(var(--background))]">
      <TopBar />
      <ConnectionBanner />
      <div className="flex">
        <Sidebar />
        <main className="flex-1 p-6 ml-60 mt-16">
          {children}
        </main>
      </div>
    </div>
  );
}
```

**File**: `frontend/components/layout/TopBar.tsx`

```typescript
export function TopBar() {
  return (
    <header className="fixed top-0 left-0 right-0 h-16 bg-[rgb(var(--surface))] border-b border-[rgb(var(--border))] z-50">
      <div className="flex items-center justify-between h-full px-6">
        <div className="flex items-center space-x-4">
          <h1 className="text-xl font-semibold text-[rgb(var(--text-primary))]">
            ☁️ ProxiCloud
          </h1>
        </div>
        <div className="flex items-center space-x-4">
          <span className="text-sm text-[rgb(var(--text-secondary))]">
            Node: pve
          </span>
        </div>
      </div>
    </header>
  );
}
```

---

## 🧪 Testing

### Backend Test Example

**File**: `backend/internal/proxmox/client_test.go`

```go
package proxmox

import (
    "testing"
)

func TestNewClient(t *testing.T) {
    client := NewClient(
        "https://localhost:8006/api2/json",
        "test-token",
        "pve",
        true,
    )

    if client == nil {
        t.Fatal("Expected client to be created")
    }

    if client.node != "pve" {
        t.Errorf("Expected node 'pve', got '%s'", client.node)
    }
}
```

---

## 🚀 Deployment

### Build Script

**File**: `deploy/build-binaries.sh`

```bash
#!/bin/bash
set -e

echo "Building ProxiCloud binaries..."

# Create dist directory
mkdir -p dist

# Build backend
echo "Building backend..."
cd backend
GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o ../dist/proxicloud-api-linux-amd64 cmd/api/main.go
GOOS=linux GOARCH=arm64 go build -ldflags="-s -w" -o ../dist/proxicloud-api-linux-arm64 cmd/api/main.go
cd ..

# Build frontend
echo "Building frontend..."
cd frontend
npm ci
npm run build
tar -czf ../dist/proxicloud-frontend.tar.gz .next/ public/ package.json package-lock.json next.config.js
cd ..

# Generate checksums
echo "Generating checksums..."
cd dist
sha256sum * > checksums.txt
cd ..

echo "Build complete! Artifacts in dist/"
```

Make it executable:
```bash
chmod +x deploy/build-binaries.sh
```

---

This implementation guide provides the foundation. Refer to PROJECT_SUMMARY.md for the complete task list!
