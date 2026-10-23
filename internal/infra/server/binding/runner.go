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

func (r *RunnerService) Plan(ctx context.Context, request dto.RunTargetReference) (*dto.RunPlan, error) {
	return r.runner.Plan(ctx, request)
}

func (r *RunnerService) SavePreferences(ctx context.Context, request dto.SaveRunnerPreferences) error {
	return r.runner.SavePreferences(ctx, request)
}

func (r *RunnerService) List(ctx context.Context, request dto.RunTargetReference) ([]dto.RunSummaryView, error) {
	return r.runner.List(ctx, request)
}

func (r *RunnerService) Response(ctx context.Context, request dto.RunRowReference) (*dto.SendResult, error) {
	return r.runner.Response(ctx, request)
}

func (r *RunnerService) Delete(ctx context.Context, request dto.RunReference) error {
	return r.runner.Delete(ctx, request)
}

func (r *RunnerService) Clear(ctx context.Context, request dto.RunTargetReference) error {
	return r.runner.Clear(ctx, request)
}

func (r *RunnerService) Export(ctx context.Context, request dto.ExportRun) error {
	return r.runner.Export(ctx, request)
}

func (r *RunnerService) ServiceShutdown() error {
	return r.runner.Abandon(context.Background())
}
