package scriptmemberkind

import (
	"strings"
	"testing"
)

func TestGetScriptMemberKind(t *testing.T) {
	cases := []struct {
		raw  string
		want any
	}{
		{raw: "object", want: Object},
		{raw: "function", want: Function},
		{raw: "property", want: Property},
	}
	for _, tc := range cases {
		for _, raw := range []string{tc.raw, strings.ToUpper(tc.raw), " " + tc.raw + " "} {
			if got := GetScriptMemberKind(raw); got != tc.want {
				t.Errorf("GetScriptMemberKind(%q) = %q, want %q", raw, got, tc.want)
			}
		}
	}
	if got := GetScriptMemberKind("not-a-value"); got != Unknown {
		t.Errorf("GetScriptMemberKind(unknown) = %q, want unknown", got)
	}
}

func TestScriptMemberKindEqualsAndString(t *testing.T) {
	value := Function
	if !value.Equals(strings.ToUpper("function")) || value.Equals("functionx") {
		t.Errorf("Equals is not a case-insensitive comparison")
	}
	if value.String() != "function" {
		t.Errorf("String() = %q", value.String())
	}
}
