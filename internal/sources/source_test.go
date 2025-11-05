package sources

import (
	"go-touch/internal/types"
	"os"
	"testing"
)

func TestNewTextSource_Dummy(t *testing.T) {
	tests := []struct {
		name       string
		sourceType string
		wantType   string
	}{
		{
			name:       "dummy lowercase",
			sourceType: "dummy",
			wantType:   "*sources.DummySource",
		},
		{
			name:       "Dummy capitalized",
			sourceType: "Dummy",
			wantType:   "*sources.DummySource",
		},
		{
			name:       "dummy_source",
			sourceType: "dummy_source",
			wantType:   "*sources.DummySource",
		},
		{
			name:       "DummySource",
			sourceType: "DummySource",
			wantType:   "*sources.DummySource",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := types.TextConfig{
				Source: tt.sourceType,
			}

			source, err := NewTextSource(tt.sourceType, config)

			if err != nil {
				t.Errorf("NewTextSource() unexpected error: %v", err)
				return
			}

			if source == nil {
				t.Errorf("NewTextSource() returned nil source")
				return
			}

			// Verify it's a DummySource
			_, ok := source.(*DummySource)
			if !ok {
				t.Errorf("NewTextSource() returned %T, want *DummySource", source)
			}
		})
	}
}

func TestNewTextSource_LLM_WithAPIKey(t *testing.T) {
	// Setup environment
	originalKey := os.Getenv("GOTOUCH_LLM_API_KEY")
	defer func() {
		if originalKey != "" {
			os.Setenv("GOTOUCH_LLM_API_KEY", originalKey)
		} else {
			os.Unsetenv("GOTOUCH_LLM_API_KEY")
		}
	}()
	os.Setenv("GOTOUCH_LLM_API_KEY", "test-key-123")

	tests := []struct {
		name       string
		sourceType string
	}{
		{"llm lowercase", "llm"},
		{"LLM uppercase", "LLM"},
		{"llm_source", "llm_source"},
		{"LLMSource", "LLMSource"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := types.TextConfig{
				Source: tt.sourceType,
				LLM: types.LLMConfig{
					Provider:       "anthropic",
					Model:          "claude-3-5-haiku-latest",
					TimeoutSeconds: 5,
				},
			}

			source, err := NewTextSource(tt.sourceType, config)

			if err != nil {
				t.Errorf("NewTextSource() unexpected error: %v", err)
				return
			}

			if source == nil {
				t.Errorf("NewTextSource() returned nil source")
				return
			}

			// Verify it's an LLMSource
			_, ok := source.(*LLMSource)
			if !ok {
				t.Errorf("NewTextSource() returned %T, want *LLMSource", source)
			}
		})
	}
}

func TestNewTextSource_LLM_WithoutAPIKey_FallbackToDummy(t *testing.T) {
	// Remove API key
	originalKey := os.Getenv("GOTOUCH_LLM_API_KEY")
	defer func() {
		if originalKey != "" {
			os.Setenv("GOTOUCH_LLM_API_KEY", originalKey)
		}
	}()
	os.Unsetenv("GOTOUCH_LLM_API_KEY")

	// Remove api-key file if it exists
	os.Remove("api-key")

	config := types.TextConfig{
		Source: "llm",
		LLM: types.LLMConfig{
			Provider:        "anthropic",
			Model:           "claude-3-5-haiku-latest",
			TimeoutSeconds:  5,
			FallbackToDummy: true,
		},
	}

	source, err := NewTextSource("llm", config)

	if err != nil {
		t.Errorf("NewTextSource() unexpected error: %v", err)
		return
	}

	// Should fallback to DummySource
	_, ok := source.(*DummySource)
	if !ok {
		t.Errorf("NewTextSource() returned %T, want *DummySource (fallback)", source)
	}
}

func TestNewTextSource_LLM_WithoutAPIKey_NoFallback(t *testing.T) {
	// Remove API key
	originalKey := os.Getenv("GOTOUCH_LLM_API_KEY")
	defer func() {
		if originalKey != "" {
			os.Setenv("GOTOUCH_LLM_API_KEY", originalKey)
		}
	}()
	os.Unsetenv("GOTOUCH_LLM_API_KEY")

	// Remove api-key file if it exists
	os.Remove("api-key")

	config := types.TextConfig{
		Source: "llm",
		LLM: types.LLMConfig{
			Provider:        "anthropic",
			Model:           "claude-3-5-haiku-latest",
			TimeoutSeconds:  5,
			FallbackToDummy: false,
		},
	}

	source, err := NewTextSource("llm", config)

	if err == nil {
		t.Errorf("NewTextSource() expected error, got nil")
	}

	if source != nil {
		t.Errorf("NewTextSource() expected nil source, got %T", source)
	}
}

func TestNewTextSource_UnknownSource(t *testing.T) {
	config := types.TextConfig{
		Source: "unknown",
	}

	source, err := NewTextSource("unknown", config)

	if err == nil {
		t.Errorf("NewTextSource() expected error for unknown source, got nil")
	}

	if source != nil {
		t.Errorf("NewTextSource() expected nil source for unknown type, got %T", source)
	}

	// Check error message contains the unknown type
	expectedMsg := "unknown source type: unknown"
	if err.Error() != expectedMsg {
		t.Errorf("NewTextSource() error = %q, want %q", err.Error(), expectedMsg)
	}
}
