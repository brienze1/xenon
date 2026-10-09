package fileutils

import (
	"strings"
	"testing"
)

func TestSafeName(t *testing.T) {
	cases := []struct {
		name string
		want string
	}{
		{name: "Create invoice", want: "create-invoice"},
		{name: "GET /users/:id?", want: "get-users-id"},
		{name: `a<b>c:d"e/f\g|h?i*j`, want: "a-b-c-d-e-f-g-h-i-j"},
		{name: "tab\there\nnew   line", want: "tab-here-new-line"},
		{name: "  ..--Trim me--..  ", want: "trim-me"},
		{name: "Café", want: "café"},
		{name: "ÉCOLE", want: "école"},
		{name: "", want: "untitled"},
		{name: "???", want: "untitled"},
		{name: "CON", want: "con_"},
		{name: "lpt9", want: "lpt9_"},
		{name: "aux.backup", want: "aux_.backup"},
		{name: "console", want: "console"},
		{name: "v1.2", want: "v1.2"},
		{name: strings.Repeat("é", 70), want: strings.Repeat("é", 64)},
		{name: strings.Repeat("a", 63) + " b", want: strings.Repeat("a", 63)},
		{name: strings.Repeat("a", 64), want: strings.Repeat("a", 64)},
	}
	for _, tc := range cases {
		if got := SafeName(tc.name); got != tc.want {
			t.Errorf("SafeName(%q) = %q, want %q", tc.name, got, tc.want)
		}
	}
}
