package handlers

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/MasonD-007/proxicloud/backend/internal/analytics"
	"github.com/MasonD-007/proxicloud/backend/internal/cache"
	"github.com/MasonD-007/proxicloud/backend/internal/proxmox"
	"github.com/gorilla/mux"
)

// Handler holds the Proxmox client, cache, analytics, and project store
type Handler struct {
	client       *proxmox.Client
	cache        *cache.Cache
	analytics    *analytics.Analytics
	projectStore *proxmox.ProjectStore
}

// NewHandler creates a new handler
func NewHandler(client *proxmox.Client, cache *cache.Cache, analytics *analytics.Analytics, projectStore *proxmox.ProjectStore) *Handler {
	return &Handler{
		client:       client,
		cache:        cache,
		analytics:    analytics,
		projectStore: projectStore,
	}
}

// generateID generates a random hex ID for projects
func generateID() string {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		// Fallback to timestamp-based ID if random generation fails
		return fmt.Sprintf("%x", time.Now().UnixNano())
	}
	return hex.EncodeToString(b)
}

// respondJSON sends a JSON response
func respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		log.Printf("Failed to encode JSON response: %v", err)
	}
}

// respondJSONWithCache sends a JSON response with cache status header
func respondJSONWithCache(w http.ResponseWriter, status int, data interface{}, cached bool) {
	w.Header().Set("Content-Type", "application/json")
	if cached {
		w.Header().Set("X-Cache-Status", "HIT")
	} else {
		w.Header().Set("X-Cache-Status", "MISS")
	}
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		log.Printf("Failed to encode JSON response: %v", err)
	}
}

// respondError sends an error response
func respondError(w http.ResponseWriter, status int, message string) {
	respondJSON(w, status, map[string]string{"error": message})
}

// Health handles health check requests
func (h *Handler) Health(w http.ResponseWriter, r *http.Request) {
	respondJSON(w, http.StatusOK, map[string]string{"status": "healthy"})
}

