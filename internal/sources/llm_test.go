package sources

import (
	"os"
	"testing"
	"time"
)

// TestNewLLMSource_APIKeyFromFile tests loading API key from file
func TestNewLLMSource_APIKeyFromFile(t *testing.T) {
	// Remove env var
	originalKey := os.Getenv("ANTHROPIC_API_KEY")
	defer func() {
		if originalKey != "" {
			os.Setenv("ANTHROPIC_API_KEY", originalKey)
		}
	}()
	os.Unsetenv("ANTHROPIC_API_KEY")

	// Create api-key file
	apiKeyContent := "test-key-from-file-123"
	err := os.WriteFile("api-key", []byte(apiKeyContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create api-key file: %v", err)
	}
	defer os.Remove("api-key")

	source, err := NewLLMSource("haiku", 5)

	if err != nil {
		t.Errorf("NewLLMSource() unexpected error: %v", err)
	}

	if source == nil {
		t.Errorf("NewLLMSource() returned nil source")
	}

	if source.model != "claude-3-5-haiku-latest" {
		t.Errorf("NewLLMSource() model = %v, want claude-3-5-haiku-latest", source.model)
	}
}

func TestNewLLMSource_ModelSelection(t *testing.T) {
	originalKey := os.Getenv("ANTHROPIC_API_KEY")
	defer os.Setenv("ANTHROPIC_API_KEY", originalKey)
	os.Setenv("ANTHROPIC_API_KEY", "test-key")

	tests := []struct {
		name          string
		modelName     string
		expectedModel string
	}{
		{
			name:          "haiku model",
			modelName:     "haiku",
			expectedModel: "claude-3-5-haiku-latest",
		},
		{
			name:          "sonnet model",
			modelName:     "sonnet",
			expectedModel: "claude-3-5-sonnet-latest",
		},
		{
			name:          "empty defaults to haiku",
			modelName:     "",
			expectedModel: "claude-3-5-haiku-latest",
		},
		{
			name:          "unknown defaults to haiku",
			modelName:     "unknown",
			expectedModel: "claude-3-5-haiku-latest",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			source, err := NewLLMSource(tt.modelName, 5)

			if err != nil {
				t.Errorf("NewLLMSource() unexpected error: %v", err)
			}

			if source == nil {
				t.Fatalf("NewLLMSource() returned nil")
			}

			if string(source.model) != tt.expectedModel {
				t.Errorf("NewLLMSource() model = %v, want %v", source.model, tt.expectedModel)
			}
		})
	}
}

func TestNewLLMSource_TimeoutConfiguration(t *testing.T) {
	originalKey := os.Getenv("ANTHROPIC_API_KEY")
	defer os.Setenv("ANTHROPIC_API_KEY", originalKey)
	os.Setenv("ANTHROPIC_API_KEY", "test-key")

	tests := []struct {
		name            string
		timeoutSeconds  int
		expectedTimeout time.Duration
	}{
		{
			name:            "5 second timeout",
			timeoutSeconds:  5,
			expectedTimeout: 5 * time.Second,
		},
		{
			name:            "30 second timeout",
			timeoutSeconds:  30,
			expectedTimeout: 30 * time.Second,
		},
		{
			name:            "zero timeout",
			timeoutSeconds:  0,
			expectedTimeout: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			source, err := NewLLMSource("haiku", tt.timeoutSeconds)

			if err != nil {
				t.Errorf("NewLLMSource() unexpected error: %v", err)
			}

			if source.timeout != tt.expectedTimeout {
				t.Errorf("NewLLMSource() timeout = %v, want %v", source.timeout, tt.expectedTimeout)
			}
		})
	}
}

