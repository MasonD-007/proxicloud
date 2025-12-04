#!/bin/bash
# ProxiCloud Uninstall Script
# Removes ProxiCloud installation and optionally cleans up data

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Installation paths
INSTALL_DIR="/opt/proxicloud"
DATA_DIR="/var/lib/proxicloud"
CONFIG_DIR="/etc/proxicloud"
SYSTEMD_DIR="/etc/systemd/system"
LOG_DIR="/var/log/proxicloud"
BIN_DIR="/usr/local/bin"

# Service names
API_SERVICE="proxicloud-api.service"
FRONTEND_SERVICE="proxicloud-frontend.service"

echo -e "${RED}═══════════════════════════════════════${NC}"
echo -e "${RED}ProxiCloud Uninstall Script${NC}"
echo -e "${RED}═══════════════════════════════════════${NC}"
echo ""

# Function to print colored output
print_info() {
    echo -e "${YELLOW}[i]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[✓]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[!]${NC} $1"
}

print_error() {
    echo -e "${RED}[✗]${NC} $1"
}

# Check if running as root
if [ "$EUID" -ne 0 ]; then
    print_error "This script must be run as root (use sudo)"
    exit 1
fi

# Warning
echo -e "${YELLOW}WARNING: This will remove ProxiCloud from your system${NC}"
echo ""
print_info "This script will:"
echo "  - Stop ProxiCloud services"
echo "  - Remove systemd services"
echo "  - Remove application files from $INSTALL_DIR"
echo "  - Optionally remove data and configuration"
echo ""

# Confirm uninstall
read -p "Are you sure you want to uninstall ProxiCloud? (yes/no): " CONFIRM

if [ "$CONFIRM" != "yes" ]; then
    print_info "Uninstall cancelled"
    exit 0
fi

echo ""

# Stop services
print_info "Stopping services..."

if systemctl is-active --quiet $API_SERVICE; then
    systemctl stop $API_SERVICE
    print_success "Stopped $API_SERVICE"
else
    print_info "$API_SERVICE is not running"
fi

if systemctl is-active --quiet $FRONTEND_SERVICE; then
    systemctl stop $FRONTEND_SERVICE
    print_success "Stopped $FRONTEND_SERVICE"
else
    print_info "$FRONTEND_SERVICE is not running"
fi

# Disable services
print_info "Disabling services..."

if systemctl is-enabled --quiet $API_SERVICE 2>/dev/null; then
    systemctl disable $API_SERVICE
    print_success "Disabled $API_SERVICE"
fi

if systemctl is-enabled --quiet $FRONTEND_SERVICE 2>/dev/null; then
    systemctl disable $FRONTEND_SERVICE
    print_success "Disabled $FRONTEND_SERVICE"
fi

# Remove systemd service files
print_info "Removing systemd service files..."

if [ -f "$SYSTEMD_DIR/$API_SERVICE" ]; then
    rm "$SYSTEMD_DIR/$API_SERVICE"
    print_success "Removed $API_SERVICE"
fi

if [ -f "$SYSTEMD_DIR/$FRONTEND_SERVICE" ]; then
    rm "$SYSTEMD_DIR/$FRONTEND_SERVICE"
    print_success "Removed $FRONTEND_SERVICE"
fi

systemctl daemon-reload
print_success "Reloaded systemd"

# Remove symlinks
print_info "Removing symlinks..."

if [ -L "$BIN_DIR/proxicloud-api" ]; then
    rm "$BIN_DIR/proxicloud-api"
    print_success "Removed proxicloud-api symlink"
fi

if [ -L "$BIN_DIR/proxicloud" ]; then
    rm "$BIN_DIR/proxicloud"
    print_success "Removed proxicloud symlink"
fi

# Remove application files
print_info "Removing application files..."

if [ -d "$INSTALL_DIR" ]; then
    rm -rf "$INSTALL_DIR"
    print_success "Removed $INSTALL_DIR"
else
    print_info "$INSTALL_DIR not found"
fi

# Ask about data removal
echo ""
print_warning "Data and Configuration Removal"
echo ""
echo "The following directories may contain your data:"
echo "  - $DATA_DIR (cache and analytics databases)"
echo "  - $CONFIG_DIR (configuration files)"
echo "  - $LOG_DIR (log files)"
echo ""
print_warning "Removing these will permanently delete your ProxiCloud data!"
echo ""

