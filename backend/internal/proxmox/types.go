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
	IPAddress string  `json:"ip_address,omitempty"` // IP address (extracted from network config)
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
	VNetID     string `json:"-"`                    // Internal: VNet ID to use (set by handler from project)
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

// ProjectNetwork represents network configuration for a project
type ProjectNetwork struct {
	Subnet          string `json:"subnet,omitempty"`            // CIDR notation (e.g., "192.168.1.0/24")
	Gateway         string `json:"gateway,omitempty"`           // Gateway IP (e.g., "192.168.1.1")
	Nameserver      string `json:"nameserver,omitempty"`        // DNS server (e.g., "8.8.8.8")
	VLanTag         int    `json:"vlan_tag,omitempty"`          // Optional VLAN tag
	VNetID          string `json:"vnet_id,omitempty"`           // Proxmox VNet ID (set by system)
	Zone            string `json:"zone,omitempty"`              // SDN Zone (set by system)
	AutoCreatedZone bool   `json:"auto_created_zone,omitempty"` // Whether the zone was auto-created by us
}

// Project represents a logical grouping of containers
type Project struct {
	ID               string          `json:"id"`
	Name             string          `json:"name"`
	Description      string          `json:"description,omitempty"`
	Tags             []string        `json:"tags,omitempty"`
	Network          *ProjectNetwork `json:"network,omitempty"`
	ContainerIDStart *int            `json:"container_id_start,omitempty"` // Start of container ID range (e.g., 200)
	ContainerIDEnd   *int            `json:"container_id_end,omitempty"`   // End of container ID range (e.g., 299)
	CreatedAt        int64           `json:"created_at"`
	UpdatedAt        int64           `json:"updated_at"`
}

// CreateProjectRequest holds parameters for creating a new project
type CreateProjectRequest struct {
	Name             string          `json:"name"`
	Description      string          `json:"description,omitempty"`
	Tags             []string        `json:"tags,omitempty"`
	Network          *ProjectNetwork `json:"network,omitempty"`
	ContainerIDStart *int            `json:"container_id_start,omitempty"` // Start of container ID range (e.g., 200)
	ContainerIDEnd   *int            `json:"container_id_end,omitempty"`   // End of container ID range (e.g., 299)
}

// UpdateProjectRequest holds parameters for updating a project
type UpdateProjectRequest struct {
	Name             string          `json:"name,omitempty"`
	Description      string          `json:"description,omitempty"`
	Tags             []string        `json:"tags,omitempty"`
	Network          *ProjectNetwork `json:"network,omitempty"`
	ContainerIDStart *int            `json:"container_id_start,omitempty"` // Start of container ID range
	ContainerIDEnd   *int            `json:"container_id_end,omitempty"`   // End of container ID range
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
	Active       *bool    `json:"-"`                       // Set when storage is accessible (parsed manually)
	Enabled      *bool    `json:"-"`                       // Set when storage is enabled (parsed manually)
	Avail        *int64   `json:"avail,omitempty"`         // Available storage space in bytes
	Total        *int64   `json:"total,omitempty"`         // Total storage space in bytes
	Used         *int64   `json:"used,omitempty"`          // Used storage space in bytes
	UsedFraction *float64 `json:"used_fraction,omitempty"` // Used fraction (used/total)
	Shared       *bool    `json:"-"`                       // Shared flag from storage configuration (parsed manually)
}

// UnmarshalJSON custom unmarshaler for Storage to handle bool/int fields
func (s *Storage) UnmarshalJSON(data []byte) error {
	// First unmarshal into a map to handle raw values
	var raw map[string]interface{}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	// Parse string fields
	if v, ok := raw["storage"].(string); ok {
		s.Storage = v
	}
	if v, ok := raw["type"].(string); ok {
		s.Type = v
	}
	if v, ok := raw["content"].(string); ok {
		s.Content = v
	}

	// Parse numeric fields
	if v, ok := raw["avail"].(float64); ok {
		val := int64(v)
		s.Avail = &val
	}
	if v, ok := raw["total"].(float64); ok {
		val := int64(v)
		s.Total = &val
	}
	if v, ok := raw["used"].(float64); ok {
		val := int64(v)
		s.Used = &val
	}
	if v, ok := raw["used_fraction"].(float64); ok {
		s.UsedFraction = &v
	}

	// Parse Active field (can be bool or int)
	if v, ok := raw["active"]; ok {
		s.Active = parseBoolOrIntValue(v)
	}

	// Parse Enabled field (can be bool or int)
	if v, ok := raw["enabled"]; ok {
		s.Enabled = parseBoolOrIntValue(v)
	}

	// Parse Shared field (can be bool or int)
	if v, ok := raw["shared"]; ok {
		s.Shared = parseBoolOrIntValue(v)
	}

	return nil
}

// parseBoolOrIntValue parses a value that can be either a bool or an int (0/1)
func parseBoolOrIntValue(v interface{}) *bool {
	// Check if it's a bool
	if b, ok := v.(bool); ok {
		return &b
	}

	// Check if it's a number (float64 is the JSON number type)
	if n, ok := v.(float64); ok {
		val := n != 0
		return &val
	}

	// Failed to parse, return nil
	return nil
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

// SDNZone represents a Proxmox SDN zone
type SDNZone struct {
	Zone    string `json:"zone"`
	Type    string `json:"type"`
	Pending bool   `json:"pending,omitempty"`
	Nodes   string `json:"nodes,omitempty"`
	MTU     int    `json:"mtu,omitempty"`
}
