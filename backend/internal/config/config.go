package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// Load reads and parses the configuration file
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	// Override with environment variables if present
	if port := os.Getenv("PORT"); port != "" {
		if _, err := fmt.Sscanf(port, "%d", &cfg.Server.Port); err != nil {
			return nil, fmt.Errorf("invalid PORT environment variable: %w", err)
		}
	}
	if host := os.Getenv("HOST"); host != "" {
		cfg.Server.Host = host
	}
	if proxmoxHost := os.Getenv("PROXMOX_HOST"); proxmoxHost != "" {
		cfg.Proxmox.Host = proxmoxHost
	}
	if proxmoxNode := os.Getenv("PROXMOX_NODE"); proxmoxNode != "" {
		cfg.Proxmox.Node = proxmoxNode
	}
	if tokenID := os.Getenv("PROXMOX_TOKEN_ID"); tokenID != "" {
		cfg.Proxmox.TokenID = tokenID
	}
	if tokenSecret := os.Getenv("PROXMOX_TOKEN_SECRET"); tokenSecret != "" {
		cfg.Proxmox.TokenSecret = tokenSecret
	}

	return &cfg, nil
}

// Validate checks if the configuration is valid
func (c *Config) Validate() error {
	if c.Server.Port <= 0 || c.Server.Port > 65535 {
		return fmt.Errorf("invalid server port: %d", c.Server.Port)
	}

	if c.Proxmox.Host == "" {
		return fmt.Errorf("proxmox host is required")
	}

	if c.Proxmox.Node == "" {
		return fmt.Errorf("proxmox node is required")
	}

	if c.Proxmox.TokenID == "" {
		return fmt.Errorf("proxmox token_id is required")
	}

	if c.Proxmox.TokenSecret == "" {
		return fmt.Errorf("proxmox token_secret is required")
	}

	return nil
}
