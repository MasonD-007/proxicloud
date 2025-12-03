#!/bin/bash
set -e

# ProxiCloud One-Line Installer
# Usage: curl -fsSL https://raw.githubusercontent.com/MasonD-007/proxicloud/main/deploy/install.sh | bash

echo "==================================="
echo "ProxiCloud Installer"
echo "==================================="
echo ""

# Check if running on Proxmox
if [ ! -f /etc/pve/.version ]; then
    echo "Error: This script must be run on a Proxmox VE node"
    exit 1
fi

# Check for required tools
echo "Checking prerequisites..."
command -v go >/dev/null 2>&1 || { echo "Error: Go is not installed. Please install Go 1.21+"; exit 1; }
command -v node >/dev/null 2>&1 || { echo "Error: Node.js is not installed. Please install Node.js 18+"; exit 1; }

# Create installation directory
INSTALL_DIR="/opt/proxicloud"
echo "Creating installation directory: $INSTALL_DIR"
mkdir -p $INSTALL_DIR
cd $INSTALL_DIR

# Clone repository (or download release in production)
echo "Downloading ProxiCloud..."
if [ -d ".git" ]; then
    echo "Repository already exists, pulling latest changes..."
    git pull
else
    git clone https://github.com/MasonD-007/proxicloud.git .
fi

# Build backend
echo "Building backend..."
cd backend
go build -o proxicloud-api cmd/api/main.go
chmod +x proxicloud-api

# Build frontend
echo "Building frontend..."
cd ../frontend
npm install
npm run build

# Create config directory
echo "Setting up configuration..."
mkdir -p /etc/proxicloud
if [ ! -f /etc/proxicloud/config.yaml ]; then
    cp ../deploy/config/config.example.yaml /etc/proxicloud/config.yaml
    echo "Configuration file created at /etc/proxicloud/config.yaml"
    echo "Please edit this file with your Proxmox credentials"
fi

# Install systemd services
echo "Installing systemd services..."
cp ../deploy/systemd/proxicloud-frontend.service /etc/systemd/system/
cat > /etc/systemd/system/proxicloud-backend.service << 'SERVICE'
[Unit]
Description=ProxiCloud Backend API
After=network.target

[Service]
Type=simple
User=root
WorkingDirectory=/opt/proxicloud/backend
ExecStart=/opt/proxicloud/backend/proxicloud-api
Restart=always
RestartSec=10
Environment="CONFIG_PATH=/etc/proxicloud/config.yaml"

[Install]
WantedBy=multi-user.target
SERVICE

# Reload systemd
systemctl daemon-reload

# Enable and start services
echo "Enabling services..."
systemctl enable proxicloud-backend
systemctl enable proxicloud-frontend

echo "Starting services..."
systemctl start proxicloud-backend
systemctl start proxicloud-frontend

# Get node IP
NODE_IP=$(hostname -I | awk '{print $1}')

echo ""
echo "==================================="
echo "Installation Complete!"
echo "==================================="
echo ""
echo "ProxiCloud is now running!"
echo ""
echo "Backend API: http://$NODE_IP:8080"
echo "Frontend UI: http://$NODE_IP:3000"
echo ""
echo "Next steps:"
echo "1. Edit /etc/proxicloud/config.yaml with your Proxmox credentials"
echo "2. Restart services: systemctl restart proxicloud-backend"
echo "3. Access the UI at http://$NODE_IP:3000"
echo ""
echo "Service management:"
echo "  - Check status: systemctl status proxicloud-backend proxicloud-frontend"
echo "  - View logs: journalctl -u proxicloud-backend -f"
echo "  - Stop: systemctl stop proxicloud-backend proxicloud-frontend"
echo ""
