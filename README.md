# ☁️ ProxiCloud

**A self-hosted AWS-style management console for Proxmox VE**

ProxiCloud brings the simplicity and elegance of AWS EC2 console to your Proxmox infrastructure. Manage LXC containers with a modern, dark-themed web interface complete with real-time analytics and monitoring.

![License](https://img.shields.io/badge/license-MIT-blue.svg)
![Go Version](https://img.shields.io/badge/go-1.21+-00ADD8.svg)
![Next.js](https://img.shields.io/badge/next.js-15.0+-000000.svg)

---

## 🚀 Features

### Container Management
- ✅ **One-click creation** - Launch LXC containers with curated OS templates
- ✅ **Smart defaults** - Auto-generated names, sensible resource allocation
- ✅ **Bulk operations** - Start, stop, reboot, or delete multiple containers
- ✅ **Live status** - Real-time container state and resource usage
- ✅ **Template library** - Curated collection of popular Linux distributions

### Real-time Analytics
- 📊 **Historical metrics** - 30 days of CPU, RAM, disk, and network data
- 📈 **Interactive charts** - Time-series graphs with zoom and filtering
- 🔄 **Auto-refresh** - Metrics collected every 30 seconds
- 💾 **Efficient storage** - SQLite-based analytics with automatic cleanup

### Modern UI/UX
- 🌙 **Dark theme** - AWS-inspired interface optimized for long sessions
- 📱 **Responsive design** - Works on desktop, tablet, and mobile
- ⚡ **Fast & lightweight** - Built with Go and Next.js for maximum performance
- 🔌 **Offline resilience** - Shows cached data when Proxmox is unreachable

### Production Ready
- 🛡️ **Secure** - Token-based authentication with Proxmox API
- 🔧 **Easy deployment** - One-line installer script
- 📦 **Pre-built binaries** - No build dependencies required
- 🔄 **Auto-restart** - Systemd services with failure recovery

---

## 📸 Screenshots

> Screenshots coming soon!

---

## ⚡ Quick Start

### One-Line Installation

Run this command on your **Proxmox VE node** as root:

```bash
bash <(curl -fsSL https://raw.githubusercontent.com/yourusername/proxicloud/main/deploy/install.sh)
```

The installer will:
1. Detect your Proxmox node configuration
2. Prompt you to create an API token
3. Download and install ProxiCloud
4. Set up systemd services
5. Start the web interface

After installation, access ProxiCloud at:
```
http://YOUR-PROXMOX-IP:3000
```

---

## 📋 Requirements

### Proxmox VE Node
- Proxmox VE 7.0 or higher
- Root access (or sudo)
- Ports 3000 and 8080 available

### Network
- LAN access to Proxmox node
- Internet access for downloading binaries (during installation)

### API Token
- Proxmox API token with following privileges:
  - `VM.Allocate`
  - `VM.Config.Disk`
  - `VM.Config.Memory`
  - `VM.PowerMgmt`
  - `Datastore.Allocate`

---

## 📖 Documentation

- **[Installation Guide](docs/INSTALLATION.md)** - Detailed setup instructions
- **[Configuration Reference](docs/CONFIGURATION.md)** - All configuration options
- **[API Documentation](docs/API.md)** - REST API endpoints and examples
- **[Development Guide](docs/DEVELOPMENT.md)** - Build and contribute
- **[Architecture Overview](docs/ARCHITECTURE.md)** - Technical design details

---

## 🏗️ Architecture

```
┌─────────────────┐
│  Browser (LAN)  │
└────────┬────────┘
         │
         ▼
┌─────────────────────────┐
│  Next.js Frontend       │
│  Port 3000              │
│  - AWS-style UI         │
│  - Real-time updates    │
└────────┬────────────────┘
         │ HTTP/JSON
         ▼
┌─────────────────────────┐
│  Go API Server          │
│  Port 8080              │
│  - RESTful endpoints    │
│  - Metrics collector    │
│  - SQLite analytics     │
└────────┬────────────────┘
         │ HTTPS
         ▼
┌─────────────────────────┐
│  Proxmox VE API         │
│  Port 8006              │
│  - Container CRUD       │
│  - Resource metrics     │
└─────────────────────────┘
```

**Tech Stack:**
- **Backend**: Go 1.21+ with Gorilla Mux
- **Frontend**: Next.js 15 (App Router) + TypeScript + Tailwind CSS
- **Database**: SQLite for analytics
- **Deployment**: Systemd services on Proxmox node

---

## 🎯 Supported Features

### Current Release (v1.0)
- ✅ LXC container management (create, start, stop, reboot, delete)
- ✅ Real-time analytics (CPU, RAM, disk, network)
- ✅ Curated template library
- ✅ Auto-generated container names
- ✅ Offline mode with cached data
- ✅ Dark theme UI
- ✅ One-line installer

### Planned Features (v2.0)
- 🔜 VM (KVM) support
- 🔜 Container snapshots and backups
- 🔜 User authentication and multi-user support
- 🔜 Email/webhook alerts for resource thresholds
- 🔜 Container migration between nodes
- 🔜 Custom templates and cloud-init support
- 🔜 Load balancer configuration helper
- 🔜 Docker integration

---

## 🤝 Contributing

Contributions are welcome! Please read our [Development Guide](docs/DEVELOPMENT.md) to get started.

### Development Setup

1. Clone the repository:
```bash
git clone https://github.com/yourusername/proxicloud.git
cd proxicloud
```

2. Start the backend:
```bash
cd backend
go run cmd/api/main.go
```

3. Start the frontend:
```bash
cd frontend
npm install
npm run dev
```

4. Access at `http://localhost:3000`

---

## 📝 License

ProxiCloud is open-source software licensed under the [MIT License](LICENSE).

---

## 🙏 Acknowledgments

- Inspired by AWS EC2 Console
- Built for the Proxmox VE community
- Thanks to [Proxmox VE Helper Scripts](https://tteck.github.io/Proxmox/) for installer inspiration

---

## 📞 Support

- **Documentation**: [docs/](docs/)
- **Issues**: [GitHub Issues](https://github.com/yourusername/proxicloud/issues)
- **Discussions**: [GitHub Discussions](https://github.com/yourusername/proxicloud/discussions)

---

## ⭐ Star History

If you find ProxiCloud useful, please consider giving it a star on GitHub!

---

**Made with ❤️ for the Proxmox community**
