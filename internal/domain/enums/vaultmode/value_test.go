package vaultmode

import (
	"strings"
	"testing"
)

func TestGetVaultMode(t *testing.T) {
	cases := []struct {
		raw  string
		want any
	}{
		{raw: "keychain", want: Keychain},
		{raw: "passphrase", want: Passphrase},
		{raw: "none", want: None},
	}
	for _, tc := range cases {
		for _, raw := range []string{tc.raw, strings.ToUpper(tc.raw), " " + tc.raw + " "} {
			if got := GetVaultMode(raw); got != tc.want {
				t.Errorf("GetVaultMode(%q) = %q, want %q", raw, got, tc.want)
			}
		}
	}
	if got := GetVaultMode("not-a-value"); got != Unknown {
		t.Errorf("GetVaultMode(unknown) = %q, want unknown", got)
	}
}

func TestVaultModeEqualsAndString(t *testing.T) {
	value := Keychain
	if !value.Equals(strings.ToUpper("keychain")) || value.Equals("keychainx") {
		t.Errorf("Equals is not a case-insensitive comparison")
	}
	if value.String() != "keychain" {
		t.Errorf("String() = %q", value.String())
	}
}
