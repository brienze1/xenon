package usecase

import (
	"context"

	"github.com/brienze1/xenon/internal/domain/entity"
)

type FetchBranch interface {
	Fetch(ctx context.Context, access entity.GitAccess, clonePath, branch string) (*entity.BranchHeads, error)
}
