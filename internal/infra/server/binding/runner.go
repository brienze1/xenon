package binding

import (
	"context"

	"github.com/brienze1/xenon/internal/integration/adapter"
	"github.com/brienze1/xenon/internal/integration/entrypoint/dto"
)

type RunnerService struct {
	runner adapter.RunnerController
}

func NewRunnerService(runner adapter.RunnerController) *RunnerService {
	return &RunnerService{runner: runner}
}

func (r *RunnerService) Start(ctx context.Context, request dto.StartRun) (*dto.RunSummaryView, error) {
	return r.runner.Start(ctx, request)
}

func (r *RunnerService) Get(ctx context.Context, request dto.RunReference) (*dto.Run, error) {
	return r.runner.Get(ctx, request)
}

func (r *RunnerService) Active(ctx context.Context, request dto.WorkspaceReference) ([]dto.RunSummaryView, error) {
	return r.runner.Active(ctx, request)
}

func (r *RunnerService) Pause(ctx context.Context, request dto.RunReference) (*dto.RunSummaryView, error) {
	return r.runner.Pause(ctx, request)
}

func (r *RunnerService) Resume(ctx context.Context, request dto.RunReference) (*dto.RunSummaryView, error) {
	return r.runner.Resume(ctx, request)
}

func (r *RunnerService) Stop(ctx context.Context, request dto.RunReference) (*dto.RunSummaryView, error) {
	return r.runner.Stop(ctx, request)
}
