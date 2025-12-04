package proxmox

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
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
