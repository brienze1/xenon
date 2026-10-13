package scriptownerkind

import (
	"strings"
	"testing"
)

func TestGetScriptOwnerKind(t *testing.T) {
	cases := []struct {
		raw  string
		want any
	}{
		{raw: "collection", want: Collection},
		{raw: "folder", want: Folder},
		{raw: "request", want: Request},
	}
	for _, tc := range cases {
		for _, raw := range []string{tc.raw, strings.ToUpper(tc.raw), " " + tc.raw + " "} {
			if got := GetScriptOwnerKind(raw); got != tc.want {
				t.Errorf("GetScriptOwnerKind(%q) = %q, want %q", raw, got, tc.want)
			}
		}
	}
	if got := GetScriptOwnerKind("not-a-value"); got != Unknown {
		t.Errorf("GetScriptOwnerKind(unknown) = %q, want unknown", got)
	}
}

func TestScriptOwnerKindEqualsAndString(t *testing.T) {
	value := Collection
	if !value.Equals(strings.ToUpper("collection")) || value.Equals("collectionx") {
		t.Errorf("Equals is not a case-insensitive comparison")
	}
	if value.String() != "collection" {
		t.Errorf("String() = %q", value.String())
	}
}
