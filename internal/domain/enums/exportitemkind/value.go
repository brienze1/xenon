package exportitemkind

import (
	"strings"

	"github.com/brienze1/xenon/internal/domain/enums"
)

const (
	Workspace   enums.ExportItemKind = "workspace"
	Collection  enums.ExportItemKind = "collection"
	Folder      enums.ExportItemKind = "folder"
	Request     enums.ExportItemKind = "request"
	Environment enums.ExportItemKind = "environment"
	Globals     enums.ExportItemKind = "globals"
	Unknown     enums.ExportItemKind = "unknown"
)

func GetExportItemKind(s string) enums.ExportItemKind {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "workspace":
		return Workspace
	case "collection":
		return Collection
	case "folder":
		return Folder
	case "request":
		return Request
	case "environment":
		return Environment
	case "globals":
		return Globals
	default:
		return Unknown
	}
}
