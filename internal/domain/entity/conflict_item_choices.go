package entity

import (
	"github.com/brienze1/xenon/internal/domain/enums"
	"github.com/brienze1/xenon/internal/domain/enums/conflictkind"
	"github.com/brienze1/xenon/internal/domain/enums/conflictresolution"
)

// Choices lists the resolutions offered for the conflict's kind; Accepts still decides whether a choice applies.
func (c ConflictItem) Choices() []enums.ConflictResolution {
	switch c.Kind {
	case conflictkind.Field:
		return []enums.ConflictResolution{conflictresolution.Mine, conflictresolution.Theirs}
	case conflictkind.DeleteModify:
		return []enums.ConflictResolution{conflictresolution.Keep, conflictresolution.Delete}
	default:
		return []enums.ConflictResolution{conflictresolution.Mine, conflictresolution.Theirs, conflictresolution.Both}
	}
}

// ChosenResolution is empty while the conflict is unresolved.
func (c ConflictItem) ChosenResolution() enums.ConflictResolution {
	return chosenResolution(c.Resolution)
}

func chosenResolution(resolution enums.ConflictResolution) enums.ConflictResolution {
	if resolution == conflictresolution.Unresolved {
		return ""
	}
	return resolution
}
