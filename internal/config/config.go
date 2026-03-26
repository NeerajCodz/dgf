package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"github.com/NeerajCodz/dgf/pkg/types"
)

const (
	configDirName  = "dgf"
	configFileName = "config.json"
)

// GetConfigDir returns the platform-specific config directory
func GetConfigDir() (string, error) {
	var baseDir string

	switch runtime.GOOS {
	case "windows":
		baseDir = os.Getenv("APPDATA")
		if baseDir == "" {
			baseDir = os.Getenv("LOCALAPPDATA")
		}
	case "darwin":
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		baseDir = filepath.Join(home, "Library", "Application Support")
	default: // Linux and others
		baseDir = os.Getenv("XDG_CONFIG_HOME")
		if baseDir == "" {
			home, err := os.UserHomeDir()
			if err != nil {
				return "", err
			}
			baseDir = filepath.Join(home, ".config")
		}
	}

	if baseDir == "" {
		return "", fmt.Errorf("could not determine config directory")
	}

	return filepath.Join(baseDir, configDirName), nil
}

// GetConfigPath returns the full path to the config file
func GetConfigPath() (string, error) {
	dir, err := GetConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, configFileName), nil
}

// Load loads the configuration from disk
func Load() (types.Config, error) {
	cfg := types.DefaultConfig()

	path, err := GetConfigPath()
	if err != nil {
		return cfg, err
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return cfg, nil // Return defaults if file doesn't exist
		}
		return cfg, fmt.Errorf("failed to read config: %v", err)
	}

	if err := json.Unmarshal(data, &cfg); err != nil {
		return cfg, fmt.Errorf("failed to parse config: %v", err)
	}

	return cfg, nil
}

// Save saves the configuration to disk
func Save(cfg types.Config) error {
	dir, err := GetConfigDir()
	if err != nil {
		return err
	}

	// Create config directory if it doesn't exist
	if err := os.MkdirAll(dir, 0700); err != nil {
		return fmt.Errorf("failed to create config directory: %v", err)
	}

	path := filepath.Join(dir, configFileName)
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %v", err)
	}

	// Write with secure permissions
	if err := os.WriteFile(path, data, 0600); err != nil {
		return fmt.Errorf("failed to write config: %v", err)
	}

	return nil
}

// Set sets a configuration value by key
func Set(key, value string) error {
	cfg, err := Load()
	if err != nil {
		return err
	}

	switch key {
	case "token", "github_token":
		cfg.GithubToken = value
	case "download_path", "path":
		cfg.DownloadPath = value
	case "ascii_mode", "ascii":
		cfg.ASCIIMode = value == "true" || value == "1"
	case "workers":
		var workers int
		fmt.Sscanf(value, "%d", &workers)
		if workers > 0 {
			cfg.Workers = workers
		}
	default:
		return fmt.Errorf("unknown config key: %s", key)
	}

	return Save(cfg)
}

// Get gets a configuration value by key
func Get(key string) (string, error) {
	cfg, err := Load()
	if err != nil {
		return "", err
	}

	switch key {
	case "token", "github_token":
		if cfg.GithubToken != "" {
			return "***" + cfg.GithubToken[max(0, len(cfg.GithubToken)-4):], nil
		}
		return "", nil
	case "download_path", "path":
		return cfg.DownloadPath, nil
	case "ascii_mode", "ascii":
		if cfg.ASCIIMode {
			return "true", nil
		}
		return "false", nil
	case "workers":
		return fmt.Sprintf("%d", cfg.Workers), nil
	default:
		return "", fmt.Errorf("unknown config key: %s", key)
	}
}

// Show returns all configuration as a formatted string
func Show() (string, error) {
	cfg, err := Load()
	if err != nil {
		return "", err
	}

	token := "(not set)"
	if cfg.GithubToken != "" {
		token = "***" + cfg.GithubToken[max(0, len(cfg.GithubToken)-4):]
	}

	return fmt.Sprintf(`dgf configuration:
  github_token:  %s
  download_path: %s
  ascii_mode:    %v
  workers:       %d
`, token, cfg.DownloadPath, cfg.ASCIIMode, cfg.Workers), nil
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
