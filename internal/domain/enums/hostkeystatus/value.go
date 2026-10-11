package hostkeystatus

import (
	"strings"

	"github.com/brienze1/xenon/internal/domain/enums"
)

const (
	Trusted      enums.HostKeyStatus = "trusted"
	Unrecognized enums.HostKeyStatus = "unrecognized"
	Changed      enums.HostKeyStatus = "changed"
	Unknown      enums.HostKeyStatus = "unknown"
)

func GetHostKeyStatus(s string) enums.HostKeyStatus {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "trusted":
		return Trusted
	case "unrecognized":
		return Unrecognized
	case "changed":
		return Changed
	default:
		return Unknown
	}
}
