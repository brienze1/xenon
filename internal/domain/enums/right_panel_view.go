package enums

import "strings"

type RightPanelView string

func (r RightPanelView) Equals(raw string) bool {
	return strings.EqualFold(string(r), raw)
}

func (r RightPanelView) String() string {
	return string(r)
}
