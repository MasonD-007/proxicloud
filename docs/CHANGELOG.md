# Changelog

All notable changes to ProxiCloud will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### In Progress (v0.6.0 - Core Implementation)
- Backend core implementation complete
- Frontend core implementation complete
- Basic container management working
- Dashboard and container list views
- One-line installer script

### Planned for v1.0
- Analytics system with metrics collection
- Offline mode with caching
- Container detail view
- Create container form UI
- Advanced middleware (logging, recovery)
- Additional deployment scripts

### Planned for v2.0+
- VM (KVM) support
- User authentication and multi-user support
- Container snapshots and backups
- WebSocket real-time updates
- Email/webhook alerts
- Custom templates and cloud-init support
- Load balancer configuration wizard
- Container migration between nodes

---

## [0.6.0] - 2024-12-03 (Development Milestone)

### Added - Backend Core
- Configuration system with YAML parsing and environment variable overrides
- Proxmox HTTP client with TLS support for self-signed certificates
- Complete LXC container operations (list, get, create, start, stop, reboot, delete)
- Auto VMID assignment for new containers
- Template listing from Proxmox storage
- RESTful API handlers for all container operations
- CORS middleware for cross-origin requests
- Server setup with Gorilla Mux router
- Health check endpoint

### Added - Frontend Core  
- Next.js 15 with App Router and TypeScript
- Dark theme design system with custom color tokens
- Layout components (TopBar with branding, Sidebar with navigation)
- UI component library (Button, Card, Input, Select, Badge)
- Dashboard page with summary statistics and recent containers
- Containers list page with start/stop/reboot/delete actions
- API client with typed TypeScript interfaces
- Utility functions for formatting bytes, CPU, uptime

### Added - Deployment
- One-line installer script for Proxmox nodes
- Simplified configuration example
- Systemd service files for backend and frontend
- Quick start documentation

### Built
- Backend binary (9.2MB, Go 1.21)
- Frontend static export (Next.js production build)

### Documentation
- Updated PROJECT_SUMMARY.md with implementation status
- Updated CHANGELOG.md with development milestone
- Created QUICK_START.md for testing guide

---

## [1.0.0] - TBD (Target Release)

### Added
- Initial release
- LXC container management (create, start, stop, reboot, delete)
- Real-time analytics (CPU, RAM, disk, network)
- SQLite-based metrics storage with 30-day retention
- Curated template library with featured OS templates
- Auto-generated container names with customization option
- Offline mode with cached data display
- Dark theme UI inspired by AWS Console
- AWS-style navigation (top bar + left sidebar)
- One-line installer script for Proxmox nodes
- RESTful API with comprehensive endpoints
- Background metrics collector (30-second interval)
- Systemd service integration
- Pre-built binaries for amd64 and arm64

### Documentation
- Installation guide with step-by-step instructions
- Configuration reference with all options
- API documentation with examples
- Development guide for contributors
- Architecture overview with technical details

### Infrastructure
- GitHub Actions workflows for build and release
- Automated binary builds for multiple architectures
- Release automation with checksums

---

## Version History

- **v1.0.0** - Initial release (TBD)

---

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md) for details on how to contribute to ProxiCloud.

## Release Notes Format

When creating a new release, follow this format:

```markdown
## [X.Y.Z] - YYYY-MM-DD

### Added
- New features

### Changed
- Changes in existing functionality

### Deprecated
- Features that will be removed in future versions

### Removed
- Features that have been removed

### Fixed
- Bug fixes

### Security
- Security improvements
```

---

[Unreleased]: https://github.com/MasonD-007/proxicloud/compare/v1.0.0...HEAD
[1.0.0]: https://github.com/MasonD-007/proxicloud/releases/tag/v1.0.0
