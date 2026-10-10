package consoleentrykind

import (
	"strings"
	"testing"
)

func TestGetConsoleEntryKind(t *testing.T) {
	cases := []struct {
		raw  string
		want any
	}{
		{raw: "request", want: Request},
		{raw: "log", want: Log},
		{raw: "error", want: Error},
		{raw: "test_failure", want: TestFailure},
		{raw: "info", want: Info},
	}
	for _, tc := range cases {
		for _, raw := range []string{tc.raw, strings.ToUpper(tc.raw), " " + tc.raw + " "} {
			if got := GetConsoleEntryKind(raw); got != tc.want {
				t.Errorf("GetConsoleEntryKind(%q) = %q, want %q", raw, got, tc.want)
			}
		}
	}
	if got := GetConsoleEntryKind("not-a-value"); got != Unknown {
		t.Errorf("GetConsoleEntryKind(unknown) = %q, want unknown", got)
	}
}

func TestConsoleEntryKindEqualsAndString(t *testing.T) {
	value := Request
	if !value.Equals(strings.ToUpper("request")) || value.Equals("requestx") {
		t.Errorf("Equals is not a case-insensitive comparison")
	}
	if value.String() != "request" {
		t.Errorf("String() = %q", value.String())
	}
}
