// Package copilot – loader.go handles loading configuration from YAML files.
package copilot

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// LoadConfigFromFile reads and parses a YAML configuration file.
// Returns the parsed Config or an error.
func LoadConfigFromFile(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading config file: %w", err)
	}

	return ParseConfig(data)
}

// ParseConfig parses YAML bytes into a Config.
// Starts with defaults and overlays values from the YAML.
func ParseConfig(data []byte) (*Config, error) {
	// Start with defaults.
	cfg := DefaultConfig()

	// Parse the raw YAML to handle the top-level structure.
	var raw map[string]any
	if err := yaml.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("parsing config YAML: %w", err)
	}

	// Re-marshal and unmarshal into Config to get proper merging.
	// This handles the flat structure where keys like "access", "channels"
	// are at the root level.
	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("mapping config: %w", err)
	}

	return cfg, nil
}

// SaveConfigToFile writes a Config as YAML to the specified path.
func SaveConfigToFile(cfg *Config, path string) error {
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("marshaling config: %w", err)
	}

	if err := os.WriteFile(path, data, 0o644); err != nil {
		return fmt.Errorf("writing config file: %w", err)
	}

	return nil
}

// FindConfigFile searches for config files in standard locations.
// Returns the path of the first found, or empty string.
func FindConfigFile() string {
	candidates := []string{
		"config.yaml",
		"config.yml",
		"copilot.yaml",
		"copilot.yml",
		"configs/config.yaml",
		"configs/copilot.yaml",
	}

	for _, path := range candidates {
		if _, err := os.Stat(path); err == nil {
			return path
		}
	}

	return ""
}
