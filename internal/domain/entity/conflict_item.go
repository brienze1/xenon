package entity

import (
	"github.com/brienze1/xenon/internal/domain/enums"
	"github.com/brienze1/xenon/internal/domain/enums/conflictkind"
	"github.com/brienze1/xenon/internal/domain/enums/conflictresolution"
	"github.com/brienze1/xenon/internal/domain/enums/itemkind"
)

const theirCopySuffix = " (theirs)"

type ConflictItem struct {
	Versions    ConflictVersions
	ItemID      string
	ItemName    string
	ItemKind    enums.ItemKind
	Kind        enums.ConflictKind
	Resolution  enums.ConflictResolution
	Path        string
	Base        string
	Mine        string
	Theirs      string
	Fields      []FieldConflict
	Uncommitted bool
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

func NewUncommittedConflictItem(ref ItemRef, plan PathPlan) ConflictItem {
	item := NewConflictItem(ref, plan, nil)
	item.Uncommitted = true
	return item
}

func (c ConflictItem) WithVersions(versions ConflictVersions) ConflictItem {
	c.Versions = versions
	return c
}

func (c ConflictItem) Accepts(resolution enums.ConflictResolution, fieldPath string) bool {
	switch c.Kind {
	case conflictkind.Field:
		return c.hasField(fieldPath) && isSideChoice(resolution)
	case conflictkind.DeleteModify:
		return fieldPath == "" && (resolution == conflictresolution.Keep || resolution == conflictresolution.Delete)
	default:
		return fieldPath == "" && (isSideChoice(resolution) || (resolution == conflictresolution.Both && c.CanKeepBoth()))
	}
}

func (c ConflictItem) Resolved(resolution enums.ConflictResolution, fieldPath string) ConflictItem {
	if c.Kind != conflictkind.Field {
		c.Resolution = resolution
		return c
	}
	fields := make([]FieldConflict, 0, len(c.Fields))
	for _, field := range c.Fields {
		if field.Path == fieldPath {
			field.Resolution = resolution
		}
		fields = append(fields, field)
	}
	c.Fields = fields
	c.Resolution = conflictresolution.Unresolved
	if c.fieldsResolved() {
		c.Resolution = conflictresolution.Merged
	}
	return c
}

func (c ConflictItem) IsResolved() bool {
	return c.Resolution != conflictresolution.Unresolved
}

func (c ConflictItem) CanKeepBoth() bool {
	return c.ItemKind == itemkind.Request || c.ItemKind == itemkind.Environment
}

func (c ConflictItem) KeepsBoth() bool {
	return c.Resolution == conflictresolution.Both
}

func (c ConflictItem) IsFieldMerge() bool {
	return c.Kind == conflictkind.Field
}

// Result is the blob kept at the conflict's path; an empty result removes the file.
func (c ConflictItem) Result() string {
	switch c.Resolution {
	case conflictresolution.Theirs:
		return c.Theirs
	case conflictresolution.Delete:
		return ""
	case conflictresolution.Keep:
		return c.modifiedSide()
	default:
		return c.Mine
	}
}

func (c ConflictItem) CopyName() string {
	return c.ItemName + theirCopySuffix
}

func (c ConflictItem) FieldResolutions() map[string]enums.ConflictResolution {
	resolutions := make(map[string]enums.ConflictResolution, len(c.Fields))
	for _, field := range c.Fields {
		resolutions[field.Path] = field.Resolution
	}
	return resolutions
}

func (c ConflictItem) modifiedSide() string {
	if c.Mine != "" {
		return c.Mine
	}
	return c.Theirs
}

func (c ConflictItem) hasField(fieldPath string) bool {
	for _, field := range c.Fields {
		if field.Path == fieldPath {
			return true
		}
	}
	return false
}

func (c ConflictItem) fieldsResolved() bool {
	for _, field := range c.Fields {
		if field.Resolution == conflictresolution.Unresolved {
			return false
		}
	}
	return true
}

func isSideChoice(resolution enums.ConflictResolution) bool {
	return resolution == conflictresolution.Mine || resolution == conflictresolution.Theirs
}

func (c ConflictItem) WithFields(fields []FieldConflict) ConflictItem {
	c.Fields = fields
	return c
}
