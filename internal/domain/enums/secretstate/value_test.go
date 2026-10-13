package secretstate

import (
	"strings"
	"testing"
)

func TestGetSecretState(t *testing.T) {
	cases := []struct {
		raw  string
		want any
	}{
		{raw: "set", want: Set},
		{raw: "missing", want: Missing},
		{raw: "locked", want: Locked},
	}
	for _, tc := range cases {
		for _, raw := range []string{tc.raw, strings.ToUpper(tc.raw), " " + tc.raw + " "} {
			if got := GetSecretState(raw); got != tc.want {
				t.Errorf("GetSecretState(%q) = %q, want %q", raw, got, tc.want)
			}
		}
	}
	if got := GetSecretState("not-a-value"); got != Unknown {
		t.Errorf("GetSecretState(unknown) = %q, want unknown", got)
	}
}

func TestSecretStateEqualsAndString(t *testing.T) {
	value := Set
	if !value.Equals(strings.ToUpper("set")) || value.Equals("setx") {
		t.Errorf("Equals is not a case-insensitive comparison")
	}
	if value.String() != "set" {
		t.Errorf("String() = %q", value.String())
	}
}
