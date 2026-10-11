package dropposition

import (
	"strings"

	"github.com/brienze1/xenon/internal/domain/enums"
)

const (
	Before  enums.DropPosition = "before"
	Inside  enums.DropPosition = "inside"
	After   enums.DropPosition = "after"
	Unknown enums.DropPosition = "unknown"
)

func GetDropPosition(s string) enums.DropPosition {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "before":
		return Before
	case "inside":
		return Inside
	case "after":
		return After
	default:
		return Unknown
	}
}
