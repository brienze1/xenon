package teststatus

import (
	"strings"
	"testing"
)

func TestGetTestStatus(t *testing.T) {
	cases := []struct {
		raw  string
		want any
	}{
		{raw: "passed", want: Passed},
		{raw: "failed", want: Failed},
		{raw: "skipped", want: Skipped},
	}
	for _, tc := range cases {
		for _, raw := range []string{tc.raw, strings.ToUpper(tc.raw), " " + tc.raw + " "} {
			if got := GetTestStatus(raw); got != tc.want {
				t.Errorf("GetTestStatus(%q) = %q, want %q", raw, got, tc.want)
			}
		}
	}
	if got := GetTestStatus("not-a-value"); got != Unknown {
		t.Errorf("GetTestStatus(unknown) = %q, want unknown", got)
	}
}

func TestTestStatusEqualsAndString(t *testing.T) {
	value := Passed
	if !value.Equals(strings.ToUpper("passed")) || value.Equals("passedx") {
		t.Errorf("Equals is not a case-insensitive comparison")
	}
	if value.String() != "passed" {
		t.Errorf("String() = %q", value.String())
	}
}
