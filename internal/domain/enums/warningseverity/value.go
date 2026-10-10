package warningseverity

import (
	"strings"

	"github.com/brienze1/xenon/internal/domain/enums"
)

const (
	Info    enums.WarningSeverity = "info"
	Warning enums.WarningSeverity = "warning"
	Unknown enums.WarningSeverity = "unknown"
)

func GetWarningSeverity(s string) enums.WarningSeverity {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "info":
		return Info
	case "warning":
		return Warning
	default:
		return Unknown
	}
}
