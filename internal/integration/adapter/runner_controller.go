package adapter

import (
	"context"

	"github.com/brienze1/xenon/internal/integration/entrypoint/dto"
)

type RunnerController interface {
	Start(ctx context.Context, request dto.StartRun) (*dto.RunSummaryView, error)
	Get(ctx context.Context, request dto.RunReference) (*dto.Run, error)
	Active(ctx context.Context, request dto.WorkspaceReference) ([]dto.RunSummaryView, error)
	Pause(ctx context.Context, request dto.RunReference) (*dto.RunSummaryView, error)
	Resume(ctx context.Context, request dto.RunReference) (*dto.RunSummaryView, error)
	Stop(ctx context.Context, request dto.RunReference) (*dto.RunSummaryView, error)
}
