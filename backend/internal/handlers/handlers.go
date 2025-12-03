package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/MasonD-007/proxicloud/backend/internal/analytics"
	"github.com/MasonD-007/proxicloud/backend/internal/cache"
	"github.com/MasonD-007/proxicloud/backend/internal/proxmox"
	"github.com/gorilla/mux"
)

// Handler holds the Proxmox client, cache, and analytics
type Handler struct {
	client    *proxmox.Client
	cache     *cache.Cache
	analytics *analytics.Analytics
}

// NewHandler creates a new handler
func NewHandler(client *proxmox.Client, cache *cache.Cache, analytics *analytics.Analytics) *Handler {
	return &Handler{
		client:    client,
		cache:     cache,
		analytics: analytics,
	}
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
	containers, err := h.client.GetContainers()
	if err != nil {
		// Try to get from cache if Proxmox is down
		if h.cache != nil {
			cached, cacheErr := h.cache.GetContainers()
			if cacheErr == nil {
				log.Printf("Serving containers from cache (Proxmox error: %v)", err)
				respondJSONWithCache(w, http.StatusOK, cached, true)
				return
			}
		}
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Cache the containers
	if h.cache != nil {
		if err := h.cache.SetContainers(containers); err != nil {
			log.Printf("Failed to cache containers: %v", err)
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

	// Get next available VMID
	vmid, err := h.client.GetNextVMID()
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	if err := h.client.CreateContainer(vmid, req); err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
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

	respondJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}

// GetTemplates lists available templates
func (h *Handler) GetTemplates(w http.ResponseWriter, r *http.Request) {
	templates, err := h.client.GetTemplates()
	if err != nil {
		// Try to get from cache if Proxmox is down
		if h.cache != nil {
			cached, cacheErr := h.cache.GetTemplates()
			if cacheErr == nil {
				log.Printf("Serving templates from cache (Proxmox error: %v)", err)
				respondJSONWithCache(w, http.StatusOK, cached, true)
				return
			}
		}
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Cache the templates
	if h.cache != nil {
		if err := h.cache.SetTemplates(templates); err != nil {
			log.Printf("Failed to cache templates: %v", err)
		}
	}

	respondJSONWithCache(w, http.StatusOK, templates, false)
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
