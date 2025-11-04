package config

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
)

// GetConfigDir returns the platform-specific config directory for GoTouch
func GetConfigDir() (string, error) {
	var configDir string

	switch runtime.GOOS {
	case "windows":
		// Windows: %APPDATA%\gotouch
		appData := os.Getenv("APPDATA")
		if appData == "" {
			return "", fmt.Errorf("APPDATA environment variable not set")
		}
		configDir = filepath.Join(appData, "gotouch")

	case "darwin", "linux":
		// macOS/Linux: ~/.config/gotouch (XDG Base Directory)
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("failed to get home directory: %w", err)
		}

		// Check XDG_CONFIG_HOME first
		xdgConfigHome := os.Getenv("XDG_CONFIG_HOME")
		if xdgConfigHome != "" {
			configDir = filepath.Join(xdgConfigHome, "gotouch")
		} else {
			configDir = filepath.Join(homeDir, ".config", "gotouch")
		}

	default:
		return "", fmt.Errorf("unsupported platform: %s", runtime.GOOS)
	}

	return configDir, nil
}

// GetDataDir returns the platform-specific data directory for GoTouch
func GetDataDir() (string, error) {
	var dataDir string

	switch runtime.GOOS {
	case "windows":
		// Windows: %LOCALAPPDATA%\gotouch
		localAppData := os.Getenv("LOCALAPPDATA")
		if localAppData == "" {
			// Fallback to APPDATA
			appData := os.Getenv("APPDATA")
			if appData == "" {
				return "", fmt.Errorf("LOCALAPPDATA and APPDATA environment variables not set")
			}
			dataDir = filepath.Join(appData, "gotouch")
		} else {
			dataDir = filepath.Join(localAppData, "gotouch")
		}

	case "darwin", "linux":
		// macOS/Linux: ~/.local/share/gotouch (XDG Base Directory)
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("failed to get home directory: %w", err)
		}

		// Check XDG_DATA_HOME first
		xdgDataHome := os.Getenv("XDG_DATA_HOME")
		if xdgDataHome != "" {
			dataDir = filepath.Join(xdgDataHome, "gotouch")
		} else {
			dataDir = filepath.Join(homeDir, ".local", "share", "gotouch")
		}

	default:
		return "", fmt.Errorf("unsupported platform: %s", runtime.GOOS)
	}

	return dataDir, nil
}

// EnsureDir creates the directory if it doesn't exist
func EnsureDir(dir string) error {
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
	}
	return nil
}

// FindConfigFile searches for config file in the following order:
// 1. Explicit path (if provided)
// 2. Config directory (~/.config/gotouch/config.yaml)
// 3. Current directory (./config.yaml)
// Returns the path to the config file, or empty string if not found
func FindConfigFile(explicitPath string) (string, error) {
	// 1. Check explicit path first
	if explicitPath != "" {
		if _, err := os.Stat(explicitPath); err == nil {
			return explicitPath, nil
		}
		return "", fmt.Errorf("config file not found at specified path: %s", explicitPath)
	}

	// 2. Check config directory
	configDir, err := GetConfigDir()
	if err == nil {
		configPath := filepath.Join(configDir, "config.yaml")
		if _, err := os.Stat(configPath); err == nil {
			return configPath, nil
		}
	}

	// 3. Check current directory
	currentDirPath := "config.yaml"
	if _, err := os.Stat(currentDirPath); err == nil {
		return currentDirPath, nil
	}

	// Config file not found anywhere
	return "", nil
}

// GetDefaultConfigPath returns the default config file path in the config directory
func GetDefaultConfigPath() (string, error) {
	configDir, err := GetConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(configDir, "config.yaml"), nil
}

// GetDefaultStatsPath returns the default stats file path in the data directory
func GetDefaultStatsPath() (string, error) {
	dataDir, err := GetDataDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dataDir, "user_stats.json"), nil
}

// FindAPIKeyFile searches for api-key file in the following order:
// 1. Config directory (~/.config/gotouch/api-key)
// 2. Current directory (./api-key)
// Returns the path to the api-key file, or empty string if not found
func FindAPIKeyFile() string {
	// 1. Check config directory
	configDir, err := GetConfigDir()
	if err == nil {
		apiKeyPath := filepath.Join(configDir, "api-key")
		if _, err := os.Stat(apiKeyPath); err == nil {
			return apiKeyPath
		}
	}

	// 2. Check current directory
	currentDirPath := "api-key"
	if _, err := os.Stat(currentDirPath); err == nil {
		return currentDirPath
	}

	return ""
}
