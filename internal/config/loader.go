package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// LoadConfig loads configuration from file
// Priority: 1. Provided path, 2. ./check-projects.yml, 3. ~/check-projects.yml
func LoadConfig(configPath string) (*Config, error) {
	var paths []string

	if configPath != "" {
		paths = append(paths, configPath)
	}

	// Local config
	if localPath := "check-projects.yml"; fileExists(localPath) {
		paths = append(paths, localPath)
	}

	// Global config
	home, err := os.UserHomeDir()
	if err == nil {
		globalPath := filepath.Join(home, "check-projects.yml")
		if fileExists(globalPath) {
			paths = append(paths, globalPath)
		}
	}

	if len(paths) == 0 {
		return nil, fmt.Errorf("no configuration file found (searched: ./check-projects.yml, ~/check-projects.yml)")
	}

	// Load the first available config
	cfg, err := loadFromFile(paths[0])
	if err != nil {
		return nil, err
	}
	cfg.ConfigPath = paths[0]
	return cfg, nil
}

func loadFromFile(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file %s: %w", path, err)
	}

	config := DefaultConfig()
	if err := yaml.Unmarshal(data, config); err != nil {
		return nil, fmt.Errorf("failed to parse config file %s: %w", path, err)
	}

	return config, nil
}

// SaveConfig saves the configuration back to file
func SaveConfig(cfg *Config) error {
	if cfg.ConfigPath == "" {
		return fmt.Errorf("config path is not set")
	}

	if cfg.IsFiltered {
		return fmt.Errorf("cannot save filtered config (use without --category to save)")
	}

	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(cfg.ConfigPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file %s: %w", cfg.ConfigPath, err)
	}

	return nil
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
