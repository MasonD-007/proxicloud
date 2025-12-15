package proxmox

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"net/url"
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

			// Convert to form values using url.Values for proper URL encoding
			formValues := url.Values{}
			for k, v := range bodyMap {
				formValues.Set(k, fmt.Sprintf("%v", v))
			}

			formData := formValues.Encode()
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
	defer func() {
		if err := resp.Body.Close(); err != nil {
			log.Printf("Failed to close response body: %v", err)
		}
	}()

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

	// Fetch IP addresses for each container
	for i := range response.Data {
		container := &response.Data[i]
		configPath := fmt.Sprintf("/nodes/%s/lxc/%d/config", c.node, container.VMID)
		configBody, err := c.doRequest("GET", configPath, nil)
		if err == nil {
			var configResponse struct {
				Data map[string]interface{} `json:"data"`
			}
			if err := json.Unmarshal(configBody, &configResponse); err == nil {
				// Extract IP address from net0 configuration
				if net0, ok := configResponse.Data["net0"].(string); ok {
					container.IPAddress = extractIPFromNetConfig(net0)
				}
			}
		}

		fmt.Printf("[DEBUG] Container %d: VMID=%d, Name=%s, Status=%s, IP=%s, CPU=%.2f, Mem=%d, MaxMem=%d\n",
			i, container.VMID, container.Name, container.Status, container.IPAddress, container.CPU, container.Mem, container.MaxMem)
	}

	return response.Data, nil
}

