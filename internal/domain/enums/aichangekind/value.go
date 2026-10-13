package aichangekind

import (
	"strings"

	"github.com/brienze1/xenon/internal/domain/enums"
)

const (
	Tree       enums.AIChangeKind = "tree"
	Request    enums.AIChangeKind = "request"
	Variables  enums.AIChangeKind = "variables"
	History    enums.AIChangeKind = "history"
	Cookies    enums.AIChangeKind = "cookies"
	Workspaces enums.AIChangeKind = "workspaces"
	Unknown    enums.AIChangeKind = "unknown"
)

func GetAIChangeKind(s string) enums.AIChangeKind {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "tree":
		return Tree
	case "request":
		return Request
	case "variables":
		return Variables
	case "history":
		return History
	case "cookies":
		return Cookies
	case "workspaces":
		return Workspaces
	default:
		return Unknown
	}
}
