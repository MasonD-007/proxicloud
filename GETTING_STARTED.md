# Getting Started with ProxiCloud Development

This guide will help you get started when you're ready to begin implementation.

---

## 🎯 Pre-Implementation Checklist

Before starting development, ensure you have:

- [ ] **Go 1.21+** installed (`go version`)
- [ ] **Node.js 18+** installed (`node --version`)
- [ ] **Access to a Proxmox VE node** for testing
- [ ] **Proxmox API token** created (see docs/INSTALLATION.md)
- [ ] **Git configured** with your name and email
- [ ] **Code editor** set up (VS Code recommended)

---

## 🚀 Quick Start Commands

### Initial Setup

```bash
# Clone repository (if not already done)
git clone https://github.com/yourusername/proxicloud.git
cd proxicloud

# Create backend directory structure
mkdir -p backend/cmd/api
mkdir -p backend/internal/{config,proxmox,analytics,cache,handlers,middleware,server}

# Create frontend directory structure  
mkdir -p frontend/{app,components,lib,public}
```

### Backend Setup

```bash
cd backend

# Initialize Go module
go mod init github.com/yourusername/proxicloud/backend

# Add dependencies
go get github.com/gorilla/mux
go get github.com/mattn/go-sqlite3
go get gopkg.in/yaml.v3
go get github.com/rs/cors

# Create main.go
cat > cmd/api/main.go <<'EOF'
package main

import (
    "fmt"
    "log"
    "net/http"
)

func main() {
    http.HandleFunc("/api/health", func(w http.ResponseWriter, r *http.Request) {
        w.Header().Set("Content-Type", "application/json")
        w.WriteHeader(http.StatusOK)
        fmt.Fprintf(w, `{"status":"healthy"}`)
    })

    log.Println("Starting ProxiCloud API server on :8080")
    log.Fatal(http.ListenAndServe(":8080", nil))
}
EOF

# Test run
go run cmd/api/main.go
```

### Frontend Setup

```bash
cd frontend

# Initialize Next.js project
npx create-next-app@latest . --typescript --tailwind --app --no-src-dir --import-alias "@/*"

# Install additional dependencies
npm install recharts
npm install lucide-react
npm install date-fns

# Create API client stub
cat > lib/api.ts <<'EOF'
const API_URL = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080/api';

export async function healthCheck() {
  const response = await fetch(`${API_URL}/health`);
  return response.json();
}
EOF

# Test run
npm run dev
```

---

## 📂 Recommended Development Order

### Week 1: Backend Foundation

**Day 1-2: Core Setup**
1. Initialize Go project
2. Create configuration system
3. Implement Proxmox API client
4. Add authentication

**Day 3-4: Container Management**
5. Implement container list endpoint
6. Implement container create endpoint
7. Implement start/stop/reboot endpoints
8. Implement delete endpoint

**Day 5: Testing**
9. Test against real Proxmox node
10. Fix bugs and edge cases

### Week 2: Analytics & Frontend

**Day 1-2: Analytics System**
1. Set up SQLite database
2. Implement metrics collector
3. Create analytics endpoints
4. Test data collection

**Day 3-5: Frontend**
5. Create layout and navigation
6. Build dashboard page
7. Build container list page
8. Build create container form

### Week 3: Polish & Deploy

**Day 1-2: Advanced Features**
1. Implement analytics charts
2. Add offline mode
3. Implement auto-name generation

**Day 3-4: Deployment**
4. Create installer script
5. Test deployment on Proxmox
6. Write build scripts

**Day 5: Release**
7. Final testing
8. Create release
9. Publish documentation

---

## 🔧 Development Environment Setup

### VS Code Extensions (Recommended)

```json
{
  "recommendations": [
    "golang.go",
    "bradlc.vscode-tailwindcss",
    "esbenp.prettier-vscode",
    "dbaeumer.vscode-eslint",
    "formulahendry.auto-rename-tag",
    "ms-vscode.vscode-typescript-next"
  ]
}
```

