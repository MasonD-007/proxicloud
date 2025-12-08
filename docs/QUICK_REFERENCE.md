# ProxiCloud - Quick Reference Guide

## 🎯 Recent Updates

### Custom VMID Feature ✅ NEW!
Users can now optionally specify their own container ID (VMID) when creating containers.

- **Default behavior:** Leave VMID field empty for auto-assignment
- **Custom VMID:** Enter a number between 100-999999999
- **Validation:** Checks if VMID is already in use
- **See:** `CUSTOM_VMID_FEATURE.md` for full documentation

---

## 🐛 Bug Fixes Applied

### 1. Container Creation Error - FIXED ✅
**Error:** `json: cannot unmarshal string into Go struct field .data of type int`

**Fix:** Updated `GetNextVMID()` to handle both string and integer responses from Proxmox API

**Files Changed:**
- `backend/internal/proxmox/client.go`

### 2. POST Request Format - FIXED ✅
**Problem:** Proxmox API expects `application/x-www-form-urlencoded` for POST requests

**Fix:** Updated `doRequest()` to automatically convert POST/PUT requests to form-encoded format

**Files Changed:**
- `backend/internal/proxmox/client.go`

---

## 🚀 Quick Start

### Starting the Backend

```bash
# Method 1: Using systemd (Production)
sudo systemctl start proxicloud-api
sudo systemctl status proxicloud-api

# Method 2: Manual start (Development)
cd backend
export CONFIG_PATH="/etc/proxicloud/config.yaml"
export CACHE_PATH="/var/lib/proxicloud/cache.db"
export ANALYTICS_PATH="/var/lib/proxicloud/analytics.db"
./bin/api
```

### Starting the Frontend

```bash
cd frontend
npm run dev
# or for production
npm run build && npm start
```

### Access the Application

- **Frontend:** http://YOUR_SERVER:3000
- **API:** http://YOUR_SERVER:8080/api
- **Health Check:** http://YOUR_SERVER:8080/api/health

---

## 📝 Configuration

### config.yaml Location

- **Production:** `/etc/proxicloud/config.yaml`
- **Development:** `./config.test.yaml`

### Required Settings

```yaml
server:
  port: 8080
  host: "0.0.0.0"

proxmox:
  host: "https://YOUR_PROXMOX_IP:8006"
  node: "YOUR_NODE_NAME"
  token_id: "root@pam!proxicloud"
  token_secret: "YOUR_SECRET_TOKEN"
  insecure: true  # Set to false if using valid SSL certificate
```

---

## 🔧 Common Tasks

### Rebuild Backend

```bash
cd backend
make build
```

### Rebuild Frontend

```bash
cd frontend
npm run build
```

### View Logs

```bash
# Backend logs (systemd)
sudo journalctl -u proxicloud-api -f

# Backend logs (manual)
tail -f /tmp/proxicloud-api.log

# Frontend logs
npm run dev  # Shows logs in console
```

### Restart Services

```bash
# Restart both services
sudo systemctl restart proxicloud-api
sudo systemctl restart proxicloud-frontend

# Check status
sudo systemctl status proxicloud-api proxicloud-frontend
```

---

## 🎨 Features

### Container Management
- ✅ List all containers
- ✅ View container details
- ✅ Create new containers (with auto or custom VMID) ⭐ NEW!
- ✅ Start/Stop/Reboot containers
- ✅ Delete containers
- ✅ Real-time status updates

### Analytics
- ✅ CPU usage metrics
- ✅ Memory usage tracking
- ✅ Disk I/O monitoring
- ✅ Network traffic stats
- ✅ 30-day data retention

### System Features
- ✅ Offline mode with cache
- ✅ Auto-retry with backoff
- ✅ Dark theme UI
- ✅ Responsive design
- ✅ Error handling

---

## 🔍 Troubleshooting

### Backend Won't Start

**Check 1: Configuration file exists**
```bash
ls -la /etc/proxicloud/config.yaml
```

**Check 2: Proxmox credentials are correct**
```bash
# Test Proxmox connection
curl -k -H "Authorization: PVEAPIToken=USER@REALM!TOKENID=SECRET" \
  https://YOUR_PROXMOX:8006/api2/json/cluster/resources
```

**Check 3: Ports are available**
```bash
sudo lsof -i :8080
sudo lsof -i :3000
```

### Frontend Shows "Network Error"

**Check 1: Backend is running**
```bash
curl http://localhost:8080/api/health
# Should return: {"status":"healthy"}
```

**Check 2: CORS is configured**
- Backend automatically allows all origins
- Check browser console for CORS errors

**Check 3: API URL is correct**
- Frontend uses: `${window.location.protocol}//${window.location.hostname}:8080/api`
- Make sure backend is accessible on port 8080

### Container Creation Fails

