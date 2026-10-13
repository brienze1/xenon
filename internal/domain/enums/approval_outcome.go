package enums

import "strings"

type ApprovalOutcome string

func (v ApprovalOutcome) Equals(raw string) bool {
	return strings.EqualFold(string(v), raw)
}

func (v ApprovalOutcome) String() string {
	return string(v)
}
