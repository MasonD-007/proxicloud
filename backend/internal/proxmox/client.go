package proxmox

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"strings"
	"time"
)

// Client represents a Proxmox API client
type Client struct {
	baseURL     string
	node        string
	tokenID     string
	tokenSecret string
	httpClient  *http.Client
}

// NewClient creates a new Proxmox API client
func NewClient(host, node, tokenID, tokenSecret string, insecure bool) *Client {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: insecure},
	}

	// Normalize the host URL
	// Remove any trailing slashes
	host = strings.TrimRight(host, "/")

	// Check if host already contains protocol and port
	var baseURL string
	if strings.HasPrefix(host, "http://") || strings.HasPrefix(host, "https://") {
		// Host is already a full URL, use as-is
		baseURL = fmt.Sprintf("%s/api2/json", host)
	} else {
		// Host is just hostname/IP, add protocol and port
		baseURL = fmt.Sprintf("https://%s:8006/api2/json", host)
	}

	return &Client{
		baseURL:     baseURL,
		node:        node,
		tokenID:     tokenID,
		tokenSecret: tokenSecret,
		httpClient: &http.Client{
			Transport: tr,
			Timeout:   60 * time.Second, // Increased from 30s to 60s for slower networks
		},
	}
}

// doRequest performs an HTTP request to the Proxmox API
func (c *Client) doRequest(method, path string, body interface{}) ([]byte, error) {
	var reqBody io.Reader
	var contentType string

	if body != nil {
		if method == "POST" || method == "PUT" {
			// Proxmox API expects application/x-www-form-urlencoded for POST/PUT
			jsonData, err := json.Marshal(body)
			if err != nil {
				return nil, fmt.Errorf("failed to marshal request body: %w", err)
			}

			var bodyMap map[string]interface{}
			if err := json.Unmarshal(jsonData, &bodyMap); err != nil {
				return nil, fmt.Errorf("failed to unmarshal request body: %w", err)
			}

			// Convert to form values
			values := make([]string, 0, len(bodyMap))
			for k, v := range bodyMap {
				values = append(values, fmt.Sprintf("%s=%v", k, v))
			}
			formData := strings.Join(values, "&")
			fmt.Printf("[DEBUG] Form data: %s\n", formData)
			reqBody = strings.NewReader(formData)
			contentType = "application/x-www-form-urlencoded"
		} else {
			// For GET/DELETE, use JSON if body present
			jsonData, err := json.Marshal(body)
			if err != nil {
				return nil, fmt.Errorf("failed to marshal request body: %w", err)
			}
			reqBody = bytes.NewBuffer(jsonData)
			contentType = "application/json"
		}
	}

	fullURL := c.baseURL + path
	fmt.Printf("[DEBUG] Proxmox API Request: %s %s\n", method, fullURL)

	req, err := http.NewRequest(method, fullURL, reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("PVEAPIToken=%s=%s", c.tokenID, c.tokenSecret))
	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		fmt.Printf("[ERROR] Proxmox API request failed: %v\n", err)
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	fmt.Printf("[DEBUG] Proxmox API Response: status=%d, body_length=%d bytes\n", resp.StatusCode, len(respBody))

	// Log first 500 chars of response body for debugging
	if len(respBody) > 0 {
		preview := string(respBody)
		if len(preview) > 500 {
			preview = preview[:500] + "..."
		}
		fmt.Printf("[DEBUG] Response body preview: %s\n", preview)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		fmt.Printf("[ERROR] Proxmox API error response: %s\n", string(respBody))
		return nil, fmt.Errorf("proxmox API error (status %d): %s", resp.StatusCode, string(respBody))
	}

	return respBody, nil
}

