package config

// Config represents the application configuration
type Config struct {
	Server  ServerConfig  `yaml:"server"`
	Proxmox ProxmoxConfig `yaml:"proxmox"`
}

// ServerConfig holds server-specific configuration
type ServerConfig struct {
	Port int    `yaml:"port"`
	Host string `yaml:"host"`
}

// ProxmoxConfig holds Proxmox API configuration
type ProxmoxConfig struct {
	Host        string `yaml:"host"`
	Node        string `yaml:"node"`
	TokenID     string `yaml:"token_id"`
	TokenSecret string `yaml:"token_secret"`
	Insecure    bool   `yaml:"insecure"`
}