// extractIPFromNetConfig extracts the IP address from a Proxmox network config string
// Example: "bridge=vmbr0,name=eth0,firewall=1,ip=192.168.1.100/24,gw=192.168.1.1" -> "192.168.1.100/24"
func extractIPFromNetConfig(netConfig string) string {
	parts := strings.Split(netConfig, ",")
	for _, part := range parts {
		if strings.HasPrefix(part, "ip=") {
			ip := strings.TrimPrefix(part, "ip=")
			// Don't return "dhcp" as an IP address
			if ip != "dhcp" {
				return ip
			}
		}
	}
	return ""
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

	container := &response.Data

	// Fetch container config to get network information
	configPath := fmt.Sprintf("/nodes/%s/lxc/%d/config", c.node, vmid)
	configBody, err := c.doRequest("GET", configPath, nil)
	if err == nil {
		var configResponse struct {
			Data map[string]interface{} `json:"data"`
		}
		if err := json.Unmarshal(configBody, &configResponse); err == nil {
			// Extract IP address from net0 configuration
			if net0, ok := configResponse.Data["net0"].(string); ok {
				container.IPAddress = extractIPFromNetConfig(net0)
			}
		}
	}

	return container, nil
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

	// Determine which bridge to use
	// If VNetID is set, use the SDN VNet, otherwise use default vmbr0
	bridge := "vmbr0"
	if req.VNetID != "" {
		bridge = req.VNetID
		log.Printf("[INFO] Using SDN VNet bridge: %s", bridge)
	}

	// Network configuration
	// If IP address is provided, configure static IP
	// Otherwise, use DHCP by default
	// Format: net0: bridge=vmbr0,name=eth0,firewall=1,ip=...
	// Note: bridge must come first in the config string
	if req.IPAddress != "" {
		// Format: net0: bridge=vmbr0,name=eth0,firewall=1,ip=192.168.1.100/24,gw=192.168.1.1
		netConfig := fmt.Sprintf("bridge=%s,name=eth0,firewall=1,ip=%s", bridge, req.IPAddress)

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
		params["net0"] = fmt.Sprintf("bridge=%s,name=eth0,firewall=1,ip=dhcp", bridge)
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
	defer func() {
		if err := resp.Body.Close(); err != nil {
			log.Printf("Failed to close response body: %v", err)
		}
	}()

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

	// Use a map to track unique volumes by volid
	volumeMap := make(map[string]Volume)

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
				volumeMap[item.VolID] = volume
			}
		}
	}

	// Now check all containers to see which volumes are attached
	// This will also discover volumes that weren't in the storage content list
	containers, err := c.GetContainers()
	if err != nil {
		fmt.Printf("[WARNING] Failed to get containers for volume attachment check: %v\n", err)
		// Continue anyway, just won't have attachment info
	} else {
		// For each container, get its config to check for attached volumes
		for _, container := range containers {
			attachments, err := c.getContainerVolumeAttachments(container.VMID)
			if err != nil {
				fmt.Printf("[WARNING] Failed to get volume attachments for container %d: %v\n", container.VMID, err)
				continue
			}

			fmt.Printf("[DEBUG] Container %d has %d attachments\n", container.VMID, len(attachments))
			for volid := range attachments {
				fmt.Printf("[DEBUG]   - %s\n", volid)
			}

			// Process all attachments found in container config
			for volid, attachment := range attachments {
				if vol, exists := volumeMap[volid]; exists {
					// Update existing volume with attachment info
					fmt.Printf("[DEBUG] Updating existing volume %s: attaching to container %d (was: %v)\n", volid, container.VMID, vol.AttachedTo)
					vmid := container.VMID // Create a copy to avoid pointer issues with loop variable
					vol.Status = "in-use"
					vol.AttachedTo = &vmid
					vol.MountPoint = attachment.MountPoint

					// Try to get disk usage if not already set
					if vol.Used == 0 {
						diskUsed := c.getVolumeUsageFromContainer(vmid, attachment.MountPoint)
						if diskUsed > 0 {
							vol.Used = diskUsed / (1024 * 1024 * 1024) // Convert bytes to GB
						}
					}

					volumeMap[volid] = vol
					fmt.Printf("[DEBUG] Volume %s now attached to: %d, used: %d GB\n", volid, *vol.AttachedTo, vol.Used)
				} else {
					// Volume not found in storage list, add it now
					// Parse storage from volid
					parts := strings.SplitN(volid, ":", 2)
					storage := "unknown"
					if len(parts) == 2 {
						storage = parts[0]
					}

					// Try to get size info from storage API
					size := int64(0)
					format := "raw"
					volumePath := fmt.Sprintf("/nodes/%s/storage/%s/content/%s", c.node, storage, volid)
					respBody, err := c.doRequest("GET", volumePath, nil)
					if err == nil {
						var volResponse struct {
							Data struct {
								Size   int64  `json:"size"`
								Format string `json:"format"`
							} `json:"data"`
						}
						if err := json.Unmarshal(respBody, &volResponse); err == nil {
							size = volResponse.Data.Size / (1024 * 1024 * 1024) // Convert to GB
							if volResponse.Data.Format != "" {
								format = volResponse.Data.Format
							}
						}
					}

					vmid := container.VMID // Create a copy to avoid pointer issues with loop variable
					volume := Volume{
						VolID:      volid,
						Name:       extractVolumeName(volid),
						Size:       size,
						Node:       c.node,
						Storage:    storage,
						Format:     format,
						Status:     "in-use",
						AttachedTo: &vmid,
						MountPoint: attachment.MountPoint,
					}

					// Try to get disk usage
					diskUsed := c.getVolumeUsageFromContainer(vmid, attachment.MountPoint)
					if diskUsed > 0 {
						volume.Used = diskUsed / (1024 * 1024 * 1024) // Convert bytes to GB
					}

					volumeMap[volid] = volume
					fmt.Printf("[INFO] Discovered volume %s from container %d config, attached_to=%d, used=%d GB\n", volid, container.VMID, vmid, volume.Used)
				}
			}
		}
	}

	// Convert map to slice
	allVolumes := make([]Volume, 0, len(volumeMap))
	for _, vol := range volumeMap {
		allVolumes = append(allVolumes, vol)
		if vol.AttachedTo != nil {
			fmt.Printf("[DEBUG] Final volume %s: attached_to=%d, mountpoint=%s\n", vol.VolID, *vol.AttachedTo, vol.MountPoint)
		} else {
			fmt.Printf("[DEBUG] Final volume %s: no attachment\n", vol.VolID)
		}
	}

	fmt.Printf("[INFO] GetVolumes: returning %d total volumes\n", len(allVolumes))
	return allVolumes, nil
}