Save as `.vscode/extensions.json`

### VS Code Settings

```json
{
  "go.useLanguageServer": true,
  "go.formatTool": "gofmt",
  "editor.formatOnSave": true,
  "editor.codeActionsOnSave": {
    "source.fixAll.eslint": true
  },
  "[typescript]": {
    "editor.defaultFormatter": "esbenp.prettier-vscode"
  },
  "[typescriptreact]": {
    "editor.defaultFormatter": "esbenp.prettier-vscode"
  }
}
```

Save as `.vscode/settings.json`

---

## 🎨 Design Tokens (For Frontend)

Add to `frontend/app/globals.css`:

```css
@tailwind base;
@tailwind components;
@tailwind utilities;

@layer base {
  :root {
    --background: 15 20 25;           /* #0f1419 */
    --surface: 22 24 29;              /* #16181d */
    --surface-elevated: 30 33 38;     /* #1e2126 */
    --border: 42 45 53;               /* #2a2d35 */
    
    --text-primary: 255 255 255;      /* #ffffff */
    --text-secondary: 156 163 175;    /* #9ca3af */
    --text-muted: 107 114 128;        /* #6b7280 */
    
    --success: 16 185 129;            /* #10b981 */
    --warning: 245 158 11;            /* #f59e0b */
    --error: 239 68 68;               /* #ef4444 */
    --info: 59 130 246;               /* #3b82f6 */
    
    --primary: 14 165 233;            /* #0ea5e9 */
    --primary-hover: 2 132 199;       /* #0284c7 */
    --accent: 139 92 246;             /* #8b5cf6 */
  }
}

body {
  background-color: rgb(var(--background));
  color: rgb(var(--text-primary));
}
```

---

## 📋 Implementation Checklist

### Backend (backend/)

#### Configuration
- [ ] `internal/config/config.go` - Load YAML config
- [ ] `internal/config/types.go` - Config structs
- [ ] Environment variable support
- [ ] Validate configuration on startup

#### Proxmox Client
- [ ] `internal/proxmox/client.go` - HTTP client
- [ ] `internal/proxmox/lxc.go` - Container operations
- [ ] `internal/proxmox/metrics.go` - RRD data
- [ ] `internal/proxmox/templates.go` - Template listing
- [ ] `internal/proxmox/types.go` - Response types
- [ ] Error handling and retries

#### Handlers
- [ ] `internal/handlers/containers.go` - CRUD endpoints
- [ ] `internal/handlers/metrics.go` - Analytics endpoints
- [ ] `internal/handlers/dashboard.go` - Summary stats
- [ ] `internal/handlers/health.go` - Health check
- [ ] `internal/handlers/templates.go` - Template endpoints

#### Middleware
- [ ] `internal/middleware/cors.go` - CORS support
- [ ] `internal/middleware/logging.go` - Request logging
- [ ] `internal/middleware/recovery.go` - Panic recovery

#### Server
- [ ] `internal/server/server.go` - Server setup
- [ ] `internal/server/routes.go` - Route registration

#### Analytics
- [ ] `internal/analytics/store.go` - SQLite operations
- [ ] `internal/analytics/collector.go` - Background collector
- [ ] `internal/analytics/retention.go` - Data cleanup
- [ ] Database schema creation
- [ ] Indexes for performance

#### Cache
- [ ] `internal/cache/cache.go` - Offline mode cache
- [ ] Cache update logic
- [ ] Cache retrieval on error

#### Main
- [ ] `cmd/api/main.go` - Application entry point

---

### Frontend (frontend/)

#### Layout
- [ ] `components/layout/TopBar.tsx`
- [ ] `components/layout/Sidebar.tsx`
- [ ] `components/layout/Layout.tsx`
- [ ] `components/layout/ConnectionBanner.tsx`

