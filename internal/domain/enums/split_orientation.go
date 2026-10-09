package enums

import "strings"

type SplitOrientation string

func (s SplitOrientation) Equals(raw string) bool {
	return strings.EqualFold(string(s), raw)
}

func (s SplitOrientation) String() string {
	return string(s)
}