read -p "Do you want to remove data and configuration? (yes/no): " REMOVE_DATA

if [ "$REMOVE_DATA" = "yes" ]; then
    print_info "Removing data and configuration..."
    
    if [ -d "$DATA_DIR" ]; then
        rm -rf "$DATA_DIR"
        print_success "Removed $DATA_DIR"
    fi
    
    if [ -d "$CONFIG_DIR" ]; then
        rm -rf "$CONFIG_DIR"
        print_success "Removed $CONFIG_DIR"
    fi
    
    if [ -d "$LOG_DIR" ]; then
        rm -rf "$LOG_DIR"
        print_success "Removed $LOG_DIR"
    fi
    
    print_success "Data and configuration removed"
else
    print_info "Data and configuration preserved:"
    echo "  - $DATA_DIR"
    echo "  - $CONFIG_DIR"
    echo "  - $LOG_DIR"
    echo ""
    print_info "To manually remove later, run:"
    echo "  sudo rm -rf $DATA_DIR $CONFIG_DIR $LOG_DIR"
fi

# Remove user (optional)
echo ""
if id "proxicloud" &>/dev/null; then
    read -p "Remove 'proxicloud' user? (yes/no): " REMOVE_USER
    
    if [ "$REMOVE_USER" = "yes" ]; then
        userdel proxicloud 2>/dev/null || true
        print_success "Removed proxicloud user"
    else
        print_info "User 'proxicloud' preserved"
    fi
fi

# Remove Go and Node.js (optional)
echo ""
print_warning "Development Dependencies Removal"
echo ""
echo "ProxiCloud installed Go and Node.js as dependencies."
echo ""
print_warning "Removing these will affect other applications that may use them!"
echo ""

read -p "Do you want to remove Go? (yes/no): " REMOVE_GO

if [ "$REMOVE_GO" = "yes" ]; then
    print_info "Removing Go..."
    
    if [ -d "/usr/local/go" ]; then
        rm -rf /usr/local/go
        print_success "Removed Go installation from /usr/local/go"
    fi
    
    # Remove Go path from /etc/profile
    if grep -q "/usr/local/go/bin" /etc/profile 2>/dev/null; then
        sed -i '/\/usr\/local\/go\/bin/d' /etc/profile
        print_success "Removed Go from PATH in /etc/profile"
    fi
    
    print_success "Go removed"
else
    print_info "Go preserved"
fi

echo ""
read -p "Do you want to remove Node.js? (yes/no): " REMOVE_NODE

if [ "$REMOVE_NODE" = "yes" ]; then
    print_info "Removing Node.js..."
    
    # Remove Node.js and npm
    if command -v apt-get >/dev/null 2>&1; then
        apt-get remove -y nodejs npm 2>/dev/null || true
        apt-get autoremove -y 2>/dev/null || true
        print_success "Removed Node.js and npm"
    fi
    
    # Remove NodeSource repository
    if [ -f "/etc/apt/sources.list.d/nodesource.list" ]; then
        rm -f /etc/apt/sources.list.d/nodesource.list
        print_success "Removed NodeSource repository"
    fi
    
    # Remove Node.js related directories
    rm -rf /usr/lib/node_modules 2>/dev/null || true
    rm -rf /usr/local/lib/node_modules 2>/dev/null || true
    
    print_success "Node.js removed"
else
    print_info "Node.js preserved"
fi

# Final cleanup
print_info "Performing final cleanup..."

# Remove any temporary files
rm -rf /tmp/proxicloud* 2>/dev/null || true

# Remove any lock files
rm -f /var/lock/proxicloud* 2>/dev/null || true

# Remove any PID files
rm -f /var/run/proxicloud* 2>/dev/null || true

print_success "Cleanup complete"

# Summary
echo ""
echo -e "${GREEN}═══════════════════════════════════════${NC}"
echo -e "${GREEN}Uninstall Complete${NC}"
echo -e "${GREEN}═══════════════════════════════════════${NC}"
echo ""
print_success "ProxiCloud has been removed from your system"
echo ""

if [ "$REMOVE_DATA" != "yes" ]; then
    print_info "Your data has been preserved in:"
    echo "  - $DATA_DIR"
    echo "  - $CONFIG_DIR"
    echo "  - $LOG_DIR"
    echo ""
    print_info "To reinstall ProxiCloud later, your data will be available"
fi

echo "Thank you for using ProxiCloud!"
echo ""
