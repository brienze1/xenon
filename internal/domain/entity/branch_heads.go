package entity

import (
	"github.com/brienze1/xenon/internal/domain/enums"
	"github.com/brienze1/xenon/internal/domain/enums/divergence"
)

type BranchHeads struct {
	Local  string
	Remote string
	Base   string
	Ahead  int
	Behind int
}

func (b BranchHeads) HasLocalCommits() bool {
	return b.Local != ""
}

func (b BranchHeads) Divergence() enums.Divergence {
	switch {
	case b.Local == b.Remote:
		return divergence.UpToDate
	case b.Remote == "" || b.Base == b.Remote:
		return divergence.Ahead
	case b.Local == "" || b.Base == b.Local:
		return divergence.FastForward
	default:
		return divergence.Diverged
	}
}

func (b BranchHeads) HasNothingToPull() bool {
	state := b.Divergence()
	return state == divergence.UpToDate || state == divergence.Ahead
}

func (b BranchHeads) IsFastForward() bool {
	return b.Divergence() == divergence.FastForward
}

func (b BranchHeads) IsDiverged() bool {
	return b.Divergence() == divergence.Diverged
}
