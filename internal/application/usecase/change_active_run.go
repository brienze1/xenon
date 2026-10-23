package usecase

import (
	"context"

	"github.com/brienze1/xenon/internal/domain/entity"
)

type ChangeActiveRun interface {
	Change(ctx context.Context, localId, runId string, change func(*entity.Run)) (*entity.Run, bool)
}
