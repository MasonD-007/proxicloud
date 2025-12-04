# ProxiCloud Quick Reference Card

One-page reference for ProxiCloud development.

---

## 📦 Project Structure

```
proxicloud/
├── backend/           Go API server
├── frontend/          Next.js web app
├── docs/             Complete documentation
├── deploy/           Deployment scripts
└── .github/          CI/CD workflows
```

---

## 🚀 Quick Start

```bash
# Backend
cd backend && go mod download && go run cmd/api/main.go

# Frontend
cd frontend && npm install && npm run dev
```

---

## 📚 Documentation Files

| File | Purpose |
|------|---------|
| `README.md` | Project overview |
| `GETTING_STARTED.md` | How to start development |
| `IMPLEMENTATION_GUIDE.md` | Code templates |
| `PROJECT_SUMMARY.md` | Complete implementation plan |
| `docs/INSTALLATION.md` | Installation guide |
| `docs/CONFIGURATION.md` | Config reference |
| `docs/API.md` | API documentation |
| `docs/DEVELOPMENT.md` | Development guide |
| `docs/ARCHITECTURE.md` | Technical architecture |

---

## 🎯 Key Specifications

| Feature | Specification |
|---------|--------------|
| **Backend** | Go 1.21+, Gorilla Mux, SQLite |
| **Frontend** | Next.js 15, TypeScript, Tailwind |
| **API Token** | Root token, privilege separation disabled |
| **Defaults** | 2 cores, 1GB RAM, 9GB disk |
| **Naming** | `proxicloud-001`, `proxicloud-002`, ... |
| **Analytics** | 30s collection, 30-day retention |
| **Theme** | Dark mode, AWS-style navigation |
| **Ports** | Backend: 8080, Frontend: 3000 |
| **Config** | `/etc/proxicloud/config.yaml` |

---

## 🔧 Common Commands

### Backend

```bash
make dev          # Run with hot reload
make build        # Build binary
make test         # Run tests
make lint         # Run linter
make clean        # Clean artifacts
```

### Frontend

```bash
make dev          # Development server
make build        # Production build
make lint         # Run ESLint
make type-check   # TypeScript check
make clean        # Clean artifacts
```

---

## 🗺️ Implementation Roadmap

### Week 1: Backend Core (18 hours)
1. Config system (2h)
2. Proxmox client (4h)
3. Container CRUD (6h)
4. Analytics system (4h)
5. Testing (2h)

### Week 2: Frontend (12 hours)
1. Layout components (2h)
2. Dashboard page (2h)
3. Container list (3h)
4. Create form (3h)
5. Analytics page (2h)

### Week 3: Polish & Deploy (6 hours)
1. Offline mode (2h)
2. Deployment scripts (3h)
3. Final testing (1h)

**Total**: ~36 hours (3-4 weeks part-time)

---

## 📂 Core Files to Implement

### Backend (15 files)

```
cmd/api/main.go                          Entry point
internal/config/config.go                Config loader
internal/config/types.go                 Config structs
internal/proxmox/client.go               HTTP client
internal/proxmox/lxc.go                  Container ops
internal/proxmox/metrics.go              RRD fetching
internal/analytics/store.go              SQLite
internal/analytics/collector.go          Metrics collection
internal/cache/cache.go                  Offline mode
internal/handlers/containers.go          Container endpoints
internal/handlers/metrics.go             Analytics endpoints
internal/middleware/cors.go              CORS
internal/middleware/logging.go           Logging
internal/server/server.go                HTTP server
internal/server/routes.go                Routes
```

### Frontend (20 files)

```
app/layout.tsx                           Root layout
app/page.tsx                             Dashboard
app/containers/page.tsx                  Container list
app/containers/create/page.tsx           Create form
app/analytics/page.tsx                   Analytics
components/layout/TopBar.tsx             Top nav
components/layout/Sidebar.tsx            Sidebar
components/layout/ConnectionBanner.tsx   Offline banner
components/containers/ContainerCard.tsx  Card view
components/containers/CreateForm.tsx     Form
components/analytics/MetricsChart.tsx    Charts
components/ui/Button.tsx                 Button
components/ui/Card.tsx                   Card
lib/api.ts                               API client
lib/types.ts                             Types
```

---

## 🎨 Design System

### Colors

