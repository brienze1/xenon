package adapter

import (
	"context"

	"github.com/brienze1/xenon/internal/domain/entity"
)

type ManageRunnerService interface {
	Start(ctx context.Context, start entity.RunStart, appVersion string) (*entity.Run, error)
	Get(ctx context.Context, localId, runId string) (*entity.Run, error)
	Active(ctx context.Context, localId string) ([]entity.Run, error)
	Pause(ctx context.Context, localId, runId string) (*entity.Run, error)
	Resume(ctx context.Context, localId, runId string) (*entity.Run, error)
	Stop(ctx context.Context, localId, runId string) (*entity.Run, error)
}
