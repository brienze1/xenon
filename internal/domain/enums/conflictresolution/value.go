package conflictresolution

import (
	"strings"

	"github.com/brienze1/xenon/internal/domain/enums"
)

const (
	Unresolved enums.ConflictResolution = "unresolved"
	Mine       enums.ConflictResolution = "mine"
	Theirs     enums.ConflictResolution = "theirs"
	Both       enums.ConflictResolution = "both"
	Keep       enums.ConflictResolution = "keep"
	Delete     enums.ConflictResolution = "delete"
	Unknown    enums.ConflictResolution = "unknown"
)

func GetConflictResolution(s string) enums.ConflictResolution {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "unresolved":
		return Unresolved
	case "mine":
		return Mine
	case "theirs":
		return Theirs
	case "both":
		return Both
	case "keep":
		return Keep
	case "delete":
		return Delete
	default:
		return Unknown
	}
}
