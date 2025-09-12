package sources

import "fmt"

type TextSource interface {
	GetText() (string, error)
}

func NewTextSource(sourceType string) (TextSource, error) {
	switch sourceType {
	case "dummy", "Dummy", "dummy_source", "DummySource":
		return &DummySource{}, nil
	case "llm", "LLM", "llm_source", "LLMSource":
		return &LLMSource{}, nil
	// case "wiki", "Wiki", "wiki_source", "WikiSource":
	// 	return &WikiSource{}, nil
	default:
		return nil, fmt.Errorf("unknown source type: %s", sourceType)
	}
}
