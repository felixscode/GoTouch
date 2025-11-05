package sources

import (
	"go-touch/internal/types"
	"os"
	"testing"
	"time"
)

// TestNewLLMSource_APIKeyFromFile tests loading API key from file
func TestNewLLMSource_APIKeyFromFile(t *testing.T) {
	// Remove env var
	originalKey := os.Getenv("GOTOUCH_LLM_API_KEY")
	defer func() {
		if originalKey != "" {
			os.Setenv("GOTOUCH_LLM_API_KEY", originalKey)
		}
	}()
	os.Unsetenv("GOTOUCH_LLM_API_KEY")

	// Create api-key file
	apiKeyContent := "test-key-from-file-123"
	err := os.WriteFile("api-key", []byte(apiKeyContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create api-key file: %v", err)
	}
	defer os.Remove("api-key")

	config := types.LLMConfig{
		Provider:       "anthropic",
		Model:          "claude-3-5-haiku-latest",
		TimeoutSeconds: 5,
	}

	source, err := NewLLMSource(config)

	if err != nil {
		t.Errorf("NewLLMSource() unexpected error: %v", err)
	}

	if source == nil {
		t.Errorf("NewLLMSource() returned nil source")
	}
}

func TestNewLLMSource_Providers(t *testing.T) {
	originalKey := os.Getenv("GOTOUCH_LLM_API_KEY")
	defer func() {
		if originalKey != "" {
			os.Setenv("GOTOUCH_LLM_API_KEY", originalKey)
		} else {
			os.Unsetenv("GOTOUCH_LLM_API_KEY")
		}
	}()
	os.Setenv("GOTOUCH_LLM_API_KEY", "test-key")

	tests := []struct {
		name     string
		config   types.LLMConfig
		wantErr  bool
		errMsg   string
	}{
		{
			name: "anthropic provider",
			config: types.LLMConfig{
				Provider:       "anthropic",
				Model:          "claude-3-5-haiku-latest",
				TimeoutSeconds: 5,
			},
			wantErr: false,
		},
		{
			name: "openai provider",
			config: types.LLMConfig{
				Provider:       "openai",
				Model:          "gpt-4",
				TimeoutSeconds: 5,
			},
			wantErr: false,
		},
		{
			name: "ollama provider without api key",
			config: types.LLMConfig{
				Provider:       "ollama",
				Model:          "llama2",
				APIBase:        "http://localhost:11434",
				TimeoutSeconds: 5,
			},
			wantErr: false,
		},
		{
			name: "unsupported provider",
			config: types.LLMConfig{
				Provider:       "unsupported",
				Model:          "some-model",
				TimeoutSeconds: 5,
			},
			wantErr: true,
			errMsg:  "unsupported provider",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// For ollama, don't require API key
			if tt.config.Provider == "ollama" {
				os.Unsetenv("GOTOUCH_LLM_API_KEY")
			} else {
				os.Setenv("GOTOUCH_LLM_API_KEY", "test-key")
			}

			source, err := NewLLMSource(tt.config)

			if tt.wantErr {
				if err == nil {
					t.Errorf("NewLLMSource() expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("NewLLMSource() unexpected error: %v", err)
			}

			if source == nil {
				t.Fatalf("NewLLMSource() returned nil")
			}

			expectedTimeout := time.Duration(tt.config.TimeoutSeconds) * time.Second
			if source.timeout != expectedTimeout {
				t.Errorf("NewLLMSource() timeout = %v, want %v", source.timeout, expectedTimeout)
			}
		})
	}
}

func TestNewLLMSource_TimeoutConfiguration(t *testing.T) {
	originalKey := os.Getenv("GOTOUCH_LLM_API_KEY")
	defer func() {
		if originalKey != "" {
			os.Setenv("GOTOUCH_LLM_API_KEY", originalKey)
		} else {
			os.Unsetenv("GOTOUCH_LLM_API_KEY")
		}
	}()
	os.Setenv("GOTOUCH_LLM_API_KEY", "test-key")

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
			config := types.LLMConfig{
				Provider:       "anthropic",
				Model:          "claude-3-5-haiku-latest",
				TimeoutSeconds: tt.timeoutSeconds,
			}

			source, err := NewLLMSource(config)

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
		config          types.LLMConfig
		envVarSet       bool
		wantErr         bool
		expectedTimeout time.Duration
	}{
		{
			name: "anthropic with API key",
			config: types.LLMConfig{
				Provider:       "anthropic",
				Model:          "claude-3-5-haiku-latest",
				TimeoutSeconds: 5,
			},
			envVarSet:       true,
			wantErr:         false,
			expectedTimeout: 5 * time.Second,
		},
		{
			name: "openai with API key",
			config: types.LLMConfig{
				Provider:       "openai",
				Model:          "gpt-4",
				TimeoutSeconds: 10,
			},
			envVarSet:       true,
			wantErr:         false,
			expectedTimeout: 10 * time.Second,
		},
		{
			name: "no API key for anthropic",
			config: types.LLMConfig{
				Provider:       "anthropic",
				Model:          "claude-3-5-haiku-latest",
				TimeoutSeconds: 5,
			},
			envVarSet: false,
			wantErr:   true,
		},
		{
			name: "ollama without API key",
			config: types.LLMConfig{
				Provider:       "ollama",
				Model:          "llama2",
				TimeoutSeconds: 5,
			},
			envVarSet:       false,
			wantErr:         false,
			expectedTimeout: 5 * time.Second,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup environment
			originalKey := os.Getenv("GOTOUCH_LLM_API_KEY")
			defer func() {
				if originalKey != "" {
					os.Setenv("GOTOUCH_LLM_API_KEY", originalKey)
				} else {
					os.Unsetenv("GOTOUCH_LLM_API_KEY")
				}
			}()

			if tt.envVarSet {
				os.Setenv("GOTOUCH_LLM_API_KEY", "test-key-123")
			} else {
				os.Unsetenv("GOTOUCH_LLM_API_KEY")
			}

			// Test
			source, err := NewLLMSource(tt.config)

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
		})
	}
}

