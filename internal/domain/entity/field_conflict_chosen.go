package entity

import "github.com/brienze1/xenon/internal/domain/enums"

// ChosenResolution is empty while the field is unresolved.
func (f FieldConflict) ChosenResolution() enums.ConflictResolution {
	return chosenResolution(f.Resolution)
}
