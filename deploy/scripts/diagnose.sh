#!/bin/bash

# ProxiCloud Diagnostic Script
# Run this to diagnose connection issues

echo "=================================="
echo "ProxiCloud Diagnostic Tool"
echo "=================================="
echo ""

# Check if running on Proxmox
if [ -f /etc/pve/.version ]; then
    echo "✓ Running on Proxmox VE"
    PROXMOX_MODE=true
    CONFIG_PATH="/etc/proxicloud/config.yaml"
    BINARY_PATH="/opt/proxicloud/backend/proxicloud-api"
else
    echo "○ Running in development mode (not on Proxmox)"
    PROXMOX_MODE=false
    CONFIG_PATH="$(pwd)/config.test.yaml"
    BINARY_PATH="$(pwd)/backend/proxicloud-api"
fi

echo ""
echo "Configuration Check:"
echo "--------------------"

# Check config file
if [ -f "$CONFIG_PATH" ]; then
    echo "✓ Config file exists: $CONFIG_PATH"
    echo ""
    echo "Config contents:"
    cat "$CONFIG_PATH"
    echo ""
else
    echo "✗ Config file NOT found: $CONFIG_PATH"
    exit 1
fi

echo ""
echo "Backend Binary Check:"
echo "--------------------"

# Check binary
if [ -f "$BINARY_PATH" ]; then
    echo "✓ Binary exists: $BINARY_PATH"
    if [ -x "$BINARY_PATH" ]; then
        echo "✓ Binary is executable"
    else
        echo "✗ Binary is NOT executable"
        echo "  Fix with: chmod +x $BINARY_PATH"
    fi
else
    echo "✗ Binary NOT found: $BINARY_PATH"
    echo "  Build it with: cd backend && go build -o proxicloud-api cmd/api/main.go"
    exit 1
fi

echo ""
echo "Service Status Check:"
echo "--------------------"

if [ "$PROXMOX_MODE" = true ]; then
    # Check systemd service
    if systemctl is-active --quiet proxicloud-api; then
        echo "✓ Backend service is running"
    else
        echo "✗ Backend service is NOT running"
        echo ""
        echo "Service status:"
        systemctl status proxicloud-api --no-pager
        echo ""
        echo "Recent logs:"
        journalctl -u proxicloud-api -n 20 --no-pager
        exit 1
    fi
    
    if systemctl is-active --quiet proxicloud-frontend; then
        echo "✓ Frontend service is running"
    else
        echo "✗ Frontend service is NOT running"
        echo ""
        echo "Service status:"
        systemctl status proxicloud-frontend --no-pager
        echo ""
        echo "Recent logs:"
        journalctl -u proxicloud-frontend -n 20 --no-pager
    fi
else
    # Check if processes are running
    if pgrep -f "proxicloud-api" > /dev/null; then
        echo "✓ Backend process is running (PID: $(pgrep -f proxicloud-api))"
    else
        echo "✗ Backend process is NOT running"
        echo "  Start it with: ./start.sh"
        exit 1
    fi
    
    if pgrep -f "next-server" > /dev/null || pgrep -f "npm.*dev" > /dev/null; then
        echo "✓ Frontend process is running"
    else
        echo "✗ Frontend process is NOT running"
        echo "  Start it with: ./start.sh"
    fi
fi

echo ""
echo "Network Check:"
echo "--------------------"

# Check if backend port is listening
if lsof -i :8080 > /dev/null 2>&1 || netstat -an 2>/dev/null | grep -q ":8080.*LISTEN"; then
    echo "✓ Port 8080 is listening"
else
    echo "✗ Port 8080 is NOT listening"
    echo "  Backend may not be running or failed to start"
    exit 1
fi

# Check if frontend port is listening
if lsof -i :3000 > /dev/null 2>&1 || netstat -an 2>/dev/null | grep -q ":3000.*LISTEN"; then
    echo "✓ Port 3000 is listening"
else
    echo "○ Port 3000 is NOT listening (frontend may not be running)"
fi

# Test backend API
echo ""
echo "API Health Check:"
echo "--------------------"

if command -v curl > /dev/null; then
    RESPONSE=$(curl -s -w "\n%{http_code}" http://localhost:8080/api/health 2>&1)
    HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
    BODY=$(echo "$RESPONSE" | head -n-1)
    
    if [ "$HTTP_CODE" = "200" ]; then
        echo "✓ Backend API is responding"
        echo "  Response: $BODY"
    else
        echo "✗ Backend API is NOT responding correctly"
        echo "  HTTP Code: $HTTP_CODE"
        echo "  Response: $BODY"
        exit 1
    fi
else
    echo "○ curl not available, skipping API test"
fi

echo ""
echo "=================================="
echo "Diagnosis Complete!"
echo "=================================="
echo ""

if [ "$PROXMOX_MODE" = true ]; then
    echo "Everything looks good! Access ProxiCloud at:"
    NODE_IP=$(hostname -I | awk '{print $1}')
    echo "  Frontend: http://$NODE_IP:3000"
    echo "  Backend:  http://$NODE_IP:8080/api"
else
    echo "Everything looks good! Access ProxiCloud at:"
    echo "  Frontend: http://localhost:3000"
    echo "  Backend:  http://localhost:8080/api"
fi
