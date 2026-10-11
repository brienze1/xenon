package usecase

import (
	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/internal/domain/enums"
	"github.com/brienze1/xenon/internal/domain/enums/divergence"
)

func ClassifyDivergence(heads entity.BranchHeads) enums.Divergence {
	switch {
	case heads.Local == heads.Remote:
		return divergence.UpToDate
	case heads.Remote == "" || heads.Base == heads.Remote:
		return divergence.Ahead
	case heads.Local == "" || heads.Base == heads.Local:
		return divergence.FastForward
	default:
		return divergence.Diverged
	}
}
