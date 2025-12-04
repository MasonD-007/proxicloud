# ProxiCloud Deployment Scripts

This directory contains utility scripts for deploying and managing ProxiCloud.

## Scripts Overview

### 1. `build-binaries.sh`
**Multi-architecture build script for creating release binaries**

Builds ProxiCloud for multiple platforms:
- Linux AMD64
- Linux ARM64
- macOS AMD64 (Intel)
- macOS ARM64 (Apple Silicon)

**Usage:**
```bash
./deploy/scripts/build-binaries.sh
```

**Features:**
- Builds backend binaries for all platforms
- Builds and packages frontend
- Creates release archives with all necessary files
- Generates SHA256 checksums
- Includes configuration templates and systemd services

**Output:**
- Creates `build/` directory with:
  - Platform-specific binaries
  - Packaged release archives (`.tar.gz`)
  - SHA256SUMS file for verification

**Requirements:**
- Go 1.21+
- Node.js 18+
- Build tools (gcc for CGO)

### 2. `diagnose.sh`
**System diagnostics and troubleshooting tool**

Checks ProxiCloud installation and configuration for issues.

**Usage:**
```bash
./deploy/scripts/diagnose.sh
```

### 3. `setup-token.sh`
**Interactive Proxmox API token configuration helper**

Guides users through creating and configuring Proxmox API tokens.

**Usage:**
```bash
./deploy/scripts/setup-token.sh
```

**Features:**
- Step-by-step instructions for token creation
- Interactive prompts for configuration
- Optional connection testing
- Automatic config.yaml generation
- Security reminders and best practices

**Process:**
1. Collects Proxmox host information
2. Provides web UI instructions
3. Gathers token credentials
4. Tests connection (optional)
5. Generates configuration file
6. Shows security recommendations

**Output:**
- Creates `config.yaml` with:
  - Proxmox connection settings
  - API token credentials
  - Server configuration
  - Cache and analytics settings
  - Default container settings

### 3. `dev.sh`
**Development server for debugging without binaries**

Runs ProxiCloud directly from source code on your Proxmox node for testing and debugging.

**Usage:**
```bash
./deploy/scripts/dev.sh
```

**Features:**
- Runs backend with `go run` (no compilation needed)
- Runs frontend with `npm run dev` (hot reload enabled)
- Shows all errors and stack traces in real-time
- Live reload on file changes
- Uses temporary data directories by default
- Graceful shutdown with Ctrl+C
- Process monitoring and auto-cleanup

**Requirements:**
- Go 1.21+
- Node.js 18+
- Configuration file (config.yaml)

**Environment Variables:**
- `CONFIG_FILE` - Path to config file (default: ./config.test.yaml)
- `CACHE_PATH` - Cache database location (default: /tmp/proxicloud-dev/cache.db)
- `ANALYTICS_PATH` - Analytics database location (default: /tmp/proxicloud-dev/analytics.db)
- `BACKEND_PORT` - Backend port (default: 8080)
- `FRONTEND_PORT` - Frontend port (default: 3000)

**Example:**
```bash
# Basic usage
./deploy/scripts/dev.sh

# Custom config and ports
CONFIG_FILE=/etc/proxicloud/config.yaml FRONTEND_PORT=4000 ./deploy/scripts/dev.sh

# For help
./deploy/scripts/dev.sh --help
```

**Benefits:**
- See compile errors immediately
- Debug with full stack traces
- Test changes without rebuilding binaries
- Hot reload speeds up development
- No need to restart services manually

### 4. `uninstall.sh`
**Clean removal script with optional data preservation**

Safely removes ProxiCloud from the system.

**Usage:**
```bash
sudo ./deploy/scripts/uninstall.sh
```

**Features:**
- Stops running services
- Disables systemd services
- Removes application files
- Optional data removal
- Optional user removal
- Cleanup of temporary files

**What it removes:**
- ProxiCloud binaries from `/opt/proxicloud`
- Systemd service files
- Binary symlinks in `/usr/local/bin`

**Optional removals:**
- Data directory (`/var/lib/proxicloud`)
- Configuration files (`/etc/proxicloud`)
- Log files (`/var/log/proxicloud`)
- System user (`proxicloud`)

**Safety:**
- Requires confirmation before proceeding
- Asks separately about data removal
- Preserves data by default
- Shows what will be preserved

## Quick Start Guide

### Building for Release

