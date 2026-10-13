package usecase

import (
	"context"

	"github.com/brienze1/xenon/internal/domain/entity"
)

type PushBranch interface {
	Push(ctx context.Context, access entity.GitAccess, clonePath, branch, expectedRemote string) error
}