// Dashboard returns dashboard statistics
func (h *Handler) Dashboard(w http.ResponseWriter, r *http.Request) {
	containers, err := h.client.GetContainers()
	if err != nil {
		// Try to get from cache if Proxmox is down
		if h.cache != nil {
			cached, cacheErr := h.cache.GetDashboard()
			if cacheErr == nil {
				log.Printf("Serving dashboard from cache (Proxmox error: %v)", err)
				respondJSONWithCache(w, http.StatusOK, cached, true)
				return
			}
		}
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	stats := struct {
		TotalContainers   int     `json:"total_containers"`
		RunningContainers int     `json:"running_containers"`
		StoppedContainers int     `json:"stopped_containers"`
		TotalCPU          float64 `json:"total_cpu"`
		TotalMemory       int64   `json:"total_memory"`
		UsedMemory        int64   `json:"used_memory"`
		TotalDisk         int64   `json:"total_disk"`
		UsedDisk          int64   `json:"used_disk"`
	}{}

	stats.TotalContainers = len(containers)
	for _, c := range containers {
		if c.Status == "running" {
			stats.RunningContainers++
		} else {
			stats.StoppedContainers++
		}
		stats.TotalCPU += c.CPU
		stats.TotalMemory += c.MaxMem
		stats.UsedMemory += c.Mem
		stats.TotalDisk += c.MaxDisk
		stats.UsedDisk += c.Disk
	}

	// Cache the stats
	if h.cache != nil {
		if err := h.cache.SetDashboard(stats); err != nil {
			log.Printf("Failed to cache dashboard: %v", err)
		}
	}

	respondJSONWithCache(w, http.StatusOK, stats, false)
}

// ListContainers lists all containers
func (h *Handler) ListContainers(w http.ResponseWriter, r *http.Request) {
	log.Printf("[DEBUG] ListContainers handler called")

	containers, err := h.client.GetContainers()
	if err != nil {
		log.Printf("[ERROR] GetContainers failed: %v", err)
		// Try to get from cache if Proxmox is down
		if h.cache != nil {
			cached, cacheErr := h.cache.GetContainers()
			if cacheErr == nil {
				log.Printf("[INFO] Serving containers from cache (Proxmox error: %v)", err)
				respondJSONWithCache(w, http.StatusOK, cached, true)
				return
			}
			log.Printf("[ERROR] Cache retrieval also failed: %v", cacheErr)
		}
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	log.Printf("[INFO] Successfully retrieved %d containers from Proxmox", len(containers))
	if len(containers) > 0 {
		log.Printf("[DEBUG] Sample container data: %+v", containers[0])
	} else {
		log.Printf("[WARNING] Proxmox returned empty container list - this may indicate no containers exist or an API issue")
	}

	// Enrich with project information
	if h.projectStore != nil {
		for i := range containers {
			projectID := h.projectStore.GetContainerProject(containers[i].VMID)
			containers[i].ProjectID = projectID
		}
	}

	// Cache the containers
	if h.cache != nil {
		if err := h.cache.SetContainers(containers); err != nil {
			log.Printf("[ERROR] Failed to cache containers: %v", err)
		}
	}

	respondJSONWithCache(w, http.StatusOK, containers, false)
}

// GetContainer gets a specific container
func (h *Handler) GetContainer(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	vmid, err := strconv.Atoi(vars["vmid"])
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid vmid")
		return
	}

	container, err := h.client.GetContainer(vmid)
	if err != nil {
		// Try to get from cache if Proxmox is down
		if h.cache != nil {
			cached, cacheErr := h.cache.GetContainer(vmid)
			if cacheErr == nil {
				log.Printf("Serving container %d from cache (Proxmox error: %v)", vmid, err)
				respondJSONWithCache(w, http.StatusOK, cached, true)
				return
			}
		}
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Enrich with project information
	if h.projectStore != nil {
		container.ProjectID = h.projectStore.GetContainerProject(vmid)
	}

	// Cache the container
	if h.cache != nil {
		if err := h.cache.SetContainer(*container); err != nil {
			log.Printf("Failed to cache container %d: %v", vmid, err)
		}
	}

	respondJSONWithCache(w, http.StatusOK, container, false)
}

// CreateContainer creates a new container
func (h *Handler) CreateContainer(w http.ResponseWriter, r *http.Request) {
	var req proxmox.CreateContainerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	var vmid int
	var err error

	// Check if user provided a custom VMID
	if req.VMID != nil && *req.VMID > 0 {
		vmid = *req.VMID
		log.Printf("[INFO] Using user-specified VMID: %d", vmid)

		// If container is assigned to a project with ID range, validate it's within range
		if req.ProjectID != "" && h.projectStore != nil {
			project, err := h.projectStore.GetProject(req.ProjectID)
			if err == nil && project.ContainerIDStart != nil && project.ContainerIDEnd != nil {
				if vmid < *project.ContainerIDStart || vmid > *project.ContainerIDEnd {
					respondError(w, http.StatusBadRequest, fmt.Sprintf("VMID %d is outside project's container ID range %d-%d", vmid, *project.ContainerIDStart, *project.ContainerIDEnd))
					return
				}
			}
		}

		// Verify the VMID is not already in use
		_, err := h.client.GetContainer(vmid)
		if err == nil {
			// Container exists with this VMID
			respondError(w, http.StatusConflict, fmt.Sprintf("VMID %d is already in use", vmid))
			return
		}
		// If error is not nil, the VMID is likely available (or there's another issue)
		// We'll let Proxmox handle the final validation
	} else {
		// If container is assigned to a project with ID range, use project's range
		if req.ProjectID != "" && h.projectStore != nil {
			project, err := h.projectStore.GetProject(req.ProjectID)
			if err == nil && project.ContainerIDStart != nil && project.ContainerIDEnd != nil {
				vmid, err = h.projectStore.GetNextContainerIDInRange(req.ProjectID)
				if err != nil {
					respondError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to allocate container ID from project range: %v", err))
					return
				}
				log.Printf("[INFO] Using project-allocated VMID: %d (from range %d-%d)", vmid, *project.ContainerIDStart, *project.ContainerIDEnd)
			} else {
				// Project doesn't have ID range, use global next VMID
				vmid, err = h.client.GetNextVMID()
				if err != nil {
					respondError(w, http.StatusInternalServerError, err.Error())
					return
				}
				log.Printf("[INFO] Using auto-generated VMID: %d", vmid)
			}
		} else {
			// No project or no projectStore, get next available VMID
			vmid, err = h.client.GetNextVMID()
			if err != nil {
				respondError(w, http.StatusInternalServerError, err.Error())
				return
			}
			log.Printf("[INFO] Using auto-generated VMID: %d", vmid)
		}
	}

	// If container is assigned to a project with network configuration, use project's VNet
	if req.ProjectID != "" && h.projectStore != nil {
		project, err := h.projectStore.GetProject(req.ProjectID)
		if err == nil && project.Network != nil && project.Network.VNetID != "" {
			log.Printf("[INFO] Container assigned to project %s with VNet %s", req.ProjectID, project.Network.VNetID)

			// If user didn't specify network config, use project defaults
			if req.IPAddress == "" && req.Gateway == "" {
				// Use project's gateway if available
				if project.Network.Gateway != "" {
					req.Gateway = project.Network.Gateway
				}
				// Use project's nameserver if available
				if project.Network.Nameserver != "" {
					req.Nameserver = project.Network.Nameserver
				}
				log.Printf("[INFO] Applied project network defaults: gateway=%s, nameserver=%s", req.Gateway, req.Nameserver)
			}

			// Set the VNet ID to use for container creation
			req.VNetID = project.Network.VNetID
		}
	}

	if err := h.client.CreateContainer(vmid, req); err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Assign to project if specified
	if req.ProjectID != "" && h.projectStore != nil {
		if err := h.projectStore.AssignContainer(vmid, req.ProjectID); err != nil {
			log.Printf("[WARNING] Failed to assign container %d to project %s: %v", vmid, req.ProjectID, err)
			// Don't fail the whole request, just log the error
		}
	}

	respondJSON(w, http.StatusCreated, map[string]int{"vmid": vmid})
}

// StartContainer starts a container
func (h *Handler) StartContainer(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	vmid, err := strconv.Atoi(vars["vmid"])
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid vmid")
		return
	}

	if err := h.client.StartContainer(vmid); err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{"status": "started"})
}

