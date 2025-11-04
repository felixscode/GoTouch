package sources

import "testing"

func TestDummySource_GetText(t *testing.T) {
	source := &DummySource{}

	text, err := source.GetText()

	// Should not return error
	if err != nil {
		t.Errorf("GetText() unexpected error: %v", err)
	}

	// Should return non-empty text
	if text == "" {
		t.Errorf("GetText() returned empty text")
	}

	// Should return expected dummy text
	expected := "The quick brown fox jumps over the lazy dog near the old wooden bridge. "
	if text != expected {
		t.Errorf("GetText() = %q, want %q", text, expected)
	}

	// Should be consistent (same text every time)
	text2, err2 := source.GetText()
	if err2 != nil {
		t.Errorf("GetText() second call unexpected error: %v", err2)
	}
	if text != text2 {
		t.Errorf("GetText() not consistent: first=%q, second=%q", text, text2)
	}
}
