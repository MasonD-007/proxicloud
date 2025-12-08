# ProxiCloud Debugging Guide

## Issue: Empty Container List

### Symptoms
Your Proxmox API returns `{"data":[]}` even though you have 2 LXC containers running.

### Debug Output Analysis
```
[DEBUG] GetContainers: requesting path=/nodes/server/lxc
[DEBUG] Proxmox API Response: status=200, body_length=11 bytes
[DEBUG] Response body preview: {"data":[]}
```

**Key Observations:**
- ✅ HTTP 200 (authentication is working)
- ✅ Node name in path is "server" (not "pve")
- ❌ Empty data array (containers not found)

### Root Cause
**Node name mismatch** - The config file being used has a different node name than expected.

The debug log shows the request goes to `/nodes/server/lxc`, which means:
1. Your actual Proxmox node name is "**server**"
2. The config file being used in production has `node: "server"`
3. But your `config.test.yaml` has `node: "pve"`

### Solution

You need to find and update the correct config file being used by your running service.

## Diagnostic Scripts

### 1. Find Your Node Name
```bash
./deploy/scripts/find-node.sh
```

This script will:
- Connect to your Proxmox server
- List all available nodes
- Show their status and resources
- List containers on each node
- Tell you the exact node name to use

**Interactive prompts:**
- Proxmox host (e.g., 192.168.10.69)
- Token ID
- Token secret
- SSL verification preference

### 2. Test Connection
```bash
./deploy/scripts/test-connection.sh /path/to/config.yaml
```

This script will:
- Test API connectivity
- Verify node configuration
- List containers
- Show storage availability

### 3. Development Server
```bash
CONFIG_FILE=/etc/proxicloud/config.yaml ./deploy/scripts/dev.sh
```

This script will:
- Run backend and frontend from source
- Show all errors in real-time
- Use the specified config file
- Enable hot reload

## Quick Fix Steps

### Step 1: Find Your Actual Node Name
```bash
# On your Proxmox node
./deploy/scripts/find-node.sh
```

### Step 2: Locate Your Running Config
```bash
# Check which config file the service is using
ps aux | grep proxicloud-api

# Or check the systemd service
systemctl cat proxicloud-api | grep CONFIG_PATH
```

### Step 3: Update the Config
```bash
# Edit the config file being used
sudo nano /etc/proxicloud/config.yaml

# Change this line to match your actual node name:
# node: "pve"  # Change this to "server"
```

### Step 4: Restart Service
```bash
sudo systemctl restart proxicloud-api
sudo systemctl restart proxicloud-frontend
```

### Step 5: Verify
```bash
# Check logs
journalctl -u proxicloud-api -f

# Or test directly
curl http://localhost:8080/api/containers
```

## Common Node Names

- **pve** - Default Proxmox installation
- **server** - Custom hostname
- **proxmox** - Common custom name
- **pve01**, **pve02** - Cluster nodes

## How to Find Node Name Manually

### Method 1: From Proxmox Web UI
1. Log into Proxmox web interface
2. Look at the left sidebar under "Datacenter"
3. The node names are listed there

### Method 2: From Proxmox CLI
```bash
# SSH into your Proxmox host
ssh root@192.168.10.69

# Check hostname
hostname

# List nodes
pvesh get /nodes

# List containers on specific node
pvesh get /nodes/YOUR_NODE_NAME/lxc
```

### Method 3: From API
```bash
# Get list of all nodes
curl -k -H "Authorization: PVEAPIToken=root@pam!proxicloud=YOUR_SECRET" \
  https://192.168.10.69:8006/api2/json/nodes

# The response will show all node names in the "data" array
```

## Configuration File Locations

Common locations where config files might be:

1. **System-wide**: `/etc/proxicloud/config.yaml`
2. **Development**: `./config.test.yaml`
3. **Custom location**: Set via `CONFIG_PATH` environment variable
4. **Service override**: Check systemd service file

To find which config is being used:
```bash
# Check environment variable
echo $CONFIG_PATH

# Check systemd service
systemctl show proxicloud-api | grep CONFIG_PATH

# Check running process
ps aux | grep proxicloud-api | grep -o 'CONFIG_PATH=[^ ]*'
```

## Verification Checklist

After updating the config, verify:

- [ ] Node name matches Proxmox hostname
- [ ] API returns containers: `curl http://localhost:8080/api/containers`
- [ ] Frontend shows containers at `http://localhost:3000`
- [ ] Logs show container count: `journalctl -u proxicloud-api | grep containers`

## Still Not Working?

If containers still don't show up:

1. **Check API Token Permissions**
   - Token needs `PVEAuditor` role or higher
   - "Privilege Separation" should be UNCHECKED

2. **Verify Containers Exist**
   ```bash
   ssh root@your-proxmox
   pct list
   ```

3. **Check Container Type**
   - ProxiCloud only shows **LXC** containers
   - VMs (QEMU) are not shown
   - Verify with: `qm list` (VMs) vs `pct list` (LXC)

4. **Test API Directly**
   ```bash
   curl -k -H "Authorization: PVEAPIToken=YOUR_TOKEN_ID=YOUR_SECRET" \
     https://YOUR_HOST:8006/api2/json/nodes/YOUR_NODE/lxc
   ```

## Debug Mode

For maximum debugging output:

```bash
# Run with debug logging
CONFIG_FILE=/etc/proxicloud/config.yaml \
DEBUG=true \
./deploy/scripts/dev.sh
```

This will show:
- Every API request
- Full response bodies
- Cache operations
- Database queries
- Frontend API calls

## Need More Help?

1. Run the diagnostic script: `./deploy/scripts/find-node.sh`
2. Check the logs: `journalctl -u proxicloud-api -n 100`
3. Test the API directly with curl
4. Verify container type (LXC vs VM)
5. Open an issue on GitHub with debug output