// StopContainer stops a container
func (h *Handler) StopContainer(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	vmid, err := strconv.Atoi(vars["vmid"])
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid vmid")
		return
	}

	if err := h.client.StopContainer(vmid); err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{"status": "stopped"})
}

// RebootContainer reboots a container
func (h *Handler) RebootContainer(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	vmid, err := strconv.Atoi(vars["vmid"])
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid vmid")
		return
	}

	if err := h.client.RebootContainer(vmid); err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{"status": "rebooting"})
}

// DeleteContainer deletes a container
func (h *Handler) DeleteContainer(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	vmid, err := strconv.Atoi(vars["vmid"])
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid vmid")
		return
	}

	if err := h.client.DeleteContainer(vmid); err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Remove container from project assignment if it was assigned
	if h.projectStore != nil {
		if err := h.projectStore.AssignContainer(vmid, ""); err != nil {
			log.Printf("[WARNING] Failed to remove container %d from project assignment: %v", vmid, err)
			// Don't fail the request, just log the warning
		}
	}

	respondJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}

// GetTemplates lists available templates
func (h *Handler) GetTemplates(w http.ResponseWriter, r *http.Request) {
	log.Printf("[DEBUG] GetTemplates handler called")

	templates, err := h.client.GetTemplates()
	if err != nil {
		log.Printf("[ERROR] GetTemplates failed: %v", err)
		// Try to get from cache if Proxmox is down
		if h.cache != nil {
			cached, cacheErr := h.cache.GetTemplates()
			if cacheErr == nil {
				log.Printf("[INFO] Serving templates from cache (Proxmox error: %v)", err)
				respondJSONWithCache(w, http.StatusOK, cached, true)
				return
			}
			log.Printf("[ERROR] Cache retrieval also failed: %v", cacheErr)
		}
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	log.Printf("[INFO] Successfully retrieved %d templates from Proxmox", len(templates))
	if len(templates) > 0 {
		log.Printf("[DEBUG] Sample template data: %+v", templates[0])
	} else {
		log.Printf("[WARNING] Proxmox returned empty template list - check storage configuration and permissions")
	}

	// Cache the templates
	if h.cache != nil {
		if err := h.cache.SetTemplates(templates); err != nil {
			log.Printf("[ERROR] Failed to cache templates: %v", err)
		}
	}

	respondJSONWithCache(w, http.StatusOK, templates, false)
}

// UploadTemplate uploads a new container template
func (h *Handler) UploadTemplate(w http.ResponseWriter, r *http.Request) {
	log.Printf("[DEBUG] UploadTemplate handler called")

	// Parse multipart form (max 5GB)
	if err := r.ParseMultipartForm(5 << 30); err != nil {
		log.Printf("[ERROR] Failed to parse multipart form: %v", err)
		respondError(w, http.StatusBadRequest, "failed to parse upload form")
		return
	}

	// Get the file from the form
	file, header, err := r.FormFile("file")
	if err != nil {
		log.Printf("[ERROR] Failed to get file from form: %v", err)
		respondError(w, http.StatusBadRequest, "file field is required")
		return
	}
	defer func() {
		if err := file.Close(); err != nil {
			log.Printf("Failed to close file: %v", err)
		}
	}()

	// Get storage parameter (default to "local")
	storage := r.FormValue("storage")
	if storage == "" {
		storage = "local"
	}

	log.Printf("[INFO] Uploading template: filename=%s, size=%d bytes, storage=%s", header.Filename, header.Size, storage)

	// Validate file extension
	filename := header.Filename
	validExtensions := []string{".tar.gz", ".tar.xz", ".tar.zst", ".tar.bz2", ".tgz"}
	valid := false
	for _, ext := range validExtensions {
		if strings.HasSuffix(strings.ToLower(filename), ext) {
			valid = true
			break
		}
	}
	if !valid {
		respondError(w, http.StatusBadRequest, "invalid file format. Must be .tar.gz, .tar.xz, .tar.zst, .tar.bz2, or .tgz")
		return
	}

	// Read file data
	fileData, err := io.ReadAll(file)
	if err != nil {
		log.Printf("[ERROR] Failed to read file data: %v", err)
		respondError(w, http.StatusInternalServerError, "failed to read file data")
		return
	}

	// Upload to Proxmox
	if err := h.client.UploadTemplate(storage, filename, fileData); err != nil {
		log.Printf("[ERROR] Failed to upload template: %v", err)
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	log.Printf("[INFO] Template uploaded successfully: %s", filename)
	respondJSON(w, http.StatusOK, map[string]string{
		"status":   "success",
		"filename": filename,
		"storage":  storage,
	})
}

// GetContainerMetrics returns time-series metrics for a container
func (h *Handler) GetContainerMetrics(w http.ResponseWriter, r *http.Request) {
	if h.analytics == nil {
		respondError(w, http.StatusServiceUnavailable, "analytics not available")
		return
	}

	vars := mux.Vars(r)
	vmid, err := strconv.Atoi(vars["vmid"])
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid vmid")
		return
	}

	// Parse query parameters
	hoursStr := r.URL.Query().Get("hours")
	hours := 24 // Default to 24 hours
	if hoursStr != "" {
		if h, err := strconv.Atoi(hoursStr); err == nil && h > 0 {
			hours = h
		}
	}

	limitStr := r.URL.Query().Get("limit")
	limit := 1000 // Default limit
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}

	// Calculate time range
	end := time.Now()
	start := end.Add(-time.Duration(hours) * time.Hour)

	metrics, err := h.analytics.GetMetrics(vmid, start, end, limit)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, metrics)
}

