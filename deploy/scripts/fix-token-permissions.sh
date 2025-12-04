#!/bin/bash
# Fix Proxmox API Token Permissions
# Run this on your Proxmox server to give the token proper permissions

set -e

# Colors
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
RED='\033[0;31m'
NC='\033[0m'

echo -e "${GREEN}=== ProxiCloud API Token Permission Fix ===${NC}"
echo ""

# Check if we're root
if [ "$EUID" -ne 0 ]; then 
    echo -e "${RED}Error: This script must be run as root${NC}"
    echo "Please run: sudo $0"
    exit 1
fi

# Check if pvesh exists
if ! command -v pvesh &> /dev/null; then
    echo -e "${RED}Error: This must be run on a Proxmox server${NC}"
    exit 1
fi

TOKEN_ID="root@pam!proxicloud"

echo -e "${BLUE}Current token:${NC} $TOKEN_ID"
echo ""

# Check if token exists
if ! pvesh get /access/users/root@pam/token/proxicloud &>/dev/null; then
    echo -e "${RED}Error: Token 'proxicloud' does not exist${NC}"
    echo ""
    echo "Create it first with:"
    echo "  pveum user token add root@pam proxicloud --privsep=0"
    exit 1
fi

echo -e "${YELLOW}Checking current permissions...${NC}"
echo ""

# Show current token info
pvesh get /access/users/root@pam/token/proxicloud

echo ""
echo -e "${GREEN}=== Fix #1: Ensure Privilege Separation is OFF ===${NC}"
echo ""
echo "Privilege separation must be disabled for the token to inherit"
echo "the user's (root) permissions."
echo ""

# Disable privilege separation
pveum user token modify root@pam proxicloud --privsep=0

echo -e "${GREEN}[✓]${NC} Privilege separation disabled"
echo ""

echo -e "${GREEN}=== Fix #2: Grant Administrator Role ===${NC}"
echo ""
echo "Ensuring token has Administrator role on / (root path)"
echo ""

# Grant Administrator role to the token
pveum acl modify / --tokens root@pam!proxicloud --roles Administrator

echo -e "${GREEN}[✓]${NC} Administrator role granted"
echo ""

echo -e "${GREEN}=== Verification ===${NC}"
echo ""

# Test the token
TOKEN_SECRET=$(cat /etc/pve/priv/token.cfg | grep "root@pam!proxicloud" | cut -d: -f2 | tr -d ' ')

if [ -z "$TOKEN_SECRET" ]; then
    echo -e "${YELLOW}Warning: Could not read token secret from config${NC}"
    echo "You may need to test manually"
else
    echo "Testing API access..."
    
    # Test nodes endpoint
    if curl -k -s -f \
        -H "Authorization: PVEAPIToken=root@pam!proxicloud=${TOKEN_SECRET}" \
        "https://localhost:8006/api2/json/nodes" > /dev/null; then
        echo -e "${GREEN}[✓]${NC} Can access /nodes"
    else
        echo -e "${RED}[✗]${NC} Cannot access /nodes"
    fi
    
    # Test containers endpoint
    NODE=$(hostname)
    if curl -k -s -f \
        -H "Authorization: PVEAPIToken=root@pam!proxicloud=${TOKEN_SECRET}" \
        "https://localhost:8006/api2/json/nodes/${NODE}/lxc" > /dev/null; then
        echo -e "${GREEN}[✓]${NC} Can access /nodes/${NODE}/lxc"
        
        # Count containers
        CONTAINER_COUNT=$(curl -k -s \
            -H "Authorization: PVEAPIToken=root@pam!proxicloud=${TOKEN_SECRET}" \
            "https://localhost:8006/api2/json/nodes/${NODE}/lxc" | \
            python3 -c "import sys, json; data=json.load(sys.stdin); print(len(data.get('data', [])))")
        
        echo -e "${GREEN}[✓]${NC} Found ${CONTAINER_COUNT} containers via API"
    else
        echo -e "${RED}[✗]${NC} Cannot access /nodes/${NODE}/lxc"
    fi
fi

echo ""
echo -e "${GREEN}=== Summary ===${NC}"
echo ""
echo "Token configuration updated:"
echo "  • Privilege separation: DISABLED (privsep=0)"
echo "  • Role: Administrator"
echo "  • Path: / (full access)"
echo ""
echo "Your token should now be able to:"
echo "  ✓ List all nodes"
echo "  ✓ List all containers"
echo "  ✓ Create/start/stop/delete containers"
echo "  ✓ Access all Proxmox resources"
echo ""
echo -e "${YELLOW}Next steps:${NC}"
echo "1. Test with: ./test-connection.sh /etc/proxicloud/config.yaml"
echo "2. Or run: ./find-node.sh (should now show containers)"
echo "3. Or restart ProxiCloud: systemctl restart proxicloud-api"
echo ""
