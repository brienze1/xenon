package envcolor

import (
	"strings"
	"testing"
)

func TestGetEnvColor(t *testing.T) {
	cases := []struct {
		raw  string
		want any
	}{
		{raw: "none", want: None},
		{raw: "green", want: Green},
		{raw: "cyan", want: Cyan},
		{raw: "magenta", want: Magenta},
		{raw: "yellow", want: Yellow},
		{raw: "red", want: Red},
	}
	for _, tc := range cases {
		for _, raw := range []string{tc.raw, strings.ToUpper(tc.raw), " " + tc.raw + " "} {
			if got := GetEnvColor(raw); got != tc.want {
				t.Errorf("GetEnvColor(%q) = %q, want %q", raw, got, tc.want)
			}
		}
	}
	if got := GetEnvColor("not-a-value"); got != Unknown {
		t.Errorf("GetEnvColor(unknown) = %q, want unknown", got)
	}
}

func TestEnvColorEqualsAndString(t *testing.T) {
	value := None
	if !value.Equals(strings.ToUpper("none")) || value.Equals("nonex") {
		t.Errorf("Equals is not a case-insensitive comparison")
	}
	if value.String() != "none" {
		t.Errorf("String() = %q", value.String())
	}
}