// GetContainerMetricsSummary returns aggregated metrics for a container
func (h *Handler) GetContainerMetricsSummary(w http.ResponseWriter, r *http.Request) {
	if h.analytics == nil {
		respondError(w, http.StatusServiceUnavailable, "analytics not available")
		return
	}

	vars := mux.Vars(r)
	vmid, err := strconv.Atoi(vars["vmid"])
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid vmid")
		return
	}

	// Parse query parameters
	hoursStr := r.URL.Query().Get("hours")
	hours := 24 // Default to 24 hours
	if hoursStr != "" {
		if h, err := strconv.Atoi(hoursStr); err == nil && h > 0 {
			hours = h
		}
	}

	// Calculate time range
	end := time.Now()
	start := end.Add(-time.Duration(hours) * time.Hour)

	summary, err := h.analytics.GetMetricsSummary(vmid, start, end)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, summary)
}

// GetAnalyticsStats returns overall analytics statistics
func (h *Handler) GetAnalyticsStats(w http.ResponseWriter, r *http.Request) {
	if h.analytics == nil {
		respondError(w, http.StatusServiceUnavailable, "analytics not available")
		return
	}

	count, err := h.analytics.GetMetricsCount()
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	stats := map[string]interface{}{
		"total_metrics": count,
		"enabled":       true,
	}

	respondJSON(w, http.StatusOK, stats)
}

// ListVolumes lists all persistent volumes
func (h *Handler) ListVolumes(w http.ResponseWriter, r *http.Request) {
	log.Printf("[DEBUG] ListVolumes handler called")

	volumes, err := h.client.GetVolumes()
	if err != nil {
		log.Printf("[ERROR] GetVolumes failed: %v", err)
		// Try to get from cache if Proxmox is down
		if h.cache != nil {
			cached, cacheErr := h.cache.GetVolumes()
			if cacheErr == nil {
				log.Printf("[INFO] Serving volumes from cache (Proxmox error: %v)", err)
				respondJSONWithCache(w, http.StatusOK, cached, true)
				return
			}
			log.Printf("[ERROR] Cache retrieval also failed: %v", cacheErr)
		}
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	log.Printf("[INFO] Successfully retrieved %d volumes from Proxmox", len(volumes))

	// Cache the volumes
	if h.cache != nil {
		if err := h.cache.SetVolumes(volumes); err != nil {
			log.Printf("[ERROR] Failed to cache volumes: %v", err)
		}
	}

	respondJSONWithCache(w, http.StatusOK, volumes, false)
}

// GetVolume gets a specific volume
func (h *Handler) GetVolume(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	volid := vars["volid"]

	if volid == "" {
		respondError(w, http.StatusBadRequest, "volid is required")
		return
	}

	volume, err := h.client.GetVolume(volid)
	if err != nil {
		// Try to get from cache if Proxmox is down
		if h.cache != nil {
			cached, cacheErr := h.cache.GetVolume(volid)
			if cacheErr == nil {
				log.Printf("Serving volume %s from cache (Proxmox error: %v)", volid, err)
				respondJSONWithCache(w, http.StatusOK, cached, true)
				return
			}
		}
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Cache the volume
	if h.cache != nil {
		if err := h.cache.SetVolume(*volume); err != nil {
			log.Printf("Failed to cache volume %s: %v", volid, err)
		}
	}

	respondJSONWithCache(w, http.StatusOK, volume, false)
}

// CreateVolume creates a new persistent volume
func (h *Handler) CreateVolume(w http.ResponseWriter, r *http.Request) {
	var req proxmox.CreateVolumeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	// Validate required fields
	if req.Name == "" {
		respondError(w, http.StatusBadRequest, "name is required")
		return
	}
	if req.Size <= 0 {
		respondError(w, http.StatusBadRequest, "size must be greater than 0")
		return
	}

	volume, err := h.client.CreateVolume(req)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusCreated, volume)
}

// DeleteVolume deletes a volume
func (h *Handler) DeleteVolume(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	volid := vars["volid"]

	if volid == "" {
		respondError(w, http.StatusBadRequest, "volid is required")
		return
	}

	if err := h.client.DeleteVolume(volid); err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}

// AttachVolume attaches a volume to a container
func (h *Handler) AttachVolume(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	volid := vars["volid"]
	vmidStr := vars["vmid"]

	if volid == "" || vmidStr == "" {
		respondError(w, http.StatusBadRequest, "volid and vmid are required")
		return
	}

	vmid, err := strconv.Atoi(vmidStr)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid vmid")
		return
	}

	var req proxmox.AttachVolumeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		// If body is empty, just use default values
		req = proxmox.AttachVolumeRequest{VMID: vmid}
	} else {
		req.VMID = vmid // Override with path parameter
	}

	if err := h.client.AttachVolume(volid, req); err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{"status": "attached"})
}

