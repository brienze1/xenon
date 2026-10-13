package snippetphase

import (
	"strings"
	"testing"
)

func TestGetSnippetPhase(t *testing.T) {
	cases := []struct {
		raw  string
		want any
	}{
		{raw: "any", want: Any},
		{raw: "preRequest", want: PreRequest},
		{raw: "postResponse", want: PostResponse},
	}
	for _, tc := range cases {
		for _, raw := range []string{tc.raw, strings.ToUpper(tc.raw), " " + tc.raw + " "} {
			if got := GetSnippetPhase(raw); got != tc.want {
				t.Errorf("GetSnippetPhase(%q) = %q, want %q", raw, got, tc.want)
			}
		}
	}
	if got := GetSnippetPhase("not-a-value"); got != Unknown {
		t.Errorf("GetSnippetPhase(unknown) = %q, want unknown", got)
	}
}

func TestSnippetPhaseEqualsAndString(t *testing.T) {
	value := Any
	if !value.Equals(strings.ToUpper("any")) || value.Equals("anyx") {
		t.Errorf("Equals is not a case-insensitive comparison")
	}
	if value.String() != "any" {
		t.Errorf("String() = %q", value.String())
	}
}
