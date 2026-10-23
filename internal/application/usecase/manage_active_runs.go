package usecase

import (
	"context"

	"github.com/brienze1/xenon/internal/domain/entity"
)

type ManageActiveRuns interface {
	Claim(ctx context.Context, run entity.Run, control *entity.RunControl) error
	Change(ctx context.Context, localId, runId string, change func(*entity.Run)) (*entity.ActiveRun, bool)
	Release(ctx context.Context, localId, runId string)
	Find(ctx context.Context, localId, runId string) (*entity.ActiveRun, bool)
	ListWorkspace(ctx context.Context, localId string) []entity.ActiveRun
	ListActive(ctx context.Context) []entity.ActiveRun
}
