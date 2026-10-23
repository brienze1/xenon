package adapter

import (
	"context"

	"github.com/brienze1/xenon/internal/integration/entrypoint/dto"
)

type RunnerController interface {
	Plan(ctx context.Context, request dto.RunTargetReference) (*dto.RunPlan, error)
	SavePreferences(ctx context.Context, request dto.SaveRunnerPreferences) error
	Start(ctx context.Context, request dto.StartRun) (*dto.RunSummaryView, error)
	Get(ctx context.Context, request dto.RunReference) (*dto.Run, error)
	List(ctx context.Context, request dto.RunTargetReference) ([]dto.RunSummaryView, error)
	Active(ctx context.Context, request dto.WorkspaceReference) ([]dto.RunSummaryView, error)
	Pause(ctx context.Context, request dto.RunReference) (*dto.RunSummaryView, error)
	Resume(ctx context.Context, request dto.RunReference) (*dto.RunSummaryView, error)
	Stop(ctx context.Context, request dto.RunReference) (*dto.RunSummaryView, error)
	Response(ctx context.Context, request dto.RunRowReference) (*dto.SendResult, error)
	Delete(ctx context.Context, request dto.RunReference) error
	Clear(ctx context.Context, request dto.RunTargetReference) error
	Export(ctx context.Context, request dto.ExportRun) error
	Abandon(ctx context.Context) error
}
