package config

import (
	"os"
	"path/filepath"
	"testing"

	"gopkg.in/yaml.v3"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	// Verify default values
	if cfg.Text.Source != "dummy" {
		t.Errorf("DefaultConfig() Source = %v, want 'dummy'", cfg.Text.Source)
	}

	if cfg.Text.LLM.Provider != "anthropic" {
		t.Errorf("DefaultConfig() Provider = %v, want 'anthropic'", cfg.Text.LLM.Provider)
	}

	if cfg.Text.LLM.Model != "claude-3-5-haiku-latest" {
		t.Errorf("DefaultConfig() Model = %v, want 'claude-3-5-haiku-latest'", cfg.Text.LLM.Model)
	}

	if cfg.Text.LLM.PregenerateThreshold != 20 {
		t.Errorf("DefaultConfig() PregenerateThreshold = %v, want 20", cfg.Text.LLM.PregenerateThreshold)
	}

	if !cfg.Text.LLM.FallbackToDummy {
		t.Errorf("DefaultConfig() FallbackToDummy = %v, want true", cfg.Text.LLM.FallbackToDummy)
	}

	if cfg.Text.LLM.TimeoutSeconds != 5 {
		t.Errorf("DefaultConfig() TimeoutSeconds = %v, want 5", cfg.Text.LLM.TimeoutSeconds)
	}

	if cfg.Text.LLM.MaxRetries != 1 {
		t.Errorf("DefaultConfig() MaxRetries = %v, want 1", cfg.Text.LLM.MaxRetries)
	}

	if cfg.Ui.Theme != "default" {
		t.Errorf("DefaultConfig() Theme = %v, want 'default'", cfg.Ui.Theme)
	}

	if cfg.Stats.FileDir == "" {
		t.Errorf("DefaultConfig() FileDir should not be empty")
	}
}

func TestCreateDefaultConfigFile(t *testing.T) {
	// Save original environment
	oldHome := os.Getenv("HOME")
	oldXDG := os.Getenv("XDG_CONFIG_HOME")
	defer func() {
		os.Setenv("HOME", oldHome)
		if oldXDG != "" {
			os.Setenv("XDG_CONFIG_HOME", oldXDG)
		} else {
			os.Unsetenv("XDG_CONFIG_HOME")
		}
	}()

	// Create a temporary directory for testing
	tmpDir := t.TempDir()

	// Set XDG_CONFIG_HOME to temp directory
	os.Setenv("XDG_CONFIG_HOME", tmpDir)
	os.Setenv("HOME", tmpDir)

	// Test creating config file
	configPath, err := CreateDefaultConfigFile()
	if err != nil {
		t.Fatalf("CreateDefaultConfigFile() error = %v", err)
	}

	if configPath == "" {
		t.Fatal("CreateDefaultConfigFile() returned empty path")
	}

	// Verify file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Errorf("Config file was not created at %s", configPath)
	}

	// Verify file contents are valid YAML
	data, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("Failed to read created config file: %v", err)
	}

	var cfg map[string]interface{}
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		t.Errorf("Created config file is not valid YAML: %v", err)
	}

	// Test calling again should return existing file
	configPath2, err := CreateDefaultConfigFile()
	if err != nil {
		t.Fatalf("CreateDefaultConfigFile() second call error = %v", err)
	}

	if configPath != configPath2 {
		t.Errorf("CreateDefaultConfigFile() returned different paths: %v vs %v", configPath, configPath2)
	}
}

func TestLoadOrCreateConfig(t *testing.T) {
	// Save original environment
	oldHome := os.Getenv("HOME")
	oldXDG := os.Getenv("XDG_CONFIG_HOME")
	defer func() {
		os.Setenv("HOME", oldHome)
		if oldXDG != "" {
			os.Setenv("XDG_CONFIG_HOME", oldXDG)
		} else {
			os.Unsetenv("XDG_CONFIG_HOME")
		}
	}()

	tmpDir := t.TempDir()
	os.Setenv("XDG_CONFIG_HOME", tmpDir)
	os.Setenv("HOME", tmpDir)

	t.Run("load existing config", func(t *testing.T) {
		// Create a test config file
		testConfigDir := filepath.Join(tmpDir, "test-existing")
		if err := os.MkdirAll(testConfigDir, 0755); err != nil {
			t.Fatalf("Failed to create test directory: %v", err)
		}

		testConfigPath := filepath.Join(testConfigDir, "test-config.yaml")
		testConfig := DefaultConfig()
		testConfig.Text.Source = "test-source"

		data, err := yaml.Marshal(testConfig)
		if err != nil {
			t.Fatalf("Failed to marshal test config: %v", err)
		}

		if err := os.WriteFile(testConfigPath, data, 0644); err != nil {
			t.Fatalf("Failed to write test config: %v", err)
		}

		// Load the config
		cfg, path, err := LoadOrCreateConfig(testConfigPath)
		if err != nil {
			t.Fatalf("LoadOrCreateConfig() error = %v", err)
		}

		if cfg == nil {
			t.Fatal("LoadOrCreateConfig() returned nil config")
		}

		if path != testConfigPath {
			t.Errorf("LoadOrCreateConfig() path = %v, want %v", path, testConfigPath)
		}

		if cfg.Text.Source != "test-source" {
			t.Errorf("LoadOrCreateConfig() Source = %v, want 'test-source'", cfg.Text.Source)
		}
	})

	t.Run("create new config when none exists", func(t *testing.T) {
		// Use a fresh temp directory
		newTmpDir := t.TempDir()
		os.Setenv("XDG_CONFIG_HOME", newTmpDir)

		cfg, path, err := LoadOrCreateConfig("")
		if err != nil {
			t.Fatalf("LoadOrCreateConfig() error = %v", err)
		}

		if cfg == nil {
			t.Fatal("LoadOrCreateConfig() returned nil config")
		}

		// Should have created a config file or returned in-memory config
		if path != "" {
			// If path is returned, file should exist
			if _, err := os.Stat(path); os.IsNotExist(err) {
				t.Errorf("Config file should exist at %s", path)
			}
		}

		// Config should have default values
		if cfg.Text.Source != "dummy" {
			t.Errorf("LoadOrCreateConfig() Source = %v, want 'dummy'", cfg.Text.Source)
		}
	})

	t.Run("error on invalid explicit path", func(t *testing.T) {
		_, _, err := LoadOrCreateConfig("/nonexistent/path/config.yaml")
		if err == nil {
			t.Error("LoadOrCreateConfig() should return error for non-existent explicit path")
		}
	})
}
