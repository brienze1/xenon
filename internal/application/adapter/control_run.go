package adapter

import (
	"context"

	"github.com/brienze1/xenon/internal/domain/entity"
)

type ControlRunService interface {
	Pause(ctx context.Context, localId, runId string) (*entity.Run, error)
	Resume(ctx context.Context, localId, runId string) (*entity.Run, error)
	Stop(ctx context.Context, localId, runId string) (*entity.Run, error)
	Abandon(ctx context.Context)
}