```css
--background: #0f1419        /* Main background */
--surface: #16181d           /* Cards/panels */
--border: #2a2d35            /* Borders */
--text-primary: #ffffff      /* Primary text */
--text-secondary: #9ca3af    /* Secondary text */
--success: #10b981           /* Running */
--error: #ef4444             /* Stopped/Error */
--primary: #0ea5e9           /* ProxiCloud blue */
```

### Typography

```
Font: System default
Sizes: text-sm, text-base, text-lg, text-xl, text-2xl
Weights: font-normal, font-medium, font-semibold
```

---

## 🔍 Debugging

### Backend

```bash
# Enable debug logging
export PROXICLOUD_LOGGING_LEVEL=debug

# View logs
journalctl -u proxicloud-api -f

# Test API
curl http://localhost:8080/api/health
```

### Frontend

```bash
# Check for errors
npm run lint
npm run type-check

# View in browser
open http://localhost:3000
```

---

## 📡 API Endpoints

| Endpoint | Method | Purpose |
|----------|--------|---------|
| `/api/health` | GET | Health check |
| `/api/containers` | GET | List containers |
| `/api/containers` | POST | Create container |
| `/api/containers/:id` | GET | Get container |
| `/api/containers/:id/start` | POST | Start container |
| `/api/containers/:id/stop` | POST | Stop container |
| `/api/containers/:id` | DELETE | Delete container |
| `/api/analytics/:id/cpu` | GET | CPU metrics |
| `/api/analytics/:id/memory` | GET | Memory metrics |

---

## 🐛 Common Issues

### Backend won't start
- Check config file exists
- Verify API token format
- Ensure Proxmox is reachable

### Frontend build fails
- Run `npm install`
- Check Node.js version (>= 18)
- Clear `.next/` cache

### Can't connect to Proxmox
- Verify API token has permissions
- Check firewall rules
- Test with `curl`

---

## 🚀 Releases & GitHub Features

### Create a Release

```bash
# 1. Update changelog
vim CHANGELOG.md

# 2. Commit changes
git add CHANGELOG.md
git commit -m "docs: Update changelog for v1.0.0"
git push origin main

# 3. Create and push tag (triggers release workflow)
git tag -a v1.0.0 -m "Release v1.0.0 - Initial stable release"
git push origin v1.0.0
```

### GitHub CLI Commands

```bash
# Issues
gh issue create --title "Bug: ..." --body "..."
gh issue list
gh issue view 123

# Pull Requests
gh pr create --fill
gh pr list
gh pr checkout 123
gh pr review 123 --approve

# Releases
gh release list
gh release view v1.0.0
gh release download v1.0.0

# Workflows
gh workflow list
gh run list
gh run watch
```

### Commit Message Format

```bash
feat: Add new feature          # New feature
fix: Fix bug                   # Bug fix
docs: Update docs              # Documentation
style: Format code             # Code style
refactor: Refactor code        # Code restructuring
perf: Improve performance      # Performance improvement
test: Add tests                # Testing
build: Update dependencies     # Build changes
ci: Update workflows           # CI/CD changes
chore: Update config           # Maintenance tasks
```

**For detailed release instructions, see [RELEASE_GUIDE.md](RELEASE_GUIDE.md)**

---

## 🔗 Useful Links

- **Proxmox API Docs**: https://pve.proxmox.com/pve-docs/api-viewer/
- **Go Docs**: https://go.dev/doc/
- **Next.js Docs**: https://nextjs.org/docs
- **Tailwind Docs**: https://tailwindcss.com/docs
- **GitHub CLI**: https://cli.github.com/

---

## ✅ Pre-Launch Checklist

- [ ] All tests pass
- [ ] Documentation complete
- [ ] Installer tested
- [ ] Manual testing on Proxmox
- [ ] Screenshots added to README
- [ ] CHANGELOG updated
- [ ] Version tagged

---

## 🎉 You're Ready!

Everything is documented and ready for implementation.

**Next step**: `cd backend && make dev`

---

**Quick Links**:
- [Full Implementation Plan](PROJECT_SUMMARY.md)
- [Getting Started Guide](GETTING_STARTED.md)
- [Code Templates](IMPLEMENTATION_GUIDE.md)
- [Development Guide](docs/DEVELOPMENT.md)
