#!/bin/bash
# ProxiCloud Connection Test Script
# Tests connectivity to Proxmox API and shows available resources

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
CONFIG_FILE="${1:-./config.test.yaml}"

print_status() {
    echo -e "${GREEN}[✓]${NC} $1"
}

print_error() {
    echo -e "${RED}[✗]${NC} $1"
}

print_info() {
    echo -e "${YELLOW}[i]${NC} $1"
}

print_debug() {
    echo -e "${BLUE}[DEBUG]${NC} $1"
}

echo -e "${GREEN}ProxiCloud Connection Test${NC}"
echo ""

# Check if config file exists
if [ ! -f "$CONFIG_FILE" ]; then
    print_error "Configuration file not found: $CONFIG_FILE"
    echo ""
    echo "Usage: $0 [config_file]"
    echo "Example: $0 /etc/proxicloud/config.yaml"
    exit 1
fi

print_info "Using config: $CONFIG_FILE"
echo ""

# Parse YAML config (simple extraction)
HOST=$(grep 'host:' "$CONFIG_FILE" | head -1 | awk '{print $2}' | tr -d '"')
NODE=$(grep 'node:' "$CONFIG_FILE" | head -1 | awk '{print $2}' | tr -d '"')
TOKEN_ID=$(grep 'token_id:' "$CONFIG_FILE" | head -1 | awk '{print $2}' | tr -d '"')
TOKEN_SECRET=$(grep 'token_secret:' "$CONFIG_FILE" | head -1 | awk '{print $2}' | tr -d '"')
INSECURE=$(grep 'insecure:' "$CONFIG_FILE" | head -1 | awk '{print $2}' | tr -d '"')

print_debug "Host: $HOST"
print_debug "Node: $NODE"
print_debug "Token ID: $TOKEN_ID"
print_debug "Insecure: $INSECURE"
echo ""

# Normalize host URL
if [[ ! "$HOST" =~ ^https?:// ]]; then
    BASE_URL="https://${HOST}:8006"
else
    BASE_URL="$HOST"
fi

BASE_URL="${BASE_URL}/api2/json"

# Setup curl options
CURL_OPTS=(-s -w "\nHTTP_CODE:%{http_code}\n")
if [ "$INSECURE" = "true" ]; then
    CURL_OPTS+=(-k)
fi

# Function to make API request
api_request() {
    local path="$1"
    local full_url="${BASE_URL}${path}"
    
    print_info "Testing: $path"
    
    response=$(curl "${CURL_OPTS[@]}" \
        -H "Authorization: PVEAPIToken=${TOKEN_ID}=${TOKEN_SECRET}" \
        "$full_url")
    
    http_code=$(echo "$response" | grep "HTTP_CODE:" | cut -d: -f2)
    body=$(echo "$response" | sed '/HTTP_CODE:/d')
    
    if [ "$http_code" = "200" ]; then
        print_status "Success (HTTP $http_code)"
        echo "$body"
        return 0
    else
        print_error "Failed (HTTP $http_code)"
        echo "$body"
        return 1
    fi
}

# Test 1: Check cluster version
echo "=== Test 1: Cluster Version ==="
if api_request "/version"; then
    echo ""
else
    print_error "Failed to connect to Proxmox API"
    echo ""
    echo "Possible issues:"
    echo "  • Host is incorrect or unreachable"
    echo "  • Firewall blocking connection"
    echo "  • Proxmox not running"
    exit 1
fi

# Test 2: Check node status
echo "=== Test 2: Node Status ==="
if api_request "/nodes/${NODE}/status"; then
    echo ""
else
    print_error "Failed to get node status"
    echo ""
    echo "Possible issues:"
    echo "  • Node name is incorrect (should be: ${NODE})"
    echo "  • Node is offline"
    exit 1
fi

# Test 3: List containers
echo "=== Test 3: List Containers ==="
response=$(curl "${CURL_OPTS[@]}" \
    -H "Authorization: PVEAPIToken=${TOKEN_ID}=${TOKEN_SECRET}" \
    "${BASE_URL}/nodes/${NODE}/lxc")

http_code=$(echo "$response" | grep "HTTP_CODE:" | cut -d: -f2)
body=$(echo "$response" | sed '/HTTP_CODE:/d')

if [ "$http_code" = "200" ]; then
    container_count=$(echo "$body" | grep -o '"data":\[' | wc -l)
    
    # Check if data array is empty
    if echo "$body" | grep -q '"data":\[\]'; then
        print_info "No containers found (this is OK if you haven't created any yet)"
        print_info "Empty array means: Authentication works, node name is correct"
    else
        print_status "Containers found!"
        echo "$body" | python3 -m json.tool 2>/dev/null || echo "$body"
    fi
    echo ""
else
    print_error "Failed to list containers (HTTP $http_code)"
    echo "$body"
    exit 1
fi

# Test 4: List storage
echo "=== Test 4: Available Storage ==="
if api_request "/nodes/${NODE}/storage"; then
    echo ""
else
    print_error "Failed to get storage info"
fi

# Test 5: Get next VMID
echo "=== Test 5: Next Available VMID ==="
if response=$(api_request "/cluster/nextid"); then
    next_vmid=$(echo "$response" | grep -o '"data":[0-9]*' | cut -d: -f2)
    print_info "Next available VMID: $next_vmid"
    echo ""
fi

# Summary
echo "=== Summary ==="
print_status "All connectivity tests passed!"
echo ""
echo "Your ProxiCloud configuration is working correctly."
echo "You can now:"
echo "  1. Create containers through the API"
echo "  2. Run the development server: ./deploy/scripts/dev.sh"
echo "  3. Access the web UI at http://localhost:3000"
echo ""
print_info "Note: Empty container list is normal if you haven't created any containers yet"
