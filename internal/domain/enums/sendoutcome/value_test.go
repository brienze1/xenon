package sendoutcome

import (
	"strings"
	"testing"
)

func TestGetSendOutcome(t *testing.T) {
	cases := []struct {
		raw  string
		want any
	}{
		{raw: "sent", want: Sent},
		{raw: "skipped_by_script", want: SkippedByScript},
		{raw: "stopped_by_script_error", want: StoppedByScriptError},
		{raw: "cancelled", want: Cancelled},
		{raw: "error", want: Error},
	}
	for _, tc := range cases {
		for _, raw := range []string{tc.raw, strings.ToUpper(tc.raw), " " + tc.raw + " "} {
			if got := GetSendOutcome(raw); got != tc.want {
				t.Errorf("GetSendOutcome(%q) = %q, want %q", raw, got, tc.want)
			}
		}
	}
	if got := GetSendOutcome("not-a-value"); got != Unknown {
		t.Errorf("GetSendOutcome(unknown) = %q, want unknown", got)
	}
}

func TestSendOutcomeEqualsAndString(t *testing.T) {
	value := Sent
	if !value.Equals(strings.ToUpper("sent")) || value.Equals("sentx") {
		t.Errorf("Equals is not a case-insensitive comparison")
	}
	if value.String() != "sent" {
		t.Errorf("String() = %q", value.String())
	}
}
