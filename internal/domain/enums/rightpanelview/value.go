package rightpanelview

import (
	"strings"

	"github.com/brienze1/xenon/internal/domain/enums"
)

const (
	Info      enums.RightPanelView = "info"
	Variables enums.RightPanelView = "variables"
	Code      enums.RightPanelView = "code"
	Unknown   enums.RightPanelView = "unknown"
)

func GetRightPanelView(s string) enums.RightPanelView {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "info":
		return Info
	case "variables":
		return Variables
	case "code":
		return Code
	default:
		return Unknown
	}
}