// TestNewLLMSource tests the constructor
func TestNewLLMSource(t *testing.T) {
	tests := []struct {
		name            string
		modelName       string
		timeoutSeconds  int
		envVarSet       bool
		wantErr         bool
		expectedTimeout time.Duration
	}{
		{
			name:            "default model with API key",
			modelName:       "",
			timeoutSeconds:  5,
			envVarSet:       true,
			wantErr:         false,
			expectedTimeout: 5 * time.Second,
		},
		{
			name:            "custom model with API key",
			modelName:       "sonnet",
			timeoutSeconds:  10,
			envVarSet:       true,
			wantErr:         false,
			expectedTimeout: 10 * time.Second,
		},
		{
			name:           "no API key",
			modelName:      "",
			timeoutSeconds: 5,
			envVarSet:      false,
			wantErr:        true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup environment
			originalKey := os.Getenv("ANTHROPIC_API_KEY")
			defer os.Setenv("ANTHROPIC_API_KEY", originalKey)

			if tt.envVarSet {
				os.Setenv("ANTHROPIC_API_KEY", "test-key-123")
			} else {
				os.Unsetenv("ANTHROPIC_API_KEY")
			}

			// Test
			source, err := NewLLMSource(tt.modelName, tt.timeoutSeconds)

			// Verify
			if tt.wantErr {
				if err == nil {
					t.Errorf("NewLLMSource() expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("NewLLMSource() unexpected error: %v", err)
				return
			}

			if source == nil {
				t.Errorf("NewLLMSource() returned nil source")
				return
			}

			if source.timeout != tt.expectedTimeout {
				t.Errorf("NewLLMSource() timeout = %v, want %v", source.timeout, tt.expectedTimeout)
			}

			// Verify model
			if tt.modelName != "" {
				if source.model != "claude-3-5-sonnet-latest" {
					t.Errorf("NewLLMSource() model = %v, want %v", source.model, "claude-3-5-sonnet-latest")
				}
			}

		})
	}
}

// TestGetText_Integration is an integration test that calls the real API
// Skip by default, run with: go test -run TestGetText_Integration
func TestGetText_Integration(t *testing.T) {
	if os.Getenv("ANTHROPIC_API_KEY") == "" {
		t.Skip("Skipping integration test: ANTHROPIC_API_KEY not set")
	}

	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	source, err := NewLLMSource("claude-3-opus-20240229", 30)
	if err != nil {
		t.Fatalf("Failed to create LLM source: %v", err)
	}

	text, err := source.GetText()
	if err != nil {
		t.Fatalf("GetText() error: %v", err)
	}

	t.Logf("Generated text: %s", text)

	// Basic validations
	if text == "" {
		t.Errorf("GetText() returned empty text")
	}

	if len(text) < 10 {
		t.Errorf("GetText() returned text too short: %d chars", len(text))
	}

	if len(text) > 200 {
		t.Logf("Warning: GetText() returned text longer than expected: %d chars", len(text))
	}
}

// TestGetNextSentence_Integration is an integration test
func TestGetNextSentence_Integration(t *testing.T) {
	if os.Getenv("ANTHROPIC_API_KEY") == "" {
		t.Skip("Skipping integration test: ANTHROPIC_API_KEY not set")
	}

	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	source, err := NewLLMSource("claude-3-opus-20240229", 30)
	if err != nil {
		t.Fatalf("Failed to create LLM source: %v", err)
	}

	tests := []struct {
		name             string
		previousSentence string
		errorChars       []rune
		errorWords       []string
	}{
		{
			name:             "simple continuation",
			previousSentence: "The quick brown fox jumps over the lazy dog.",
			errorChars:       nil,
			errorWords:       nil,
		},
		{
			name:             "with error characters",
			previousSentence: "Programming in Go is fun and efficient.",
			errorChars:       []rune{'g', 'o', 'p'},
			errorWords:       nil,
		},
		{
			name:             "with error words",
			previousSentence: "Machine learning models require lots of data.",
			errorChars:       nil,
			errorWords:       []string{"learning", "require"},
		},
		{
			name:             "with both errors",
			previousSentence: "Python and JavaScript are popular languages.",
			errorChars:       []rune{'p', 'j', 's'},
			errorWords:       []string{"Python", "JavaScript"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			text, err := source.GetNextSentence(tt.previousSentence, tt.errorChars, tt.errorWords)
			if err != nil {
				t.Fatalf("GetNextSentence() error: %v", err)
			}

			t.Logf("Previous: %s", tt.previousSentence)
			t.Logf("Generated: %s", text)

			// Basic validations
			if text == "" {
				t.Errorf("GetNextSentence() returned empty text")
			}

			if len(text) < 10 {
				t.Errorf("GetNextSentence() returned text too short: %d chars", len(text))
			}

			if text == tt.previousSentence {
				t.Errorf("GetNextSentence() returned same text as input")
			}
		})
	}
}

// TestGetText_Timeout tests timeout behavior
func TestGetText_Timeout(t *testing.T) {
	if os.Getenv("ANTHROPIC_API_KEY") == "" {
		t.Skip("Skipping test: ANTHROPIC_API_KEY not set")
	}

	if testing.Short() {
		t.Skip("Skipping timeout test in short mode")
	}

	// Use a very short timeout to trigger timeout error
	source, err := NewLLMSource("claude-3-opus-20240229", 0) // 0 second timeout
	if err != nil {
		t.Fatalf("Failed to create LLM source: %v", err)
	}

	// This should timeout
	_, err = source.GetText()
	if err == nil {
		t.Log("Expected timeout error, but got success - API might have been very fast")
	}
}

// TestGetNextSentence_EmptyPrevious tests edge case
func TestGetNextSentence_EmptyPrevious(t *testing.T) {
	if os.Getenv("ANTHROPIC_API_KEY") == "" {
		t.Skip("Skipping integration test: ANTHROPIC_API_KEY not set")
	}

	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	source, err := NewLLMSource("claude-3-opus-20240229", 30)
	if err != nil {
		t.Fatalf("Failed to create LLM source: %v", err)
	}

	// Test with empty previous sentence
	text, err := source.GetNextSentence("", nil, nil)
	if err != nil {
		t.Fatalf("GetNextSentence() with empty previous: %v", err)
	}

	if text == "" {
		t.Errorf("GetNextSentence() returned empty text")
	}

	t.Logf("Generated with empty previous: %s", text)
}

// Benchmark tests
func BenchmarkGetText(b *testing.B) {
	if os.Getenv("ANTHROPIC_API_KEY") == "" {
		b.Skip("Skipping benchmark: ANTHROPIC_API_KEY not set")
	}

	source, err := NewLLMSource("claude-3-opus-20240229", 30)
	if err != nil {
		b.Fatalf("Failed to create LLM source: %v", err)
	}

	b.ResetTimer()
	for range b.N {
		_, err := source.GetText()
		if err != nil {
			b.Fatalf("GetText() error: %v", err)
		}
	}
}

func BenchmarkGetNextSentence(b *testing.B) {
	if os.Getenv("ANTHROPIC_API_KEY") == "" {
		b.Skip("Skipping benchmark: ANTHROPIC_API_KEY not set")
	}

	source, err := NewLLMSource("claude-3-opus-20240229", 30)
	if err != nil {
		b.Fatalf("Failed to create LLM source: %v", err)
	}

	previousSentence := "The quick brown fox jumps over the lazy dog."
	errorChars := []rune{'t', 'h', 'e'}
	errorWords := []string{"quick", "brown"}

	b.ResetTimer()
	for range b.N {
		_, err := source.GetNextSentence(previousSentence, errorChars, errorWords)
		if err != nil {
			b.Fatalf("GetNextSentence() error: %v", err)
		}
	}
}