**Check 1: Proxmox permissions**
- API token needs VM.Allocate permission
- Check token permissions in Proxmox UI

**Check 2: Storage available**
```bash
# Check storage in Proxmox
pvesm status
```

**Check 3: Template exists**
- Verify template is downloaded in Proxmox
- Check storage for .tar.gz or .tar.zst files

**Check 4: Custom VMID conflicts** ⭐ NEW!
- If using custom VMID, ensure it's not already in use
- Try auto-assign by leaving VMID field empty
- Error message will show: "VMID XXX is already in use"

### Empty Container List

**Check 1: Node name is correct**
```bash
# List nodes
pvesh get /nodes
```

**Check 2: Containers exist**
```bash
# List containers manually
pct list
```

**Check 3: API token has permissions**
- Token needs: VM.Audit, VM.Allocate, VM.Config.Disk, VM.Config.Memory, VM.PowerMgmt

---

## 📊 API Endpoints

### Health & Dashboard
- `GET /api/health` - Health check
- `GET /api/dashboard` - Dashboard statistics

### Containers
- `GET /api/containers` - List all containers
- `GET /api/containers/{vmid}` - Get container details
- `POST /api/containers` - Create new container (optionally with custom VMID) ⭐ NEW!
- `DELETE /api/containers/{vmid}` - Delete container
- `POST /api/containers/{vmid}/start` - Start container
- `POST /api/containers/{vmid}/stop` - Stop container
- `POST /api/containers/{vmid}/reboot` - Reboot container

### Templates
- `GET /api/templates` - List available templates

### Analytics
- `GET /api/analytics/stats` - Analytics statistics
- `GET /api/containers/{vmid}/metrics` - Container metrics
- `GET /api/containers/{vmid}/metrics/summary` - Metrics summary

---

## 🎓 Best Practices

### Container Naming
- Use lowercase letters, numbers, and hyphens only
- Start with a letter
- Keep names short and descriptive
- Example: `web-server-01`, `db-primary`, `cache-redis`

### VMID Organization ⭐ NEW!
With custom VMID support, you can organize containers by ID range:
- **100-199:** Web servers
- **200-299:** Databases
- **300-399:** Cache/Queue systems
- **400-499:** Development
- **500-599:** Staging
- **600-699:** Monitoring
- **700-799:** Backups
- **800-899:** Testing

### Resource Allocation
- **Small workload:** 1-2 cores, 512-1024 MB RAM, 8-16 GB disk
- **Medium workload:** 2-4 cores, 2-4 GB RAM, 32-64 GB disk
- **Large workload:** 4-8 cores, 8-16 GB RAM, 128-256 GB disk

### Security
- ✅ Use unprivileged containers when possible
- ✅ Set strong root passwords or use SSH keys
- ✅ Don't start containers on boot unless necessary
- ✅ Regularly update container templates
- ✅ Use API tokens instead of username/password

---

## 📚 Documentation Files

- `README.md` - Project overview
- `QUICK_REFERENCE.md` - This file
- `FIXES_APPLIED.md` - Recent bug fixes
- `CUSTOM_VMID_FEATURE.md` - Custom VMID feature documentation ⭐ NEW!
- `PROJECT_SUMMARY.md` - Complete project status
- `docs/INSTALLATION.md` - Installation guide
- `docs/API.md` - API documentation
- `docs/CONFIGURATION.md` - Configuration reference

---

## 🆘 Getting Help

### Check Logs
1. Backend: `journalctl -u proxicloud-api -n 100`
2. Frontend: Browser DevTools Console
3. Proxmox: `/var/log/pve/tasks/`

### Common Log Messages

**Good:**
```
[INFO] Successfully retrieved N containers from Proxmox
[INFO] Using auto-generated VMID: 150
[INFO] Using user-specified VMID: 200  ⭐ NEW!
[DEBUG] Successfully parsed nextid as string and converted: 150
```

**Warnings:**
```
[WARNING] Proxmox returned empty container list
[INFO] Serving containers from cache (Proxmox error: ...)
```

**Errors:**
```
[ERROR] Failed to parse nextid response
[ERROR] Proxmox API request failed
[ERROR] GetContainers failed: context deadline exceeded
[ERROR] VMID 200 is already in use  ⭐ NEW!
```

---

## 🔄 Update Process

1. **Pull latest changes:**
   ```bash
   git pull origin main
   ```

2. **Rebuild backend:**
   ```bash
   cd backend && make build
   ```

3. **Rebuild frontend:**
   ```bash
   cd frontend && npm install && npm run build
   ```

4. **Restart services:**
   ```bash
   sudo systemctl restart proxicloud-api
   sudo systemctl restart proxicloud-frontend
   ```

---

**Last Updated:** December 4, 2025
**Version:** 1.0.0
**Status:** ✅ Production Ready with Custom VMID Support
