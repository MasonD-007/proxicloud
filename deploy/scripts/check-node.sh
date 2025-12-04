#!/bin/bash
# Quick script to check your actual Proxmox node name and see containers
# Run this directly on your Proxmox server

set -e

# Colors
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

echo -e "${GREEN}=== Proxmox Node Information ===${NC}"
echo ""

# Get hostname (this is usually the node name)
HOSTNAME=$(hostname)
echo -e "${BLUE}System Hostname:${NC} $HOSTNAME"
echo ""

# Check if pvesh command exists (we're on Proxmox)
if ! command -v pvesh &> /dev/null; then
    echo "Error: This script must be run on a Proxmox server"
    echo "pvesh command not found"
    exit 1
fi

# Get node list
echo -e "${GREEN}=== Available Nodes ===${NC}"
pvesh get /nodes --output-format json | python3 -c "
import sys, json
nodes = json.load(sys.stdin)
for node in nodes:
    name = node.get('node', 'unknown')
    status = node.get('status', 'unknown')
    icon = '🟢' if status == 'online' else '🔴'
    print(f'{icon} {name} (Status: {status})')
"
echo ""

# Get actual node name from pvesh
NODE_NAME=$(pvesh get /nodes --output-format json | python3 -c "
import sys, json
nodes = json.load(sys.stdin)
if len(nodes) > 0:
    print(nodes[0]['node'])
" 2>/dev/null || echo "$HOSTNAME")

echo -e "${BLUE}Detected Node Name:${NC} ${GREEN}$NODE_NAME${NC}"
echo ""

# List LXC containers
echo -e "${GREEN}=== LXC Containers on '$NODE_NAME' ===${NC}"
pvesh get /nodes/$NODE_NAME/lxc --output-format json 2>/dev/null | python3 -c "
import sys, json
try:
    containers = json.load(sys.stdin)
    if len(containers) == 0:
        print('No LXC containers found')
    else:
        print(f'Found {len(containers)} container(s):\n')
        for ct in containers:
            vmid = ct.get('vmid', '?')
            name = ct.get('name', 'unknown')
            status = ct.get('status', 'unknown')
            icon = '🟢' if status == 'running' else '🔴'
            print(f'{icon} VMID {vmid}: {name} ({status})')
            
            if 'mem' in ct and 'maxmem' in ct:
                mem_mb = ct['mem'] / (1024**2)
                max_mb = ct['maxmem'] / (1024**2)
                mem_pct = (ct['mem'] / ct['maxmem']) * 100
                print(f'   Memory: {mem_mb:.0f}MB / {max_mb:.0f}MB ({mem_pct:.1f}%)')
            
            if 'cpus' in ct:
                print(f'   CPUs: {ct[\"cpus\"]}')
            print()
except:
    print('Error reading container data')
" || pct list

echo ""
echo -e "${GREEN}=== Configuration for ProxiCloud ===${NC}"
echo ""
echo "Use this in your config.yaml:"
echo ""
echo "proxmox:"
echo "  host: \"$(hostname -I | awk '{print $1}')\""
echo "  node: \"${NODE_NAME}\""
echo "  token_id: \"root@pam!proxicloud\""
echo "  token_secret: \"YOUR_SECRET_HERE\""
echo "  insecure: true"
echo ""