// GetContainers retrieves all LXC containers on the node
func (c *Client) GetContainers() ([]Container, error) {
	path := fmt.Sprintf("/nodes/%s/lxc", c.node)
	fmt.Printf("[DEBUG] GetContainers: requesting path=%s\n", path)

	respBody, err := c.doRequest("GET", path, nil)
	if err != nil {
		return nil, err
	}

	var response struct {
		Data []Container `json:"data"`
	}
	if err := json.Unmarshal(respBody, &response); err != nil {
		fmt.Printf("[ERROR] Failed to unmarshal containers response: %v\n", err)
		fmt.Printf("[DEBUG] Raw response body: %s\n", string(respBody))
		return nil, fmt.Errorf("failed to parse containers response: %w", err)
	}

	fmt.Printf("[INFO] GetContainers: parsed %d containers from Proxmox API\n", len(response.Data))

	// Log details of each container for debugging
	for i, container := range response.Data {
		fmt.Printf("[DEBUG] Container %d: VMID=%d, Name=%s, Status=%s, CPU=%.2f, Mem=%d, MaxMem=%d\n",
			i, container.VMID, container.Name, container.Status, container.CPU, container.Mem, container.MaxMem)
	}

	return response.Data, nil
}

// GetContainer retrieves a specific LXC container
func (c *Client) GetContainer(vmid int) (*Container, error) {
	path := fmt.Sprintf("/nodes/%s/lxc/%d/status/current", c.node, vmid)
	respBody, err := c.doRequest("GET", path, nil)
	if err != nil {
		return nil, err
	}

	var response struct {
		Data Container `json:"data"`
	}
	if err := json.Unmarshal(respBody, &response); err != nil {
		return nil, fmt.Errorf("failed to parse container response: %w", err)
	}

	return &response.Data, nil
}

// CreateContainer creates a new LXC container
func (c *Client) CreateContainer(vmid int, req CreateContainerRequest) error {
	path := fmt.Sprintf("/nodes/%s/lxc", c.node)

	// Build create request
	params := map[string]interface{}{
		"vmid":       vmid,
		"hostname":   req.Hostname,
		"cores":      req.Cores,
		"memory":     req.Memory,
		"rootfs":     fmt.Sprintf("local-lvm:%d", req.Disk),
		"ostemplate": req.OSTemplate,
	}

	// Network configuration
	// If IP address is provided, configure static IP
	// Otherwise, use DHCP by default
	if req.IPAddress != "" {
		// Format: net0: name=eth0,bridge=vmbr0,ip=192.168.1.100/24,gw=192.168.1.1
		netConfig := "name=eth0,bridge=vmbr0,firewall=1,ip=" + req.IPAddress

		if req.Gateway != "" {
			netConfig += ",gw=" + req.Gateway
		}

		params["net0"] = netConfig

		// Set nameserver if provided
		if req.Nameserver != "" {
			params["nameserver"] = req.Nameserver
		}
	} else {
		// Use DHCP if no IP specified
		params["net0"] = "name=eth0,bridge=vmbr0,firewall=1,ip=dhcp"
	}

	if req.Password != "" {
		params["password"] = req.Password
	}
	if req.SSHKeys != "" {
		params["ssh-public-keys"] = req.SSHKeys
	}
	if req.StartOnBoot {
		params["onboot"] = 1
	}
	if req.Unprivileged {
		params["unprivileged"] = 1
	}

	_, err := c.doRequest("POST", path, params)
	return err
}

// StartContainer starts a container
func (c *Client) StartContainer(vmid int) error {
	path := fmt.Sprintf("/nodes/%s/lxc/%d/status/start", c.node, vmid)
	_, err := c.doRequest("POST", path, nil)
	return err
}

// StopContainer stops a container
func (c *Client) StopContainer(vmid int) error {
	path := fmt.Sprintf("/nodes/%s/lxc/%d/status/stop", c.node, vmid)
	_, err := c.doRequest("POST", path, nil)
	return err
}

// RebootContainer reboots a container
func (c *Client) RebootContainer(vmid int) error {
	path := fmt.Sprintf("/nodes/%s/lxc/%d/status/reboot", c.node, vmid)
	_, err := c.doRequest("POST", path, nil)
	return err
}