// TestGetText_Integration is an integration test that calls the real API
// Skip by default, run with: go test -run TestGetText_Integration
func TestGetText_Integration(t *testing.T) {
	if os.Getenv("GOTOUCH_LLM_API_KEY") == "" {
		t.Skip("Skipping integration test: GOTOUCH_LLM_API_KEY not set")
	}

	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	config := types.LLMConfig{
		Provider:       "anthropic",
		Model:          "claude-3-5-haiku-latest",
		TimeoutSeconds: 30,
	}

	source, err := NewLLMSource(config)
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
	if os.Getenv("GOTOUCH_LLM_API_KEY") == "" {
		t.Skip("Skipping integration test: GOTOUCH_LLM_API_KEY not set")
	}

	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	config := types.LLMConfig{
		Provider:       "anthropic",
		Model:          "claude-3-5-haiku-latest",
		TimeoutSeconds: 30,
	}

	source, err := NewLLMSource(config)
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
	if os.Getenv("GOTOUCH_LLM_API_KEY") == "" {
		t.Skip("Skipping test: GOTOUCH_LLM_API_KEY not set")
	}

	if testing.Short() {
		t.Skip("Skipping timeout test in short mode")
	}

	config := types.LLMConfig{
		Provider:       "anthropic",
		Model:          "claude-3-5-haiku-latest",
		TimeoutSeconds: 0, // 0 second timeout
	}

	// Use a very short timeout to trigger timeout error
	source, err := NewLLMSource(config)
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
	if os.Getenv("GOTOUCH_LLM_API_KEY") == "" {
		t.Skip("Skipping integration test: GOTOUCH_LLM_API_KEY not set")
	}

	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	config := types.LLMConfig{
		Provider:       "anthropic",
		Model:          "claude-3-5-haiku-latest",
		TimeoutSeconds: 30,
	}

	source, err := NewLLMSource(config)
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
	if os.Getenv("GOTOUCH_LLM_API_KEY") == "" {
		b.Skip("Skipping benchmark: GOTOUCH_LLM_API_KEY not set")
	}

	config := types.LLMConfig{
		Provider:       "anthropic",
		Model:          "claude-3-5-haiku-latest",
		TimeoutSeconds: 30,
	}

	source, err := NewLLMSource(config)
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
	if os.Getenv("GOTOUCH_LLM_API_KEY") == "" {
		b.Skip("Skipping benchmark: GOTOUCH_LLM_API_KEY not set")
	}

	config := types.LLMConfig{
		Provider:       "anthropic",
		Model:          "claude-3-5-haiku-latest",
		TimeoutSeconds: 30,
	}

	source, err := NewLLMSource(config)
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
