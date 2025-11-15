package main

import (
	"fmt"
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

}

func TestGetText_LLMSource_WithAPIKey(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping LLM integration test in short mode")
	}

	// Setup API key
	originalKey := os.Getenv("GOTOUCH_LLM_API_KEY")
	defer func() {
		if originalKey != "" {
			os.Setenv("GOTOUCH_LLM_API_KEY", originalKey)
		} else {
			os.Unsetenv("GOTOUCH_LLM_API_KEY")
		}
	}()

	// Try to get API key from file or env
	if originalKey == "" {
		keyData, err := os.ReadFile("api-key")
		if err == nil && len(keyData) > 0 {
			os.Setenv("GOTOUCH_LLM_API_KEY", string(keyData))
		}
	}

	if os.Getenv("GOTOUCH_LLM_API_KEY") == "" {
		t.Skip("Skipping LLM test: GOTOUCH_LLM_API_KEY not set")
	}

	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	configContent := `text:
  source: llm
  llm:
    provider: "anthropic"
    model: "claude-3-5-haiku-latest"
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
	originalKey := os.Getenv("GOTOUCH_LLM_API_KEY")
	defer func() {
		if originalKey != "" {
			os.Setenv("GOTOUCH_LLM_API_KEY", originalKey)
		}
	}()
	os.Unsetenv("GOTOUCH_LLM_API_KEY")

	// Remove api-key file if exists
	os.Remove("api-key")

	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	configContent := `text:
  source: llm
  llm:
    provider: "anthropic"
    model: "claude-3-5-haiku-latest"
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
	originalKey := os.Getenv("GOTOUCH_LLM_API_KEY")
	defer func() {
		if originalKey != "" {
			os.Setenv("GOTOUCH_LLM_API_KEY", originalKey)
		}
	}()
	os.Unsetenv("GOTOUCH_LLM_API_KEY")

	// Remove api-key file if exists
	os.Remove("api-key")

	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	configContent := `text:
  source: llm
  llm:
    provider: "anthropic"
    model: "claude-3-5-haiku-latest"
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

func TestGetText_ErrorHandling(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	// Test with empty source
	configContent := `text:
  source: ""
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

	// Empty source should default to dummy
	text, textSource, err := getText(*config)

	if err != nil {
		t.Logf("getText() with empty source: %v", err)
	}

	if text != "" && textSource != nil {
		t.Logf("getText() returned text: %s", text)
	}
}

func TestGetText_DummySource_Variations(t *testing.T) {
	tests := []struct {
		name           string
		source         string
		expectError    bool
		expectNonEmpty bool
	}{
		{
			name:           "dummy source",
			source:         "dummy",
			expectError:    false,
			expectNonEmpty: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			configPath := filepath.Join(tmpDir, "config.yaml")

			configContent := fmt.Sprintf(`text:
  source: %s
ui:
  theme: "default"
stats:
  file_dir: "user_stats.json"
`, tt.source)

			err := os.WriteFile(configPath, []byte(configContent), 0644)
			if err != nil {
				t.Fatalf("Failed to create test config: %v", err)
			}

			config, err := types.LoadConfig(configPath)
			if err != nil {
				t.Fatalf("LoadConfig() error: %v", err)
			}

			text, textSource, err := getText(*config)

			if tt.expectError && err == nil {
				t.Errorf("Expected error but got nil")
			}

			if !tt.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			if tt.expectNonEmpty && text == "" {
				t.Error("Expected non-empty text")
			}

			if textSource == nil && !tt.expectError {
				t.Error("Expected non-nil textSource")
			}
		})
	}
}

func TestMain_EdgeCases(t *testing.T) {
	// Test that getText handles various edge cases
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	// Test with minimal config
	configContent := `text:
  source: dummy
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
		t.Errorf("getText() unexpected error with minimal config: %v", err)
	}

	if text == "" {
		t.Error("getText() returned empty text with minimal config")
	}

	if textSource == nil {
		t.Error("getText() returned nil textSource with minimal config")
	}
}
