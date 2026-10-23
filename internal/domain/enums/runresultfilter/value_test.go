package runresultfilter

import (
	"strings"
	"testing"
)

func TestGetRunResultFilter(t *testing.T) {
	cases := []struct {
		raw  string
		want any
	}{
		{raw: "all", want: All},
		{raw: "passed", want: Passed},
		{raw: "failed", want: Failed},
		{raw: "skipped", want: Skipped},
	}
	for _, tc := range cases {
		for _, raw := range []string{tc.raw, strings.ToUpper(tc.raw), " " + tc.raw + " "} {
			if got := GetRunResultFilter(raw); got != tc.want {
				t.Errorf("GetRunResultFilter(%q) = %q, want %q", raw, got, tc.want)
			}
		}
	}
	if got := GetRunResultFilter("not-a-value"); got != Unknown {
		t.Errorf("GetRunResultFilter(unknown) = %q, want unknown", got)
	}
}