// DetachVolume detaches a volume from a container
func (h *Handler) DetachVolume(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	volid := vars["volid"]
	vmidStr := vars["vmid"]

	if volid == "" || vmidStr == "" {
		respondError(w, http.StatusBadRequest, "volid and vmid are required")
		return
	}

	vmid, err := strconv.Atoi(vmidStr)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid vmid")
		return
	}

	var req proxmox.DetachVolumeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		// If body is empty, just use default values
		req = proxmox.DetachVolumeRequest{VMID: vmid}
	} else {
		req.VMID = vmid // Override with path parameter
	}

	if err := h.client.DetachVolume(volid, req); err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{"status": "detached"})
}

// CreateSnapshot creates a snapshot of a volume
func (h *Handler) CreateSnapshot(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	volid := vars["volid"]

	if volid == "" {
		respondError(w, http.StatusBadRequest, "volid is required")
		return
	}

	var req proxmox.CreateSnapshotRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Name == "" {
		respondError(w, http.StatusBadRequest, "snapshot name is required")
		return
	}

	snapshot, err := h.client.CreateSnapshot(volid, req)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusCreated, snapshot)
}

// ListSnapshots lists all snapshots for a volume
func (h *Handler) ListSnapshots(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	volid := vars["volid"]

	if volid == "" {
		respondError(w, http.StatusBadRequest, "volid is required")
		return
	}

	snapshots, err := h.client.GetSnapshots(volid)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, snapshots)
}

// RestoreSnapshot restores a volume from a snapshot
func (h *Handler) RestoreSnapshot(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	volid := vars["volid"]

	if volid == "" {
		respondError(w, http.StatusBadRequest, "volid is required")
		return
	}

	var req proxmox.RestoreSnapshotRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.SnapshotName == "" {
		respondError(w, http.StatusBadRequest, "snapshot_name is required")
		return
	}

	if err := h.client.RestoreSnapshot(volid, req); err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{"status": "restored"})
}

// CloneSnapshot clones a volume from a snapshot
func (h *Handler) CloneSnapshot(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	volid := vars["volid"]

	if volid == "" {
		respondError(w, http.StatusBadRequest, "volid is required")
		return
	}

	var req proxmox.CloneSnapshotRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.SnapshotName == "" {
		respondError(w, http.StatusBadRequest, "snapshot_name is required")
		return
	}

	if req.NewName == "" {
		respondError(w, http.StatusBadRequest, "new_name is required")
		return
	}

	volume, err := h.client.CloneSnapshot(volid, req)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusCreated, volume)
}

// ListProjects lists all projects
func (h *Handler) ListProjects(w http.ResponseWriter, r *http.Request) {
	if h.projectStore == nil {
		respondError(w, http.StatusServiceUnavailable, "project store not available")
		return
	}

	projects, err := h.projectStore.ListProjects()
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, projects)
}

