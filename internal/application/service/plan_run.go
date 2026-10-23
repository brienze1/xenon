package service

import (
	"context"

	"github.com/brienze1/xenon/internal/application/adapter"
	"github.com/brienze1/xenon/internal/application/usecase"
	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/pkg/logger"
)

type planRun struct {
	planner *runPlanner
}

func NewPlanRunService(
	findWorkspaceRegistry usecase.FindWorkspaceRegistry,
	loadTree usecase.LoadTree,
	findRequest usecase.FindRequest,
	findLocalVariables usecase.FindLocalVariables,
	listEnvironments usecase.ListEnvironments,
	findActiveRuns usecase.FindActiveRuns,
	deleteRuns usecase.DeleteRuns,
	releaseResponse usecase.ReleaseResponse,
	findRunnerPreferences usecase.FindRunnerPreferences,
	saveRunnerPreferences usecase.SaveRunnerPreferences,
	deleteRunnerPreferences usecase.DeleteRunnerPreferences,
	findIterationData usecase.FindIterationData,
) adapter.PlanRunService {
	return &planRun{
		planner: newRunPlanner(
			findWorkspaceRegistry,
			loadTree,
			findRequest,
			findLocalVariables,
			listEnvironments,
			findActiveRuns,
			deleteRuns,
			releaseResponse,
			findRunnerPreferences,
			saveRunnerPreferences,
			deleteRunnerPreferences,
			findIterationData,
		),
	}
}

func (p *planRun) Plan(ctx context.Context, localId, itemId string) (*entity.RunPlan, error) {
	logger.Info(ctx, "Started", map[string]any{"localId": localId, "itemId": itemId})

	root, tree, err := p.planner.workspaceTree(ctx, localId)
	if err != nil {
		return nil, err
	}
	plan, err := p.planner.plan(ctx, localId, root, *tree, itemId)
	if err != nil {
		return nil, err
	}

	logger.Info(ctx, "Finished", map[string]any{"localId": localId, "itemId": itemId, "rows": len(plan.Rows), "dataFileMissing": plan.DataFileMissing})
	return plan, nil
}
