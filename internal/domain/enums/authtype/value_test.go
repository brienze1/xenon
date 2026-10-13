package authtype

import (
	"strings"
	"testing"
)

func TestGetAuthType(t *testing.T) {
	cases := []struct {
		raw  string
		want any
	}{
		{raw: "inherit", want: Inherit},
		{raw: "none", want: None},
		{raw: "bearer", want: Bearer},
		{raw: "basic", want: Basic},
		{raw: "apikey", want: ApiKey},
		{raw: "oauth2", want: OAuth2},
		{raw: "awsv4", want: AwsV4},
		{raw: "digest", want: Digest},
	}
	for _, tc := range cases {
		for _, raw := range []string{tc.raw, strings.ToUpper(tc.raw), " " + tc.raw + " "} {
			if got := GetAuthType(raw); got != tc.want {
				t.Errorf("GetAuthType(%q) = %q, want %q", raw, got, tc.want)
			}
		}
	}
	if got := GetAuthType("not-a-value"); got != Unknown {
		t.Errorf("GetAuthType(unknown) = %q, want unknown", got)
	}
}

func TestAuthTypeEqualsAndString(t *testing.T) {
	value := Inherit
	if !value.Equals(strings.ToUpper("inherit")) || value.Equals("unknownx") {
		t.Errorf("Equals is not a case-insensitive comparison")
	}
	if value.String() != "inherit" {
		t.Errorf("String() = %q", value.String())
	}
}
