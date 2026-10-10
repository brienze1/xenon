package referencestatus

import (
	"strings"

	"github.com/brienze1/xenon/internal/domain/enums"
)

const (
	Resolved      enums.ReferenceStatus = "resolved"
	Unresolved    enums.ReferenceStatus = "unresolved"
	MissingSecret enums.ReferenceStatus = "missing_secret"
	LockedSecret  enums.ReferenceStatus = "locked_secret"
	Dynamic       enums.ReferenceStatus = "dynamic"
	Unknown       enums.ReferenceStatus = "unknown"
)

func GetReferenceStatus(s string) enums.ReferenceStatus {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "resolved":
		return Resolved
	case "unresolved":
		return Unresolved
	case "missing_secret":
		return MissingSecret
	case "locked_secret":
		return LockedSecret
	case "dynamic":
		return Dynamic
	default:
		return Unknown
	}
}
