package runstatus

import (
	"strings"
	"testing"
)

func TestGetRunStatus(t *testing.T) {
	cases := []struct {
		raw  string
		want any
	}{
		{raw: "running", want: Running},
		{raw: "paused", want: Paused},
		{raw: "completed", want: Completed},
		{raw: "stopped", want: Stopped},
		{raw: "stopped_on_error", want: StoppedOnError},
	}
	for _, tc := range cases {
		for _, raw := range []string{tc.raw, strings.ToUpper(tc.raw), " " + tc.raw + " "} {
			if got := GetRunStatus(raw); got != tc.want {
				t.Errorf("GetRunStatus(%q) = %q, want %q", raw, got, tc.want)
			}
		}
	}
	if got := GetRunStatus("not-a-value"); got != Unknown {
		t.Errorf("GetRunStatus(unknown) = %q, want unknown", got)
	}
}
