package usecase

import (
	"context"

	"github.com/brienze1/xenon/internal/domain/entity"
)

type FindActiveRuns interface {
	FindById(ctx context.Context, localId, runId string) (*entity.Run, bool)
	FindByWorkspace(ctx context.Context, localId string) []entity.Run
}
