package runexportformat

import (
	"strings"
	"testing"
)

func TestGetRunExportFormat(t *testing.T) {
	cases := []struct {
		raw  string
		want any
	}{
		{raw: "json", want: Json},
		{raw: "junit", want: Junit},
	}
	for _, tc := range cases {
		for _, raw := range []string{tc.raw, strings.ToUpper(tc.raw), " " + tc.raw + " "} {
			if got := GetRunExportFormat(raw); got != tc.want {
				t.Errorf("GetRunExportFormat(%q) = %q, want %q", raw, got, tc.want)
			}
		}
	}
	if got := GetRunExportFormat("not-a-value"); got != Unknown {
		t.Errorf("GetRunExportFormat(unknown) = %q, want unknown", got)
	}
}
