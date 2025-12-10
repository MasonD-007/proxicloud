package proxmox

import "encoding/json"

// Container represents an LXC container
type Container struct {
	VMID      int     `json:"vmid"`
	Name      string  `json:"name"`
	Status    string  `json:"status"`
	Node      string  `json:"node"`
	CPU       float64 `json:"cpu"`
	Mem       int64   `json:"mem"`
	MaxMem    int64   `json:"maxmem"`
	Disk      int64   `json:"disk"`
	MaxDisk   int64   `json:"maxdisk"`
	Uptime    int64   `json:"uptime"`
	Template  string  `json:"template,omitempty"`
	OS        string  `json:"os,omitempty"`
	ProjectID string  `json:"project_id,omitempty"` // Associated project ID
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
	ProjectID    string `json:"project_id,omitempty"` // Optional: assign to project
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

// Volume represents a persistent storage volume (ZFS zvol)
type Volume struct {
	VolID      string `json:"volid"`
	Name       string `json:"name"`
	Size       int64  `json:"size"` // Size in GB
	Used       int64  `json:"used"` // Used space in GB
	Node       string `json:"node"`
	Storage    string `json:"storage"`               // Storage pool (e.g., local-lvm, local-zfs)
	Type       string `json:"type"`                  // ssd or hdd
	Format     string `json:"format"`                // raw, qcow2, etc.
	Status     string `json:"status"`                // available, in-use, error
	AttachedTo *int   `json:"attached_to,omitempty"` // VMID if attached
	MountPoint string `json:"mountpoint,omitempty"`  // Mount point if attached (mp0-mp9)
	CreatedAt  int64  `json:"created_at,omitempty"`  // Unix timestamp
}

// CreateVolumeRequest holds parameters for creating a new volume
type CreateVolumeRequest struct {
	Name    string `json:"name"`
	Size    int    `json:"size"`           // Size in GB
	Storage string `json:"storage"`        // Storage pool (default: local-lvm)
	Type    string `json:"type"`           // ssd or hdd (default: ssd)
	Node    string `json:"node,omitempty"` // Optional: specific node
}

// AttachVolumeRequest holds parameters for attaching a volume to a container
type AttachVolumeRequest struct {
	VMID       int    `json:"vmid"`
	MountPoint string `json:"mountpoint,omitempty"` // Optional: mp0-mp9 (auto-detect if not provided)
}

// DetachVolumeRequest holds parameters for detaching a volume
type DetachVolumeRequest struct {
	VMID  int  `json:"vmid"`
	Force bool `json:"force,omitempty"` // Force detach even if container is running
}

// Snapshot represents a volume snapshot
type Snapshot struct {
	Name        string `json:"name"`
	VolID       string `json:"volid"`
	Description string `json:"description,omitempty"`
	CreatedAt   int64  `json:"created_at"`
	Size        int64  `json:"size"`             // Snapshot size in GB
	Parent      string `json:"parent,omitempty"` // Parent snapshot name
}

// CreateSnapshotRequest holds parameters for creating a volume snapshot
type CreateSnapshotRequest struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
}

// RestoreSnapshotRequest holds parameters for restoring a volume from snapshot
type RestoreSnapshotRequest struct {
	SnapshotName string `json:"snapshot_name"`
}

// CloneSnapshotRequest holds parameters for cloning a volume from snapshot
type CloneSnapshotRequest struct {
	SnapshotName string `json:"snapshot_name"`
	NewName      string `json:"new_name"`
	Storage      string `json:"storage,omitempty"` // Optional: different storage pool
}

// Project represents a logical grouping of containers
type Project struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Description string   `json:"description,omitempty"`
	Tags        []string `json:"tags,omitempty"`
	CreatedAt   int64    `json:"created_at"`
	UpdatedAt   int64    `json:"updated_at"`
}

// CreateProjectRequest holds parameters for creating a new project
type CreateProjectRequest struct {
	Name        string   `json:"name"`
	Description string   `json:"description,omitempty"`
	Tags        []string `json:"tags,omitempty"`
}

// UpdateProjectRequest holds parameters for updating a project
type UpdateProjectRequest struct {
	Name        string   `json:"name,omitempty"`
	Description string   `json:"description,omitempty"`
	Tags        []string `json:"tags,omitempty"`
}

// AssignProjectRequest holds parameters for assigning a container to a project
type AssignProjectRequest struct {
	ProjectID string `json:"project_id"` // Empty string means "No Project"
}

// Storage represents a Proxmox storage datastore
type Storage struct {
	Storage      string   `json:"storage"`                 // The storage identifier
	Type         string   `json:"type"`                    // Storage type
	Content      string   `json:"content"`                 // Allowed storage content types
	Active       *bool    `json:"active,omitempty"`        // Set when storage is accessible
	Enabled      *bool    `json:"enabled,omitempty"`       // Set when storage is enabled (not disabled)
	Avail        *int64   `json:"avail,omitempty"`         // Available storage space in bytes
	Total        *int64   `json:"total,omitempty"`         // Total storage space in bytes
	Used         *int64   `json:"used,omitempty"`          // Used storage space in bytes
	UsedFraction *float64 `json:"used_fraction,omitempty"` // Used fraction (used/total)
	Shared       *bool    `json:"shared,omitempty"`        // Shared flag from storage configuration
}

// GetStorageRequest holds optional parameters for querying storage
type GetStorageRequest struct {
	Content string `json:"content,omitempty"` // Only list stores which support this content type
	Enabled *bool  `json:"enabled,omitempty"` // Only list stores which are enabled
	Format  *bool  `json:"format,omitempty"`  // Include information about formats
	Storage string `json:"storage,omitempty"` // Only list status for specified storage
	Target  string `json:"target,omitempty"`  // If target is different to 'node', we only lists shared storages
}

// ProxmoxResponse is the generic response from Proxmox API
type ProxmoxResponse struct {
	Data json.RawMessage `json:"data"`
}
