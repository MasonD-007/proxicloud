#!/bin/bash
# ProxiCloud Analytics Diagnostic Script
# Run this on your Proxmox server to diagnose analytics issues

set +e  # Don't exit on error

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}════════════════════════════════════════════${NC}"
echo -e "${BLUE}ProxiCloud Analytics Diagnostic Tool${NC}"
echo -e "${BLUE}════════════════════════════════════════════${NC}"
echo ""

# Get paths from environment or use defaults
ANALYTICS_PATH="${ANALYTICS_PATH:-/tmp/proxicloud-dev/analytics.db}"
CACHE_PATH="${CACHE_PATH:-/tmp/proxicloud-dev/cache.db}"
CONFIG_PATH="${CONFIG_PATH:-/etc/proxicloud/config.yaml}"

echo -e "${YELLOW}[1] Environment Variables${NC}"
echo "   ANALYTICS_PATH: ${ANALYTICS_PATH}"
echo "   CACHE_PATH: ${CACHE_PATH}"
echo "   CONFIG_PATH: ${CONFIG_PATH}"
echo ""

# Check directory
ANALYTICS_DIR=$(dirname "$ANALYTICS_PATH")
echo -e "${YELLOW}[2] Directory Check${NC}"
if [ -d "$ANALYTICS_DIR" ]; then
    echo -e "   ${GREEN}✓${NC} Directory exists: $ANALYTICS_DIR"
    ls -la "$ANALYTICS_DIR"
else
    echo -e "   ${RED}✗${NC} Directory does NOT exist: $ANALYTICS_DIR"
    echo -e "   ${YELLOW}→${NC} Creating directory..."
    mkdir -p "$ANALYTICS_DIR" && echo -e "   ${GREEN}✓${NC} Created successfully" || echo -e "   ${RED}✗${NC} Failed to create"
fi
echo ""

# Check permissions
echo -e "${YELLOW}[3] Permissions Check${NC}"
if [ -w "$ANALYTICS_DIR" ]; then
    echo -e "   ${GREEN}✓${NC} Directory is writable"
else
    echo -e "   ${RED}✗${NC} Directory is NOT writable"
    echo "   Owner: $(stat -c '%U:%G' "$ANALYTICS_DIR" 2>/dev/null || stat -f '%Su:%Sg' "$ANALYTICS_DIR" 2>/dev/null)"
    echo "   Current user: $(whoami)"
fi
echo ""

# Check if files exist
echo -e "${YELLOW}[4] Database Files${NC}"
if [ -f "$ANALYTICS_PATH" ]; then
    echo -e "   ${GREEN}✓${NC} Analytics DB exists: $ANALYTICS_PATH"
    echo "   Size: $(du -h "$ANALYTICS_PATH" | cut -f1)"
    if command -v sqlite3 &> /dev/null; then
        echo "   Records: $(sqlite3 "$ANALYTICS_PATH" "SELECT COUNT(*) FROM metrics" 2>/dev/null || echo "Error reading")"
    fi
else
    echo -e "   ${YELLOW}!${NC} Analytics DB does NOT exist (will be created on backend start)"
fi

if [ -f "$CACHE_PATH" ]; then
    echo -e "   ${GREEN}✓${NC} Cache DB exists: $CACHE_PATH"
    echo "   Size: $(du -h "$CACHE_PATH" | cut -f1)"
else
    echo -e "   ${YELLOW}!${NC} Cache DB does NOT exist (will be created on backend start)"
fi
echo ""

# Check if backend is running
echo -e "${YELLOW}[5] Backend Process${NC}"
if pgrep -f "go run.*main.go" > /dev/null; then
    echo -e "   ${GREEN}✓${NC} Backend is RUNNING"
    echo "   PID(s): $(pgrep -f 'go run.*main.go' | tr '\n' ' ')"
else
    echo -e "   ${RED}✗${NC} Backend is NOT running"
fi
echo ""

