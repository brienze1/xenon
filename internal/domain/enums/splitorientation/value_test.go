package splitorientation

import (
	"strings"
	"testing"
)

func TestGetSplitOrientation(t *testing.T) {
	cases := []struct {
		raw  string
		want any
	}{
		{raw: "stacked", want: Stacked},
		{raw: "side-by-side", want: SideBySide},
	}
	for _, tc := range cases {
		for _, raw := range []string{tc.raw, strings.ToUpper(tc.raw), " " + tc.raw + " "} {
			if got := GetSplitOrientation(raw); got != tc.want {
				t.Errorf("GetSplitOrientation(%q) = %q, want %q", raw, got, tc.want)
			}
		}
	}
	if got := GetSplitOrientation("not-a-value"); got != Unknown {
		t.Errorf("GetSplitOrientation(unknown) = %q, want unknown", got)
	}
}

func TestSplitOrientationEqualsAndString(t *testing.T) {
	value := Stacked
	if !value.Equals(strings.ToUpper("stacked")) || value.Equals("unknownx") {
		t.Errorf("Equals is not a case-insensitive comparison")
	}
	if value.String() != "stacked" {
		t.Errorf("String() = %q", value.String())
	}
}
