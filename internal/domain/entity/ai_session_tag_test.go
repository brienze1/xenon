package entity

import "testing"

func TestAISessionTagIsTheFirstSixCharacters(t *testing.T) {
	tests := map[string]string{
		"ABCDEFGHIJ": "ABCDEF",
		"ABCDEF":     "ABCDEF",
		"ABC":        "ABC",
		"":           "",
	}
	for sessionId, want := range tests {
		if got := AISessionTag(sessionId); got != want {
			t.Errorf("AISessionTag(%q) = %q, want %q", sessionId, got, want)
		}
	}
}
