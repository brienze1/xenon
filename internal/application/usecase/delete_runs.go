package usecase

import (
	"context"

	"github.com/brienze1/xenon/internal/domain/entity"
)

type DeleteRuns interface {
	DeleteRun(ctx context.Context, localId, runId string) (*entity.Run, error)
	ClearRuns(ctx context.Context, localId, targetId string) ([]entity.Run, error)
	DeleteOrphanRuns(ctx context.Context, localId string, targetIds []string) ([]entity.Run, error)
}
