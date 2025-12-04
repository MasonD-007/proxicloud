#!/bin/bash
set -e

# ProxiCloud One-Line Installer
# Usage: curl -fsSL https://raw.githubusercontent.com/MasonD-007/proxicloud/main/deploy/install.sh | bash

echo "==================================="
echo "ProxiCloud Installer"
echo "==================================="
echo ""

# Configuration
GITHUB_REPO="MasonD-007/proxicloud"
INSTALL_DIR="/opt/proxicloud"
CONFIG_DIR="/etc/proxicloud"
DATA_DIR="/var/lib/proxicloud"
LOG_DIR="/var/log/proxicloud"

# Check if running on Proxmox
if [ ! -f /etc/pve/.version ]; then
    echo "Error: This script must be run on a Proxmox VE node"
    exit 1
fi

# Detect architecture
ARCH=$(uname -m)
case $ARCH in
    x86_64)
        BINARY_ARCH="amd64"
        ;;
    aarch64|arm64)
        BINARY_ARCH="arm64"
        ;;
    *)
        echo "Error: Unsupported architecture: $ARCH"
        echo "Supported architectures: x86_64 (amd64), aarch64/arm64"
        exit 1
        ;;
esac

echo "Detected architecture: $ARCH ($BINARY_ARCH)"

# Install required runtime dependencies
echo "Installing runtime dependencies..."
apt-get update -qq

# Install Node.js if missing (needed for frontend)
if ! command -v node >/dev/null 2>&1; then
    echo "Installing Node.js 20..."
    curl -fsSL https://deb.nodesource.com/setup_20.x | bash -
    apt-get install -y nodejs
    echo "Node.js installed: $(node -v)"
else
    NODE_VERSION=$(node -v | cut -d'v' -f2 | cut -d'.' -f1)
    if [ "$NODE_VERSION" -lt 18 ]; then
        echo "Node.js version too old ($(node -v)). Upgrading to Node.js 20..."
        curl -fsSL https://deb.nodesource.com/setup_20.x | bash -
        apt-get install -y nodejs
        echo "Node.js upgraded: $(node -v)"
    else
        echo "Node.js already installed: $(node -v)"
    fi
fi

# Install curl and tar (usually present, but just in case)
apt-get install -y curl tar

echo "Dependencies installed successfully."
echo ""

