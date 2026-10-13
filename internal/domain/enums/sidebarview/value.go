package sidebarview

import (
	"strings"

	"github.com/brienze1/xenon/internal/domain/enums"
)

const (
	Collections  enums.SidebarView = "collections"
	History      enums.SidebarView = "history"
	Environments enums.SidebarView = "environments"
	Unknown      enums.SidebarView = "unknown"
)

func GetSidebarView(s string) enums.SidebarView {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "collections":
		return Collections
	case "history":
		return History
	case "environments":
		return Environments
	default:
		return Unknown
	}
}
