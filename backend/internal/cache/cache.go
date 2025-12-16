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
		if closeErr := db.Close(); closeErr != nil {
			log.Printf("Failed to close database after initialization error: %v", closeErr)
		}
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

	CREATE TABLE IF NOT EXISTS volumes (
		volid TEXT PRIMARY KEY,
		data TEXT NOT NULL,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS storage (
		id INTEGER PRIMARY KEY DEFAULT 1,
		data TEXT NOT NULL,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	CREATE INDEX IF NOT EXISTS idx_containers_updated ON containers(updated_at);
	CREATE INDEX IF NOT EXISTS idx_volumes_updated ON volumes(updated_at);
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
	defer func() {
		if err := tx.Rollback(); err != nil && err != sql.ErrTxDone {
			log.Printf("Failed to rollback transaction: %v", err)
		}
	}()

	// Clear old data
	if _, err := tx.Exec("DELETE FROM containers"); err != nil {
		return err
	}

	// Insert new data
	stmt, err := tx.Prepare("INSERT INTO containers (vmid, data, updated_at) VALUES (?, ?, ?)")
	if err != nil {
		return err
	}
	defer func() {
		if closeErr := stmt.Close(); closeErr != nil {
			log.Printf("Failed to close statement: %v", closeErr)
		}
	}()

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
	defer func() {
		if closeErr := rows.Close(); closeErr != nil {
			log.Printf("Failed to close rows: %v", closeErr)
		}
	}()

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

// SetVolumes caches the list of volumes
func (c *Cache) SetVolumes(volumes []proxmox.Volume) error {
	tx, err := c.db.Begin()
	if err != nil {
		return err
	}
	defer func() {
		if err := tx.Rollback(); err != nil && err != sql.ErrTxDone {
			log.Printf("Failed to rollback transaction: %v", err)
		}
	}()

	// Clear old data
	if _, err := tx.Exec("DELETE FROM volumes"); err != nil {
		return err
	}

	// Insert new data
	stmt, err := tx.Prepare("INSERT INTO volumes (volid, data, updated_at) VALUES (?, ?, ?)")
	if err != nil {
		return err
	}
	defer func() {
		if closeErr := stmt.Close(); closeErr != nil {
			log.Printf("Failed to close statement: %v", closeErr)
		}
	}()

	now := time.Now()
	for _, volume := range volumes {
		data, err := json.Marshal(volume)
		if err != nil {
			log.Printf("Failed to marshal volume %s: %v", volume.VolID, err)
			continue
		}

		if _, err := stmt.Exec(volume.VolID, string(data), now); err != nil {
			log.Printf("Failed to cache volume %s: %v", volume.VolID, err)
		}
	}

	return tx.Commit()
}

// GetVolumes retrieves cached volumes
func (c *Cache) GetVolumes() ([]proxmox.Volume, error) {
	rows, err := c.db.Query("SELECT data FROM volumes ORDER BY volid")
	if err != nil {
		return nil, err
	}
	defer func() {
		if closeErr := rows.Close(); closeErr != nil {
			log.Printf("Failed to close rows: %v", closeErr)
		}
	}()

	var volumes []proxmox.Volume
	for rows.Next() {
		var data string
		if err := rows.Scan(&data); err != nil {
			log.Printf("Failed to scan volume: %v", err)
			continue
		}

		var volume proxmox.Volume
		if err := json.Unmarshal([]byte(data), &volume); err != nil {
			log.Printf("Failed to unmarshal volume: %v", err)
			continue
		}

		volumes = append(volumes, volume)
	}

	return volumes, rows.Err()
}

// SetVolume caches a single volume
func (c *Cache) SetVolume(volume proxmox.Volume) error {
	data, err := json.Marshal(volume)
	if err != nil {
		return err
	}

	_, err = c.db.Exec(
		"INSERT OR REPLACE INTO volumes (volid, data, updated_at) VALUES (?, ?, ?)",
		volume.VolID, string(data), time.Now(),
	)
	return err
}

// GetVolume retrieves a cached volume
func (c *Cache) GetVolume(volid string) (*proxmox.Volume, error) {
	var data string
	err := c.db.QueryRow("SELECT data FROM volumes WHERE volid = ?", volid).Scan(&data)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("volume %s not found in cache", volid)
		}
		return nil, err
	}

	var volume proxmox.Volume
	if err := json.Unmarshal([]byte(data), &volume); err != nil {
		return nil, err
	}

	return &volume, nil
}

// DeleteVolume removes a volume from cache
func (c *Cache) DeleteVolume(volid string) error {
	_, err := c.db.Exec("DELETE FROM volumes WHERE volid = ?", volid)
	return err
}

// SetStorage caches storage information
func (c *Cache) SetStorage(storages []proxmox.Storage) error {
	data, err := json.Marshal(storages)
	if err != nil {
		return err
	}

	_, err = c.db.Exec(
		"INSERT OR REPLACE INTO storage (id, data, updated_at) VALUES (1, ?, ?)",
		string(data), time.Now(),
	)
	return err
}

// GetStorage retrieves cached storage information
func (c *Cache) GetStorage() ([]proxmox.Storage, error) {
	var data string
	err := c.db.QueryRow("SELECT data FROM storage WHERE id = 1").Scan(&data)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("storage not found in cache")
		}
		return nil, err
	}

	var storages []proxmox.Storage
	if err := json.Unmarshal([]byte(data), &storages); err != nil {
		return nil, err
	}

	return storages, nil
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
	_, err := c.db.Exec("DELETE FROM containers; DELETE FROM dashboard; DELETE FROM templates; DELETE FROM volumes; DELETE FROM storage")
	return err
}
