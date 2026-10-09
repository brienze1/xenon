package splitorientation

import (
	"strings"

	"github.com/brienze1/xenon/internal/domain/enums"
)

const (
	Stacked    enums.SplitOrientation = "stacked"
	SideBySide enums.SplitOrientation = "side-by-side"
	Unknown    enums.SplitOrientation = "unknown"
)

func GetSplitOrientation(s string) enums.SplitOrientation {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "stacked":
		return Stacked
	case "side-by-side":
		return SideBySide
	default:
		return Unknown
	}
}
