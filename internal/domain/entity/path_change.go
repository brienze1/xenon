package entity

import (
	"github.com/brienze1/xenon/internal/domain/enums"
	"github.com/brienze1/xenon/internal/domain/enums/changetype"
)

type PathChange struct {
	Path   string
	Change enums.ChangeType
}

func (p PathChange) IsDeletion() bool {
	return p.Change == changetype.Deleted
}
