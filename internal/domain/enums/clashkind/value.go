package clashkind

import (
	"strings"

	"github.com/brienze1/xenon/internal/domain/enums"
)

const (
	Collection  enums.ClashKind = "collection"
	Folder      enums.ClashKind = "folder"
	Request     enums.ClashKind = "request"
	Environment enums.ClashKind = "environment"
	Variable    enums.ClashKind = "variable"
	Unknown     enums.ClashKind = "unknown"
)

func GetClashKind(s string) enums.ClashKind {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "collection":
		return Collection
	case "folder":
		return Folder
	case "request":
		return Request
	case "environment":
		return Environment
	case "variable":
		return Variable
	default:
		return Unknown
	}
}