// CreateProject creates a new project
func (h *Handler) CreateProject(w http.ResponseWriter, r *http.Request) {
	if h.projectStore == nil {
		respondError(w, http.StatusServiceUnavailable, "project store not available")
		return
	}

	var req proxmox.CreateProjectRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Name == "" {
		respondError(w, http.StatusBadRequest, "name is required")
		return
	}

	// Generate project ID first (we'll use this for SDN naming)
	projectID := generateID()

	// If network configuration is provided, validate and create SDN resources
	if req.Network != nil && req.Network.Subnet != "" {
		log.Printf("[INFO] Creating SDN network for project %s: subnet=%s, gateway=%s", req.Name, req.Network.Subnet, req.Network.Gateway)

		// Validate subnet format
		if !proxmox.IsValidCIDR(req.Network.Subnet) {
			respondError(w, http.StatusBadRequest, "invalid subnet CIDR format (must be network address like 10.0.1.0/24)")
			return
		}

		// Validate gateway is required
		if req.Network.Gateway == "" {
			respondError(w, http.StatusBadRequest, "gateway is required when subnet is specified")
			return
		}

		// Validate gateway is within subnet
		if err := proxmox.ValidateGatewayInSubnet(req.Network.Subnet, req.Network.Gateway); err != nil {
			respondError(w, http.StatusBadRequest, fmt.Sprintf("gateway validation failed: %v", err))
			return
		}

		// Generate unique zone/vnet identifier from project ID
		// Format: "prj" + first 5 chars of projectID = 8 chars total
		sdnID := proxmox.GenerateSDNIdentifier(projectID, req.Name)
		zone := sdnID
		vnetID := sdnID

		log.Printf("[INFO] Generated SDN identifiers: zone=%s, vnet=%s for project %s", zone, vnetID, req.Name)

		// Create per-project SDN zone with DHCP enabled
		log.Printf("[INFO] Creating dedicated SDN zone: %s with DHCP enabled", zone)
		if err := h.client.CreateSDNZone(zone, "simple", "", true); err != nil {
			log.Printf("[ERROR] Failed to create SDN zone: %v", err)
			respondError(w, http.StatusInternalServerError, fmt.Sprintf("failed to create SDN zone: %v", err))
			return
		}

		// Create VNet in the new zone
		log.Printf("[INFO] Creating VNet: %s in zone: %s", vnetID, zone)
		if err := h.client.CreateVNet(vnetID, zone, req.Network.VLanTag); err != nil {
			log.Printf("[ERROR] Failed to create VNet: %v, attempting to cleanup zone", err)
			// Cleanup zone on failure
			if cleanupErr := h.client.DeleteSDNZone(zone); cleanupErr != nil {
				log.Printf("[WARNING] Failed to cleanup zone after VNet creation failure: %v", cleanupErr)
			}
			respondError(w, http.StatusInternalServerError, fmt.Sprintf("failed to create VNet: %v", err))
			return
		}

		// Calculate DHCP range
		dhcpRange, err := proxmox.CalculateDHCPRange(req.Network.Subnet, req.Network.Gateway)
		if err != nil {
			log.Printf("[ERROR] Failed to calculate DHCP range: %v", err)
			// Cleanup VNet and zone on failure
			h.client.DeleteVNet(vnetID)
			h.client.DeleteSDNZone(zone)
			respondError(w, http.StatusInternalServerError, fmt.Sprintf("failed to calculate DHCP range: %v", err))
			return
		}

		// Create Subnet with DHCP
		log.Printf("[INFO] Creating subnet: %s with gateway: %s and DHCP range: %s", req.Network.Subnet, req.Network.Gateway, dhcpRange)
		if err := h.client.CreateSubnet(vnetID, req.Network.Subnet, req.Network.Gateway, true, dhcpRange); err != nil {
			// Try to cleanup VNet and zone if subnet creation fails
			log.Printf("[ERROR] Failed to create subnet: %v, attempting to cleanup VNet and zone", err)
			h.client.DeleteVNet(vnetID)
			h.client.DeleteSDNZone(zone)
			respondError(w, http.StatusInternalServerError, fmt.Sprintf("failed to create subnet: %v", err))
			return
		}

		// Apply SDN configuration
		log.Printf("[INFO] Applying SDN configuration")
		if err := h.client.ApplySDNConfig(); err != nil {
			log.Printf("[ERROR] Failed to apply SDN config: %v", err)
			// Note: We don't fail the whole operation here as the resources were created
			// The user can manually apply the config from Proxmox GUI
			log.Printf("[WARNING] SDN configuration not applied, please apply manually from Proxmox GUI")
		}

		// Store SDN identifiers in the network config
		req.Network.VNetID = vnetID
		req.Network.Zone = zone
		req.Network.AutoCreatedZone = true // We always auto-create per-project zones now

		log.Printf("[INFO] Successfully created SDN network for project %s", req.Name)
	}

	project, err := h.projectStore.CreateProjectWithID(projectID, req)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusCreated, project)
}

// GetProject gets a project by ID
func (h *Handler) GetProject(w http.ResponseWriter, r *http.Request) {
	if h.projectStore == nil {
		respondError(w, http.StatusServiceUnavailable, "project store not available")
		return
	}

	vars := mux.Vars(r)
	id := vars["id"]

	project, err := h.projectStore.GetProject(id)
	if err != nil {
		respondError(w, http.StatusNotFound, "project not found")
		return
	}

	respondJSON(w, http.StatusOK, project)
}

// UpdateProject updates a project
func (h *Handler) UpdateProject(w http.ResponseWriter, r *http.Request) {
	if h.projectStore == nil {
		respondError(w, http.StatusServiceUnavailable, "project store not available")
		return
	}

	vars := mux.Vars(r)
	id := vars["id"]

	var req proxmox.UpdateProjectRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	project, err := h.projectStore.UpdateProject(id, req)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, project)
}

