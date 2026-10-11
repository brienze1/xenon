package conflictkind

import (
	"strings"

	"github.com/brienze1/xenon/internal/domain/enums"
)

const (
	BothModified enums.ConflictKind = "both_modified"
	DeleteModify enums.ConflictKind = "delete_modify"
	Field        enums.ConflictKind = "field"
	Unknown      enums.ConflictKind = "unknown"
)

func GetConflictKind(s string) enums.ConflictKind {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "both_modified":
		return BothModified
	case "delete_modify":
		return DeleteModify
	case "field":
		return Field
	default:
		return Unknown
	}
}
