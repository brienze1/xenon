package consolelevel

import (
	"strings"
	"testing"
)

func TestGetConsoleLevel(t *testing.T) {
	cases := []struct {
		raw  string
		want any
	}{
		{raw: "log", want: Log},
		{raw: "info", want: Info},
		{raw: "warn", want: Warn},
		{raw: "error", want: Error},
		{raw: "debug", want: Debug},
		{raw: "clear", want: Clear},
	}
	for _, tc := range cases {
		for _, raw := range []string{tc.raw, strings.ToUpper(tc.raw), " " + tc.raw + " "} {
			if got := GetConsoleLevel(raw); got != tc.want {
				t.Errorf("GetConsoleLevel(%q) = %q, want %q", raw, got, tc.want)
			}
		}
	}
	if got := GetConsoleLevel("not-a-value"); got != Unknown {
		t.Errorf("GetConsoleLevel(unknown) = %q, want unknown", got)
	}
}

func TestConsoleLevelEqualsAndString(t *testing.T) {
	value := Log
	if !value.Equals(strings.ToUpper("log")) || value.Equals("logx") {
		t.Errorf("Equals is not a case-insensitive comparison")
	}
	if value.String() != "log" {
		t.Errorf("String() = %q", value.String())
	}
}
