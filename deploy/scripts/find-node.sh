#!/bin/bash
# Find the actual Proxmox node name
# This helps when you get empty container lists

set -e

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

print_status() {
    echo -e "${GREEN}[✓]${NC} $1"
}

print_error() {
    echo -e "${RED}[✗]${NC} $1"
}

print_info() {
    echo -e "${YELLOW}[i]${NC} $1"
}

echo -e "${GREEN}Proxmox Node Name Finder${NC}"
echo ""

# Get parameters
read -p "Enter Proxmox host (e.g., 192.168.10.69): " HOST

# Validate host is not empty
if [ -z "$HOST" ]; then
    print_error "Host cannot be empty"
    exit 1
fi

read -p "Enter token ID (e.g., root@pam!proxicloud): " TOKEN_ID

# Validate token ID is not empty
if [ -z "$TOKEN_ID" ]; then
    print_error "Token ID cannot be empty"
    exit 1
fi

read -sp "Enter token secret: " TOKEN_SECRET
echo ""

# Validate token secret is not empty
if [ -z "$TOKEN_SECRET" ]; then
    print_error "Token secret cannot be empty"
    exit 1
fi

read -p "Skip SSL verification? (y/n): " SKIP_SSL

# Setup curl options
CURL_OPTS=(-s)
if [ "$SKIP_SSL" = "y" ] || [ "$SKIP_SSL" = "Y" ]; then
    CURL_OPTS+=(-k)
fi

# Normalize host URL
if [[ ! "$HOST" =~ ^https?:// ]]; then
    BASE_URL="https://${HOST}:8006/api2/json"
else
    BASE_URL="${HOST}/api2/json"
fi

print_info "Connecting to: $BASE_URL"
echo ""

# Get list of nodes
print_info "Fetching node list..."
response=$(curl "${CURL_OPTS[@]}" \
    -H "Authorization: PVEAPIToken=${TOKEN_ID}=${TOKEN_SECRET}" \
    "${BASE_URL}/nodes" 2>&1)

if [ $? -ne 0 ]; then
    print_error "Failed to connect to Proxmox"
    echo "$response"
    exit 1
fi

# Check if response is valid JSON
if ! echo "$response" | python3 -m json.tool &>/dev/null; then
    print_error "Invalid response from Proxmox API"
    echo ""
    echo "Response:"
    echo "$response"
    exit 1
fi

print_status "Connected successfully!"
echo ""

# Parse nodes
echo "=== Available Nodes ==="
echo "$response" | python3 -c "
import sys, json
data = json.load(sys.stdin)
if 'data' in data:
    nodes = data['data']
    if len(nodes) == 0:
        print('No nodes found!')
    else:
        for node in nodes:
            name = node.get('node', 'unknown')
            status = node.get('status', 'unknown')
            online = '🟢' if status == 'online' else '🔴'
            print(f'{online} Node: {name} (Status: {status})')
            
            # Show node details
            if 'mem' in node and 'maxmem' in node:
                mem_used = node['mem'] / (1024**3)  # Convert to GB
                mem_total = node['maxmem'] / (1024**3)
                mem_pct = (node['mem'] / node['maxmem']) * 100
                print(f'   Memory: {mem_used:.1f}GB / {mem_total:.1f}GB ({mem_pct:.1f}%)')
            
            if 'cpu' in node and 'maxcpu' in node:
                cpu_pct = node['cpu'] * 100
                print(f'   CPU: {cpu_pct:.1f}% ({node[\"maxcpu\"]} cores)')
            
            if 'uptime' in node:
                days = node['uptime'] // 86400
                hours = (node['uptime'] % 86400) // 3600
                print(f'   Uptime: {days}d {hours}h')
            
            print()
else:
    print('Unexpected response format')
    print(json.dumps(data, indent=2))
"

# Get the first node name
NODE_NAME=$(echo "$response" | python3 -c "
import sys, json
data = json.load(sys.stdin)
if 'data' in data and len(data['data']) > 0:
    print(data['data'][0]['node'])
" 2>/dev/null)

if [ -n "$NODE_NAME" ]; then
    echo "=== Testing Node: $NODE_NAME ==="
    print_info "Checking for LXC containers on node '$NODE_NAME'..."
    
    containers=$(curl "${CURL_OPTS[@]}" \
        -H "Authorization: PVEAPIToken=${TOKEN_ID}=${TOKEN_SECRET}" \
        "${BASE_URL}/nodes/${NODE_NAME}/lxc")
    
    echo "$containers" | python3 -c "
import sys, json
data = json.load(sys.stdin)
if 'data' in data:
    containers = data['data']
    print(f'Found {len(containers)} container(s)')
    print()
    for ct in containers:
        vmid = ct.get('vmid', '?')
        name = ct.get('name', 'unknown')
        status = ct.get('status', 'unknown')
        icon = '🟢' if status == 'running' else '🔴'
        print(f'{icon} VMID {vmid}: {name} ({status})')
        
        if 'mem' in ct and 'maxmem' in ct:
            mem_mb = ct['mem'] / (1024**2)
            max_mb = ct['maxmem'] / (1024**2)
            print(f'   Memory: {mem_mb:.0f}MB / {max_mb:.0f}MB')
        
        if 'disk' in ct and 'maxdisk' in ct:
            disk_gb = ct['disk'] / (1024**3)
            max_gb = ct['maxdisk'] / (1024**3)
            print(f'   Disk: {disk_gb:.1f}GB / {max_gb:.1f}GB')
        print()
"
    
    echo ""
    echo "=== Configuration ==="
    print_status "Your Proxmox node name is: ${GREEN}${NODE_NAME}${NC}"
    echo ""
    echo "Update your config.yaml with:"
    echo ""
    echo "proxmox:"
    echo "  host: \"${HOST}\""
    echo "  node: \"${NODE_NAME}\""
    echo "  token_id: \"${TOKEN_ID}\""
    echo "  token_secret: \"${TOKEN_SECRET}\""
    echo "  insecure: true"
    echo ""
fi
