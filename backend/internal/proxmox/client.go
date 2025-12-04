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
			Timeout:   30 * time.Second,
		},
	}
}

// doRequest performs an HTTP request to the Proxmox API
func (c *Client) doRequest(method, path string, body interface{}) ([]byte, error) {
	var reqBody io.Reader
	if body != nil {
		jsonData, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request body: %w", err)
		}
		reqBody = bytes.NewBuffer(jsonData)
	}

	req, err := http.NewRequest(method, c.baseURL+path, reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("PVEAPIToken=%s=%s", c.tokenID, c.tokenSecret))
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("proxmox API error (status %d): %s", resp.StatusCode, string(respBody))
	}

	return respBody, nil
}

// GetContainers retrieves all LXC containers on the node
func (c *Client) GetContainers() ([]Container, error) {
	path := fmt.Sprintf("/nodes/%s/lxc", c.node)
	respBody, err := c.doRequest("GET", path, nil)
	if err != nil {
		return nil, err
	}

	var response struct {
		Data []Container `json:"data"`
	}
	if err := json.Unmarshal(respBody, &response); err != nil {
		return nil, fmt.Errorf("failed to parse containers response: %w", err)
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
	path := fmt.Sprintf("/nodes/%s/storage/local/content", c.node)
	respBody, err := c.doRequest("GET", path, nil)
	if err != nil {
		return nil, err
	}

	var response struct {
		Data []Template `json:"data"`
	}
	if err := json.Unmarshal(respBody, &response); err != nil {
		return nil, fmt.Errorf("failed to parse templates response: %w", err)
	}

	// Filter for templates only
	var templates []Template
	for _, t := range response.Data {
		if t.Content == "vztmpl" {
			templates = append(templates, t)
		}
	}

	return templates, nil
}

// GetNextVMID returns the next available VMID
func (c *Client) GetNextVMID() (int, error) {
	path := "/cluster/nextid"
	respBody, err := c.doRequest("GET", path, nil)
	if err != nil {
		return 0, err
	}

	var response struct {
		Data int `json:"data"`
	}
	if err := json.Unmarshal(respBody, &response); err != nil {
		return 0, fmt.Errorf("failed to parse nextid response: %w", err)
	}

	return response.Data, nil
}