# Get the latest release version
echo "Fetching latest release information..."
LATEST_RELEASE=$(curl -s https://api.github.com/repos/$GITHUB_REPO/releases/latest | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/')

if [ -z "$LATEST_RELEASE" ]; then
    echo "Error: Could not fetch latest release information"
    echo "Please check your internet connection or try again later"
    exit 1
fi

echo "Latest release: $LATEST_RELEASE"
echo ""

# Create installation directories
echo "Creating installation directories..."
mkdir -p $INSTALL_DIR/{backend,frontend}
mkdir -p $CONFIG_DIR
mkdir -p $DATA_DIR
mkdir -p $LOG_DIR

# Download backend binary
echo "Downloading backend binary (linux-${BINARY_ARCH})..."
BACKEND_URL="https://github.com/$GITHUB_REPO/releases/download/$LATEST_RELEASE/proxicloud-api-linux-${BINARY_ARCH}"
curl -fsSL -o $INSTALL_DIR/backend/proxicloud-api "$BACKEND_URL"

if [ ! -f $INSTALL_DIR/backend/proxicloud-api ]; then
    echo "Error: Failed to download backend binary"
    exit 1
fi

chmod +x $INSTALL_DIR/backend/proxicloud-api
echo "Backend binary downloaded and installed."

# Download diagnostic script
echo "Downloading diagnostic script..."
DIAGNOSE_URL="https://raw.githubusercontent.com/$GITHUB_REPO/main/diagnose.sh"
curl -fsSL -o $INSTALL_DIR/diagnose.sh "$DIAGNOSE_URL"

if [ ! -f $INSTALL_DIR/diagnose.sh ]; then
    echo "Warning: Failed to download diagnostic script (non-critical)"
else
    chmod +x $INSTALL_DIR/diagnose.sh
    echo "Diagnostic script downloaded."
fi

# Download frontend package
echo "Downloading frontend package..."
FRONTEND_URL="https://github.com/$GITHUB_REPO/releases/download/$LATEST_RELEASE/proxicloud-frontend.tar.gz"
curl -fsSL -o /tmp/proxicloud-frontend.tar.gz "$FRONTEND_URL"

if [ ! -f /tmp/proxicloud-frontend.tar.gz ]; then
    echo "Error: Failed to download frontend package"
    exit 1
fi

# Extract frontend
echo "Extracting frontend..."
tar -xzf /tmp/proxicloud-frontend.tar.gz -C $INSTALL_DIR/frontend
rm /tmp/proxicloud-frontend.tar.gz

# Get node IP for API URL
NODE_IP=$(hostname -I | awk '{print $1}')

echo "Frontend setup complete."
echo ""

# Create configuration if it doesn't exist
if [ ! -f $CONFIG_DIR/config.yaml ]; then
    echo "==================================="
    echo "Configuration Setup"
    echo "==================================="
    echo ""
    echo "Please provide your Proxmox configuration details:"
    echo ""
    
    # Prompt for Proxmox Host (can be IP, hostname, or full URL)
    echo "Proxmox Host Configuration:"
    echo "  - Enter IP address (e.g., 192.168.1.100)"
    echo "  - Enter hostname (e.g., pve.local)"
    echo "  - Or full URL (e.g., https://192.168.1.100:8006)"
    read -p "Proxmox Host: " PROXMOX_HOST
    
    # Prompt for Proxmox Node Name
    read -p "Proxmox Node Name [$(hostname)]: " NODE_NAME
    NODE_NAME=${NODE_NAME:-$(hostname)}
    
    # Prompt for API Token ID
    read -p "API Token ID (e.g., root@pam!proxicloud): " TOKEN_ID
    
    # Prompt for API Token Secret (hidden input)
    echo -n "API Token Secret: "
    read -s TOKEN_SECRET
    echo ""
    
    # Prompt for verify SSL (default: true for self-signed certs)
    read -p "Skip SSL verification? (true/false) [true]: " INSECURE
    INSECURE=${INSECURE:-true}
    
    # Prompt for API server settings
    read -p "API Server Host [0.0.0.0]: " API_HOST
    API_HOST=${API_HOST:-0.0.0.0}
    
    read -p "API Server Port [8080]: " API_PORT
    API_PORT=${API_PORT:-8080}
    
    # Create configuration file with user input
    cat > $CONFIG_DIR/config.yaml << EOF
# ProxiCloud Configuration File
# Generated by installer on $(date)

# API Server Configuration
server:
  host: "$API_HOST"
  port: $API_PORT

# Proxmox Configuration
proxmox:
  host: "$PROXMOX_HOST"
  node: "$NODE_NAME"
  token_id: "$TOKEN_ID"
  token_secret: "$TOKEN_SECRET"
  insecure: $INSECURE
EOF
    
    echo ""
    echo "Configuration file created at: $CONFIG_DIR/config.yaml"
    echo ""
else
    echo "Configuration file already exists at: $CONFIG_DIR/config.yaml"
    echo ""
fi

# Install systemd services
echo "Installing systemd services..."

# Create API service
cat > /etc/systemd/system/proxicloud-api.service << 'SERVICE'
[Unit]
Description=ProxiCloud API Server
Documentation=https://github.com/MasonD-007/proxicloud
After=network-online.target
Wants=network-online.target

[Service]
Type=simple
User=root
Group=root

# Path to the ProxiCloud API binary
ExecStart=/opt/proxicloud/backend/proxicloud-api

# Working directory
WorkingDirectory=/opt/proxicloud/backend

# Configuration file location
Environment="CONFIG_PATH=/etc/proxicloud/config.yaml"

# Restart policy
Restart=on-failure
RestartSec=5s

# Security settings
NoNewPrivileges=true
PrivateTmp=true

# Logging
StandardOutput=journal
StandardError=journal
SyslogIdentifier=proxicloud-api

[Install]
WantedBy=multi-user.target
SERVICE

# Create frontend service
cat > /etc/systemd/system/proxicloud-frontend.service << SERVICE
[Unit]
Description=ProxiCloud Frontend Server
Documentation=https://github.com/MasonD-007/proxicloud
After=network-online.target proxicloud-api.service
Wants=network-online.target
Requires=proxicloud-api.service

[Service]
Type=simple
User=root
Group=root

# Path to Node.js and the Next.js standalone server
ExecStart=/usr/bin/node server.js

# Working directory (where server.js is located)
WorkingDirectory=/opt/proxicloud/frontend

# Environment variables
Environment="NODE_ENV=production"
Environment="PORT=3000"
Environment="HOSTNAME=0.0.0.0"

# Restart policy
Restart=on-failure
RestartSec=5s

# Security settings
NoNewPrivileges=true
PrivateTmp=true

# Logging
StandardOutput=journal
StandardError=journal
SyslogIdentifier=proxicloud-frontend

[Install]
WantedBy=multi-user.target
SERVICE

# Reload systemd
systemctl daemon-reload

# Enable and start services
echo "Enabling services..."
systemctl enable proxicloud-api
systemctl enable proxicloud-frontend

echo "Starting services..."
systemctl start proxicloud-api
systemctl start proxicloud-frontend

echo ""
echo "==================================="
echo "Installation Complete!"
echo "==================================="
echo ""
echo "ProxiCloud $LATEST_RELEASE is now running!"
echo ""
echo "Access URLs:"
echo "  Frontend: http://$NODE_IP:3000"
echo "  Backend:  http://$NODE_IP:8080/api"
echo ""
echo "Service Management:"
echo "  Status:  systemctl status proxicloud-api proxicloud-frontend"
echo "  Logs:    journalctl -u proxicloud-api -f"
echo "  Restart: systemctl restart proxicloud-api proxicloud-frontend"
echo "  Stop:    systemctl stop proxicloud-api proxicloud-frontend"
echo ""
echo "Diagnostics:"
echo "  Run:     $INSTALL_DIR/diagnose.sh"
echo ""
echo "Configuration:"
echo "  Config:  $CONFIG_DIR/config.yaml"
echo "  Data:    $DATA_DIR"
echo "  Logs:    $LOG_DIR"
echo ""
echo "Documentation:"
echo "  https://github.com/$GITHUB_REPO/blob/main/docs/"
echo ""
