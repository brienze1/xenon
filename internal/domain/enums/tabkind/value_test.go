package tabkind

import (
	"strings"
	"testing"
)

func TestGetTabKind(t *testing.T) {
	cases := []struct {
		raw  string
		want any
	}{
		{raw: "request", want: Request},
		{raw: "example", want: Example},
		{raw: "environment", want: Environment},
		{raw: "globals", want: Globals},
		{raw: "collection", want: Collection},
		{raw: "folder", want: Folder},
	}
	for _, tc := range cases {
		for _, raw := range []string{tc.raw, strings.ToUpper(tc.raw), " " + tc.raw + " "} {
			if got := GetTabKind(raw); got != tc.want {
				t.Errorf("GetTabKind(%q) = %q, want %q", raw, got, tc.want)
			}
		}
	}
	if got := GetTabKind("not-a-value"); got != Unknown {
		t.Errorf("GetTabKind(unknown) = %q, want unknown", got)
	}
}

func TestTabKindEqualsAndString(t *testing.T) {
	value := Environment
	if !value.Equals(strings.ToUpper("environment")) || value.Equals("unknownx") {
		t.Errorf("Equals is not a case-insensitive comparison")
	}
	if value.String() != "environment" {
		t.Errorf("String() = %q", value.String())
	}
}