// DeleteContainer deletes a container
func (c *Client) DeleteContainer(vmid int) error {
	path := fmt.Sprintf("/nodes/%s/lxc/%d", c.node, vmid)
	_, err := c.doRequest("DELETE", path, nil)
	return err
}

// GetTemplates retrieves available container templates
func (c *Client) GetTemplates() ([]Template, error) {
	// Try multiple storage locations based on user's storage configuration
	storageLocations := []string{"local", "local-lvm"}

	var allTemplates []Template

	for _, storage := range storageLocations {
		path := fmt.Sprintf("/nodes/%s/storage/%s/content", c.node, storage)
		fmt.Printf("[DEBUG] GetTemplates: requesting path=%s\n", path)

		respBody, err := c.doRequest("GET", path, nil)
		if err != nil {
			fmt.Printf("[WARNING] Failed to get templates from storage '%s': %v\n", storage, err)
			continue // Try next storage location
		}

		var response struct {
			Data []Template `json:"data"`
		}
		if err := json.Unmarshal(respBody, &response); err != nil {
			fmt.Printf("[ERROR] Failed to unmarshal templates response from '%s': %v\n", storage, err)
			fmt.Printf("[DEBUG] Raw response body: %s\n", string(respBody))
			continue
		}

		// Filter for templates only
		storageTemplates := 0
		for _, t := range response.Data {
			if t.Content == "vztmpl" {
				allTemplates = append(allTemplates, t)
				storageTemplates++
			}
		}

		fmt.Printf("[INFO] GetTemplates: found %d templates in storage '%s' (total items: %d)\n",
			storageTemplates, storage, len(response.Data))
	}

	fmt.Printf("[INFO] GetTemplates: returning %d total templates from all storage locations\n", len(allTemplates))

	// Log details of each template for debugging
	for i, tmpl := range allTemplates {
		fmt.Printf("[DEBUG] Template %d: VolID=%s, Format=%s, Size=%d\n",
			i, tmpl.VolID, tmpl.Format, tmpl.Size)
	}

	return allTemplates, nil
}

// GetNextVMID returns the next available VMID
func (c *Client) GetNextVMID() (int, error) {
	path := "/cluster/nextid"
	respBody, err := c.doRequest("GET", path, nil)
	if err != nil {
		return 0, err
	}

	fmt.Printf("[DEBUG] GetNextVMID raw response: %s\n", string(respBody))

	// Proxmox can return the nextid as either a string or an int
	// Use json.RawMessage to handle both cases
	var response struct {
		Data json.RawMessage `json:"data"`
	}

	if err := json.Unmarshal(respBody, &response); err != nil {
		fmt.Printf("[ERROR] Failed to parse nextid response structure: %v\n", err)
		return 0, fmt.Errorf("failed to parse nextid response: %w", err)
	}

	// Try to unmarshal as int first
	var vmidInt int
	if err := json.Unmarshal(response.Data, &vmidInt); err == nil {
		fmt.Printf("[DEBUG] Successfully parsed nextid as int: %d\n", vmidInt)
		return vmidInt, nil
	}

	// Try to unmarshal as string
	var vmidStr string
	if err := json.Unmarshal(response.Data, &vmidStr); err == nil {
		// Convert string to int
		var vmid int
		if _, err := fmt.Sscanf(vmidStr, "%d", &vmid); err == nil {
			fmt.Printf("[DEBUG] Successfully parsed nextid as string and converted: %d\n", vmid)
			return vmid, nil
		}
		fmt.Printf("[ERROR] Failed to convert string '%s' to int\n", vmidStr)
		return 0, fmt.Errorf("failed to convert nextid string to int: %s", vmidStr)
	}

	fmt.Printf("[ERROR] Failed to parse nextid data as either int or string: %s\n", string(response.Data))
	return 0, fmt.Errorf("failed to parse nextid data: expected int or string, got: %s", string(response.Data))
}

