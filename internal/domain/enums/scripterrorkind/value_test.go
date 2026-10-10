package scripterrorkind

import (
	"strings"
	"testing"
)

func TestGetScriptErrorKind(t *testing.T) {
	cases := []struct {
		raw  string
		want any
	}{
		{raw: "script_syntax", want: Syntax},
		{raw: "script_runtime", want: Runtime},
		{raw: "script_timeout", want: Timeout},
		{raw: "script_memory", want: Memory},
		{raw: "script_stack_overflow", want: StackOverflow},
		{raw: "script_unhandled_rejection", want: UnhandledRejection},
		{raw: "script_unsupported_api", want: UnsupportedApi},
	}
	for _, tc := range cases {
		for _, raw := range []string{tc.raw, strings.ToUpper(tc.raw), " " + tc.raw + " "} {
			if got := GetScriptErrorKind(raw); got != tc.want {
				t.Errorf("GetScriptErrorKind(%q) = %q, want %q", raw, got, tc.want)
			}
		}
	}
	if got := GetScriptErrorKind("not-a-value"); got != Unknown {
		t.Errorf("GetScriptErrorKind(unknown) = %q, want unknown", got)
	}
}

func TestScriptErrorKindEqualsAndString(t *testing.T) {
	value := Syntax
	if !value.Equals(strings.ToUpper("script_syntax")) || value.Equals("script_syntaxx") {
		t.Errorf("Equals is not a case-insensitive comparison")
	}
	if value.String() != "script_syntax" {
		t.Errorf("String() = %q", value.String())
	}
}
