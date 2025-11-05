package sources

import (
	"fmt"
	"go-touch/internal/types"
)

type TextSource interface {
	GetText() (string, error)
}

func NewTextSource(sourceType string, config types.TextConfig) (TextSource, error) {
	switch sourceType {
	case "dummy", "Dummy", "dummy_source", "DummySource":
		return &DummySource{}, nil
	case "llm", "LLM", "llm_source", "LLMSource":
		llmSource, err := NewLLMSource(config.LLM)
		if err != nil {
			if config.LLM.FallbackToDummy {
				fmt.Printf("Warning: Failed to initialize LLM source (%v), falling back to dummy\n", err)
				return &DummySource{}, nil
			}
			return nil, fmt.Errorf("failed to initialize LLM source: %w", err)
		}
		return llmSource, nil
	// case "wiki", "Wiki", "wiki_source", "WikiSource":
	// 	return &WikiSource{}, nil
	default:
		return nil, fmt.Errorf("unknown source type: %s", sourceType)
	}
}
