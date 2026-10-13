package linkmode

import (
	"strings"

	"github.com/brienze1/xenon/internal/domain/enums"
)

const (
	Unspecified     enums.LinkMode = ""
	AdoptRepo       enums.LinkMode = "adopt_repo"
	ReplaceWithMine enums.LinkMode = "replace_with_mine"
	PlaceMine       enums.LinkMode = "place_mine"
	Unknown         enums.LinkMode = "unknown"
)

func GetLinkMode(s string) enums.LinkMode {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "":
		return Unspecified
	case "adopt_repo":
		return AdoptRepo
	case "replace_with_mine":
		return ReplaceWithMine
	case "place_mine":
		return PlaceMine
	default:
		return Unknown
	}
}
