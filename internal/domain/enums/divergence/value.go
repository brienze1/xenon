package divergence

import (
	"strings"

	"github.com/brienze1/xenon/internal/domain/enums"
)

const (
	UpToDate    enums.Divergence = "up_to_date"
	FastForward enums.Divergence = "fast_forward"
	Ahead       enums.Divergence = "ahead"
	Diverged    enums.Divergence = "diverged"
	Unknown     enums.Divergence = "unknown"
)

func GetDivergence(s string) enums.Divergence {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "up_to_date":
		return UpToDate
	case "fast_forward":
		return FastForward
	case "ahead":
		return Ahead
	case "diverged":
		return Diverged
	default:
		return Unknown
	}
}