// DeleteProject deletes a project
func (h *Handler) DeleteProject(w http.ResponseWriter, r *http.Request) {
	if h.projectStore == nil {
		respondError(w, http.StatusServiceUnavailable, "project store not available")
		return
	}

	vars := mux.Vars(r)
	id := vars["id"]

	// Get project first to check for network config
	project, err := h.projectStore.GetProject(id)
	if err != nil {
		respondError(w, http.StatusNotFound, "project not found")
		return
	}

	// Get all containers to verify which ones still exist
	existingContainers, err := h.client.GetContainers()
	if err != nil {
		log.Printf("[WARNING] Failed to get containers for cleanup check: %v", err)
		// Continue anyway, as this is just a cleanup attempt
	}

	// Create a map of existing container VMIDs for quick lookup
	existingVMIDs := make(map[int]bool)
	for _, c := range existingContainers {
		existingVMIDs[c.VMID] = true
	}

	// Get containers assigned to this project
	assignedVMIDs := h.projectStore.GetProjectContainers(id)

	// Clean up any stale assignments (containers that no longer exist in Proxmox)
	activeContainers := 0
	for _, vmid := range assignedVMIDs {
		if !existingVMIDs[vmid] {
			log.Printf("[INFO] Cleaning up stale container assignment: VMID %d in project %s", vmid, id)
			if err := h.projectStore.AssignContainer(vmid, ""); err != nil {
				log.Printf("[WARNING] Failed to clean up stale assignment for VMID %d: %v", vmid, err)
			}
		} else {
			activeContainers++
		}
	}

	// Check if any active containers still exist - block deletion if they do
	if activeContainers > 0 {
		respondError(w, http.StatusConflict,
			fmt.Sprintf("cannot delete project: %d container(s) still assigned. Please delete or reassign containers first.", activeContainers))
		return
	}

	// If project has SDN network, clean it up before deleting the project
	// Order of operations: subnet -> vnet -> zone -> apply SDN config
	if project.Network != nil && project.Network.VNetID != "" {
		log.Printf("[INFO] Cleaning up SDN resources for project %s", project.Name)

		vnetID := project.Network.VNetID
		zoneID := project.Network.Zone

		// Step 1: Delete subnet(s) from vnet first
		// This must be done before deleting the vnet itself
		if project.Network.Subnet != "" {
			log.Printf("[INFO] Deleting subnet: %s from VNet: %s", project.Network.Subnet, vnetID)
			if err := h.client.DeleteSubnet(vnetID, project.Network.Subnet); err != nil {
				log.Printf("[WARNING] Failed to delete subnet: %v (continuing with cleanup)", err)
				// Continue even if subnet deletion fails - the VNet deletion might cascade
			}
		}

		// Step 2: Delete vnet
		// This can only succeed if all subnets are removed
		log.Printf("[INFO] Deleting VNet: %s", vnetID)
		if err := h.client.DeleteVNet(vnetID); err != nil {
			log.Printf("[WARNING] Failed to delete VNet: %v (continuing with cleanup)", err)
			// Continue to try zone deletion
		}

		// Step 3: Delete zone (if auto-created by us)
		// This can only succeed if all vnets are removed from the zone
		if project.Network.AutoCreatedZone {
			log.Printf("[INFO] Deleting auto-created SDN zone: %s", zoneID)
			if err := h.client.DeleteSDNZone(zoneID); err != nil {
				log.Printf("[WARNING] Failed to delete SDN zone: %v", err)
				// Don't fail the whole operation
			}
		}

		// Step 4: Apply SDN config to commit all changes
		// This is equivalent to clicking "Apply" in the Proxmox UI
		log.Printf("[INFO] Applying SDN configuration after cleanup")
		if err := h.client.ApplySDNConfig(); err != nil {
			log.Printf("[WARNING] Failed to apply SDN config after cleanup: %v", err)
		} else {
			log.Printf("[INFO] SDN configuration applied successfully")
		}
	}

	// Delete project from store
	if err := h.projectStore.DeleteProject(id); err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	log.Printf("[INFO] Successfully deleted project %s and cleaned up SDN resources", project.Name)
	respondJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}

