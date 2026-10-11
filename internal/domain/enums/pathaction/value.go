package pathaction

import (
	"strings"

	"github.com/brienze1/xenon/internal/domain/enums"
)

const (
	Keep            enums.PathAction = "keep"
	TakeMine        enums.PathAction = "take_mine"
	TakeTheirs      enums.PathAction = "take_theirs"
	MergeStructural enums.PathAction = "merge_structural"
	BothModified    enums.PathAction = "both_modified"
	DeleteModify    enums.PathAction = "delete_modify"
	Unknown         enums.PathAction = "unknown"
)

func GetPathAction(s string) enums.PathAction {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "keep":
		return Keep
	case "take_mine":
		return TakeMine
	case "take_theirs":
		return TakeTheirs
	case "merge_structural":
		return MergeStructural
	case "both_modified":
		return BothModified
	case "delete_modify":
		return DeleteModify
	default:
		return Unknown
	}
}
