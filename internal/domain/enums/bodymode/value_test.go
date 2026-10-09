package bodymode

import (
	"strings"
	"testing"
)

func TestGetBodyMode(t *testing.T) {
	cases := []struct {
		raw  string
		want any
	}{
		{raw: "none", want: None},
		{raw: "formdata", want: FormData},
		{raw: "urlencoded", want: UrlEncoded},
		{raw: "raw", want: Raw},
		{raw: "binary", want: Binary},
		{raw: "graphql", want: GraphQl},
	}
	for _, tc := range cases {
		for _, raw := range []string{tc.raw, strings.ToUpper(tc.raw), " " + tc.raw + " "} {
			if got := GetBodyMode(raw); got != tc.want {
				t.Errorf("GetBodyMode(%q) = %q, want %q", raw, got, tc.want)
			}
		}
	}
	if got := GetBodyMode("not-a-value"); got != Unknown {
		t.Errorf("GetBodyMode(unknown) = %q, want unknown", got)
	}
}

func TestBodyModeEqualsAndString(t *testing.T) {
	value := None
	if !value.Equals(strings.ToUpper("none")) || value.Equals("unknownx") {
		t.Errorf("Equals is not a case-insensitive comparison")
	}
	if value.String() != "none" {
		t.Errorf("String() = %q", value.String())
	}
}
