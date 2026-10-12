package entity

import "github.com/brienze1/xenon/internal/domain/enums"

type FieldConflict struct {
	Mine       *DocumentNode
	Theirs     *DocumentNode
	MineText   *string
	TheirsText *string
	Path       string
	Resolution enums.ConflictResolution
}

func (f FieldConflict) WithTexts(mine, theirs *string) FieldConflict {
	f.MineText = mine
	f.TheirsText = theirs
	return f
}
