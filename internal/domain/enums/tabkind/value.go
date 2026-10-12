package tabkind

import (
	"strings"

	"github.com/brienze1/xenon/internal/domain/enums"
)

const (
	Request     enums.TabKind = "request"
	Example     enums.TabKind = "example"
	Environment enums.TabKind = "environment"
	Globals     enums.TabKind = "globals"
	Unknown     enums.TabKind = "unknown"
)

func GetTabKind(s string) enums.TabKind {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "request":
		return Request
	case "example":
		return Example
	case "environment":
		return Environment
	case "globals":
		return Globals
	default:
		return Unknown
	}
}
