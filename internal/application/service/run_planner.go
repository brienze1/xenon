package service

import (
	"context"

	"github.com/brienze1/xenon/internal/application/usecase"
	"github.com/brienze1/xenon/internal/domain/entity"
)

type runPlanner struct {
	findWorkspaceRegistry   usecase.FindWorkspaceRegistry
	loadTree                usecase.LoadTree
	findRequest             usecase.FindRequest
	findLocalVariables      usecase.FindLocalVariables
	listEnvironments        usecase.ListEnvironments
	findActiveRuns          usecase.FindActiveRuns
	deleteRuns              usecase.DeleteRuns
	releaseResponse         usecase.ReleaseResponse
	findRunnerPreferences   usecase.FindRunnerPreferences
	saveRunnerPreferences   usecase.SaveRunnerPreferences
	deleteRunnerPreferences usecase.DeleteRunnerPreferences
	findIterationData       usecase.FindIterationData
}

func newRunPlanner(
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
) *runPlanner {
	return &runPlanner{
		findWorkspaceRegistry:   findWorkspaceRegistry,
		loadTree:                loadTree,
		findRequest:             findRequest,
		findLocalVariables:      findLocalVariables,
		listEnvironments:        listEnvironments,
		findActiveRuns:          findActiveRuns,
		deleteRuns:              deleteRuns,
		releaseResponse:         releaseResponse,
		findRunnerPreferences:   findRunnerPreferences,
		saveRunnerPreferences:   saveRunnerPreferences,
		deleteRunnerPreferences: deleteRunnerPreferences,
		findIterationData:       findIterationData,
	}
}

func (p *runPlanner) workspaceTree(ctx context.Context, localId string) (string, *entity.Tree, error) {
	workspace, err := p.findWorkspaceRegistry.FindWorkspace(ctx, localId)
	if err != nil {
		return "", nil, err
	}
	tree, err := p.loadTree.Load(ctx, workspace.Path)
	if err != nil {
		return "", nil, err
	}
	return workspace.Path, tree, nil
}

func (p *runPlanner) plan(ctx context.Context, localId, root string, tree entity.Tree, itemId string) (*entity.RunPlan, error) {
	container, err := tree.RunTarget(itemId)
	if err != nil {
		return nil, err
	}
	preferences, err := p.findRunnerPreferences.FindRunnerPreferences(ctx, localId)
	if err != nil {
		return nil, err
	}
	if err := p.forgetRemovedTargets(ctx, localId, tree, preferences); err != nil {
		return nil, err
	}
	sidebar, err := p.sidebarRows(ctx, root, *container)
	if err != nil {
		return nil, err
	}
	config, dataFileMissing, err := p.rememberedConfig(ctx, localId, itemId, preferences)
	if err != nil {
		return nil, err
	}
	rows := entity.ReconcileRunRows(config.Rows, sidebar)
	planned, err := p.planOf(ctx, root, entity.NewRunTarget(localId, *container), rows, config.WithRows(rows))
	if err != nil {
		return nil, err
	}
	plan := planned.WithDataFileMissing(dataFileMissing)
	return &plan, nil
}

func (p *runPlanner) forgetRemovedTargets(ctx context.Context, localId string, tree entity.Tree, preferences map[string]entity.RunnerPreferences) error {
	if stale := tree.StaleRunnerTargets(preferences); len(stale) > 0 {
		if err := p.deleteRunnerPreferences.DeleteRunnerPreferences(ctx, localId, stale); err != nil {
			return err
		}
	}
	deleted, err := p.deleteRuns.DeleteOrphanRuns(ctx, localId, tree.ContainerIDs())
	if err != nil {
		return err
	}
	return releaseRunResponses(ctx, p.releaseResponse, deleted...)
}

func (p *runPlanner) sidebarRows(ctx context.Context, root string, container entity.TreeNode) ([]entity.RunPlanRow, error) {
	rows := entity.FlattenRunRows(container)
	for index, row := range rows {
		stored, err := p.findRequest.Find(ctx, root, row.RequestID)
		if err != nil {
			return nil, err
		}
		rows[index] = row.WithAliases(stored.Aliases)
	}
	return rows, nil
}

func (p *runPlanner) rememberedConfig(ctx context.Context, localId, itemId string, preferences map[string]entity.RunnerPreferences) (entity.RunConfig, bool, error) {
	remembered, found := preferences[itemId]
	if !found {
		return entity.NewRunConfig(), false, nil
	}
	config := remembered.RememberedConfig()
	missing := false
	for _, path := range config.DataFiles() {
		if !p.findIterationData.IterationDataExists(ctx, path) {
			config = config.WithoutDataFile(path)
			missing = true
		}
	}
	if missing {
		if err := p.saveRunnerPreferences.SaveRunnerPreferences(ctx, localId, itemId, entity.NewRunnerPreferences(config)); err != nil {
			return config, missing, err
		}
	}
	return config, missing, nil
}

func (p *runPlanner) planOf(ctx context.Context, root string, target entity.RunTarget, rows []entity.RunPlanRow, config entity.RunConfig) (*entity.RunPlan, error) {
	environments, err := p.listEnvironments.List(ctx, root)
	if err != nil {
		return nil, err
	}
	local, err := p.findLocalVariables.Find(ctx, target.LocalID)
	if local, err = localVariablesOrDefaults(ctx, target.LocalID, local, err); err != nil {
		return nil, err
	}
	plan := entity.NewRunPlan(target, rows, config, environments, local.ActiveEnvironmentID).WithActiveRunOf(p.findActiveRuns.FindByWorkspace(ctx, target.LocalID))
	return &plan, nil
}
