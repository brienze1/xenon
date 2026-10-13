package changetype

import (
	"strings"

	"github.com/brienze1/xenon/internal/domain/enums"
)

const (
	Added    enums.ChangeType = "added"
	Modified enums.ChangeType = "modified"
	Deleted  enums.ChangeType = "deleted"
	Renamed  enums.ChangeType = "renamed"
	Unknown  enums.ChangeType = "unknown"
)

func GetChangeType(s string) enums.ChangeType {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "added":
		return Added
	case "modified":
		return Modified
	case "deleted":
		return Deleted
	case "renamed":
		return Renamed
	default:
		return Unknown
	}
}
