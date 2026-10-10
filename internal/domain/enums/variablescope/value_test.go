package variablescope

import (
	"strings"
	"testing"
)

func TestGetVariableScope(t *testing.T) {
	cases := []struct {
		raw  string
		want any
	}{
		{raw: "local", want: Local},
		{raw: "environment", want: Environment},
		{raw: "collection", want: Collection},
		{raw: "globals", want: Globals},
		{raw: "dynamic", want: Dynamic},
		{raw: "none", want: None},
	}
	for _, tc := range cases {
		for _, raw := range []string{tc.raw, strings.ToUpper(tc.raw), " " + tc.raw + " "} {
			if got := GetVariableScope(raw); got != tc.want {
				t.Errorf("GetVariableScope(%q) = %q, want %q", raw, got, tc.want)
			}
		}
	}
	if got := GetVariableScope("not-a-value"); got != Unknown {
		t.Errorf("GetVariableScope(unknown) = %q, want unknown", got)
	}
}

func TestVariableScopeEqualsAndString(t *testing.T) {
	value := Local
	if !value.Equals(strings.ToUpper("local")) || value.Equals("localx") {
		t.Errorf("Equals is not a case-insensitive comparison")
	}
	if value.String() != "local" {
		t.Errorf("String() = %q", value.String())
	}
}
