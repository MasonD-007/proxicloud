package proxmox

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// generateID creates a random ID for projects
func generateID() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("failed to generate random ID: %v", err)
	}
	return hex.EncodeToString(b), nil
}

// ProjectStore manages project persistence
type ProjectStore struct {
	filePath string
	mu       sync.RWMutex
	projects map[string]*Project
	vmidMap  map[int]string // Maps VMID to ProjectID
}

// NewProjectStore creates a new project store
func NewProjectStore(dataPath string) (*ProjectStore, error) {
	if dataPath == "" {
		dataPath = "/var/lib/proxicloud/projects.json"
	}

	// Ensure directory exists
	dir := filepath.Dir(dataPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create directory: %v", err)
	}

	store := &ProjectStore{
		filePath: dataPath,
		projects: make(map[string]*Project),
		vmidMap:  make(map[int]string),
	}

	// Load existing data
	if err := store.load(); err != nil {
		if !os.IsNotExist(err) {
			return nil, fmt.Errorf("failed to load projects: %v", err)
		}
		// File doesn't exist yet, that's OK
	}

	return store, nil
}

// load reads projects from disk
func (ps *ProjectStore) load() error {
	data, err := os.ReadFile(ps.filePath)
	if err != nil {
		return err
	}

	var stored struct {
		Projects map[string]*Project `json:"projects"`
		VmidMap  map[int]string      `json:"vmid_map"`
	}

	if err := json.Unmarshal(data, &stored); err != nil {
		return fmt.Errorf("failed to unmarshal projects: %v", err)
	}

	ps.projects = stored.Projects
	ps.vmidMap = stored.VmidMap

	return nil
}

// save writes projects to disk
func (ps *ProjectStore) save() error {
	stored := struct {
		Projects map[string]*Project `json:"projects"`
		VmidMap  map[int]string      `json:"vmid_map"`
	}{
		Projects: ps.projects,
		VmidMap:  ps.vmidMap,
	}

	data, err := json.MarshalIndent(stored, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal projects: %v", err)
	}

	// Write atomically via temp file
	tmpFile := ps.filePath + ".tmp"
	if err := os.WriteFile(tmpFile, data, 0644); err != nil {
		return fmt.Errorf("failed to write temp file: %v", err)
	}

	if err := os.Rename(tmpFile, ps.filePath); err != nil {
		_ = os.Remove(tmpFile)
		return fmt.Errorf("failed to rename temp file: %v", err)
	}

	return nil
}

// CreateProject creates a new project with auto-generated ID
func (ps *ProjectStore) CreateProject(req CreateProjectRequest) (*Project, error) {
	// Generate ID
	id, err := generateID()
	if err != nil {
		return nil, err
	}

	return ps.CreateProjectWithID(id, req)
}

// CreateProjectWithID creates a new project with a specific ID (used for SDN identifier generation)
func (ps *ProjectStore) CreateProjectWithID(id string, req CreateProjectRequest) (*Project, error) {
	ps.mu.Lock()
	defer ps.mu.Unlock()

	// Validate name
	if req.Name == "" {
		return nil, fmt.Errorf("project name is required")
	}

	// Check for duplicate name
	for _, p := range ps.projects {
		if p.Name == req.Name {
			return nil, fmt.Errorf("project with name '%s' already exists", req.Name)
		}
	}

	// Validate container ID range if provided
	if req.ContainerIDStart != nil && req.ContainerIDEnd != nil {
		// Temporarily unlock to call ValidateContainerIDRange (which needs read lock)
		ps.mu.Unlock()
		err := ps.ValidateContainerIDRange(id, *req.ContainerIDStart, *req.ContainerIDEnd)
		ps.mu.Lock()
		if err != nil {
			return nil, fmt.Errorf("invalid container ID range: %w", err)
		}
	} else if req.ContainerIDStart != nil || req.ContainerIDEnd != nil {
		// Both must be provided or neither
		return nil, fmt.Errorf("both container_id_start and container_id_end must be provided together")
	}

	now := time.Now().Unix()
	project := &Project{
		ID:               id,
		Name:             req.Name,
		Description:      req.Description,
		Tags:             req.Tags,
		Network:          req.Network,
		ContainerIDStart: req.ContainerIDStart,
		ContainerIDEnd:   req.ContainerIDEnd,
		CreatedAt:        now,
		UpdatedAt:        now,
	}

	ps.projects[project.ID] = project

	if err := ps.save(); err != nil {
		return nil, err
	}

	return project, nil
}

// GetProject retrieves a project by ID
func (ps *ProjectStore) GetProject(id string) (*Project, error) {
	ps.mu.RLock()
	defer ps.mu.RUnlock()

	project, exists := ps.projects[id]
	if !exists {
		return nil, fmt.Errorf("project not found: %s", id)
	}

	return project, nil
}

// ListProjects returns all projects
func (ps *ProjectStore) ListProjects() ([]*Project, error) {
	ps.mu.RLock()
	defer ps.mu.RUnlock()

	projects := make([]*Project, 0, len(ps.projects))
	for _, p := range ps.projects {
		projects = append(projects, p)
	}

	return projects, nil
}

