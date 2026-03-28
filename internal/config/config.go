package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

const (
	defaultAPIURL = "https://api.formfor.ai"
	configDir     = ".formfor"
	configFile    = "config.json"
)

// Config holds persisted CLI configuration.
type Config struct {
	APIKey string `json:"api_key,omitempty"`
	APIURL string `json:"api_url,omitempty"`
}

// configPath returns the full path to the config file.
func configPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("determining home directory: %w", err)
	}
	return filepath.Join(home, configDir, configFile), nil
}

// Load reads the config file from disk. Returns a zero-value Config if the
// file does not exist.
func Load() (*Config, error) {
	p, err := configPath()
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(p)
	if err != nil {
		if os.IsNotExist(err) {
			return &Config{}, nil
		}
		return nil, fmt.Errorf("reading config: %w", err)
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parsing config: %w", err)
	}
	return &cfg, nil
}

// Save writes the config to disk, creating the directory if needed.
func Save(cfg *Config) error {
	p, err := configPath()
	if err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(p), 0700); err != nil {
		return fmt.Errorf("creating config directory: %w", err)
	}

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("marshalling config: %w", err)
	}

	if err := os.WriteFile(p, data, 0600); err != nil {
		return fmt.Errorf("writing config: %w", err)
	}
	return nil
}

// GetAPIKey returns the API key, checking the FF_API_KEY environment variable
// first, then the config file.
func GetAPIKey() string {
	if v := os.Getenv("FF_API_KEY"); v != "" {
		return v
	}
	cfg, err := Load()
	if err != nil {
		return ""
	}
	return cfg.APIKey
}

// GetAPIURL returns the API URL, checking FF_API_URL env first, then config
// file, then the default.
func GetAPIURL() string {
	if v := os.Getenv("FF_API_URL"); v != "" {
		return v
	}
	cfg, err := Load()
	if err != nil {
		return defaultAPIURL
	}
	if cfg.APIURL != "" {
		return cfg.APIURL
	}
	return defaultAPIURL
}

// Set sets a configuration key to the given value.
func Set(key, value string) error {
	cfg, err := Load()
	if err != nil {
		return err
	}

	switch key {
	case "api-key":
		cfg.APIKey = value
	case "api-url":
		cfg.APIURL = value
	default:
		return fmt.Errorf("unknown config key: %s (valid keys: api-key, api-url)", key)
	}

	return Save(cfg)
}

// Get retrieves a configuration value by key.
func Get(key string) (string, error) {
	switch key {
	case "api-key":
		v := GetAPIKey()
		if v == "" {
			return "", fmt.Errorf("api-key is not set (use `ff config set api-key <key>` or set FF_API_KEY)")
		}
		return v, nil
	case "api-url":
		return GetAPIURL(), nil
	default:
		return "", fmt.Errorf("unknown config key: %s (valid keys: api-key, api-url)", key)
	}
}
