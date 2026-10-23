package runtype

import (
	"strings"
	"testing"
)

func TestGetRunType(t *testing.T) {
	cases := []struct {
		raw  string
		want any
	}{
		{raw: "functional", want: Functional},
		{raw: "performance", want: Performance},
	}
	for _, tc := range cases {
		for _, raw := range []string{tc.raw, strings.ToUpper(tc.raw), " " + tc.raw + " "} {
			if got := GetRunType(raw); got != tc.want {
				t.Errorf("GetRunType(%q) = %q, want %q", raw, got, tc.want)
			}
		}
	}
	if got := GetRunType("not-a-value"); got != Unknown {
		t.Errorf("GetRunType(unknown) = %q, want unknown", got)
	}
}
