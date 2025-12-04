package proxmox

import "encoding/json"

// Container represents an LXC container
type Container struct {
	VMID     int     `json:"vmid"`
	Name     string  `json:"name"`
	Status   string  `json:"status"`
	Node     string  `json:"node"`
	CPU      float64 `json:"cpu"`
	Mem      int64   `json:"mem"`
	MaxMem   int64   `json:"maxmem"`
	Disk     int64   `json:"disk"`
	MaxDisk  int64   `json:"maxdisk"`
	Uptime   int64   `json:"uptime"`
	Template string  `json:"template,omitempty"`
	OS       string  `json:"os,omitempty"`
}

// CreateContainerRequest holds parameters for creating a new container
type CreateContainerRequest struct {
	VMID         *int   `json:"vmid,omitempty"` // Optional: user can specify VMID
	Hostname     string `json:"hostname"`
	Cores        int    `json:"cores"`
	Memory       int    `json:"memory"`
	Disk         int    `json:"disk"`
	OSTemplate   string `json:"ostemplate"`
	Password     string `json:"password,omitempty"`
	SSHKeys      string `json:"ssh_keys,omitempty"`
	StartOnBoot  bool   `json:"start_on_boot,omitempty"`
	Unprivileged bool   `json:"unprivileged,omitempty"`
	// Network configuration
	IPAddress  string `json:"ip_address,omitempty"` // IP address with CIDR notation (e.g., "192.168.1.100/24")
	Gateway    string `json:"gateway,omitempty"`    // Gateway IP address (e.g., "192.168.1.1")
	Nameserver string `json:"nameserver,omitempty"` // DNS nameserver (e.g., "8.8.8.8")
}

// Template represents a container template
type Template struct {
	VolID   string `json:"volid"`
	Format  string `json:"format"`
	Size    int64  `json:"size"`
	Content string `json:"content"`
}

// ProxmoxResponse is the generic response from Proxmox API
type ProxmoxResponse struct {
	Data json.RawMessage `json:"data"`
}
