package rightpanelview

import (
	"strings"

	"github.com/brienze1/xenon/internal/domain/enums"
)

const (
	Info    enums.RightPanelView = "info"
	Unknown enums.RightPanelView = "unknown"
)

func GetRightPanelView(s string) enums.RightPanelView {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "info":
		return Info
	default:
		return Unknown
	}
}