# Check if backend API is responding
echo -e "${YELLOW}[6] Backend API Health${NC}"
if command -v curl &> /dev/null; then
    HEALTH_RESPONSE=$(curl -s http://localhost:8080/api/health 2>&1)
    if [ $? -eq 0 ]; then
        echo -e "   ${GREEN}✓${NC} Backend API is responding"
        echo "   Response: $HEALTH_RESPONSE"
    else
        echo -e "   ${RED}✗${NC} Backend API is NOT responding"
        echo "   Error: $HEALTH_RESPONSE"
    fi
else
    echo -e "   ${YELLOW}!${NC} curl not installed, skipping API check"
fi
echo ""

# Test analytics endpoint
echo -e "${YELLOW}[7] Analytics Endpoint Test${NC}"
if command -v curl &> /dev/null; then
    ANALYTICS_RESPONSE=$(curl -s http://localhost:8080/api/analytics/stats 2>&1)
    if echo "$ANALYTICS_RESPONSE" | grep -q "analytics not available"; then
        echo -e "   ${RED}✗${NC} Analytics is NOT available"
        echo "   Response: $ANALYTICS_RESPONSE"
    elif echo "$ANALYTICS_RESPONSE" | grep -q "enabled"; then
        echo -e "   ${GREEN}✓${NC} Analytics is WORKING"
        echo "   Response: $ANALYTICS_RESPONSE"
    else
        echo -e "   ${YELLOW}!${NC} Unexpected response"
        echo "   Response: $ANALYTICS_RESPONSE"
    fi
else
    echo -e "   ${YELLOW}!${NC} curl not installed, skipping endpoint check"
fi
echo ""

# Check backend logs (if systemd)
echo -e "${YELLOW}[8] Recent Backend Logs${NC}"
if command -v journalctl &> /dev/null && systemctl is-active proxicloud-api &> /dev/null; then
    echo "   Checking systemd logs..."
    journalctl -u proxicloud-api -n 20 --no-pager | grep -i "analytics\|cache\|error" || echo "   No relevant logs found"
else
    echo -e "   ${YELLOW}!${NC} Not running as systemd service or journalctl not available"
    echo "   If running dev.sh, check terminal output for errors"
fi
echo ""

# Test database creation manually
echo -e "${YELLOW}[9] Manual Database Test${NC}"
if command -v sqlite3 &> /dev/null; then
    TEST_DB="/tmp/proxicloud-test-$$.db"
    echo "   Testing SQLite database creation..."
    if sqlite3 "$TEST_DB" "CREATE TABLE test (id INTEGER); DROP TABLE test;" 2>/dev/null; then
        echo -e "   ${GREEN}✓${NC} Can create SQLite databases in /tmp"
        rm -f "$TEST_DB"
    else
        echo -e "   ${RED}✗${NC} Cannot create SQLite databases"
        echo "   This might indicate SQLite issues or permission problems"
    fi
else
    echo -e "   ${YELLOW}!${NC} sqlite3 not installed, skipping test"
fi
echo ""

# Recommendations
echo -e "${BLUE}════════════════════════════════════════════${NC}"
echo -e "${BLUE}Recommendations${NC}"
echo -e "${BLUE}════════════════════════════════════════════${NC}"
echo ""

if [ ! -d "$ANALYTICS_DIR" ]; then
    echo -e "${RED}→${NC} Create the analytics directory:"
    echo "   mkdir -p $ANALYTICS_DIR"
    echo ""
fi

if ! pgrep -f "go run.*main.go" > /dev/null; then
    echo -e "${RED}→${NC} Backend is not running. Start it with:"
    echo "   cd /path/to/proxicloud"
    echo "   ./deploy/scripts/dev.sh"
    echo ""
    echo "   Or check backend logs for startup errors"
    echo ""
fi

if [ -d "$ANALYTICS_DIR" ] && pgrep -f "go run.*main.go" > /dev/null; then
    if [ ! -f "$ANALYTICS_PATH" ]; then
        echo -e "${YELLOW}→${NC} Backend is running but databases not created:"
        echo "   1. Check backend terminal for initialization errors"
        echo "   2. Look for messages like:"
        echo "      'Analytics initialized at...'"
        echo "      'Warning: Failed to initialize analytics...'"
        echo ""
        echo "   3. Try stopping and restarting the backend:"
        echo "      Ctrl+C (in dev.sh terminal)"
        echo "      ./deploy/scripts/dev.sh"
        echo ""
    fi
fi

echo -e "${BLUE}════════════════════════════════════════════${NC}"
echo -e "${GREEN}Diagnostic complete!${NC}"
echo -e "${BLUE}════════════════════════════════════════════${NC}"