```bash
# Clone repository
git clone https://github.com/MasonD-007/proxicloud
cd proxicloud

# Build all platforms
./deploy/scripts/build-binaries.sh

# Upload to GitHub releases
gh release create v1.0.0 build/*.tar.gz build/SHA256SUMS
```

### First-Time Setup

```bash
# Download release
wget https://github.com/MasonD-007/proxicloud/releases/download/v1.0.0/proxicloud-v1.0.0-linux-amd64.tar.gz

# Extract
tar xzf proxicloud-v1.0.0-linux-amd64.tar.gz
cd proxicloud-linux-amd64

# Configure Proxmox token
./setup-token.sh

# Set permissions
chmod 600 config.yaml

# Create data directory
sudo mkdir -p /var/lib/proxicloud
sudo chown $USER:$USER /var/lib/proxicloud

# Run ProxiCloud
./proxicloud-api
```

### Uninstalling

```bash
# With data removal
sudo /opt/proxicloud/uninstall.sh
# Answer "yes" to both prompts

# Preserving data
sudo /opt/proxicloud/uninstall.sh
# Answer "yes" to uninstall, "no" to data removal
```

## Directory Structure

```
deploy/
├── config/
│   └── config.example.yaml      # Configuration template
├── scripts/
│   ├── build-binaries.sh        # Multi-arch build script
│   ├── dev.sh                   # Development server (no binaries)
│   ├── diagnose.sh              # Diagnostics tool
│   ├── setup-token.sh           # Token setup helper
│   └── uninstall.sh             # Uninstall script
├── systemd/
│   ├── proxicloud-api.service   # Backend service
│   └── proxicloud-frontend.service  # Frontend service
└── install.sh                   # One-line installer
```

## Environment Variables

### Build Script
- `VERSION` - Override version detection (default: git describe)

### Application
- `CONFIG_PATH` - Path to config.yaml
- `CACHE_PATH` - Override cache database path
- `ANALYTICS_PATH` - Override analytics database path
- `NEXT_PUBLIC_API_URL` - Frontend API URL

## Troubleshooting

### Build Issues

**Problem:** CGO build fails on macOS
```bash
# Install Xcode command line tools
xcode-select --install
```

**Problem:** Cross-compilation fails
```bash
# Install cross-compilation tools
brew install FiloSottile/musl-cross/musl-cross
```

### Setup Issues

**Problem:** Token test fails
- Verify Proxmox is accessible at the specified host
- Check that port 8006 is not blocked by firewall
- Ensure token has proper permissions (Privilege Separation unchecked)

**Problem:** Permission denied on config.yaml
```bash
# Fix permissions
chmod 600 config.yaml
```

### Uninstall Issues

**Problem:** Services still running
```bash
# Force stop services
sudo systemctl stop proxicloud-api proxicloud-frontend
sudo systemctl kill proxicloud-api proxicloud-frontend
```

**Problem:** Files remain after uninstall
```bash
# Manual cleanup
sudo rm -rf /opt/proxicloud /var/lib/proxicloud /etc/proxicloud
```

## Best Practices

### For Development
1. Use `dev.sh` for local development and debugging
2. Use `build-binaries.sh` for release builds
3. Test on multiple architectures before release
4. Verify SHA256 checksums match
5. Tag releases with semantic versioning
6. Run `diagnose.sh` to check for issues before committing

### For Deployment
1. Always use `setup-token.sh` for initial configuration
2. Set proper file permissions (600 for config.yaml)
3. Use systemd services for production
4. Enable TLS verification in production
5. Backup configuration before upgrades

### For Maintenance
1. Keep configuration files outside installation directory
2. Use data directory for databases
3. Implement regular backups
4. Monitor log files in `/var/log/proxicloud`

## Security Considerations

### Token Management
- Store tokens securely (use 600 permissions)
- Never commit tokens to version control
- Rotate tokens periodically
- Use separate tokens for different environments

### System Security
- Run services as non-root user
- Use systemd for service management
- Enable TLS verification in production
- Restrict CORS origins to known domains
- Keep ProxiCloud and dependencies updated

### Data Protection
- Backup databases regularly
- Encrypt backups at rest
- Use secure channels for data transfer
- Implement access controls

## Contributing

When adding new deployment scripts:
1. Follow bash best practices
2. Include help text and usage examples
3. Add error handling and validation
4. Test on multiple platforms
5. Update this README

## Support

For issues with deployment scripts:
1. Check the troubleshooting section
2. Review script output for error messages
3. Consult the main documentation
4. Open an issue on GitHub

## License

MIT License - See LICENSE file for details
