package sidebarview

import (
	"strings"
	"testing"
)

func TestGetSidebarView(t *testing.T) {
	cases := []struct {
		raw  string
		want any
	}{
		{raw: "collections", want: Collections},
		{raw: "history", want: History},
	}
	for _, tc := range cases {
		for _, raw := range []string{tc.raw, strings.ToUpper(tc.raw), " " + tc.raw + " "} {
			if got := GetSidebarView(raw); got != tc.want {
				t.Errorf("GetSidebarView(%q) = %q, want %q", raw, got, tc.want)
			}
		}
	}
	if got := GetSidebarView("not-a-value"); got != Unknown {
		t.Errorf("GetSidebarView(unknown) = %q, want unknown", got)
	}
}

func TestSidebarViewEqualsAndString(t *testing.T) {
	value := Collections
	if !value.Equals(strings.ToUpper("collections")) || value.Equals("unknownx") {
		t.Errorf("Equals is not a case-insensitive comparison")
	}
	if value.String() != "collections" {
		t.Errorf("String() = %q", value.String())
	}
}
