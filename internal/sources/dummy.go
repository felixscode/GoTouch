package sources

type DummySource struct{}

func (d *DummySource) GetText() (string, error) {
	var dummyText string = "The quick brown fox jumps over the lazy dog near the old wooden bridge. During summer evenings, children often play games in the park while their parents watch from comfortable benches. Technology has transformed how we communicate with friends and family across great distances. Modern computers process information at incredible speeds, making complex calculations seem effortless. Students learn new skills through interactive online platforms that adapt to individual learning styles. Fresh vegetables from local farmers markets provide essential nutrients for healthy living. Musicians create beautiful melodies using both traditional instruments and digital software. Photography captures precious memories that last forever."
	return dummyText, nil
}