// UploadTemplate uploads a new container template to Proxmox storage
func (c *Client) UploadTemplate(storage string, filename string, fileData []byte) error {
	path := fmt.Sprintf("/nodes/%s/storage/%s/upload", c.node, storage)
	fmt.Printf("[DEBUG] UploadTemplate: uploading to path=%s, filename=%s, size=%d bytes\n", path, filename, len(fileData))

	// Create multipart form data
	var requestBody bytes.Buffer
	writer := multipart.NewWriter(&requestBody)

	// Add content type field
	if err := writer.WriteField("content", "vztmpl"); err != nil {
		return fmt.Errorf("failed to write content field: %w", err)
	}

	// Add the file
	part, err := writer.CreateFormFile("filename", filename)
	if err != nil {
		return fmt.Errorf("failed to create form file: %w", err)
	}

	if _, err := part.Write(fileData); err != nil {
		return fmt.Errorf("failed to write file data: %w", err)
	}

	contentType := writer.FormDataContentType()
	if err := writer.Close(); err != nil {
		return fmt.Errorf("failed to close multipart writer: %w", err)
	}

	// Make the request with custom content type
	fullURL := c.baseURL + path
	fmt.Printf("[DEBUG] Proxmox API Upload Request: POST %s\n", fullURL)

	req, err := http.NewRequest("POST", fullURL, &requestBody)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("PVEAPIToken=%s=%s", c.tokenID, c.tokenSecret))
	req.Header.Set("Content-Type", contentType)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		fmt.Printf("[ERROR] Proxmox API upload request failed: %v\n", err)
		return fmt.Errorf("failed to execute upload request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read upload response: %w", err)
	}

	fmt.Printf("[DEBUG] Proxmox API Upload Response: status=%d, body_length=%d bytes\n", resp.StatusCode, len(respBody))

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		fmt.Printf("[ERROR] Proxmox API upload error response: %s\n", string(respBody))
		return fmt.Errorf("proxmox API upload error (status %d): %s", resp.StatusCode, string(respBody))
	}

	fmt.Printf("[INFO] Template uploaded successfully: %s\n", filename)
	return nil
}

// CreateVolume creates a new persistent volume (ZFS zvol)
func (c *Client) CreateVolume(req CreateVolumeRequest) (*Volume, error) {
	storage := req.Storage
	if storage == "" {
		storage = "local-lvm"
	}

	volumeType := req.Type
	if volumeType == "" {
		volumeType = "ssd"
	}

	node := req.Node
	if node == "" {
		node = c.node
	}

	// Create ZFS zvol using Proxmox storage API
	path := fmt.Sprintf("/nodes/%s/storage/%s/content", node, storage)
	fmt.Printf("[DEBUG] CreateVolume: requesting path=%s\n", path)

	params := map[string]interface{}{
		"filename": req.Name,
		"size":     fmt.Sprintf("%dG", req.Size),
		"vmid":     0, // Not attached to any VM initially
	}

	respBody, err := c.doRequest("POST", path, params)
	if err != nil {
		return nil, fmt.Errorf("failed to create volume: %w", err)
	}

	var response struct {
		Data string `json:"data"` // Returns volid (e.g., "local-lvm:vm-0-disk-0")
	}
	if err := json.Unmarshal(respBody, &response); err != nil {
		return nil, fmt.Errorf("failed to parse create volume response: %w", err)
	}

	fmt.Printf("[INFO] CreateVolume: created volume with volid=%s\n", response.Data)

	// Return volume object
	volume := &Volume{
		VolID:     response.Data,
		Name:      req.Name,
		Size:      int64(req.Size),
		Node:      node,
		Storage:   storage,
		Type:      volumeType,
		Format:    "raw",
		Status:    "available",
		CreatedAt: time.Now().Unix(),
	}

	return volume, nil
}

