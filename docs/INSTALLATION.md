# Installation Guide

This guide will walk you through installing ProxiCloud on your Proxmox VE node.

---

## 📋 Prerequisites

### Proxmox VE Requirements
- **Proxmox VE 7.0 or higher**
- **Root access** to the Proxmox node (or sudo privileges)
- **Available ports**: 3000 (frontend), 8080 (backend API)
- **Internet connection** (for downloading binaries during installation)

### Network Requirements
- ProxiCloud must be installed **directly on a Proxmox node** (not a remote machine)
- The web interface will be accessible from any machine on your LAN
- Ensure firewall allows incoming connections on port 3000

---

## 🚀 Quick Installation (Recommended)

### One-Line Installer

SSH into your Proxmox node as root and run:

```bash
bash <(curl -fsSL https://raw.githubusercontent.com/yourusername/proxicloud/main/deploy/install.sh)
```

Or download and inspect the script first:

```bash
curl -fsSL https://raw.githubusercontent.com/yourusername/proxicloud/main/deploy/install.sh -o install.sh
chmod +x install.sh
./install.sh
```

The installer will guide you through the setup process.

---

## 📝 Installation Steps (Detailed)

### Step 1: Create Proxmox API Token

Before running the installer, you need to create an API token in Proxmox.

1. **Log into Proxmox Web UI** (https://your-proxmox:8006)

2. **Navigate to API Tokens**:
   ```
   Datacenter → Permissions → API Tokens
   ```

3. **Click "Add" button**

4. **Fill in the form**:
   - **User**: `root@pam`
   - **Token ID**: `proxicloud` (or any name you prefer)
   - **Privilege Separation**: ❌ **Uncheck this box** (important!)
   - This allows the token to use the full privileges of the root user

5. **Click "Add"**

6. **Copy the token secret** - You'll see something like:
   ```
   root@pam!proxicloud=xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx
   ```
   
   ⚠️ **Important**: Save this token! You won't be able to see it again.

#### Alternative: Non-Root Token (Advanced)

If you prefer not to use a root token, create a dedicated user with specific privileges:

1. Create a new user:
   ```
   Datacenter → Permissions → Users → Add
   ```
   - **User name**: `proxicloud`
   - **Realm**: `Proxmox VE authentication server`
   - **Password**: (set a password)

2. Create a role with required privileges:
   ```
   Datacenter → Permissions → Roles → Create
   ```
   - **Name**: `ProxiCloud`
   - **Privileges**:
     - `VM.Allocate`
     - `VM.Config.Disk`
     - `VM.Config.Memory`
     - `VM.Config.Network`
     - `VM.PowerMgmt`
     - `VM.Audit`
     - `Datastore.Allocate`
     - `Datastore.AllocateSpace`

3. Assign the role to the user:
   ```
   Datacenter → Permissions → Add → User Permission
   ```
   - **Path**: `/`
   - **User**: `proxicloud@pve`
   - **Role**: `ProxiCloud`

4. Create API token for this user:
   ```
   Datacenter → Permissions → API Tokens → Add
   ```
   - **User**: `proxicloud@pve`
   - **Token ID**: `token1`

---

### Step 2: Run the Installer

SSH into your Proxmox node:

```bash
ssh root@your-proxmox-ip
```

Run the installer:

```bash
bash <(curl -fsSL https://raw.githubusercontent.com/yourusername/proxicloud/main/deploy/install.sh)
```

#### Installer Process

The installer will:

1. **Pre-flight checks**
   - Verify you're running on Proxmox VE
   - Check if ports 3000 and 8080 are available
   - Verify root/sudo access

2. **Node detection**
   - Auto-detect your Proxmox node name
   - Display detected configuration for confirmation

3. **API Token input**
   - Prompt you to paste your API token
   - Validate token format

4. **Download binaries**
   - Fetch the latest release from GitHub
   - Verify checksums
   - Extract to `/opt/proxicloud/`

5. **Configuration**
   - Generate `/etc/proxicloud/config.yaml`
   - Set secure file permissions (600)

6. **Database initialization**
   - Create `/var/lib/proxicloud/` directory
   - Initialize SQLite database schema

7. **Install systemd services**
   - Create `proxicloud-api.service`
   - Create `proxicloud-frontend.service`
   - Enable auto-start on boot
   - Start services

8. **Verification**
   - Test API health endpoint
   - Test frontend accessibility
   - Display success message

#### Expected Output

```
╔════════════════════════════════════════════════╗
║         ProxiCloud Installer v1.0              ║
╚════════════════════════════════════════════════╝

[✓] Running on Proxmox VE 8.1.3
[✓] Ports 3000 and 8080 are available
[✓] Detected node name: pve

Please paste your Proxmox API token:
(Format: user@realm!tokenid=secret)
> root@pam!proxicloud=xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx

[✓] Token validated
[✓] Downloading binaries from GitHub...
[✓] Installing to /opt/proxicloud/
[✓] Generating configuration...
[✓] Initializing database...
[✓] Installing systemd services...
[✓] Starting services...

╔════════════════════════════════════════════════╗
║     ProxiCloud Installed Successfully! 🎉      ║
╚════════════════════════════════════════════════╝

Access your console at:
  → http://192.168.1.100:3000

Service Status:
  Backend API:   ✓ Running (port 8080)
  Frontend:      ✓ Running (port 3000)

Logs:
  journalctl -u proxicloud-api -f
  journalctl -u proxicloud-frontend -f

Configuration:
  /etc/proxicloud/config.yaml

To uninstall:
  /opt/proxicloud/scripts/uninstall.sh
```

---

## 🔧 Post-Installation

### Verify Installation

Check service status:

```bash
systemctl status proxicloud-api
systemctl status proxicloud-frontend
```

Both services should show as `active (running)`.

### Access the Web Interface

Open your browser and navigate to:

```
http://YOUR-PROXMOX-IP:3000
```

You should see the ProxiCloud dashboard.

### View Logs

Backend API logs:
```bash
journalctl -u proxicloud-api -f
```

Frontend logs:
```bash
journalctl -u proxicloud-frontend -f
```

Application logs:
```bash
tail -f /var/log/proxicloud/app.log
```

---

## ⚙️ Configuration

The main configuration file is located at:
```
/etc/proxicloud/config.yaml
```

See [CONFIGURATION.md](CONFIGURATION.md) for all available options.

### Common Configuration Changes

#### Change Ports

Edit `/etc/proxicloud/config.yaml`:

```yaml
server:
  backend_port: 8080   # Change this
  frontend_port: 3000  # Change this
```

Restart services:
```bash
systemctl restart proxicloud-api
systemctl restart proxicloud-frontend
```

#### Change Metrics Collection Interval

```yaml
analytics:
  interval: "30s"  # Change to "60s", "5m", etc.
```

Restart backend:
```bash
systemctl restart proxicloud-api
```

#### Change Data Retention

```yaml
analytics:
  retention_days: 30  # Change to desired number of days
```

---

## 🛡️ Security Best Practices

### 1. Secure the Configuration File

The config file contains your API token. Ensure proper permissions:

```bash
chmod 600 /etc/proxicloud/config.yaml
chown root:root /etc/proxicloud/config.yaml
```

### 2. Firewall Configuration

If you have a firewall enabled, allow access to port 3000:

**Using `ufw`**:
```bash
ufw allow 3000/tcp comment "ProxiCloud Web UI"
```

**Using `iptables`**:
```bash
iptables -A INPUT -p tcp --dport 3000 -j ACCEPT
iptables-save > /etc/iptables/rules.v4
```

**Proxmox Web UI**:
1. Go to: `Node → Firewall → Add`
2. Set: Direction=In, Action=ACCEPT, Protocol=tcp, Dest. port=3000

### 3. Restrict Access by IP (Optional)

To only allow specific IPs to access ProxiCloud:

**Using `ufw`**:
```bash
ufw allow from 192.168.1.0/24 to any port 3000
```

**Using `iptables`**:
```bash
iptables -A INPUT -p tcp -s 192.168.1.0/24 --dport 3000 -j ACCEPT
iptables -A INPUT -p tcp --dport 3000 -j DROP
```

### 4. Enable HTTPS (Future Release)

HTTPS support is planned for v2.0. For now, use a reverse proxy like Nginx if HTTPS is required.

---

## 🔄 Updating ProxiCloud

### Automatic Update (Planned)

A built-in update mechanism is planned for v1.1.

### Manual Update

1. Download the latest binaries:
```bash
cd /tmp
curl -LO https://github.com/yourusername/proxicloud/releases/latest/download/proxicloud-api-linux-amd64
curl -LO https://github.com/yourusername/proxicloud/releases/latest/download/proxicloud-frontend-linux.tar.gz
```

2. Stop services:
```bash
systemctl stop proxicloud-api proxicloud-frontend
```

3. Backup current installation:
```bash
cp /opt/proxicloud/api /opt/proxicloud/api.backup
tar -czf /opt/proxicloud/frontend.backup.tar.gz -C /opt/proxicloud/frontend .
```

4. Install new binaries:
```bash
mv proxicloud-api-linux-amd64 /opt/proxicloud/api
chmod +x /opt/proxicloud/api

cd /opt/proxicloud/frontend
tar -xzf /tmp/proxicloud-frontend-linux.tar.gz
```

5. Restart services:
```bash
systemctl start proxicloud-api proxicloud-frontend
```

6. Verify:
```bash
systemctl status proxicloud-api proxicloud-frontend
```

---

## 🗑️ Uninstallation

### Using Uninstall Script

```bash
/opt/proxicloud/scripts/uninstall.sh
```

The script will prompt you whether to keep or delete:
- Configuration file
- Analytics database
- Log files

### Manual Uninstallation

1. Stop and disable services:
```bash
systemctl stop proxicloud-api proxicloud-frontend
systemctl disable proxicloud-api proxicloud-frontend
```

2. Remove service files:
```bash
rm /etc/systemd/system/proxicloud-api.service
rm /etc/systemd/system/proxicloud-frontend.service
systemctl daemon-reload
```

3. Remove binaries:
```bash
rm -rf /opt/proxicloud
```

4. Remove configuration (optional):
```bash
rm -rf /etc/proxicloud
```

5. Remove database (optional):
```bash
rm -rf /var/lib/proxicloud
```

6. Remove logs (optional):
```bash
rm -rf /var/log/proxicloud
```

---

## 🐛 Troubleshooting

### Services Won't Start

**Check service status**:
```bash
systemctl status proxicloud-api
systemctl status proxicloud-frontend
```

**Check logs**:
```bash
journalctl -u proxicloud-api -n 50
journalctl -u proxicloud-frontend -n 50
```

**Common issues**:
- Port already in use → Change ports in config
- Invalid API token → Check token format in config.yaml
- Missing permissions → Verify token has required privileges

### Can't Access Web Interface

**Check if frontend is running**:
```bash
curl http://localhost:3000
```

**Check firewall**:
```bash
# Proxmox firewall
pveversion
iptables -L -n | grep 3000

# Test from another machine
telnet your-proxmox-ip 3000
```

**Check bind address**:
Ensure `config.yaml` has:
```yaml
server:
  bind_address: "0.0.0.0"  # Not 127.0.0.1
```

### API Connection Errors

**Test API health**:
```bash
curl http://localhost:8080/api/health
```

**Test Proxmox connectivity**:
```bash
curl -k -H "Authorization: PVEAPIToken=YOUR-TOKEN" \
  https://localhost:8006/api2/json/nodes
```

**Check API token**:
- Token format: `user@realm!tokenid=secret`
- Privilege separation must be disabled
- Token must have required permissions

### Database Errors

**Check database file**:
```bash
ls -lh /var/lib/proxicloud/analytics.db
sqlite3 /var/lib/proxicloud/analytics.db ".tables"
```

**Reset database**:
```bash
systemctl stop proxicloud-api
rm /var/lib/proxicloud/analytics.db
systemctl start proxicloud-api
# Database will be recreated automatically
```

### High Resource Usage

**Check metrics collection interval**:
```yaml
analytics:
  interval: "30s"  # Increase to "60s" or "120s" if needed
```

**Check database size**:
```bash
du -sh /var/lib/proxicloud/analytics.db
```

**Manually clean old data**:
```bash
sqlite3 /var/lib/proxicloud/analytics.db \
  "DELETE FROM metrics WHERE timestamp < strftime('%s', 'now', '-7 days');"
```

---

## 📞 Getting Help

- **Documentation**: Check other docs in [docs/](.)
- **GitHub Issues**: [Report a bug](https://github.com/yourusername/proxicloud/issues)
- **Discussions**: [Ask questions](https://github.com/yourusername/proxicloud/discussions)

---

## ✅ Next Steps

After installation:
1. [Configure ProxiCloud](CONFIGURATION.md) - Customize settings
2. [Explore the API](API.md) - Integrate with scripts or automation
3. [Learn about development](DEVELOPMENT.md) - Contribute to the project