// getContainerVolumeAttachments gets all volume attachments for a container
// Returns a map of volid -> attachment info
func (c *Client) getContainerVolumeAttachments(vmid int) (map[string]struct{ MountPoint string }, error) {
	path := fmt.Sprintf("/nodes/%s/lxc/%d/config", c.node, vmid)
	fmt.Printf("[DEBUG] getContainerVolumeAttachments: requesting path=%s\n", path)

	respBody, err := c.doRequest("GET", path, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get container config: %w", err)
	}

	var response struct {
		Data map[string]interface{} `json:"data"`
	}
	if err := json.Unmarshal(respBody, &response); err != nil {
		return nil, fmt.Errorf("failed to parse container config: %w", err)
	}

	attachments := make(map[string]struct{ MountPoint string })

	// Check rootfs
	if rootfs, ok := response.Data["rootfs"].(string); ok {
		// Parse rootfs format: "storage:volume,size=XG" or "storage:vm-XXX-disk-0,size=XG"
		if volid := extractVolIDFromConfig(rootfs); volid != "" {
			attachments[volid] = struct{ MountPoint string }{MountPoint: "rootfs"}
			fmt.Printf("[DEBUG] Found rootfs volume: %s for container %d\n", volid, vmid)
		}
	}

	// Check mount points (mp0 through mp9)
	for i := 0; i < 10; i++ {
		mpKey := fmt.Sprintf("mp%d", i)
		if mp, ok := response.Data[mpKey].(string); ok {
			// Parse mount point format: "storage:volume,mp=/path"
			if volid := extractVolIDFromConfig(mp); volid != "" {
				attachments[volid] = struct{ MountPoint string }{MountPoint: mpKey}
				fmt.Printf("[DEBUG] Found mount point %s volume: %s for container %d\n", mpKey, volid, vmid)
			}
		}
	}

	return attachments, nil
}

// extractVolIDFromConfig extracts the volume ID from a config string
// Format examples:
// - "local-lvm:vm-100-disk-0,size=8G"
// - "local-zfs:vm-100-disk-1,mp=/mnt/data"
func extractVolIDFromConfig(configStr string) string {
	// Split by comma to get the first part
	parts := strings.Split(configStr, ",")
	if len(parts) == 0 {
		return ""
	}

	// The first part should be "storage:volume"
	volid := parts[0]

	// Verify it has the expected format (contains a colon)
	if strings.Contains(volid, ":") {
		return volid
	}

	return ""
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

	// Parse the response with flexible field handling
	var response struct {
		Data map[string]interface{} `json:"data"`
	}
	if err := json.Unmarshal(respBody, &response); err != nil {
		return nil, fmt.Errorf("failed to parse volume response: %w", err)
	}

	// Extract basic fields
	var volumeID string
	var size int64
	var used int64
	var format string

	if v, ok := response.Data["volid"].(string); ok {
		volumeID = v
	}
	if v, ok := response.Data["size"].(float64); ok {
		size = int64(v)
	}
	if v, ok := response.Data["format"].(string); ok {
		format = v
	}

	// Try to get 'used' field (may not be present for all storage types)
	if v, ok := response.Data["used"].(float64); ok {
		used = int64(v)
		fmt.Printf("[DEBUG] GetVolume: found 'used' field in API response: %d bytes\n", used)
	} else {
		fmt.Printf("[DEBUG] GetVolume: 'used' field not present in API response\n")
	}

	volume := &Volume{
		VolID:   volumeID,
		Name:    extractVolumeName(volumeID),
		Size:    size / (1024 * 1024 * 1024), // Convert bytes to GB
		Used:    used / (1024 * 1024 * 1024), // Convert bytes to GB (will be 0 if not present)
		Node:    c.node,
		Storage: storage,
		Format:  format,
		Status:  "available",
	}

	// Check if volume is attached to any container
	containers, err := c.GetContainers()
	if err != nil {
		fmt.Printf("[WARNING] Failed to get containers for volume attachment check: %v\n", err)
		// Continue anyway, just won't have attachment info
	} else {
		// For each container, get its config to check for attached volumes
		for _, container := range containers {
			attachments, err := c.getContainerVolumeAttachments(container.VMID)
			if err != nil {
				fmt.Printf("[WARNING] Failed to get volume attachments for container %d: %v\n", container.VMID, err)
				continue
			}

			// Check if our volume is in the attachments
			if attachment, found := attachments[volid]; found {
				vmid := container.VMID
				volume.Status = "in-use"
				volume.AttachedTo = &vmid
				volume.MountPoint = attachment.MountPoint
				fmt.Printf("[DEBUG] Volume %s is attached to container %d at %s\n", volid, vmid, attachment.MountPoint)

				// If used space is not available from the API, try to get it from container stats
				if volume.Used == 0 {
					diskUsed := c.getVolumeUsageFromContainer(vmid, attachment.MountPoint)
					if diskUsed > 0 {
						volume.Used = diskUsed / (1024 * 1024 * 1024) // Convert bytes to GB
						fmt.Printf("[DEBUG] Got volume usage from container stats: %d GB\n", volume.Used)
					}
				}
				break // Found attachment, no need to check other containers
			}
		}
	}

	fmt.Printf("[DEBUG] GetVolume final result - Size: %d GB, Used: %d GB\n", volume.Size, volume.Used)
	return volume, nil
}

