package config

import (
	"fmt"
	"go-touch/internal/types"
	"os"

	"gopkg.in/yaml.v3"
)

// DefaultConfig returns a default configuration
func DefaultConfig() types.Config {
	statsPath, _ := GetDefaultStatsPath()
	if statsPath == "" {
		statsPath = "user_stats.json"
	}

	return types.Config{
		Text: types.TextConfig{
			Source: "dummy",
			LLM: types.LLMConfig{
				Provider:             "anthropic",
				Model:                "claude-3-5-haiku-latest",
				APIBase:              "",
				PregenerateThreshold: 20,
				FallbackToDummy:      true,
				TimeoutSeconds:       5,
				MaxRetries:           1,
			},
		},
		Ui: types.UiConfig{
			Theme: "default",
		},
		Stats: types.StatsConfig{
			FileDir: statsPath,
		},
	}
}

// CreateDefaultConfigFile creates a default config.yaml file in the config directory
func CreateDefaultConfigFile() (string, error) {
	configDir, err := GetConfigDir()
	if err != nil {
		return "", fmt.Errorf("failed to get config directory: %w", err)
	}

	// Ensure config directory exists
	if err := EnsureDir(configDir); err != nil {
		return "", err
	}

	configPath, err := GetDefaultConfigPath()
	if err != nil {
		return "", fmt.Errorf("failed to determine default config path: %w", err)
	}

	// Check if config already exists
	if _, err := os.Stat(configPath); err == nil {
		return configPath, nil // Already exists
	}

	// Create default config
	defaultCfg := DefaultConfig()

	// Marshal to YAML
	data, err := yaml.Marshal(defaultCfg)
	if err != nil {
		return "", fmt.Errorf("failed to marshal default config: %w", err)
	}

	// Write to file
	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return "", fmt.Errorf("failed to write config file: %w", err)
	}

	return configPath, nil
}

// LoadOrCreateConfig loads config from the given path, or creates a default one
func LoadOrCreateConfig(explicitPath string) (*types.Config, string, error) {
	// Find existing config file
	configPath, err := FindConfigFile(explicitPath)
	if err != nil {
		return nil, "", err
	}

	// If config found, load it
	if configPath != "" {
		cfg, err := types.LoadConfig(configPath)
		if err != nil {
			return nil, "", fmt.Errorf("failed to load config from %s: %w", configPath, err)
		}
		return cfg, configPath, nil
	}

	// No config found - create default
	fmt.Println("No config file found. Creating default configuration...")
	configPath, err = CreateDefaultConfigFile()
	if err != nil {
		// If we can't create in config dir, use default in-memory config
		fmt.Printf("Warning: Failed to create config file: %v\n", err)
		fmt.Println("Using default configuration")
		cfg := DefaultConfig()
		return &cfg, "", nil
	}

	fmt.Printf("Created default config at: %s\n", configPath)

	// Load the newly created config
	cfg, err := types.LoadConfig(configPath)
	if err != nil {
		return nil, "", fmt.Errorf("failed to load newly created config: %w", err)
	}

	return cfg, configPath, nil
}
