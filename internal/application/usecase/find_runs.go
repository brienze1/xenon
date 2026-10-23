package usecase

import (
	"context"

	"github.com/brienze1/xenon/internal/domain/entity"
)

type FindRuns interface {
	FindRun(ctx context.Context, localId, runId string) (*entity.Run, error)
	FindRuns(ctx context.Context, localId, targetId string) ([]entity.Run, error)
}
