package mcptokenstate

import (
	"testing"

	"github.com/brienze1/xenon/internal/integration/entrypoint/enums"
)

func TestGetMCPTokenState(t *testing.T) {
	cases := map[string]enums.MCPTokenState{"valid": Valid, " Expired ": Expired, "revoked": Unknown, "": Unknown}
	for input, want := range cases {
		if got := GetMCPTokenState(input); got != want {
			t.Errorf("GetMCPTokenState(%q) = %s, want %s", input, got, want)
		}
	}
}

func TestMCPTokenStateEquals(t *testing.T) {
	if !Valid.Equals("VALID") || Valid.Equals("expired") {
		t.Errorf("Equals must compare ignoring case")
	}
}
