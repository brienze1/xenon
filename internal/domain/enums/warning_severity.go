package enums

import "strings"

type WarningSeverity string

func (w WarningSeverity) Equals(s string) bool {
	return strings.EqualFold(string(w), s)
}

func (w WarningSeverity) String() string {
	return string(w)
}
