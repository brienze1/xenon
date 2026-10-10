package scriptownerkind

import (
	"strings"

	"github.com/brienze1/xenon/internal/domain/enums"
)

const (
	Collection enums.ScriptOwnerKind = "collection"
	Folder     enums.ScriptOwnerKind = "folder"
	Request    enums.ScriptOwnerKind = "request"
	Unknown    enums.ScriptOwnerKind = "unknown"
)

func GetScriptOwnerKind(s string) enums.ScriptOwnerKind {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "collection":
		return Collection
	case "folder":
		return Folder
	case "request":
		return Request
	default:
		return Unknown
	}
}
