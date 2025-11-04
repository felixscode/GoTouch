package sources

import (
	"context"
	"fmt"
	"go-touch/internal/config"
	"os"
	"strings"
	"time"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"
)

type LLMSource struct {
	client  anthropic.Client
	model   anthropic.Model
	timeout time.Duration
}

func NewLLMSource(modelName string, timeoutSeconds int) (*LLMSource, error) {
	apiKey := os.Getenv("ANTHROPIC_API_KEY")
	if apiKey == "" {
		// Try loading from api-key file as fallback
		// Check config directory first, then current directory
		apiKeyPath := config.FindAPIKeyFile()
		if apiKeyPath != "" {
			keyData, err := os.ReadFile(apiKeyPath)
			if err == nil && len(keyData) > 0 {
				apiKey = strings.TrimSpace(string(keyData))
			}
		}
	}

	if apiKey == "" {
		return nil, fmt.Errorf("ANTHROPIC_API_KEY environment variable not set and api-key file not found")
	}

	client := anthropic.NewClient(
		option.WithAPIKey(apiKey),
	)

	// }
	model := anthropic.ModelClaude3_5HaikuLatest
	if modelName == "sonnet" {
		model = anthropic.ModelClaude3_5SonnetLatest
	}

	return &LLMSource{
		client:  client,
		model:   model,
		timeout: time.Duration(timeoutSeconds) * time.Second,
	}, nil
}

func (l *LLMSource) GetText() (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), l.timeout)
	defer cancel()

	prompt := `Generate a single interesting sentence for typing practice.
Make it varied content (quotes, facts, or creative).
Length: 50-80 characters.
Only output the sentence, nothing else.`

	message, err := l.client.Messages.New(ctx, anthropic.MessageNewParams{
		Model:     l.model,
		MaxTokens: 100,
		Messages: []anthropic.MessageParam{
			{
				Role: anthropic.MessageParamRoleUser,
				Content: []anthropic.ContentBlockParamUnion{
					{
						OfText: &anthropic.TextBlockParam{
							Text: prompt,
						},
					},
				},
			},
		},
	})

	if err != nil {
		return "", fmt.Errorf("API call failed: %w", err)
	}

	if len(message.Content) == 0 {
		return "", fmt.Errorf("empty response from API")
	}

	// Extract text from content blocks
	for _, block := range message.Content {
		if block.Type == "text" && block.Text != "" {
			return strings.TrimSpace(block.Text), nil
		}
	}

	return "", fmt.Errorf("no text content in response")
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

	message, err := l.client.Messages.New(ctx, anthropic.MessageNewParams{
		Model:     l.model,
		MaxTokens: 150,
		Messages: []anthropic.MessageParam{
			{
				Role: anthropic.MessageParamRoleUser,
				Content: []anthropic.ContentBlockParamUnion{
					{
						OfText: &anthropic.TextBlockParam{
							Text: promptBuilder.String(),
						},
					},
				},
			},
		},
	})

	if err != nil {
		return "", fmt.Errorf("API call failed: %w", err)
	}

	if len(message.Content) == 0 {
		return "", fmt.Errorf("empty response from API")
	}

	// Extract text from content blocks
	for _, block := range message.Content {
		if block.Type == "text" && block.Text != "" {
			return strings.TrimSpace(block.Text), nil
		}
	}

	return "", fmt.Errorf("no text content in response")
}
