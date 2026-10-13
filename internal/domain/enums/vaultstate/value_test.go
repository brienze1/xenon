package vaultstate

import (
	"strings"
	"testing"
)

func TestGetVaultState(t *testing.T) {
	cases := []struct {
		raw  string
		want any
	}{
		{raw: "empty", want: Empty},
		{raw: "unlocked", want: Unlocked},
		{raw: "locked", want: Locked},
		{raw: "broken", want: Broken},
	}
	for _, tc := range cases {
		for _, raw := range []string{tc.raw, strings.ToUpper(tc.raw), " " + tc.raw + " "} {
			if got := GetVaultState(raw); got != tc.want {
				t.Errorf("GetVaultState(%q) = %q, want %q", raw, got, tc.want)
			}
		}
	}
	if got := GetVaultState("not-a-value"); got != Unknown {
		t.Errorf("GetVaultState(unknown) = %q, want unknown", got)
	}
}

func TestVaultStateEqualsAndString(t *testing.T) {
	value := Empty
	if !value.Equals(strings.ToUpper("empty")) || value.Equals("emptyx") {
		t.Errorf("Equals is not a case-insensitive comparison")
	}
	if value.String() != "empty" {
		t.Errorf("String() = %q", value.String())
	}
}
