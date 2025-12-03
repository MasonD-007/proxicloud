# Development Guide

This guide covers everything you need to know to develop, build, and contribute to ProxiCloud.

---

## 📋 Table of Contents

- [Prerequisites](#prerequisites)
- [Development Environment Setup](#development-environment-setup)
- [Project Structure](#project-structure)
- [Running Locally](#running-locally)
- [Building](#building)
- [Testing](#testing)
- [Contributing](#contributing)
- [Release Process](#release-process)

---

## 🛠️ Prerequisites

### Required

- **Go 1.21 or higher** - [Download](https://go.dev/dl/)
- **Node.js 18+ and npm** - [Download](https://nodejs.org/)
- **Git** - [Download](https://git-scm.com/)

### Optional

- **Access to a Proxmox VE node** (for testing)
- **Docker** (for containerized development)
- **SQLite CLI** (for database inspection)
- **Make** (for build automation)

### Verify Installation

```bash
go version      # Should be 1.21+
node --version  # Should be 18+
npm --version
git --version
```

---

## 🚀 Development Environment Setup

### 1. Clone the Repository

```bash
git clone https://github.com/yourusername/proxicloud.git
cd proxicloud
```

### 2. Backend Setup

```bash
cd backend

# Initialize Go modules
go mod download

# Create development config
mkdir -p /tmp/proxicloud
cat > /tmp/proxicloud/config.yaml <<EOF
proxmox:
  api_url: "https://YOUR-PROXMOX-IP:8006/api2/json"
  api_token: "root@pam!proxicloud=YOUR-TOKEN-HERE"
  node: ""
  insecure_skip_verify: true

server:
  backend_port: 8080
  frontend_port: 3000
  bind_address: "127.0.0.1"

analytics:
  interval: "10s"
  retention_days: 7
  database_path: "/tmp/proxicloud/analytics.db"

storage:
  default_storage: "local-lvm"

logging:
  level: "debug"
  path: "/tmp/proxicloud/app.log"

defaults:
  cpu_cores: 2
  memory_mb: 1024
  disk_gb: 9
  network_bridge: "vmbr0"
  container_name_prefix: "proxicloud"
EOF

# Set config path
export PROXICLOUD_CONFIG="/tmp/proxicloud/config.yaml"
```

### 3. Frontend Setup

```bash
cd frontend

# Install dependencies
npm install

# Create environment file
cat > .env.local <<EOF
NEXT_PUBLIC_API_URL=http://localhost:8080/api
EOF
```

---

## 📁 Project Structure

```
proxicloud/
├── backend/                    # Go API server
│   ├── cmd/
│   │   └── api/
│   │       └── main.go        # Entry point
│   ├── internal/              # Private application code
│   │   ├── config/           # Configuration management
│   │   ├── proxmox/          # Proxmox API client
│   │   ├── analytics/        # Metrics collection
│   │   ├── cache/            # Offline mode cache
│   │   ├── handlers/         # HTTP handlers
│   │   ├── middleware/       # HTTP middleware
│   │   └── server/           # HTTP server setup
│   ├── go.mod
│   ├── go.sum
│   └── Makefile
│
├── frontend/                  # Next.js web app
│   ├── app/                  # App Router pages
│   ├── components/           # React components
│   ├── lib/                  # Utilities and API client
│   ├── public/               # Static assets
│   ├── package.json
│   └── tsconfig.json
│
├── deploy/                    # Deployment scripts
│   ├── install.sh
│   ├── build-binaries.sh
│   ├── systemd/
│   ├── config/
│   └── scripts/
│
├── docs/                      # Documentation
│   ├── INSTALLATION.md
│   ├── CONFIGURATION.md
│   ├── API.md
│   ├── DEVELOPMENT.md
│   └── ARCHITECTURE.md
│
├── .github/
│   └── workflows/            # GitHub Actions
│       ├── build.yml
│       └── release.yml
│
└── README.md
```

---

## 🏃 Running Locally

### Option 1: Run Both Services Separately

**Terminal 1 - Backend**:
```bash
cd backend
export PROXICLOUD_CONFIG="/tmp/proxicloud/config.yaml"
go run cmd/api/main.go
```

Expected output:
```
2024/12/03 10:00:00 INFO Loading configuration from /tmp/proxicloud/config.yaml
2024/12/03 10:00:00 INFO Connecting to Proxmox at https://192.168.1.100:8006
2024/12/03 10:00:00 INFO Database initialized at /tmp/proxicloud/analytics.db
2024/12/03 10:00:00 INFO Starting metrics collector (interval: 10s)
2024/12/03 10:00:00 INFO API server listening on 127.0.0.1:8080
```

**Terminal 2 - Frontend**:
```bash
cd frontend
npm run dev
```

Expected output:
```
   ▲ Next.js 15.0.0
   - Local:        http://localhost:3000
   - Network:      http://192.168.1.50:3000

 ✓ Ready in 2.3s
```

**Access**: Open http://localhost:3000 in your browser

### Option 2: Use Make (if available)

```bash
# Backend
cd backend
make dev

# Frontend
cd frontend
make dev
```

---

## 🔨 Building

### Backend

```bash
cd backend

# Build for current platform
go build -o bin/api cmd/api/main.go

# Build for Linux (production)
GOOS=linux GOARCH=amd64 go build -o bin/api-linux-amd64 cmd/api/main.go

# Build for ARM (e.g., Raspberry Pi)
GOOS=linux GOARCH=arm64 go build -o bin/api-linux-arm64 cmd/api/main.go
```

### Frontend

```bash
cd frontend

# Build for production
npm run build

# The build output will be in .next/ directory

# Package for deployment
tar -czf proxicloud-frontend.tar.gz .next/ public/ package.json next.config.js
```

### Using Build Script

```bash
# Build everything for release
./deploy/build-binaries.sh

# Output will be in dist/ directory:
# - proxicloud-api-linux-amd64
# - proxicloud-api-linux-arm64
# - proxicloud-frontend.tar.gz
# - checksums.txt
```

---

## 🧪 Testing

### Backend Tests

```bash
cd backend

# Run all tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Run tests with verbose output
go test -v ./...

# Run specific package tests
go test ./internal/proxmox
```

### Frontend Tests

```bash
cd frontend

# Run tests (when implemented)
npm test

# Run tests in watch mode
npm test -- --watch

# Run with coverage
npm test -- --coverage
```

### Integration Tests

```bash
# Test against real Proxmox node
cd backend
export PROXICLOUD_CONFIG="/path/to/test-config.yaml"
go test ./... -tags=integration
```

### Manual Testing Checklist

- [ ] Create container with default settings
- [ ] Create container with custom settings
- [ ] Start/stop/reboot container
- [ ] Delete container
- [ ] View dashboard statistics
- [ ] View container metrics (CPU, RAM, network, disk)
- [ ] Test offline mode (disconnect Proxmox network)
- [ ] Test name auto-generation
- [ ] Test template selection
- [ ] Test with different storage backends

---

## 🤝 Contributing

We welcome contributions! Here's how to get started:

### 1. Fork the Repository

Click "Fork" on GitHub to create your own copy.

### 2. Create a Branch

```bash
git checkout -b feature/your-feature-name

# Branch naming conventions:
# - feature/add-vm-support
# - fix/container-list-crash
# - docs/update-api-guide
# - refactor/improve-metrics-collector
```

### 3. Make Your Changes

- Follow existing code style
- Add tests for new features
- Update documentation
- Keep commits atomic and well-described

### 4. Code Style

#### Go Code Style

- Follow [Effective Go](https://go.dev/doc/effective_go)
- Use `gofmt` for formatting
- Run `go vet` to catch common issues
- Use meaningful variable names

```bash
# Format code
gofmt -w .

# Vet code
go vet ./...

# Run linter (if installed)
golangci-lint run
```

#### TypeScript/React Code Style

- Use TypeScript for type safety
- Follow React best practices
- Use functional components with hooks
- Use Prettier for formatting

```bash
# Format code
npm run format

# Lint code
npm run lint
```

### 5. Commit Guidelines

Use [Conventional Commits](https://www.conventionalcommits.org/):

```bash
# Format: <type>(<scope>): <description>

git commit -m "feat(backend): add VM support endpoints"
git commit -m "fix(frontend): resolve container list pagination"
git commit -m "docs(api): add WebSocket documentation"
git commit -m "refactor(analytics): optimize metrics query"
git commit -m "test(proxmox): add client integration tests"
```

**Types**:
- `feat` - New feature
- `fix` - Bug fix
- `docs` - Documentation changes
- `refactor` - Code refactoring
- `test` - Adding tests
- `chore` - Maintenance tasks
- `perf` - Performance improvements

### 6. Submit Pull Request

```bash
# Push your branch
git push origin feature/your-feature-name
```

1. Go to GitHub and click "New Pull Request"
2. Fill in the PR template
3. Link related issues
4. Request review

### Pull Request Checklist

- [ ] Code follows project style guidelines
- [ ] All tests pass
- [ ] New features have tests
- [ ] Documentation is updated
- [ ] Commit messages are clear
- [ ] No merge conflicts

---

## 🔍 Debugging

### Backend Debugging

#### Using Delve (Go debugger)

```bash
cd backend

# Install delve
go install github.com/go-delve/delve/cmd/dlv@latest

# Debug with delve
dlv debug cmd/api/main.go
```

#### VSCode Debug Configuration

Create `.vscode/launch.json`:

```json
{
  "version": "0.2.0",
  "configurations": [
    {
      "name": "Debug Backend",
      "type": "go",
      "request": "launch",
      "mode": "debug",
      "program": "${workspaceFolder}/backend/cmd/api",
      "env": {
        "PROXICLOUD_CONFIG": "/tmp/proxicloud/config.yaml"
      }
    }
  ]
}
```

#### Enable Debug Logging

```yaml
# config.yaml
logging:
  level: "debug"
```

### Frontend Debugging

#### Browser DevTools

- Use React DevTools extension
- Check Network tab for API calls
- Check Console for errors

#### Next.js Debug Mode

```bash
NODE_OPTIONS='--inspect' npm run dev
```

Then open `chrome://inspect` in Chrome.

---

## 📦 Dependencies

### Backend Dependencies

```bash
# View dependencies
go list -m all

# Update dependencies
go get -u ./...

# Tidy dependencies
go mod tidy

# Verify dependencies
go mod verify
```

### Frontend Dependencies

```bash
# View dependencies
npm list

# Update dependencies
npm update

# Check for outdated packages
npm outdated

# Audit for vulnerabilities
npm audit
```

---

## 🚢 Release Process

### 1. Version Bump

Update version in:
- `backend/internal/server/server.go` (version constant)
- `frontend/package.json`
- `README.md`

### 2. Update Changelog

Create or update `CHANGELOG.md`:

```markdown
## [1.1.0] - 2024-12-15

### Added
- VM support endpoints
- WebSocket real-time updates
- User authentication

### Fixed
- Container list pagination bug
- Memory leak in metrics collector

### Changed
- Improved error messages
- Updated UI theme
```

### 3. Create Git Tag

```bash
git tag -a v1.1.0 -m "Release version 1.1.0"
git push origin v1.1.0
```

### 4. GitHub Actions Build

The release workflow will automatically:
1. Build binaries for all platforms
2. Run tests
3. Create GitHub Release
4. Upload assets
5. Generate release notes

### 5. Verify Release

- Check GitHub Releases page
- Download and test binaries
- Verify checksums
- Test installer script

---

## 🛠️ Useful Commands

### Backend

```bash
# Format code
gofmt -w .

# Vet code
go vet ./...

# Run tests
go test ./...

# Build
go build cmd/api/main.go

# Run
./main
```

### Frontend

```bash
# Install dependencies
npm install

# Dev server
npm run dev

# Build
npm run build

# Start production server
npm start

# Lint
npm run lint

# Format
npm run format
```

### Database

```bash
# Connect to SQLite database
sqlite3 /var/lib/proxicloud/analytics.db

# View tables
.tables

# View schema
.schema metrics

# Query data
SELECT * FROM metrics WHERE vmid = 100 LIMIT 10;

# Exit
.exit
```

---

## 📚 Resources

### Documentation
- [Go Documentation](https://go.dev/doc/)
- [Next.js Documentation](https://nextjs.org/docs)
- [Proxmox API Documentation](https://pve.proxmox.com/pve-docs/api-viewer/)
- [Tailwind CSS Documentation](https://tailwindcss.com/docs)

### Tools
- [Postman](https://www.postman.com/) - API testing
- [SQLite Browser](https://sqlitebrowser.org/) - Database inspection
- [React DevTools](https://react.dev/learn/react-developer-tools)
- [Go Playground](https://go.dev/play/)

### Community
- [GitHub Discussions](https://github.com/yourusername/proxicloud/discussions)
- [GitHub Issues](https://github.com/yourusername/proxicloud/issues)

---

## ❓ FAQ

**Q: Can I develop on Windows?**
A: Yes, but you'll need WSL2 for the best experience with Go and shell scripts.

**Q: How do I test without a Proxmox node?**
A: You can use mocked Proxmox API responses. Check `backend/internal/proxmox/mock.go`.

**Q: Can I use a different database?**
A: SQLite is embedded for simplicity. PostgreSQL support is planned for v2.0.

**Q: How do I debug API calls to Proxmox?**
A: Set logging level to `debug` and check logs. You can also use `curl` to test Proxmox API directly.

**Q: Where are logs stored during development?**
A: Check the path in your config (default: `/tmp/proxicloud/app.log`) and systemd journal (`journalctl -u proxicloud-api`).

---

## 🐛 Reporting Issues

When reporting bugs, please include:

1. ProxiCloud version
2. Proxmox VE version
3. Operating system
4. Steps to reproduce
5. Expected vs actual behavior
6. Relevant logs
7. Screenshots (if applicable)

Use the GitHub issue template to ensure you include all necessary information.

---

## 🎉 Thank You!

Thank you for contributing to ProxiCloud! Your efforts help make self-hosted Proxmox management better for everyone.

---

**Happy coding!** 🚀
