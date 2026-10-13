package usecase

import (
	"context"

	"github.com/brienze1/xenon/internal/domain/entity"
)

type ReadBranchHeads interface {
	ReadHeads(ctx context.Context, clonePath, branch string) (*entity.BranchHeads, error)
}
