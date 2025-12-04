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

# Install git if missing
if ! command -v git >/dev/null 2>&1; then
    echo "Git not found. Installing Git..."
    apt-get update
    apt-get install -y git
    echo "Git installed successfully: $(git --version)"
else
    echo "Git is already installed: $(git --version)"
fi

# Install Go if missing
if ! command -v go >/dev/null 2>&1; then
    echo "Go not found. Installing Go..."
    GO_VERSION="1.22.6"
    cd /usr/local
    wget -q https://go.dev/dl/go${GO_VERSION}.linux-amd64.tar.gz
    rm -rf /usr/local/go
    tar -xzf go${GO_VERSION}.linux-amd64.tar.gz
    rm go${GO_VERSION}.linux-amd64.tar.gz
    echo 'export PATH=$PATH:/usr/local/go/bin' >> /etc/profile
    export PATH=$PATH:/usr/local/go/bin
    echo "Go installed successfully: $(go version)"
else
    echo "Go is already installed: $(go version)"
fi

# Install Node.js if missing
if ! command -v node >/dev/null 2>&1; then
    echo "Node.js not found. Installing Node.js 20..."
    curl -fsSL https://deb.nodesource.com/setup_20.x | bash -
    apt-get install -y nodejs
    echo "Node.js installed successfully: $(node -v)"
else
    echo "Node.js is already installed: $(node -v)"
fi

echo "Prerequisites check complete."

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
    echo ""
    echo "==================================="
    echo "Configuration Required"
    echo "==================================="
    echo ""
    echo "A configuration file has been created at:"
    echo "  /etc/proxicloud/config.yaml"
    echo ""
    echo "You MUST edit this file with your Proxmox credentials before continuing."
    echo ""
    echo "Required settings:"
    echo "  - Proxmox API URL"
    echo "  - API Token ID and Secret"
    echo "  - Server host and port settings"
    echo ""
    read -p "Press ENTER after you have edited the configuration file..."
    echo ""
    echo "Continuing with installation..."
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
