package request

import (
	"net/http"
	"testing"
)

type testGenerator struct{}

func (t *testGenerator) Generate(limit int) string {
	return "1234"
}

func Test_newShuffler(t *testing.T) {
	r, _ := http.NewRequest("GET", "https://test.com/{{some_regex_pattern_here}}", nil)
	s := newShuffler(r, func(match string) generator {
		return &testGenerator{}
	})

	s.Shuffle(r)

	expectedURL := "https://test.com/1234"
	if r.URL.String() != expectedURL {
		t.Errorf("Generated request URL does not match: %s, expected: %s", r.URL, expectedURL)
	}
}
