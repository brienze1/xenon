package consolefilter

import (
	"strings"
	"testing"
)

func TestGetConsoleFilter(t *testing.T) {
	cases := []struct {
		raw  string
		want any
	}{
		{raw: "all", want: All},
		{raw: "requests", want: Requests},
		{raw: "logs", want: Logs},
		{raw: "errors", want: Errors},
		{raw: "", want: All},
	}
	for _, tc := range cases {
		for _, raw := range []string{tc.raw, strings.ToUpper(tc.raw), " " + tc.raw + " "} {
			if got := GetConsoleFilter(raw); got != tc.want {
				t.Errorf("GetConsoleFilter(%q) = %q, want %q", raw, got, tc.want)
			}
		}
	}
	if got := GetConsoleFilter("not-a-value"); got != All {
		t.Errorf("GetConsoleFilter(unknown) = %q, want all", got)
	}
}

func TestConsoleFilterEqualsAndString(t *testing.T) {
	value := Errors
	if !value.Equals(strings.ToUpper("errors")) || value.Equals("errorsx") {
		t.Errorf("Equals is not a case-insensitive comparison")
	}
	if value.String() != "errors" {
		t.Errorf("String() = %q", value.String())
	}
}
