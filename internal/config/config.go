package config

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

// Config holds all application configuration
type Config struct {
	NetboxURL      string  `json:"netbox_url"`
	NAMURL         string  `json:"nam_url"`
	ESMURL         string  `json:"esm_url"`
	ESMUser        string  `json:"esm_user"`
	NetboxAPIToken string  `json:"-"` // Loaded from file, not JSON
	NAMAPIToken    string  `json:"-"` // Loaded from file, not JSON
	ESMPassword    string  `json:"-"` // Loaded from file, not JSON
	ESMTenantID    int     `json:"esm_tenant_id"`
	ESMOfferingID  string  `json:"esm_offering_id"`
	ESMRequesterID string  `json:"esm_requester_id"`
	ESMServiceID   string  `json:"esm_service_id"`
	ESMTeamID      string  `json:"esm_team_id"`
	SlackWebhook   string  `json:"slack_webhook_url"`
	Checks         []Check `json:"checks"`
}

// Check represents a DC check configuration
type Check struct {
	NetboxSiteID int    `json:"netbox_site_id"`
	Infra        string `json:"infra"`
	DCName       string `json:"dc_name"`
}

// LoadConfig loads configuration from files
// Expects:
// - config/config.json for URLs and check definitions
// - secrets/netbox-token for Netbox API token
// - secrets/nam-token for NAM API token
// - secrets/esm-password for ESM password
func LoadConfig() (*Config, error) {
	// Read main config file
	configData, err := os.ReadFile("config/config.json")
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var cfg Config
	if err := json.Unmarshal(configData, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config JSON: %w", err)
	}

	// Read Netbox token
	token, err := readTokenFile("secrets/netbox.secret")
	if err != nil {
		return nil, fmt.Errorf("failed to read Netbox token: %w", err)
	}
	cfg.NetboxAPIToken = token

	// Read NAM token
	token, err = readTokenFile("secrets/nam.secret")
	if err != nil {
		return nil, fmt.Errorf("failed to read NAM token: %w", err)
	}
	cfg.NAMAPIToken = token

	// Read ESM password
	password, err := readTokenFile("secrets/esm.secret")
	if err != nil {
		return nil, fmt.Errorf("failed to read ESM Password: %w", err)
	}
	cfg.ESMPassword = password

	return &cfg, nil
}

// readTokenFile reads a token from a file and trims whitespace
func readTokenFile(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(data)), nil
}
