package httpmethod

import (
	"strings"
	"testing"
)

func TestGetHttpMethod(t *testing.T) {
	cases := []struct {
		raw  string
		want any
	}{
		{raw: "GET", want: Get},
		{raw: "POST", want: Post},
		{raw: "PUT", want: Put},
		{raw: "PATCH", want: Patch},
		{raw: "DELETE", want: Delete},
		{raw: "HEAD", want: Head},
		{raw: "OPTIONS", want: Options},
	}
	for _, tc := range cases {
		for _, raw := range []string{tc.raw, strings.ToLower(tc.raw), " " + tc.raw + " "} {
			if got := GetHttpMethod(raw); got != tc.want {
				t.Errorf("GetHttpMethod(%q) = %q, want %q", raw, got, tc.want)
			}
		}
	}
	if got := GetHttpMethod("not-a-value"); got != Unknown {
		t.Errorf("GetHttpMethod(unknown) = %q, want UNKNOWN", got)
	}
}

func TestHttpMethodEqualsAndString(t *testing.T) {
	value := Get
	if !value.Equals(strings.ToLower("GET")) || value.Equals("UNKNOWNx") {
		t.Errorf("Equals is not a case-insensitive comparison")
	}
	if value.String() != "GET" {
		t.Errorf("String() = %q", value.String())
	}
}
