package sources

import (
	"context"
	"fmt"
	"go-touch/internal/config"
	"go-touch/internal/types"
	"math/rand"
	"os"
	"strings"
	"time"

	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/anthropic"
	"github.com/tmc/langchaingo/llms/ollama"
	"github.com/tmc/langchaingo/llms/openai"
)

type LLMSource struct {
	model   llms.Model
	timeout time.Duration
}

// NewLLMSource creates a new LLM source with the specified provider and configuration
func NewLLMSource(llmConfig types.LLMConfig) (*LLMSource, error) {
	// Get API key from environment variable
	apiKey := os.Getenv("GOTOUCH_LLM_API_KEY")
	if apiKey == "" {
		// Try loading from api-key file as fallback
		apiKeyPath := config.FindAPIKeyFile()
		if apiKeyPath != "" {
			keyData, err := os.ReadFile(apiKeyPath)
			if err == nil && len(keyData) > 0 {
				apiKey = strings.TrimSpace(string(keyData))
			}
		}
	}

	// Create provider-specific client
	var model llms.Model
	var err error

	switch strings.ToLower(llmConfig.Provider) {
	case "anthropic":
		if apiKey == "" {
			return nil, fmt.Errorf("GOTOUCH_LLM_API_KEY environment variable not set and api-key file not found")
		}
		model, err = anthropic.New(
			anthropic.WithModel(llmConfig.Model),
			anthropic.WithToken(apiKey),
		)
		if err != nil {
			return nil, fmt.Errorf("failed to create Anthropic client: %w", err)
		}

	case "openai":
		if apiKey == "" {
			return nil, fmt.Errorf("GOTOUCH_LLM_API_KEY environment variable not set and api-key file not found")
		}
		opts := []openai.Option{
			openai.WithModel(llmConfig.Model),
			openai.WithToken(apiKey),
		}
		if llmConfig.APIBase != "" {
			opts = append(opts, openai.WithBaseURL(llmConfig.APIBase))
		}
		model, err = openai.New(opts...)
		if err != nil {
			return nil, fmt.Errorf("failed to create OpenAI client: %w", err)
		}

	case "ollama":
		opts := []ollama.Option{
			ollama.WithModel(llmConfig.Model),
		}
		if llmConfig.APIBase != "" {
			opts = append(opts, ollama.WithServerURL(llmConfig.APIBase))
		} else {
			// Default Ollama endpoint
			opts = append(opts, ollama.WithServerURL("http://localhost:11434"))
		}
		model, err = ollama.New(opts...)
		if err != nil {
			return nil, fmt.Errorf("failed to create Ollama client: %w", err)
		}

	default:
		return nil, fmt.Errorf("unsupported provider: %s (supported: anthropic, openai, ollama)", llmConfig.Provider)
	}

	return &LLMSource{
		model:   model,
		timeout: time.Duration(llmConfig.TimeoutSeconds) * time.Second,
	}, nil
}

func (l *LLMSource) GetText() (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), l.timeout)
	defer cancel()

	// Generate a random letter (A-Z)
	randomLetter := string(rune('A' + rand.Intn(26)))

	prompt := fmt.Sprintf(`Generate a single interesting sentence for typing practice.
Make it varied content (quotes, facts, or creative).
Length: 50-80 characters.
IMPORTANT: Start the sentence with the letter "%s".
Only output the sentence, nothing else.`, randomLetter)

	response, err := llms.GenerateFromSinglePrompt(ctx, l.model, prompt)
	if err != nil {
		return "", fmt.Errorf("API call failed: %w", err)
	}

	return strings.TrimSpace(response), nil
}

func (l *LLMSource) GetNextSentence(previousSentence string, errorChars []rune, errorWords []string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), l.timeout)
	defer cancel()

	// Build prompt with context and error patterns
	var promptBuilder strings.Builder
	promptBuilder.WriteString(fmt.Sprintf("Previous sentence: \"%s\"\n\n", previousSentence))

	if len(errorChars) > 0 {
		promptBuilder.WriteString(fmt.Sprintf("User made mistakes typing these characters: %v\n", errorChars))
	}

	if len(errorWords) > 0 {
		promptBuilder.WriteString(fmt.Sprintf("User had trouble with these words: %v\n", errorWords))
	}

	promptBuilder.WriteString(`
Generate ONE sentence (50-80 characters) that:
1. Continues naturally from the previous sentence
2. Helps practice the problem characters and similar words
3. Maintains topic coherence with the previous sentence
4. Is interesting and natural to read

Only output the sentence, nothing else.`)

	response, err := llms.GenerateFromSinglePrompt(ctx, l.model, promptBuilder.String())
	if err != nil {
		return "", fmt.Errorf("API call failed: %w", err)
	}

	return strings.TrimSpace(response), nil
}
