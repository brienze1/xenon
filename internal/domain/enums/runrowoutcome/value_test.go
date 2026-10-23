package runrowoutcome

import (
	"strings"
	"testing"
)

func TestGetRunRowOutcome(t *testing.T) {
	cases := []struct {
		raw  string
		want any
	}{
		{raw: "passed", want: Passed},
		{raw: "failed", want: Failed},
		{raw: "skipped", want: Skipped},
		{raw: "error", want: Error},
		{raw: "cancelled", want: Cancelled},
	}
	for _, tc := range cases {
		for _, raw := range []string{tc.raw, strings.ToUpper(tc.raw), " " + tc.raw + " "} {
			if got := GetRunRowOutcome(raw); got != tc.want {
				t.Errorf("GetRunRowOutcome(%q) = %q, want %q", raw, got, tc.want)
			}
		}
	}
	if got := GetRunRowOutcome("not-a-value"); got != Unknown {
		t.Errorf("GetRunRowOutcome(unknown) = %q, want unknown", got)
	}
}
