package entity

import "github.com/brienze1/xenon/internal/domain/enums"

type FieldConflict struct {
	Mine       *DocumentNode
	Theirs     *DocumentNode
	Path       string
	Resolution enums.ConflictResolution
}