#### Pages
- [ ] `app/page.tsx` - Dashboard
- [ ] `app/containers/page.tsx` - Container list
- [ ] `app/containers/create/page.tsx` - Create form
- [ ] `app/containers/[id]/page.tsx` - Container detail
- [ ] `app/analytics/page.tsx` - Analytics dashboard
- [ ] `app/templates/page.tsx` - Template list

#### Components - Dashboard
- [ ] `components/dashboard/SummaryCard.tsx`
- [ ] `components/dashboard/RecentActivity.tsx`
- [ ] `components/dashboard/QuickActions.tsx`

#### Components - Containers
- [ ] `components/containers/ContainerCard.tsx`
- [ ] `components/containers/ContainerTable.tsx`
- [ ] `components/containers/CreateForm.tsx`
- [ ] `components/containers/StatusBadge.tsx`
- [ ] `components/containers/ActionButtons.tsx`

#### Components - Analytics
- [ ] `components/analytics/MetricsChart.tsx`
- [ ] `components/analytics/TimeRangeSelector.tsx`
- [ ] `components/analytics/ContainerSelector.tsx`

#### Components - UI
- [ ] `components/ui/Button.tsx`
- [ ] `components/ui/Card.tsx`
- [ ] `components/ui/Input.tsx`
- [ ] `components/ui/Select.tsx`
- [ ] `components/ui/Badge.tsx`

#### Utilities
- [ ] `lib/api.ts` - API client
- [ ] `lib/types.ts` - TypeScript types
- [ ] `lib/utils.ts` - Helper functions

---

### Deployment (deploy/)

- [ ] `deploy/install.sh` - One-line installer
- [ ] `deploy/build-binaries.sh` - Build script
- [ ] `deploy/scripts/setup-token.sh` - Token helper
- [ ] `deploy/scripts/uninstall.sh` - Uninstall script

---

## 🧪 Testing Strategy

### Backend Tests

```bash
# Unit tests
go test ./internal/config
go test ./internal/proxmox
go test ./internal/analytics

# Integration tests
go test -tags=integration ./...

# Coverage
go test -cover ./...
```

### Frontend Tests

```bash
# Lint
npm run lint

# Type check
npm run type-check

# Build test
npm run build
```

### Manual Testing Checklist

- [ ] Health endpoint responds
- [ ] List containers works
- [ ] Create container with defaults
- [ ] Create container with custom settings
- [ ] Start container
- [ ] Stop container
- [ ] Reboot container
- [ ] Delete container
- [ ] View analytics for container
- [ ] Offline mode displays cached data
- [ ] Auto-name generation works
- [ ] Template selection works

---

## 📞 Getting Help

If you get stuck during implementation:

1. **Check Documentation**: All details are in `docs/`
2. **Review API Spec**: See `docs/API.md`
3. **Check Architecture**: See `docs/ARCHITECTURE.md`
4. **Proxmox API Docs**: https://pve.proxmox.com/pve-docs/api-viewer/

---

## 🎯 Success Criteria

Your implementation will be complete when:

1. ✅ Backend API responds to all documented endpoints
2. ✅ Frontend displays all pages without errors
3. ✅ Can create, start, stop, and delete containers
4. ✅ Analytics charts display historical data
5. ✅ Offline mode works when Proxmox is unreachable
6. ✅ Installer script successfully deploys to Proxmox
7. ✅ All tests pass
8. ✅ Documentation matches implementation

---

## 🚀 Ready to Start?

When you're ready to begin:

```bash
# Start with backend
cd backend
go mod init github.com/yourusername/proxicloud/backend
# Follow backend checklist above

# Then frontend
cd frontend
npx create-next-app@latest .
# Follow frontend checklist above
```

**Good luck! You've got comprehensive documentation to guide you through every step.** 🎉

---

**Estimated Timeline**: 3-4 weeks working part-time (22-31 hours total)

**Next Step**: Start with `backend/cmd/api/main.go` and the configuration system.