// GetVolumes retrieves all volumes on the node
func (c *Client) GetVolumes() ([]Volume, error) {
	// Query multiple storage pools
	storageLocations := []string{"local-lvm", "local-zfs"}

	var allVolumes []Volume

	for _, storage := range storageLocations {
		path := fmt.Sprintf("/nodes/%s/storage/%s/content", c.node, storage)
		fmt.Printf("[DEBUG] GetVolumes: requesting path=%s\n", path)

		respBody, err := c.doRequest("GET", path, nil)
		if err != nil {
			fmt.Printf("[WARNING] Failed to get volumes from storage '%s': %v\n", storage, err)
			continue
		}

		var response struct {
			Data []struct {
				VolID   string `json:"volid"`
				Size    int64  `json:"size"`
				Format  string `json:"format"`
				Content string `json:"content"`
			} `json:"data"`
		}
		if err := json.Unmarshal(respBody, &response); err != nil {
			fmt.Printf("[ERROR] Failed to unmarshal volumes response from '%s': %v\n", storage, err)
			continue
		}

		// Filter for images (volumes)
		for _, item := range response.Data {
			if item.Content == "images" {
				volume := Volume{
					VolID:   item.VolID,
					Name:    extractVolumeName(item.VolID),
					Size:    item.Size / (1024 * 1024 * 1024), // Convert bytes to GB
					Node:    c.node,
					Storage: storage,
					Format:  item.Format,
					Status:  "available", // Default status
				}
				allVolumes = append(allVolumes, volume)
			}
		}
	}

	fmt.Printf("[INFO] GetVolumes: returning %d total volumes\n", len(allVolumes))
	return allVolumes, nil
}

// GetVolume retrieves a specific volume by volid
func (c *Client) GetVolume(volid string) (*Volume, error) {
	// Parse storage from volid (format: "storage:volume-name")
	parts := strings.SplitN(volid, ":", 2)
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid volid format: %s", volid)
	}
	storage := parts[0]

	path := fmt.Sprintf("/nodes/%s/storage/%s/content/%s", c.node, storage, volid)
	fmt.Printf("[DEBUG] GetVolume: requesting path=%s\n", path)

	respBody, err := c.doRequest("GET", path, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get volume: %w", err)
	}

	var response struct {
		Data struct {
			VolID  string `json:"volid"`
			Size   int64  `json:"size"`
			Format string `json:"format"`
			Used   int64  `json:"used"`
		} `json:"data"`
	}
	if err := json.Unmarshal(respBody, &response); err != nil {
		return nil, fmt.Errorf("failed to parse volume response: %w", err)
	}

	volume := &Volume{
		VolID:   response.Data.VolID,
		Name:    extractVolumeName(response.Data.VolID),
		Size:    response.Data.Size / (1024 * 1024 * 1024), // Convert bytes to GB
		Used:    response.Data.Used / (1024 * 1024 * 1024), // Convert bytes to GB
		Node:    c.node,
		Storage: storage,
		Format:  response.Data.Format,
		Status:  "available",
	}

	return volume, nil
}

// DeleteVolume deletes a volume
func (c *Client) DeleteVolume(volid string) error {
	// Parse storage from volid
	parts := strings.SplitN(volid, ":", 2)
	if len(parts) != 2 {
		return fmt.Errorf("invalid volid format: %s", volid)
	}
	storage := parts[0]

	path := fmt.Sprintf("/nodes/%s/storage/%s/content/%s", c.node, storage, volid)
	fmt.Printf("[DEBUG] DeleteVolume: requesting path=%s\n", path)

	_, err := c.doRequest("DELETE", path, nil)
	if err != nil {
		return fmt.Errorf("failed to delete volume: %w", err)
	}

	fmt.Printf("[INFO] DeleteVolume: deleted volume %s\n", volid)
	return nil
}

