package importitemkind

import (
	"strings"

	"github.com/brienze1/xenon/internal/domain/enums"
)

const (
	Collection  enums.ImportItemKind = "collection"
	Environment enums.ImportItemKind = "environment"
	Globals     enums.ImportItemKind = "globals"
	Request     enums.ImportItemKind = "request"
	Unknown     enums.ImportItemKind = "unknown"
)

func GetImportItemKind(s string) enums.ImportItemKind {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "collection":
		return Collection
	case "environment":
		return Environment
	case "globals":
		return Globals
	case "request":
		return Request
	default:
		return Unknown
	}
}
