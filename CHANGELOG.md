# Changelog

All notable changes to ProxiCloud will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Planned
- VM (KVM) support
- User authentication and multi-user support
- Container snapshots and backups
- WebSocket real-time updates
- Email/webhook alerts
- Custom templates and cloud-init support
- Load balancer configuration wizard
- Container migration between nodes

---

## [1.0.0] - TBD

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

[Unreleased]: https://github.com/yourusername/proxicloud/compare/v1.0.0...HEAD
[1.0.0]: https://github.com/yourusername/proxicloud/releases/tag/v1.0.0
