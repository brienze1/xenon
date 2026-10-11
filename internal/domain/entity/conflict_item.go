package entity

import "github.com/brienze1/xenon/internal/domain/enums"

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
