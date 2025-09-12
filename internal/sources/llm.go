package sources

type LLMSource struct{}

func (l *LLMSource) GetText() (string, error) {
	return "", nil
}
