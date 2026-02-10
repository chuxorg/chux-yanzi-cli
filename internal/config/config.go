package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const DefaultBaseURL = "http://localhost:8080"

// Config holds CLI configuration values loaded from disk.
type Config struct {
	BaseURL string
}

// Load reads ~/.yanzi/config.yaml and returns defaults if missing.
func Load() (Config, error) {
	cfg := Config{BaseURL: DefaultBaseURL}
	path, err := ConfigPath()
	if err != nil {
		return cfg, err
	}
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return cfg, nil
		}
		return cfg, fmt.Errorf("read config: %w", err)
	}

	lines := strings.Split(string(data), "\n")
	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			continue
		}
		parts := strings.SplitN(trimmed, ":", 2)
		if len(parts) != 2 {
			return cfg, fmt.Errorf("invalid config line %d", i+1)
		}
		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])
		value = strings.Trim(value, "\"'")
		if key == "base_url" {
			if value == "" {
				return cfg, errors.New("base_url cannot be empty")
			}
			cfg.BaseURL = value
		}
	}

	return cfg, nil
}

// ConfigPath returns the full path to ~/.yanzi/config.yaml.
func ConfigPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("resolve home: %w", err)
	}
	return filepath.Join(home, ".yanzi", "config.yaml"), nil
}

// StateDir returns the ~/.yanzi directory path.
func StateDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("resolve home: %w", err)
	}
	return filepath.Join(home, ".yanzi"), nil
}