// GetProjectContainers gets containers for a project
func (h *Handler) GetProjectContainers(w http.ResponseWriter, r *http.Request) {
	if h.projectStore == nil {
		respondError(w, http.StatusServiceUnavailable, "project store not available")
		return
	}

	vars := mux.Vars(r)
	id := vars["id"]

	log.Printf("[DEBUG] GetProjectContainers called for project ID: %s", id)

	// Get project
	project, err := h.projectStore.GetProject(id)
	if err != nil {
		log.Printf("[ERROR] Project not found: %s", id)
		respondError(w, http.StatusNotFound, "project not found")
		return
	}

	log.Printf("[DEBUG] Project found: %+v", project)

	// Get all containers
	containers, err := h.client.GetContainers()
	if err != nil {
		log.Printf("[ERROR] Failed to get containers: %v", err)
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	log.Printf("[DEBUG] Total containers from Proxmox: %d", len(containers))

	// Enrich with project information
	for i := range containers {
		projectID := h.projectStore.GetContainerProject(containers[i].VMID)
		containers[i].ProjectID = projectID
		log.Printf("[DEBUG] Container %d has project ID: %s", containers[i].VMID, projectID)
	}

	// Filter containers by project
	projectContainers := []proxmox.Container{}
	for _, c := range containers {
		if c.ProjectID == id {
			projectContainers = append(projectContainers, c)
			log.Printf("[DEBUG] Container %d matched project %s", c.VMID, id)
		}
	}

	log.Printf("[DEBUG] Filtered containers for project %s: %d", id, len(projectContainers))

	// Calculate aggregates
	totalCPU := 0
	totalMemMB := int64(0)
	usedMemMB := int64(0)
	running := 0
	stopped := 0

	for _, c := range projectContainers {
		// Note: Container doesn't have CPUCores field, we'll need to get it from container config
		// For now, skip CPU aggregation
		totalMemMB += c.MaxMem / 1024 / 1024
		usedMemMB += c.Mem / 1024 / 1024
		if c.Status == "running" {
			running++
		} else {
			stopped++
		}
	}

	result := map[string]interface{}{
		"project":    project,
		"containers": projectContainers,
		"aggregate": map[string]interface{}{
			"total_containers": len(projectContainers),
			"running":          running,
			"stopped":          stopped,
			"total_cpu_cores":  totalCPU,
			"total_memory_mb":  totalMemMB,
			"used_memory_mb":   usedMemMB,
		},
	}

	respondJSON(w, http.StatusOK, result)
}

// AssignContainerProject assigns a container to a project
func (h *Handler) AssignContainerProject(w http.ResponseWriter, r *http.Request) {
	if h.projectStore == nil {
		respondError(w, http.StatusServiceUnavailable, "project store not available")
		return
	}

	vars := mux.Vars(r)
	vmidStr := vars["vmid"]
	vmid, err := strconv.Atoi(vmidStr)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid vmid")
		return
	}

	var req proxmox.AssignProjectRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	// Verify container exists
	_, err = h.client.GetContainer(vmid)
	if err != nil {
		respondError(w, http.StatusNotFound, "container not found")
		return
	}

	// If project_id is not empty, verify project exists
	if req.ProjectID != "" {
		_, err := h.projectStore.GetProject(req.ProjectID)
		if err != nil {
			respondError(w, http.StatusNotFound, "project not found")
			return
		}
	}

	// Assign/unassign container (empty string means unassign)
	if err := h.projectStore.AssignContainer(vmid, req.ProjectID); err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{"status": "assigned"})
}

// GetStorage lists all storage datastores
func (h *Handler) GetStorage(w http.ResponseWriter, r *http.Request) {
	log.Printf("[DEBUG] GetStorage handler called")

	// Parse query parameters
	req := &proxmox.GetStorageRequest{}

	if content := r.URL.Query().Get("content"); content != "" {
		req.Content = content
	}

	if enabled := r.URL.Query().Get("enabled"); enabled != "" {
		switch enabled {
		case "1", "true":
			enabledBool := true
			req.Enabled = &enabledBool
		case "0", "false":
			enabledBool := false
			req.Enabled = &enabledBool
		}
	}

	if format := r.URL.Query().Get("format"); format != "" {
		switch format {
		case "1", "true":
			formatBool := true
			req.Format = &formatBool
		case "0", "false":
			formatBool := false
			req.Format = &formatBool
		}
	}

	if storage := r.URL.Query().Get("storage"); storage != "" {
		req.Storage = storage
	}

	if target := r.URL.Query().Get("target"); target != "" {
		req.Target = target
	}

	storages, err := h.client.GetStorage(req)
	if err != nil {
		log.Printf("[ERROR] GetStorage failed: %v", err)
		// Try to get from cache if Proxmox is down
		if h.cache != nil {
			cached, cacheErr := h.cache.GetStorage()
			if cacheErr == nil {
				log.Printf("[INFO] Serving storage from cache (Proxmox error: %v)", err)
				respondJSONWithCache(w, http.StatusOK, cached, true)
				return
			}
			log.Printf("[ERROR] Cache retrieval also failed: %v", cacheErr)
		}
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	log.Printf("[INFO] Successfully retrieved %d storage entries from Proxmox", len(storages))
	if len(storages) > 0 {
		log.Printf("[DEBUG] Sample storage data: %+v", storages[0])
	} else {
		log.Printf("[WARNING] Proxmox returned empty storage list")
	}

	// Cache the storage list
	if h.cache != nil {
		if err := h.cache.SetStorage(storages); err != nil {
			log.Printf("[ERROR] Failed to cache storage: %v", err)
		}
	}

	respondJSONWithCache(w, http.StatusOK, storages, false)
}

// GetContainerTermProxy creates a terminal proxy for container console access
func (h *Handler) GetContainerTermProxy(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	vmid, err := strconv.Atoi(vars["vmid"])
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid vmid")
		return
	}

	// Check if container exists and is running
	container, err := h.client.GetContainer(vmid)
	if err != nil {
		respondError(w, http.StatusNotFound, "container not found")
		return
	}

	if container.Status != "running" {
		respondError(w, http.StatusBadRequest, "container must be running to access console")
		return
	}

	// Create terminal proxy using the container's actual node
	proxyData, err := h.client.CreateTermProxy(container.Node, vmid)
	if err != nil {
		log.Printf("[ERROR] Failed to create terminal proxy for container %d on node %s: %v", vmid, container.Node, err)
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	log.Printf("[INFO] Terminal proxy created for container %d on node %s", vmid, container.Node)
	respondJSON(w, http.StatusOK, proxyData)
}
