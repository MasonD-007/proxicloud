package analytics

import (
	"context"
	"log"
	"time"

	"github.com/MasonD-007/proxicloud/backend/internal/proxmox"
)

// Collector handles background metrics collection
type Collector struct {
	client    *proxmox.Client
	analytics *Analytics
	interval  time.Duration
	ctx       context.Context
	cancel    context.CancelFunc
}

// NewCollector creates a new metrics collector
func NewCollector(client *proxmox.Client, analytics *Analytics, intervalSeconds int) *Collector {
	ctx, cancel := context.WithCancel(context.Background())

	return &Collector{
		client:    client,
		analytics: analytics,
		interval:  time.Duration(intervalSeconds) * time.Second,
		ctx:       ctx,
		cancel:    cancel,
	}
}

// Start begins collecting metrics in the background
func (c *Collector) Start() {
	log.Printf("Starting metrics collector (interval: %v)", c.interval)

	// Collect immediately on start
	go c.collectOnce()

	// Then collect on interval
	ticker := time.NewTicker(c.interval)
	go func() {
		for {
			select {
			case <-ticker.C:
				c.collectOnce()
			case <-c.ctx.Done():
				ticker.Stop()
				log.Println("Metrics collector stopped")
				return
			}
		}
	}()
}

// Stop stops the metrics collector
func (c *Collector) Stop() {
	log.Println("Stopping metrics collector...")
	c.cancel()
}

// collectOnce collects metrics for all containers once
func (c *Collector) collectOnce() {
	start := time.Now()

	// Get all containers
	containers, err := c.client.GetContainers()
	if err != nil {
		log.Printf("Failed to list containers for metrics collection: %v", err)
		return
	}

	if len(containers) == 0 {
		// No containers to collect metrics for
		return
	}

	var metrics []Metric
	timestamp := time.Now()

	// Collect metrics for each container
	for _, container := range containers {
		// Parse metrics from container status
		metric := Metric{
			VMID:      container.VMID,
			Timestamp: timestamp,
			Status:    container.Status,
			Uptime:    container.Uptime,
			CPUUsage:  container.CPU * 100, // Convert to percentage
			MemUsage:  container.Mem,
			MemTotal:  container.MaxMem,
			DiskUsage: container.Disk,
			DiskTotal: container.MaxDisk,
			NetIn:     0, // Network stats not directly available in list
			NetOut:    0, // Would need individual container queries
		}

		metrics = append(metrics, metric)
	}

	// Store all metrics
	if len(metrics) > 0 {
		if err := c.analytics.RecordMetrics(metrics); err != nil {
			log.Printf("Failed to record metrics: %v", err)
		} else {
			duration := time.Since(start)
			log.Printf("Collected metrics for %d containers in %v", len(metrics), duration)
		}
	}
}

// CollectForContainer collects metrics for a specific container
func (c *Collector) CollectForContainer(vmid int) error {
	container, err := c.client.GetContainer(vmid)
	if err != nil {
		return err
	}

	metric := Metric{
		VMID:      vmid,
		Timestamp: time.Now(),
		Status:    container.Status,
		Uptime:    container.Uptime,
		CPUUsage:  container.CPU * 100,
		MemUsage:  container.Mem,
		MemTotal:  container.MaxMem,
		DiskUsage: container.Disk,
		DiskTotal: container.MaxDisk,
		NetIn:     0,
		NetOut:    0,
	}

	return c.analytics.RecordMetric(metric)
}

// RunCleanup runs the cleanup task to remove old metrics
func (c *Collector) RunCleanup(retentionDays int) {
	ticker := time.NewTicker(24 * time.Hour) // Run daily
	go func() {
		// Run immediately on start
		if err := c.analytics.CleanOldMetrics(retentionDays); err != nil {
			log.Printf("Failed to clean old metrics: %v", err)
		}

		for {
			select {
			case <-ticker.C:
				if err := c.analytics.CleanOldMetrics(retentionDays); err != nil {
					log.Printf("Failed to clean old metrics: %v", err)
				}
			case <-c.ctx.Done():
				ticker.Stop()
				return
			}
		}
	}()
}