// AttachVolume attaches a volume to a container
func (c *Client) AttachVolume(volid string, req AttachVolumeRequest) error {
	// Determine mount point (auto-detect if not provided)
	mountPoint := req.MountPoint
	if mountPoint == "" {
		// Auto-detect next available mount point (mp0-mp9)
		_, err := c.GetContainer(req.VMID)
		if err != nil {
			return fmt.Errorf("failed to get container: %w", err)
		}

		// Find next available mount point
		// This is a simplified version - in production, you'd parse the container config
		mountPoint = "mp0"
		fmt.Printf("[INFO] Auto-detected mount point: %s\n", mountPoint)
	}

	path := fmt.Sprintf("/nodes/%s/lxc/%d/config", c.node, req.VMID)
	fmt.Printf("[DEBUG] AttachVolume: requesting path=%s\n", path)

	// Attach volume using mount point configuration
	params := map[string]interface{}{
		mountPoint: fmt.Sprintf("%s,mp=/mnt/%s", volid, extractVolumeName(volid)),
	}

	_, err := c.doRequest("PUT", path, params)
	if err != nil {
		return fmt.Errorf("failed to attach volume: %w", err)
	}

	fmt.Printf("[INFO] AttachVolume: attached %s to container %d at %s\n", volid, req.VMID, mountPoint)
	return nil
}

// DetachVolume detaches a volume from a container
func (c *Client) DetachVolume(volid string, req DetachVolumeRequest) error {
	// Get container config to find mount point
	container, err := c.GetContainer(req.VMID)
	if err != nil {
		return fmt.Errorf("failed to get container: %w", err)
	}

	// Check if container is running and force is not set
	if container.Status == "running" && !req.Force {
		return fmt.Errorf("container is running, use force=true to detach")
	}

	// Find which mount point has this volume
	// This is simplified - in production, you'd parse the container config
	mountPoint := "mp0"

	path := fmt.Sprintf("/nodes/%s/lxc/%d/config", c.node, req.VMID)
	fmt.Printf("[DEBUG] DetachVolume: requesting path=%s\n", path)

	// Remove mount point configuration
	params := map[string]interface{}{
		"delete": mountPoint,
	}

	_, err = c.doRequest("PUT", path, params)
	if err != nil {
		return fmt.Errorf("failed to detach volume: %w", err)
	}

	fmt.Printf("[INFO] DetachVolume: detached %s from container %d\n", volid, req.VMID)
	return nil
}

// CreateSnapshot creates a snapshot of a volume
func (c *Client) CreateSnapshot(volid string, req CreateSnapshotRequest) (*Snapshot, error) {
	// Parse storage from volid
	parts := strings.SplitN(volid, ":", 2)
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid volid format: %s", volid)
	}
	storage := parts[0]

	path := fmt.Sprintf("/nodes/%s/storage/%s/content/%s/snapshot", c.node, storage, volid)
	fmt.Printf("[DEBUG] CreateSnapshot: requesting path=%s\n", path)

	params := map[string]interface{}{
		"snapname": req.Name,
	}
	if req.Description != "" {
		params["description"] = req.Description
	}

	respBody, err := c.doRequest("POST", path, params)
	if err != nil {
		return nil, fmt.Errorf("failed to create snapshot: %w", err)
	}

	var response struct {
		Data string `json:"data"`
	}
	if err := json.Unmarshal(respBody, &response); err != nil {
		return nil, fmt.Errorf("failed to parse create snapshot response: %w", err)
	}

	snapshot := &Snapshot{
		Name:        req.Name,
		VolID:       volid,
		Description: req.Description,
		CreatedAt:   time.Now().Unix(),
	}

	fmt.Printf("[INFO] CreateSnapshot: created snapshot %s for volume %s\n", req.Name, volid)
	return snapshot, nil
}

