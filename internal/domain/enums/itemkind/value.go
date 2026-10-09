package itemkind

import (
	"strings"

	"github.com/brienze1/xenon/internal/domain/enums"
)

const (
	Workspace  enums.ItemKind = "workspace"
	Collection enums.ItemKind = "collection"
	Folder     enums.ItemKind = "folder"
	Request    enums.ItemKind = "request"
	Error      enums.ItemKind = "error"
	Unknown    enums.ItemKind = "unknown"
)

func GetItemKind(s string) enums.ItemKind {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "workspace":
		return Workspace
	case "collection":
		return Collection
	case "folder":
		return Folder
	case "request":
		return Request
	case "error":
		return Error
	default:
		return Unknown
	}
}
