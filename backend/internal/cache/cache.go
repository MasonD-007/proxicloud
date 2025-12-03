package cache

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/MasonD-007/proxicloud/backend/internal/proxmox"
	_ "github.com/mattn/go-sqlite3"
)

// Cache provides offline caching for Proxmox data
type Cache struct {
	db *sql.DB
}

// NewCache creates a new cache instance
func NewCache(dbPath string) (*Cache, error) {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open cache database: %w", err)
	}

	cache := &Cache{db: db}
	if err := cache.initialize(); err != nil {
		db.Close()
		return nil, err
	}

	return cache, nil
}

// initialize creates the cache tables if they don't exist
func (c *Cache) initialize() error {
	schema := `
	CREATE TABLE IF NOT EXISTS containers (
		vmid INTEGER PRIMARY KEY,
		data TEXT NOT NULL,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS dashboard (
		id INTEGER PRIMARY KEY DEFAULT 1,
		data TEXT NOT NULL,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS templates (
		id INTEGER PRIMARY KEY DEFAULT 1,
		data TEXT NOT NULL,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	CREATE INDEX IF NOT EXISTS idx_containers_updated ON containers(updated_at);
	`

	if _, err := c.db.Exec(schema); err != nil {
		return fmt.Errorf("failed to create cache tables: %w", err)
	}

	return nil
}

// Close closes the cache database
func (c *Cache) Close() error {
	return c.db.Close()
}

// SetContainers caches the list of containers
func (c *Cache) SetContainers(containers []proxmox.Container) error {
	tx, err := c.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Clear old data
	if _, err := tx.Exec("DELETE FROM containers"); err != nil {
		return err
	}

	// Insert new data
	stmt, err := tx.Prepare("INSERT INTO containers (vmid, data, updated_at) VALUES (?, ?, ?)")
	if err != nil {
		return err
	}
	defer stmt.Close()

	now := time.Now()
	for _, container := range containers {
		data, err := json.Marshal(container)
		if err != nil {
			log.Printf("Failed to marshal container %d: %v", container.VMID, err)
			continue
		}

		if _, err := stmt.Exec(container.VMID, string(data), now); err != nil {
			log.Printf("Failed to cache container %d: %v", container.VMID, err)
		}
	}

	return tx.Commit()
}

// GetContainers retrieves cached containers
func (c *Cache) GetContainers() ([]proxmox.Container, error) {
	rows, err := c.db.Query("SELECT data FROM containers ORDER BY vmid")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var containers []proxmox.Container
	for rows.Next() {
		var data string
		if err := rows.Scan(&data); err != nil {
			log.Printf("Failed to scan container: %v", err)
			continue
		}

		var container proxmox.Container
		if err := json.Unmarshal([]byte(data), &container); err != nil {
			log.Printf("Failed to unmarshal container: %v", err)
			continue
		}

		containers = append(containers, container)
	}

	return containers, rows.Err()
}

// SetContainer caches a single container
func (c *Cache) SetContainer(container proxmox.Container) error {
	data, err := json.Marshal(container)
	if err != nil {
		return err
	}

	_, err = c.db.Exec(
		"INSERT OR REPLACE INTO containers (vmid, data, updated_at) VALUES (?, ?, ?)",
		container.VMID, string(data), time.Now(),
	)
	return err
}

// GetContainer retrieves a cached container
func (c *Cache) GetContainer(vmid int) (*proxmox.Container, error) {
	var data string
	err := c.db.QueryRow("SELECT data FROM containers WHERE vmid = ?", vmid).Scan(&data)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("container %d not found in cache", vmid)
		}
		return nil, err
	}

	var container proxmox.Container
	if err := json.Unmarshal([]byte(data), &container); err != nil {
		return nil, err
	}

	return &container, nil
}

// DeleteContainer removes a container from cache
func (c *Cache) DeleteContainer(vmid int) error {
	_, err := c.db.Exec("DELETE FROM containers WHERE vmid = ?", vmid)
	return err
}

// SetDashboard caches dashboard stats
func (c *Cache) SetDashboard(stats interface{}) error {
	data, err := json.Marshal(stats)
	if err != nil {
		return err
	}

	_, err = c.db.Exec(
		"INSERT OR REPLACE INTO dashboard (id, data, updated_at) VALUES (1, ?, ?)",
		string(data), time.Now(),
	)
	return err
}

// GetDashboard retrieves cached dashboard stats
func (c *Cache) GetDashboard() (map[string]interface{}, error) {
	var data string
	err := c.db.QueryRow("SELECT data FROM dashboard WHERE id = 1").Scan(&data)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("dashboard not found in cache")
		}
		return nil, err
	}

	var stats map[string]interface{}
	if err := json.Unmarshal([]byte(data), &stats); err != nil {
		return nil, err
	}

	return stats, nil
}

// SetTemplates caches templates
func (c *Cache) SetTemplates(templates []proxmox.Template) error {
	data, err := json.Marshal(templates)
	if err != nil {
		return err
	}

	_, err = c.db.Exec(
		"INSERT OR REPLACE INTO templates (id, data, updated_at) VALUES (1, ?, ?)",
		string(data), time.Now(),
	)
	return err
}

// GetTemplates retrieves cached templates
func (c *Cache) GetTemplates() ([]proxmox.Template, error) {
	var data string
	err := c.db.QueryRow("SELECT data FROM templates WHERE id = 1").Scan(&data)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("templates not found in cache")
		}
		return nil, err
	}

	var templates []proxmox.Template
	if err := json.Unmarshal([]byte(data), &templates); err != nil {
		return nil, err
	}

	return templates, nil
}

// GetCacheAge returns the age of the cache in seconds
func (c *Cache) GetCacheAge() (int64, error) {
	var updatedAt time.Time
	err := c.db.QueryRow("SELECT updated_at FROM containers ORDER BY updated_at DESC LIMIT 1").Scan(&updatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return 0, nil
		}
		return 0, err
	}

	return int64(time.Since(updatedAt).Seconds()), nil
}

// Clear clears all cached data
func (c *Cache) Clear() error {
	_, err := c.db.Exec("DELETE FROM containers; DELETE FROM dashboard; DELETE FROM templates")
	return err
}
