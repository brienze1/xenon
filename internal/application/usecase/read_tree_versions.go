package usecase

import (
	"context"

	"github.com/brienze1/xenon/internal/domain/entity"
)

type ReadTreeVersions interface {
	ReadVersions(ctx context.Context, clonePath, subfolder string, heads entity.BranchHeads) (*entity.TreeVersions, error)
	ReadWorktree(ctx context.Context, clonePath, subfolder string) (map[string]string, error)
	ReadWorktreeFiles(ctx context.Context, clonePath, subfolder string, paths []string) (map[string][]byte, error)
	ReadBlob(ctx context.Context, clonePath, hash string) ([]byte, error)
}
