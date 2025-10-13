package sources

type DummySource struct{}

func (d *DummySource) GetText() (string, error) {
	var dummyText string = "The quick brown fox jumps over the lazy dog near the old wooden bridge. "
	return dummyText, nil
}
