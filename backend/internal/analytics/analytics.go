package analytics

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

// Analytics handles time-series metrics collection
type Analytics struct {
	db *sql.DB
}

// Metric represents a single metrics data point
type Metric struct {
	VMID      int       `json:"vmid"`
	Timestamp time.Time `json:"timestamp"`
	CPUUsage  float64   `json:"cpu_usage"`
	MemUsage  int64     `json:"mem_usage"`
	MemTotal  int64     `json:"mem_total"`
	DiskUsage int64     `json:"disk_usage"`
	DiskTotal int64     `json:"disk_total"`
	NetIn     int64     `json:"net_in"`
	NetOut    int64     `json:"net_out"`
	Uptime    int64     `json:"uptime"`
	Status    string    `json:"status"`
}

// MetricSummary represents aggregated metrics
type MetricSummary struct {
	VMID         int       `json:"vmid"`
	StartTime    time.Time `json:"start_time"`
	EndTime      time.Time `json:"end_time"`
	AvgCPU       float64   `json:"avg_cpu"`
	MaxCPU       float64   `json:"max_cpu"`
	AvgMemUsage  float64   `json:"avg_mem_usage"`
	MaxMemUsage  int64     `json:"max_mem_usage"`
	AvgDiskUsage float64   `json:"avg_disk_usage"`
	TotalNetIn   int64     `json:"total_net_in"`
	TotalNetOut  int64     `json:"total_net_out"`
	DataPoints   int       `json:"data_points"`
}

// NewAnalytics creates a new analytics instance
func NewAnalytics(dbPath string) (*Analytics, error) {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open analytics database: %w", err)
	}

	analytics := &Analytics{db: db}
	if err := analytics.initialize(); err != nil {
		db.Close()
		return nil, err
	}

	return analytics, nil
}

// initialize creates the analytics tables
func (a *Analytics) initialize() error {
	schema := `
	CREATE TABLE IF NOT EXISTS metrics (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		vmid INTEGER NOT NULL,
		timestamp DATETIME NOT NULL,
		cpu_usage REAL,
		mem_usage INTEGER,
		mem_total INTEGER,
		disk_usage INTEGER,
		disk_total INTEGER,
		net_in INTEGER,
		net_out INTEGER,
		uptime INTEGER,
		status TEXT,
		UNIQUE(vmid, timestamp)
	);

	CREATE INDEX IF NOT EXISTS idx_metrics_vmid ON metrics(vmid);
	CREATE INDEX IF NOT EXISTS idx_metrics_timestamp ON metrics(timestamp);
	CREATE INDEX IF NOT EXISTS idx_metrics_vmid_timestamp ON metrics(vmid, timestamp);
	`

	if _, err := a.db.Exec(schema); err != nil {
		return fmt.Errorf("failed to create analytics tables: %w", err)
	}

	return nil
}

// Close closes the analytics database
func (a *Analytics) Close() error {
	return a.db.Close()
}

// RecordMetric stores a single metric data point
func (a *Analytics) RecordMetric(metric Metric) error {
	query := `
		INSERT OR REPLACE INTO metrics 
		(vmid, timestamp, cpu_usage, mem_usage, mem_total, disk_usage, disk_total, net_in, net_out, uptime, status)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	_, err := a.db.Exec(
		query,
		metric.VMID,
		metric.Timestamp,
		metric.CPUUsage,
		metric.MemUsage,
		metric.MemTotal,
		metric.DiskUsage,
		metric.DiskTotal,
		metric.NetIn,
		metric.NetOut,
		metric.Uptime,
		metric.Status,
	)

	if err != nil {
		return fmt.Errorf("failed to record metric: %w", err)
	}

	return nil
}

// RecordMetrics stores multiple metrics at once
func (a *Analytics) RecordMetrics(metrics []Metric) error {
	tx, err := a.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare(`
		INSERT OR REPLACE INTO metrics 
		(vmid, timestamp, cpu_usage, mem_usage, mem_total, disk_usage, disk_total, net_in, net_out, uptime, status)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, metric := range metrics {
		_, err := stmt.Exec(
			metric.VMID,
			metric.Timestamp,
			metric.CPUUsage,
			metric.MemUsage,
			metric.MemTotal,
			metric.DiskUsage,
			metric.DiskTotal,
			metric.NetIn,
			metric.NetOut,
			metric.Uptime,
			metric.Status,
		)
		if err != nil {
			log.Printf("Failed to record metric for VMID %d: %v", metric.VMID, err)
		}
	}

	return tx.Commit()
}

