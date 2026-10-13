package scriptrunstatus

import (
	"strings"
	"testing"
)

func TestGetScriptRunStatus(t *testing.T) {
	cases := []struct {
		raw  string
		want any
	}{
		{raw: "ok", want: Ok},
		{raw: "error", want: Error},
		{raw: "skipped", want: Skipped},
	}
	for _, tc := range cases {
		for _, raw := range []string{tc.raw, strings.ToUpper(tc.raw), " " + tc.raw + " "} {
			if got := GetScriptRunStatus(raw); got != tc.want {
				t.Errorf("GetScriptRunStatus(%q) = %q, want %q", raw, got, tc.want)
			}
		}
	}
	if got := GetScriptRunStatus("not-a-value"); got != Unknown {
		t.Errorf("GetScriptRunStatus(unknown) = %q, want unknown", got)
	}
}

func TestScriptRunStatusEqualsAndString(t *testing.T) {
	value := Ok
	if !value.Equals(strings.ToUpper("ok")) || value.Equals("okx") {
		t.Errorf("Equals is not a case-insensitive comparison")
	}
	if value.String() != "ok" {
		t.Errorf("String() = %q", value.String())
	}
}
