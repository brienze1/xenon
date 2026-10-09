package enums

import "strings"

type SidebarView string

func (s SidebarView) Equals(raw string) bool {
	return strings.EqualFold(string(s), raw)
}

func (s SidebarView) String() string {
	return string(s)
}