// UpdateProject updates a project's metadata
func (ps *ProjectStore) UpdateProject(id string, req UpdateProjectRequest) (*Project, error) {
	ps.mu.Lock()
	defer ps.mu.Unlock()

	project, exists := ps.projects[id]
	if !exists {
		return nil, fmt.Errorf("project not found: %s", id)
	}

	// Update fields
	if req.Name != "" {
		// Check for duplicate name (excluding current project)
		for pid, p := range ps.projects {
			if pid != id && p.Name == req.Name {
				return nil, fmt.Errorf("project with name '%s' already exists", req.Name)
			}
		}
		project.Name = req.Name
	}
	if req.Description != "" {
		project.Description = req.Description
	}
	if req.Tags != nil {
		project.Tags = req.Tags
	}
	if req.Network != nil {
		project.Network = req.Network
	}
	if req.ContainerIDStart != nil {
		project.ContainerIDStart = req.ContainerIDStart
	}
	if req.ContainerIDEnd != nil {
		project.ContainerIDEnd = req.ContainerIDEnd
	}
	project.UpdatedAt = time.Now().Unix()

	if err := ps.save(); err != nil {
		return nil, err
	}

	return project, nil
}

// DeleteProject deletes a project (only if no containers assigned)
func (ps *ProjectStore) DeleteProject(id string) error {
	ps.mu.Lock()
	defer ps.mu.Unlock()

	// Check if project exists
	if _, exists := ps.projects[id]; !exists {
		return fmt.Errorf("project not found: %s", id)
	}

	// Check if any containers are assigned to this project
	for _, pid := range ps.vmidMap {
		if pid == id {
			return fmt.Errorf("cannot delete project: containers still assigned")
		}
	}

	delete(ps.projects, id)

	if err := ps.save(); err != nil {
		return err
	}

	return nil
}

// AssignContainer assigns a container to a project
func (ps *ProjectStore) AssignContainer(vmid int, projectID string) error {
	ps.mu.Lock()
	defer ps.mu.Unlock()

	// Validate project exists (unless empty string for "No Project")
	if projectID != "" {
		if _, exists := ps.projects[projectID]; !exists {
			return fmt.Errorf("project not found: %s", projectID)
		}
	}

	if projectID == "" {
		// Remove from project
		delete(ps.vmidMap, vmid)
	} else {
		ps.vmidMap[vmid] = projectID
	}

	if err := ps.save(); err != nil {
		return err
	}

	return nil
}

// GetContainerProject returns the project ID for a container
func (ps *ProjectStore) GetContainerProject(vmid int) string {
	ps.mu.RLock()
	defer ps.mu.RUnlock()

	return ps.vmidMap[vmid]
}

// GetProjectContainers returns all VMIDs assigned to a project
func (ps *ProjectStore) GetProjectContainers(projectID string) []int {
	ps.mu.RLock()
	defer ps.mu.RUnlock()

	vmids := []int{}
	for vmid, pid := range ps.vmidMap {
		if pid == projectID {
			vmids = append(vmids, vmid)
		}
	}

	return vmids
}

// ValidateContainerIDRange validates that a container ID range is valid and doesn't overlap with existing projects
func (ps *ProjectStore) ValidateContainerIDRange(projectID string, start, end int) error {
	if start <= 0 {
		return fmt.Errorf("container_id_start must be greater than 0")
	}
	if end <= 0 {
		return fmt.Errorf("container_id_end must be greater than 0")
	}
	if start > end {
		return fmt.Errorf("container_id_start (%d) must be less than or equal to container_id_end (%d)", start, end)
	}
	if start < 100 {
		return fmt.Errorf("container_id_start must be >= 100 (Proxmox reserves IDs below 100)")
	}

	// Check for overlaps with other projects
	ps.mu.RLock()
	defer ps.mu.RUnlock()

	for pid, project := range ps.projects {
		// Skip the project being updated
		if pid == projectID {
			continue
		}

		// Skip projects without ID ranges
		if project.ContainerIDStart == nil || project.ContainerIDEnd == nil {
			continue
		}

		existingStart := *project.ContainerIDStart
		existingEnd := *project.ContainerIDEnd

		// Check for overlap: ranges overlap if one starts before the other ends
		if (start >= existingStart && start <= existingEnd) ||
			(end >= existingStart && end <= existingEnd) ||
			(start <= existingStart && end >= existingEnd) {
			return fmt.Errorf("container ID range %d-%d overlaps with project '%s' range %d-%d",
				start, end, project.Name, existingStart, existingEnd)
		}
	}

	return nil
}

// GetNextContainerIDInRange returns the next available container ID within a project's range
func (ps *ProjectStore) GetNextContainerIDInRange(projectID string) (int, error) {
	ps.mu.RLock()
	defer ps.mu.RUnlock()

	project, exists := ps.projects[projectID]
	if !exists {
		return 0, fmt.Errorf("project not found: %s", projectID)
	}

	// Check if project has a container ID range configured
	if project.ContainerIDStart == nil || project.ContainerIDEnd == nil {
		return 0, fmt.Errorf("project '%s' does not have a container ID range configured", project.Name)
	}

	start := *project.ContainerIDStart
	end := *project.ContainerIDEnd

	// Get all container IDs in this project
	usedIDs := make(map[int]bool)
	for vmid, pid := range ps.vmidMap {
		if pid == projectID {
			usedIDs[vmid] = true
		}
	}

	// Find the first available ID in the range
	for id := start; id <= end; id++ {
		if !usedIDs[id] {
			return id, nil
		}
	}

	return 0, fmt.Errorf("no available container IDs in range %d-%d for project '%s'", start, end, project.Name)
}
