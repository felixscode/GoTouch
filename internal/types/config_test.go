package types

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadConfig_ValidConfig(t *testing.T) {
	// Create temporary config file
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	validConfig := `text:
  source: llm
  llm:
    model: "haiku"
    pregenerate_threshold: 20
    fallback_to_dummy: true
    timeout_seconds: 5
    max_retries: 1
ui:
  theme: "default"
stats:
  file_dir: "user_stats.json"
`

	err := os.WriteFile(configPath, []byte(validConfig), 0644)
	if err != nil {
		t.Fatalf("Failed to create test config file: %v", err)
	}

	config, err := LoadConfig(configPath)

	if err != nil {
		t.Errorf("LoadConfig() unexpected error: %v", err)
	}

	if config == nil {
		t.Fatalf("LoadConfig() returned nil config")
	}

	// Verify config values
	if config.Text.Source != "llm" {
		t.Errorf("config.Text.Source = %q, want %q", config.Text.Source, "llm")
	}

	if config.Text.LLM.Model != "haiku" {
		t.Errorf("config.Text.LLM.Model = %q, want %q", config.Text.LLM.Model, "haiku")
	}

	if config.Text.LLM.PregenerateThreshold != 20 {
		t.Errorf("config.Text.LLM.PregenerateThreshold = %d, want %d", config.Text.LLM.PregenerateThreshold, 20)
	}

	if !config.Text.LLM.FallbackToDummy {
		t.Errorf("config.Text.LLM.FallbackToDummy = %v, want %v", config.Text.LLM.FallbackToDummy, true)
	}

	if config.Text.LLM.TimeoutSeconds != 5 {
		t.Errorf("config.Text.LLM.TimeoutSeconds = %d, want %d", config.Text.LLM.TimeoutSeconds, 5)
	}

	if config.Text.LLM.MaxRetries != 1 {
		t.Errorf("config.Text.LLM.MaxRetries = %d, want %d", config.Text.LLM.MaxRetries, 1)
	}

	if config.Ui.Theme != "default" {
		t.Errorf("config.Ui.Theme = %q, want %q", config.Ui.Theme, "default")
	}

	if config.Stats.FileDir != "user_stats.json" {
		t.Errorf("config.Stats.FileDir = %q, want %q", config.Stats.FileDir, "user_stats.json")
	}
}

func TestLoadConfig_DummySource(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	dummyConfig := `text:
  source: dummy
ui:
  theme: "dark"
stats:
  file_dir: "stats.json"
`

	err := os.WriteFile(configPath, []byte(dummyConfig), 0644)
	if err != nil {
		t.Fatalf("Failed to create test config file: %v", err)
	}

	config, err := LoadConfig(configPath)

	if err != nil {
		t.Errorf("LoadConfig() unexpected error: %v", err)
	}

	if config == nil {
		t.Fatalf("LoadConfig() returned nil config")
	}

	if config.Text.Source != "dummy" {
		t.Errorf("config.Text.Source = %q, want %q", config.Text.Source, "dummy")
	}

	if config.Ui.Theme != "dark" {
		t.Errorf("config.Ui.Theme = %q, want %q", config.Ui.Theme, "dark")
	}

	if config.Stats.FileDir != "stats.json" {
		t.Errorf("config.Stats.FileDir = %q, want %q", config.Stats.FileDir, "stats.json")
	}
}

func TestLoadConfig_FileNotFound(t *testing.T) {
	config, err := LoadConfig("/nonexistent/path/config.yaml")

	if err == nil {
		t.Errorf("LoadConfig() expected error for nonexistent file, got nil")
	}

	if config != nil {
		t.Errorf("LoadConfig() expected nil config for nonexistent file, got %v", config)
	}
}

func TestLoadConfig_InvalidYAML(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "invalid.yaml")

	invalidYAML := `text:
  source: llm
  invalid yaml structure
    - broken
`

	err := os.WriteFile(configPath, []byte(invalidYAML), 0644)
	if err != nil {
		t.Fatalf("Failed to create test config file: %v", err)
	}

	config, err := LoadConfig(configPath)

	if err == nil {
		t.Errorf("LoadConfig() expected error for invalid YAML, got nil")
	}

	if config != nil {
		t.Errorf("LoadConfig() expected nil config for invalid YAML, got %v", config)
	}
}

func TestLoadConfig_EmptyFile(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "empty.yaml")

	err := os.WriteFile(configPath, []byte(""), 0644)
	if err != nil {
		t.Fatalf("Failed to create test config file: %v", err)
	}

	config, err := LoadConfig(configPath)

	// Empty YAML should parse successfully but return empty struct
	if err != nil {
		t.Errorf("LoadConfig() unexpected error for empty file: %v", err)
	}

	if config == nil {
		t.Errorf("LoadConfig() returned nil config for empty file")
	}
}

func TestLoadConfig_MinimalConfig(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "minimal.yaml")

	minimalConfig := `text:
  source: dummy
`

	err := os.WriteFile(configPath, []byte(minimalConfig), 0644)
	if err != nil {
		t.Fatalf("Failed to create test config file: %v", err)
	}

	config, err := LoadConfig(configPath)

	if err != nil {
		t.Errorf("LoadConfig() unexpected error: %v", err)
	}

	if config == nil {
		t.Fatalf("LoadConfig() returned nil config")
	}

	if config.Text.Source != "dummy" {
		t.Errorf("config.Text.Source = %q, want %q", config.Text.Source, "dummy")
	}

	// Other fields should have zero values
	if config.Ui.Theme != "" {
		t.Logf("Note: config.Ui.Theme = %q (default value)", config.Ui.Theme)
	}
}
