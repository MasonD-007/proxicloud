package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/MasonD-007/proxicloud/backend/internal/analytics"
	"github.com/MasonD-007/proxicloud/backend/internal/cache"
	"github.com/MasonD-007/proxicloud/backend/internal/config"
	"github.com/MasonD-007/proxicloud/backend/internal/handlers"
	"github.com/MasonD-007/proxicloud/backend/internal/middleware"
	"github.com/MasonD-007/proxicloud/backend/internal/proxmox"
	"github.com/gorilla/mux"
	"github.com/rs/cors"
)

func main() {
	// Load configuration
	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		configPath = "/etc/proxicloud/config.yaml"
	}

	cfg, err := config.Load(configPath)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	if err := cfg.Validate(); err != nil {
		log.Fatalf("Invalid config: %v", err)
	}

	// Create Proxmox client
	client := proxmox.NewClient(
		cfg.Proxmox.Host,
		cfg.Proxmox.Node,
		cfg.Proxmox.TokenID,
		cfg.Proxmox.TokenSecret,
		cfg.Proxmox.Insecure,
	)

	// Initialize cache
	cacheDB := os.Getenv("CACHE_PATH")
	if cacheDB == "" {
		cacheDB = "/var/lib/proxicloud/cache.db"
	}

	cacheInstance, err := cache.NewCache(cacheDB)
	if err != nil {
		log.Printf("Warning: Failed to initialize cache: %v (continuing without cache)", err)
		cacheInstance = nil
	} else {
		defer cacheInstance.Close()
		log.Printf("Cache initialized at %s", cacheDB)
	}

	// Initialize analytics
	analyticsDB := os.Getenv("ANALYTICS_PATH")
	if analyticsDB == "" {
		analyticsDB = "/var/lib/proxicloud/analytics.db"
	}

	analyticsInstance, err := analytics.NewAnalytics(analyticsDB)
	if err != nil {
		log.Printf("Warning: Failed to initialize analytics: %v (continuing without analytics)", err)
		analyticsInstance = nil
	} else {
		defer analyticsInstance.Close()
		log.Printf("Analytics initialized at %s", analyticsDB)

		// Start metrics collector
		collector := analytics.NewCollector(client, analyticsInstance, 30) // 30 second interval
		collector.Start()
		defer collector.Stop()

		// Start cleanup task (30-day retention)
		collector.RunCleanup(30)

		log.Println("Metrics collector started (30-second intervals, 30-day retention)")
	}

	// Create handlers
	h := handlers.NewHandler(client, cacheInstance, analyticsInstance)

	// Set up router
	router := mux.NewRouter()
	api := router.PathPrefix("/api").Subrouter()

	// Routes
	api.HandleFunc("/health", h.Health).Methods("GET")
	api.HandleFunc("/dashboard", h.Dashboard).Methods("GET")
	api.HandleFunc("/containers", h.ListContainers).Methods("GET")
	api.HandleFunc("/containers", h.CreateContainer).Methods("POST")
	api.HandleFunc("/containers/{vmid}", h.GetContainer).Methods("GET")
	api.HandleFunc("/containers/{vmid}", h.DeleteContainer).Methods("DELETE")
	api.HandleFunc("/containers/{vmid}/start", h.StartContainer).Methods("POST")
	api.HandleFunc("/containers/{vmid}/stop", h.StopContainer).Methods("POST")
	api.HandleFunc("/containers/{vmid}/reboot", h.RebootContainer).Methods("POST")
	api.HandleFunc("/templates", h.GetTemplates).Methods("GET")
	api.HandleFunc("/templates/upload", h.UploadTemplate).Methods("POST")

	// Analytics routes
	api.HandleFunc("/analytics/stats", h.GetAnalyticsStats).Methods("GET")
	api.HandleFunc("/containers/{vmid}/metrics", h.GetContainerMetrics).Methods("GET")
	api.HandleFunc("/containers/{vmid}/metrics/summary", h.GetContainerMetricsSummary).Methods("GET")

	// Volume routes
	api.HandleFunc("/volumes", h.ListVolumes).Methods("GET")
	api.HandleFunc("/volumes", h.CreateVolume).Methods("POST")
	api.HandleFunc("/volumes/{volid}", h.GetVolume).Methods("GET")
	api.HandleFunc("/volumes/{volid}", h.DeleteVolume).Methods("DELETE")
	api.HandleFunc("/volumes/{volid}/attach/{vmid}", h.AttachVolume).Methods("POST")
	api.HandleFunc("/volumes/{volid}/detach/{vmid}", h.DetachVolume).Methods("POST")
	api.HandleFunc("/volumes/{volid}/snapshots", h.ListSnapshots).Methods("GET")
	api.HandleFunc("/volumes/{volid}/snapshots", h.CreateSnapshot).Methods("POST")
	api.HandleFunc("/volumes/{volid}/snapshots/restore", h.RestoreSnapshot).Methods("POST")
	api.HandleFunc("/volumes/{volid}/snapshots/clone", h.CloneSnapshot).Methods("POST")

	// Set up CORS
	c := cors.New(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"*"},
		AllowCredentials: true,
	})

	handler := c.Handler(router)

	// Apply middleware
	handler = middleware.Logger(handler)
	handler = middleware.Recovery(handler)

	// Start server
	addr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)
	log.Printf("Starting ProxiCloud API server on %s", addr)
	if err := http.ListenAndServe(addr, handler); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
