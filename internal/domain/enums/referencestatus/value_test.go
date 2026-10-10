package referencestatus

import (
	"strings"
	"testing"
)

func TestGetReferenceStatus(t *testing.T) {
	cases := []struct {
		raw  string
		want any
	}{
		{raw: "resolved", want: Resolved},
		{raw: "unresolved", want: Unresolved},
		{raw: "missing_secret", want: MissingSecret},
		{raw: "locked_secret", want: LockedSecret},
		{raw: "dynamic", want: Dynamic},
	}
	for _, tc := range cases {
		for _, raw := range []string{tc.raw, strings.ToUpper(tc.raw), " " + tc.raw + " "} {
			if got := GetReferenceStatus(raw); got != tc.want {
				t.Errorf("GetReferenceStatus(%q) = %q, want %q", raw, got, tc.want)
			}
		}
	}
	if got := GetReferenceStatus("not-a-value"); got != Unknown {
		t.Errorf("GetReferenceStatus(unknown) = %q, want unknown", got)
	}
}

func TestReferenceStatusEqualsAndString(t *testing.T) {
	value := Resolved
	if !value.Equals(strings.ToUpper("resolved")) || value.Equals("resolvedx") {
		t.Errorf("Equals is not a case-insensitive comparison")
	}
	if value.String() != "resolved" {
		t.Errorf("String() = %q", value.String())
	}
}
