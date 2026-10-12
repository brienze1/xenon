package entity

import (
	"github.com/brienze1/xenon/internal/domain/enums"
	"github.com/brienze1/xenon/internal/domain/enums/conflictkind"
	"github.com/brienze1/xenon/internal/domain/enums/conflictresolution"
)

type ConflictItem struct {
	ItemID     string
	ItemName   string
	ItemKind   enums.ItemKind
	Kind       enums.ConflictKind
	Resolution enums.ConflictResolution
	Path       string
	Base       string
	Mine       string
	Theirs     string
	Fields     []FieldConflict
}

func NewConflictItem(ref ItemRef, plan PathPlan, fields []FieldConflict) ConflictItem {
	kind := conflictkind.BothModified
	switch {
	case len(fields) > 0:
		kind = conflictkind.Field
	case plan.Mine == "" || plan.Theirs == "":
		kind = conflictkind.DeleteModify
	}
	return ConflictItem{
		ItemID:     ref.ID,
		ItemName:   ref.Name,
		ItemKind:   ref.Kind,
		Kind:       kind,
		Resolution: conflictresolution.Unresolved,
		Path:       plan.Path,
		Base:       plan.Base,
		Mine:       plan.Mine,
		Theirs:     plan.Theirs,
		Fields:     fields,
	}
}
