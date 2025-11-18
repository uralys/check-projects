package config

import (
	"os"
	"path/filepath"
	"strings"
)

// Config represents the application configuration
type Config struct {
	Categories []Category `yaml:"categories"`
	Display    Display    `yaml:"display"`

	// Internal: path where config was loaded from (not serialized)
	ConfigPath string `yaml:"-"`
	// Internal: true if config was filtered (don't save to avoid losing data)
	IsFiltered bool `yaml:"-"`
}

// Category represents a project category
// Either Root (auto-scan) or Projects (explicit list) must be specified
type Category struct {
	Name     string   `yaml:"name"`
	Root     string   `yaml:"root,omitempty"`     // Auto-scan: recursively find all git repos
	Projects []string `yaml:"projects,omitempty"` // Explicit: list of full paths to repos
	Ignore   []string `yaml:"ignore,omitempty"`   // Projects to ignore in this category
}

// Display represents display options
type Display struct {
	HideClean   bool `yaml:"hide_clean"`
	HideIgnored bool `yaml:"hide_ignored"`
}

// ExpandPath expands ~ to home directory
func ExpandPath(path string) string {
	if strings.HasPrefix(path, "~") {
		home, err := os.UserHomeDir()
		if err != nil {
			return path
		}
		return filepath.Join(home, path[1:])
	}
	return path
}

// GetRootPath returns the expanded root path
func (c *Category) GetRootPath() string {
	return ExpandPath(c.Root)
}

// DefaultConfig returns a default configuration
func DefaultConfig() *Config {
	return &Config{
		Categories: []Category{},
		Display: Display{
			HideClean:   true,
			HideIgnored: true,
		},
	}
}
