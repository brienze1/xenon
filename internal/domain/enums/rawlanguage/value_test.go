package rawlanguage

import (
	"strings"
	"testing"
)

func TestGetRawLanguage(t *testing.T) {
	cases := []struct {
		raw  string
		want any
	}{
		{raw: "json", want: Json},
		{raw: "xml", want: Xml},
		{raw: "text", want: Text},
		{raw: "html", want: Html},
		{raw: "javascript", want: JavaScript},
	}
	for _, tc := range cases {
		for _, raw := range []string{tc.raw, strings.ToUpper(tc.raw), " " + tc.raw + " "} {
			if got := GetRawLanguage(raw); got != tc.want {
				t.Errorf("GetRawLanguage(%q) = %q, want %q", raw, got, tc.want)
			}
		}
	}
	if got := GetRawLanguage("not-a-value"); got != Unknown {
		t.Errorf("GetRawLanguage(unknown) = %q, want unknown", got)
	}
}

func TestRawLanguageEqualsAndString(t *testing.T) {
	value := Json
	if !value.Equals(strings.ToUpper("json")) || value.Equals("unknownx") {
		t.Errorf("Equals is not a case-insensitive comparison")
	}
	if value.String() != "json" {
		t.Errorf("String() = %q", value.String())
	}
}
