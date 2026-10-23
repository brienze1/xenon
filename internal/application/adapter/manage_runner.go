package adapter

import (
	"context"

	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/internal/domain/enums"
)

type ManageRunnerService interface {
	Plan(ctx context.Context, localId, itemId string) (*entity.RunPlan, error)
	SavePreferences(ctx context.Context, localId, itemId string, config entity.RunConfig) error
	Start(ctx context.Context, start entity.RunStart, appVersion string) (*entity.Run, error)
	Get(ctx context.Context, localId, runId string) (*entity.Run, error)
	List(ctx context.Context, localId, itemId string) ([]entity.Run, error)
	Active(ctx context.Context, localId string) ([]entity.Run, error)
	Pause(ctx context.Context, localId, runId string) (*entity.Run, error)
	Resume(ctx context.Context, localId, runId string) (*entity.Run, error)
	Stop(ctx context.Context, localId, runId string) (*entity.Run, error)
	Response(ctx context.Context, localId, runId string, index int) (*entity.Response, error)
	Delete(ctx context.Context, localId, runId string) error
	Clear(ctx context.Context, localId, itemId string) error
	Export(ctx context.Context, localId, runId string, format enums.RunExportFormat, path string) error
	Abandon(ctx context.Context)
}
