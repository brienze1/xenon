package senderrorkind

import (
	"strings"
	"testing"
)

func TestGetSendErrorKind(t *testing.T) {
	cases := []struct {
		raw  string
		want any
	}{
		{raw: "dns", want: Dns},
		{raw: "connection_refused", want: ConnectionRefused},
		{raw: "timeout", want: Timeout},
		{raw: "tls", want: Tls},
		{raw: "cancelled", want: Cancelled},
		{raw: "missing_file", want: MissingFile},
		{raw: "invalid_url", want: InvalidUrl},
	}
	for _, tc := range cases {
		for _, raw := range []string{tc.raw, strings.ToUpper(tc.raw), " " + tc.raw + " "} {
			if got := GetSendErrorKind(raw); got != tc.want {
				t.Errorf("GetSendErrorKind(%q) = %q, want %q", raw, got, tc.want)
			}
		}
	}
	if got := GetSendErrorKind("not-a-value"); got != Other {
		t.Errorf("GetSendErrorKind(unknown) = %q, want other", got)
	}
}

func TestSendErrorKindEqualsAndString(t *testing.T) {
	value := Dns
	if !value.Equals(strings.ToUpper("dns")) || value.Equals("otherx") {
		t.Errorf("Equals is not a case-insensitive comparison")
	}
	if value.String() != "dns" {
		t.Errorf("String() = %q", value.String())
	}
}
