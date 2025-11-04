package main

import (
	"go-touch/internal/types"
	"os"
	"path/filepath"
	"testing"
)

func TestGetText_DummySource(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	configContent := `text:
  source: dummy
ui:
  theme: "default"
stats:
  file_dir: "user_stats.json"
`

	err := os.WriteFile(configPath, []byte(configContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test config: %v", err)
	}

	config, err := types.LoadConfig(configPath)
	if err != nil {
		t.Fatalf("LoadConfig() error: %v", err)
	}

	text, textSource, err := getText(*config)

	if err != nil {
		t.Errorf("getText() unexpected error: %v", err)
	}

	if text == "" {
		t.Errorf("getText() returned empty text")
	}

	if textSource == nil {
		t.Errorf("getText() returned nil textSource")
	}

	// Verify text is from dummy source
	expectedText := "The quick brown fox jumps over the lazy dog near the old wooden bridge. "
	if text != expectedText {
		t.Errorf("getText() text = %q, want %q", text, expectedText)
	}
}

func TestGetText_LLMSource_WithAPIKey(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping LLM integration test in short mode")
	}

	// Setup API key
	originalKey := os.Getenv("ANTHROPIC_API_KEY")
	defer func() {
		if originalKey != "" {
			os.Setenv("ANTHROPIC_API_KEY", originalKey)
		} else {
			os.Unsetenv("ANTHROPIC_API_KEY")
		}
	}()

	// Try to get API key from file or env
	if originalKey == "" {
		keyData, err := os.ReadFile("api-key")
		if err == nil && len(keyData) > 0 {
			os.Setenv("ANTHROPIC_API_KEY", string(keyData))
		}
	}

	if os.Getenv("ANTHROPIC_API_KEY") == "" {
		t.Skip("Skipping LLM test: ANTHROPIC_API_KEY not set")
	}

	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	configContent := `text:
  source: llm
  llm:
    model: "haiku"
    timeout_seconds: 30
    fallback_to_dummy: false
ui:
  theme: "default"
stats:
  file_dir: "user_stats.json"
`

	err := os.WriteFile(configPath, []byte(configContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test config: %v", err)
	}

	config, err := types.LoadConfig(configPath)
	if err != nil {
		t.Fatalf("LoadConfig() error: %v", err)
	}

	text, textSource, err := getText(*config)

	if err != nil {
		t.Errorf("getText() unexpected error: %v", err)
	}

	if text == "" {
		t.Errorf("getText() returned empty text")
	}

	if textSource == nil {
		t.Errorf("getText() returned nil textSource")
	}

	t.Logf("Generated text: %s", text)
}

func TestGetText_LLMSource_NoAPIKey_WithFallback(t *testing.T) {
	// Remove API key
	originalKey := os.Getenv("ANTHROPIC_API_KEY")
	defer func() {
		if originalKey != "" {
			os.Setenv("ANTHROPIC_API_KEY", originalKey)
		}
	}()
	os.Unsetenv("ANTHROPIC_API_KEY")

	// Remove api-key file if exists
	os.Remove("api-key")

	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	configContent := `text:
  source: llm
  llm:
    model: "haiku"
    timeout_seconds: 5
    fallback_to_dummy: true
ui:
  theme: "default"
stats:
  file_dir: "user_stats.json"
`

	err := os.WriteFile(configPath, []byte(configContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test config: %v", err)
	}

	config, err := types.LoadConfig(configPath)
	if err != nil {
		t.Fatalf("LoadConfig() error: %v", err)
	}

	text, textSource, err := getText(*config)

	// Should not error because fallback_to_dummy is true
	if err != nil {
		t.Errorf("getText() unexpected error: %v", err)
	}

	if text == "" {
		t.Errorf("getText() returned empty text")
	}

	if textSource == nil {
		t.Errorf("getText() returned nil textSource")
	}

	// Should have fallen back to dummy text
	expectedText := "The quick brown fox jumps over the lazy dog near the old wooden bridge. "
	if text != expectedText {
		t.Logf("getText() text = %q, expected dummy fallback but got different text", text)
	}
}

func TestGetText_LLMSource_NoAPIKey_NoFallback(t *testing.T) {
	// Remove API key
	originalKey := os.Getenv("ANTHROPIC_API_KEY")
	defer func() {
		if originalKey != "" {
			os.Setenv("ANTHROPIC_API_KEY", originalKey)
		}
	}()
	os.Unsetenv("ANTHROPIC_API_KEY")

	// Remove api-key file if exists
	os.Remove("api-key")

	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	configContent := `text:
  source: llm
  llm:
    model: "haiku"
    timeout_seconds: 5
    fallback_to_dummy: false
ui:
  theme: "default"
stats:
  file_dir: "user_stats.json"
`

	err := os.WriteFile(configPath, []byte(configContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test config: %v", err)
	}

	config, err := types.LoadConfig(configPath)
	if err != nil {
		t.Fatalf("LoadConfig() error: %v", err)
	}

	_, _, err = getText(*config)

	// Should error because API key is not set and no fallback
	if err == nil {
		t.Errorf("getText() expected error for missing API key without fallback, got nil")
	}
}

func TestGetText_InvalidSource(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	configContent := `text:
  source: invalid_source
ui:
  theme: "default"
stats:
  file_dir: "user_stats.json"
`

	err := os.WriteFile(configPath, []byte(configContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test config: %v", err)
	}

	config, err := types.LoadConfig(configPath)
	if err != nil {
		t.Fatalf("LoadConfig() error: %v", err)
	}

	_, _, err = getText(*config)

	if err == nil {
		t.Errorf("getText() expected error for invalid source, got nil")
	}
}