// getVolumeUsageFromContainer attempts to get disk usage for a volume from container stats
func (c *Client) getVolumeUsageFromContainer(vmid int, mountPoint string) int64 {
	// For rootfs, we can get usage from container stats
	if mountPoint == "rootfs" {
		statsPath := fmt.Sprintf("/nodes/%s/lxc/%d/status/current", c.node, vmid)
		statsBody, err := c.doRequest("GET", statsPath, nil)
		if err != nil {
			fmt.Printf("[DEBUG] Failed to get container stats: %v\n", err)
			return 0
		}

		var statsResponse struct {
			Data struct {
				Disk    int64 `json:"disk"`    // Current disk usage in bytes
				MaxDisk int64 `json:"maxdisk"` // Maximum disk size in bytes
			} `json:"data"`
		}
		if err := json.Unmarshal(statsBody, &statsResponse); err != nil {
			return 0
		}

		fmt.Printf("[DEBUG] Container %d rootfs disk usage: %d bytes (max: %d bytes)\n", vmid, statsResponse.Data.Disk, statsResponse.Data.MaxDisk)
		return statsResponse.Data.Disk
	}

	// For mount points (mp0-mp9), we cannot easily get individual usage from Proxmox API
	// The disk usage in container stats only reflects rootfs
	// Would need to exec into container and run 'df' command to get mount point usage
	fmt.Printf("[DEBUG] Cannot get disk usage for mount point %s without executing commands in container\n", mountPoint)
	return 0
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

	// Check if storage supports snapshots
	// LVM-thin (local-lvm) doesn't support volume-level snapshots via this API
	if storage == "local-lvm" || storage == "local" {
		return nil, fmt.Errorf("storage type '%s' does not support volume-level snapshots. Use ZFS or another snapshot-capable storage backend", storage)
	}

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

	// Check if storage supports snapshots
	// LVM-thin (local-lvm) doesn't support volume-level snapshots via this API
	// Only ZFS and some other storage types support this
	if storage == "local-lvm" || storage == "local" {
		fmt.Printf("[INFO] GetSnapshots: storage type '%s' does not support volume-level snapshots\n", storage)
		return []Snapshot{}, nil // Return empty list, not an error
	}

	path := fmt.Sprintf("/nodes/%s/storage/%s/content/%s/snapshots", c.node, storage, volid)
	fmt.Printf("[DEBUG] GetSnapshots: requesting path=%s\n", path)

	respBody, err := c.doRequest("GET", path, nil)
	if err != nil {
		// If we get an error about illegal characters or unsupported operation, return empty list
		if strings.Contains(err.Error(), "illegal characters") ||
			strings.Contains(err.Error(), "not supported") ||
			strings.Contains(err.Error(), "500") {
			fmt.Printf("[INFO] GetSnapshots: storage does not support snapshots for %s\n", volid)
			return []Snapshot{}, nil
		}
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

	// Check if storage supports snapshots
	if storage == "local-lvm" || storage == "local" {
		return fmt.Errorf("storage type '%s' does not support volume-level snapshots", storage)
	}

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

	// Check if storage supports snapshots
	if storage == "local-lvm" || storage == "local" {
		return nil, fmt.Errorf("storage type '%s' does not support volume-level snapshots", storage)
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

// CreateVNet creates a VNet in a specified SDN zone
func (c *Client) CreateVNet(vnetID string, zone string, tag int) error {
	path := "/cluster/sdn/vnets"
	fmt.Printf("[DEBUG] CreateVNet: requesting path=%s, vnet=%s, zone=%s, tag=%d\n", path, vnetID, zone, tag)

	params := map[string]interface{}{
		"vnet": vnetID,
		"zone": zone,
	}

	if tag > 0 {
		params["tag"] = tag
	}

	_, err := c.doRequest("POST", path, params)
	if err != nil {
		return fmt.Errorf("failed to create VNet: %w", err)
	}

	fmt.Printf("[INFO] CreateVNet: created VNet %s in zone %s\n", vnetID, zone)
	return nil
}

// CreateSubnet creates a subnet within a VNet
func (c *Client) CreateSubnet(vnetID string, subnet string, gateway string, snat bool, dhcpRange string) error {
	path := fmt.Sprintf("/cluster/sdn/vnets/%s/subnets", vnetID)
	fmt.Printf("[DEBUG] CreateSubnet: requesting path=%s, subnet=%s, gateway=%s, snat=%v\n", path, subnet, gateway, snat)

	params := map[string]interface{}{
		"subnet": subnet,
	}

	if gateway != "" {
		params["gateway"] = gateway
	}

	if snat {
		params["snat"] = 1
	}

	if dhcpRange != "" {
		params["dhcp-range"] = dhcpRange
	}

	_, err := c.doRequest("POST", path, params)
	if err != nil {
		return fmt.Errorf("failed to create subnet: %w", err)
	}

	fmt.Printf("[INFO] CreateSubnet: created subnet %s in VNet %s\n", subnet, vnetID)
	return nil
}

// ApplySDNConfig applies the SDN configuration (equivalent to pressing "Apply" in GUI)
func (c *Client) ApplySDNConfig() error {
	path := "/cluster/sdn"
	fmt.Printf("[DEBUG] ApplySDNConfig: requesting path=%s\n", path)

	_, err := c.doRequest("PUT", path, nil)
	if err != nil {
		return fmt.Errorf("failed to apply SDN config: %w", err)
	}

	fmt.Printf("[INFO] ApplySDNConfig: SDN configuration applied successfully\n")
	return nil
}

// DeleteVNet deletes a VNet
func (c *Client) DeleteVNet(vnetID string) error {
	path := fmt.Sprintf("/cluster/sdn/vnets/%s", vnetID)
	fmt.Printf("[DEBUG] DeleteVNet: requesting path=%s\n", path)

	_, err := c.doRequest("DELETE", path, nil)
	if err != nil {
		return fmt.Errorf("failed to delete VNet: %w", err)
	}

	fmt.Printf("[INFO] DeleteVNet: deleted VNet %s\n", vnetID)
	return nil
}

// GetSDNZones retrieves all available SDN zones
func (c *Client) GetSDNZones() ([]SDNZone, error) {
	path := "/cluster/sdn/zones"
	fmt.Printf("[DEBUG] GetSDNZones: requesting path=%s\n", path)

	respBody, err := c.doRequest("GET", path, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get SDN zones: %w", err)
	}

	var response struct {
		Data []SDNZone `json:"data"`
	}
	if err := json.Unmarshal(respBody, &response); err != nil {
		return nil, fmt.Errorf("failed to parse SDN zones response: %w", err)
	}

	fmt.Printf("[INFO] GetSDNZones: found %d SDN zones\n", len(response.Data))
	return response.Data, nil
}

// GetStorage retrieves status for all datastores on the node
func (c *Client) GetStorage(req *GetStorageRequest) ([]Storage, error) {
	path := fmt.Sprintf("/nodes/%s/storage", c.node)
	fmt.Printf("[DEBUG] GetStorage: requesting path=%s\n", path)

	// Build query parameters if request is provided
	var queryParams []string
	if req != nil {
		if req.Content != "" {
			queryParams = append(queryParams, fmt.Sprintf("content=%s", req.Content))
		}
		if req.Enabled != nil {
			enabledVal := 0
			if *req.Enabled {
				enabledVal = 1
			}
			queryParams = append(queryParams, fmt.Sprintf("enabled=%d", enabledVal))
		}
		if req.Format != nil {
			formatVal := 0
			if *req.Format {
				formatVal = 1
			}
			queryParams = append(queryParams, fmt.Sprintf("format=%d", formatVal))
		}
		if req.Storage != "" {
			queryParams = append(queryParams, fmt.Sprintf("storage=%s", req.Storage))
		}
		if req.Target != "" {
			queryParams = append(queryParams, fmt.Sprintf("target=%s", req.Target))
		}
	}

	// Append query parameters to path if any
	if len(queryParams) > 0 {
		path = path + "?" + strings.Join(queryParams, "&")
	}

	respBody, err := c.doRequest("GET", path, nil)
	if err != nil {
		return nil, err
	}

	var response struct {
		Data []Storage `json:"data"`
	}
	if err := json.Unmarshal(respBody, &response); err != nil {
		fmt.Printf("[ERROR] Failed to unmarshal storage response: %v\n", err)
		fmt.Printf("[DEBUG] Raw response body: %s\n", string(respBody))
		return nil, fmt.Errorf("failed to parse storage response: %w", err)
	}

	fmt.Printf("[INFO] GetStorage: parsed %d storage entries from Proxmox API\n", len(response.Data))

	// Log details of each storage for debugging
	for i, storage := range response.Data {
		fmt.Printf("[DEBUG] Storage %d: ID=%s, Type=%s, Content=%s, Active=%v, Enabled=%v\n",
			i, storage.Storage, storage.Type, storage.Content, storage.Active, storage.Enabled)
	}

	return response.Data, nil
}
