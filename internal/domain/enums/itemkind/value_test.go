package itemkind

import (
	"strings"
	"testing"
)

func TestGetItemKind(t *testing.T) {
	cases := []struct {
		raw  string
		want any
	}{
		{raw: "workspace", want: Workspace},
		{raw: "collection", want: Collection},
		{raw: "folder", want: Folder},
		{raw: "request", want: Request},
		{raw: "error", want: Error},
	}
	for _, tc := range cases {
		for _, raw := range []string{tc.raw, strings.ToUpper(tc.raw), " " + tc.raw + " "} {
			if got := GetItemKind(raw); got != tc.want {
				t.Errorf("GetItemKind(%q) = %q, want %q", raw, got, tc.want)
			}
		}
	}
	if got := GetItemKind("not-a-value"); got != Unknown {
		t.Errorf("GetItemKind(unknown) = %q, want unknown", got)
	}
}

func TestItemKindEqualsAndString(t *testing.T) {
	value := Workspace
	if !value.Equals(strings.ToUpper("workspace")) || value.Equals("unknownx") {
		t.Errorf("Equals is not a case-insensitive comparison")
	}
	if value.String() != "workspace" {
		t.Errorf("String() = %q", value.String())
	}
}
