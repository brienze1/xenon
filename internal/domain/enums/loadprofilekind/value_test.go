package loadprofilekind

import (
	"strings"
	"testing"
)

func TestGetLoadProfileKind(t *testing.T) {
	cases := []struct {
		raw  string
		want any
	}{
		{raw: "fixed", want: Fixed},
		{raw: "ramp_up", want: RampUp},
		{raw: "spike", want: Spike},
		{raw: "peak", want: Peak},
	}
	for _, tc := range cases {
		for _, raw := range []string{tc.raw, strings.ToUpper(tc.raw), " " + tc.raw + " "} {
			if got := GetLoadProfileKind(raw); got != tc.want {
				t.Errorf("GetLoadProfileKind(%q) = %q, want %q", raw, got, tc.want)
			}
		}
	}
	if got := GetLoadProfileKind("not-a-value"); got != Unknown {
		t.Errorf("GetLoadProfileKind(unknown) = %q, want unknown", got)
	}
}