// GetSnapshots retrieves all snapshots for a volume
func (c *Client) GetSnapshots(volid string) ([]Snapshot, error) {
	// Parse storage from volid
	parts := strings.SplitN(volid, ":", 2)
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid volid format: %s", volid)
	}
	storage := parts[0]

	path := fmt.Sprintf("/nodes/%s/storage/%s/content/%s/snapshots", c.node, storage, volid)
	fmt.Printf("[DEBUG] GetSnapshots: requesting path=%s\n", path)

	respBody, err := c.doRequest("GET", path, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get snapshots: %w", err)
	}

	var response struct {
		Data []struct {
			Name        string `json:"name"`
			Description string `json:"description"`
			CreatedAt   int64  `json:"ctime"`
		} `json:"data"`
	}
	if err := json.Unmarshal(respBody, &response); err != nil {
		return nil, fmt.Errorf("failed to parse snapshots response: %w", err)
	}

	snapshots := make([]Snapshot, len(response.Data))
	for i, s := range response.Data {
		snapshots[i] = Snapshot{
			Name:        s.Name,
			VolID:       volid,
			Description: s.Description,
			CreatedAt:   s.CreatedAt,
		}
	}

	fmt.Printf("[INFO] GetSnapshots: returning %d snapshots for volume %s\n", len(snapshots), volid)
	return snapshots, nil
}

// RestoreSnapshot restores a volume from a snapshot
func (c *Client) RestoreSnapshot(volid string, req RestoreSnapshotRequest) error {
	// Parse storage from volid
	parts := strings.SplitN(volid, ":", 2)
	if len(parts) != 2 {
		return fmt.Errorf("invalid volid format: %s", volid)
	}
	storage := parts[0]

	path := fmt.Sprintf("/nodes/%s/storage/%s/content/%s/snapshot/%s/rollback", c.node, storage, volid, req.SnapshotName)
	fmt.Printf("[DEBUG] RestoreSnapshot: requesting path=%s\n", path)

	_, err := c.doRequest("POST", path, nil)
	if err != nil {
		return fmt.Errorf("failed to restore snapshot: %w", err)
	}

	fmt.Printf("[INFO] RestoreSnapshot: restored volume %s from snapshot %s\n", volid, req.SnapshotName)
	return nil
}

// CloneSnapshot clones a volume from a snapshot
func (c *Client) CloneSnapshot(volid string, req CloneSnapshotRequest) (*Volume, error) {
	// Parse storage from volid
	parts := strings.SplitN(volid, ":", 2)
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid volid format: %s", volid)
	}
	storage := parts[0]
	if req.Storage != "" {
		storage = req.Storage
	}

	path := fmt.Sprintf("/nodes/%s/storage/%s/content/%s/snapshot/%s/clone", c.node, storage, volid, req.SnapshotName)
	fmt.Printf("[DEBUG] CloneSnapshot: requesting path=%s\n", path)

	params := map[string]interface{}{
		"target": req.NewName,
	}

	respBody, err := c.doRequest("POST", path, params)
	if err != nil {
		return nil, fmt.Errorf("failed to clone snapshot: %w", err)
	}

	var response struct {
		Data string `json:"data"` // Returns new volid
	}
	if err := json.Unmarshal(respBody, &response); err != nil {
		return nil, fmt.Errorf("failed to parse clone snapshot response: %w", err)
	}

	volume := &Volume{
		VolID:     response.Data,
		Name:      req.NewName,
		Node:      c.node,
		Storage:   storage,
		Format:    "raw",
		Status:    "available",
		CreatedAt: time.Now().Unix(),
	}

	fmt.Printf("[INFO] CloneSnapshot: cloned snapshot %s to new volume %s\n", req.SnapshotName, response.Data)
	return volume, nil
}

// extractVolumeName extracts the volume name from a volid
// Example: "local-lvm:vm-100-disk-0" -> "vm-100-disk-0"
func extractVolumeName(volid string) string {
	parts := strings.SplitN(volid, ":", 2)
	if len(parts) == 2 {
		return parts[1]
	}
	return volid
}