// GetMetrics retrieves metrics for a container within a time range
func (a *Analytics) GetMetrics(vmid int, start, end time.Time, limit int) ([]Metric, error) {
	query := `
		SELECT vmid, timestamp, cpu_usage, mem_usage, mem_total, disk_usage, disk_total, net_in, net_out, uptime, status
		FROM metrics
		WHERE vmid = ? AND timestamp BETWEEN ? AND ?
		ORDER BY timestamp DESC
		LIMIT ?
	`

	rows, err := a.db.Query(query, vmid, start, end, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var metrics []Metric
	for rows.Next() {
		var m Metric
		err := rows.Scan(
			&m.VMID,
			&m.Timestamp,
			&m.CPUUsage,
			&m.MemUsage,
			&m.MemTotal,
			&m.DiskUsage,
			&m.DiskTotal,
			&m.NetIn,
			&m.NetOut,
			&m.Uptime,
			&m.Status,
		)
		if err != nil {
			log.Printf("Failed to scan metric: %v", err)
			continue
		}
		metrics = append(metrics, m)
	}

	return metrics, rows.Err()
}

// GetLatestMetric retrieves the most recent metric for a container
func (a *Analytics) GetLatestMetric(vmid int) (*Metric, error) {
	query := `
		SELECT vmid, timestamp, cpu_usage, mem_usage, mem_total, disk_usage, disk_total, net_in, net_out, uptime, status
		FROM metrics
		WHERE vmid = ?
		ORDER BY timestamp DESC
		LIMIT 1
	`

	var m Metric
	err := a.db.QueryRow(query, vmid).Scan(
		&m.VMID,
		&m.Timestamp,
		&m.CPUUsage,
		&m.MemUsage,
		&m.MemTotal,
		&m.DiskUsage,
		&m.DiskTotal,
		&m.NetIn,
		&m.NetOut,
		&m.Uptime,
		&m.Status,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("no metrics found for VMID %d", vmid)
		}
		return nil, err
	}

	return &m, nil
}

// GetMetricsSummary calculates aggregated metrics for a time range
func (a *Analytics) GetMetricsSummary(vmid int, start, end time.Time) (*MetricSummary, error) {
	query := `
		SELECT 
			vmid,
			MIN(timestamp) as start_time,
			MAX(timestamp) as end_time,
			AVG(cpu_usage) as avg_cpu,
			MAX(cpu_usage) as max_cpu,
			AVG(CAST(mem_usage AS REAL) / CAST(mem_total AS REAL) * 100) as avg_mem_usage,
			MAX(mem_usage) as max_mem_usage,
			AVG(CAST(disk_usage AS REAL) / CAST(disk_total AS REAL) * 100) as avg_disk_usage,
			SUM(net_in) as total_net_in,
			SUM(net_out) as total_net_out,
			COUNT(*) as data_points
		FROM metrics
		WHERE vmid = ? AND timestamp BETWEEN ? AND ?
		GROUP BY vmid
	`

	var summary MetricSummary
	err := a.db.QueryRow(query, vmid, start, end).Scan(
		&summary.VMID,
		&summary.StartTime,
		&summary.EndTime,
		&summary.AvgCPU,
		&summary.MaxCPU,
		&summary.AvgMemUsage,
		&summary.MaxMemUsage,
		&summary.AvgDiskUsage,
		&summary.TotalNetIn,
		&summary.TotalNetOut,
		&summary.DataPoints,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("no metrics found for VMID %d in time range", vmid)
		}
		return nil, err
	}

	return &summary, nil
}

// GetAllContainerMetrics retrieves latest metrics for all containers
func (a *Analytics) GetAllContainerMetrics() (map[int]*Metric, error) {
	query := `
		SELECT m1.vmid, m1.timestamp, m1.cpu_usage, m1.mem_usage, m1.mem_total, 
		       m1.disk_usage, m1.disk_total, m1.net_in, m1.net_out, m1.uptime, m1.status
		FROM metrics m1
		INNER JOIN (
			SELECT vmid, MAX(timestamp) as max_timestamp
			FROM metrics
			GROUP BY vmid
		) m2 ON m1.vmid = m2.vmid AND m1.timestamp = m2.max_timestamp
	`

	rows, err := a.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	metrics := make(map[int]*Metric)
	for rows.Next() {
		var m Metric
		err := rows.Scan(
			&m.VMID,
			&m.Timestamp,
			&m.CPUUsage,
			&m.MemUsage,
			&m.MemTotal,
			&m.DiskUsage,
			&m.DiskTotal,
			&m.NetIn,
			&m.NetOut,
			&m.Uptime,
			&m.Status,
		)
		if err != nil {
			log.Printf("Failed to scan metric: %v", err)
			continue
		}
		metrics[m.VMID] = &m
	}

	return metrics, rows.Err()
}

// CleanOldMetrics removes metrics older than the retention period (30 days)
func (a *Analytics) CleanOldMetrics(retentionDays int) error {
	cutoffTime := time.Now().AddDate(0, 0, -retentionDays)

	result, err := a.db.Exec("DELETE FROM metrics WHERE timestamp < ?", cutoffTime)
	if err != nil {
		return fmt.Errorf("failed to clean old metrics: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected > 0 {
		log.Printf("Cleaned %d old metric records (older than %d days)", rowsAffected, retentionDays)
	}

	return nil
}

// GetMetricsCount returns the total number of metrics stored
func (a *Analytics) GetMetricsCount() (int64, error) {
	var count int64
	err := a.db.QueryRow("SELECT COUNT(*) FROM metrics").Scan(&count)
	return count, err
}

// GetContainerMetricsCount returns the number of metrics for a specific container
func (a *Analytics) GetContainerMetricsCount(vmid int) (int64, error) {
	var count int64
	err := a.db.QueryRow("SELECT COUNT(*) FROM metrics WHERE vmid = ?", vmid).Scan(&count)
	return count, err
}

// ExportMetrics exports metrics for a container as JSON
func (a *Analytics) ExportMetrics(vmid int, start, end time.Time) ([]byte, error) {
	metrics, err := a.GetMetrics(vmid, start, end, 10000) // Limit to 10k records
	if err != nil {
		return nil, err
	}

	return json.MarshalIndent(metrics, "", "  ")
}
