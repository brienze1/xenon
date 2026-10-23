package service

import (
	"context"

	"github.com/brienze1/xenon/internal/application/usecase"
	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/pkg/logger"
)

func (m *manageRunner) Plan(ctx context.Context, localId, itemId string) (*entity.RunPlan, error) {
	logger.Info(ctx, "Started", map[string]any{"localId": localId, "itemId": itemId})

	workspace, err := m.findWorkspaceRegistry.FindWorkspace(ctx, localId)
	if err != nil {
		return nil, err
	}
	tree, err := m.loadTree.Load(ctx, workspace.Path)
	if err != nil {
		return nil, err
	}
	container, found := tree.Container(itemId)
	if !found {
		return nil, usecase.RejectUnknownRunTarget(itemId)
	}
	preferences, err := m.findRunnerPreferences.FindRunnerPreferences(ctx, localId)
	if err != nil {
		return nil, err
	}
	if err := m.forgetRemovedTargets(ctx, localId, *tree, preferences); err != nil {
		return nil, err
	}
	sidebar, err := m.sidebarRows(ctx, workspace.Path, *container)
	if err != nil {
		return nil, err
	}
	config, dataFileMissing, err := m.rememberedConfig(ctx, localId, itemId, preferences)
	if err != nil {
		return nil, err
	}
	rows := entity.ReconcileRunRows(config.Rows, sidebar)
	plan, err := m.planOf(ctx, workspace.Path, entity.NewRunTarget(localId, *container), rows, config.WithRows(rows))
	if err != nil {
		return nil, err
	}
	plan.DataFileMissing = dataFileMissing

	logger.Info(ctx, "Finished", map[string]any{"localId": localId, "itemId": itemId, "rows": len(plan.Rows), "dataFileMissing": dataFileMissing})
	return plan, nil
}

func (m *manageRunner) SavePreferences(ctx context.Context, localId, itemId string, config entity.RunConfig) error {
	logger.Info(ctx, "Started", map[string]any{"localId": localId, "itemId": itemId, "config": config})

	if _, err := m.findWorkspaceRegistry.FindWorkspace(ctx, localId); err != nil {
		return err
	}
	if err := m.saveRunnerPreferences.SaveRunnerPreferences(ctx, localId, itemId, entity.NewRunnerPreferences(config)); err != nil {
		return err
	}

	logger.Info(ctx, "Finished", map[string]any{"localId": localId, "itemId": itemId})
	return nil
}

func (m *manageRunner) forgetRemovedTargets(ctx context.Context, localId string, tree entity.Tree, preferences map[string]entity.RunnerPreferences) error {
	live := tree.ContainerIDs()
	if stale := entity.StaleRunnerPreferences(preferences, live); len(stale) > 0 {
		if err := m.saveRunnerPreferences.DeleteRunnerPreferences(ctx, localId, stale); err != nil {
			return err
		}
	}
	deleted, err := m.deleteRuns.DeleteOrphanRuns(ctx, localId, live)
	if err != nil {
		return err
	}
	m.releaseResponses(ctx, deleted...)
	return nil
}

func (m *manageRunner) sidebarRows(ctx context.Context, root string, container entity.TreeNode) ([]entity.RunPlanRow, error) {
	rows := entity.FlattenRunRows(container)
	for index, row := range rows {
		stored, err := m.findRequest.Find(ctx, root, row.RequestID)
		if err != nil {
			return nil, err
		}
		rows[index] = row.WithAliases(stored.Aliases)
	}
	return rows, nil
}

func (m *manageRunner) rememberedConfig(ctx context.Context, localId, itemId string, preferences map[string]entity.RunnerPreferences) (entity.RunConfig, bool, error) {
	remembered, found := preferences[itemId]
	if !found {
		return entity.NewRunConfig(), false, nil
	}
	config := remembered.Config.Completed()
	missing := false
	for _, path := range config.DataFiles() {
		if !m.findIterationData.IterationDataExists(ctx, path) {
			config = config.WithoutDataFile(path)
			missing = true
		}
	}
	if missing {
		if err := m.saveRunnerPreferences.SaveRunnerPreferences(ctx, localId, itemId, entity.NewRunnerPreferences(config)); err != nil {
			return config, missing, err
		}
	}
	return config, missing, nil
}

func (m *manageRunner) planOf(ctx context.Context, root string, target entity.RunTarget, rows []entity.RunPlanRow, config entity.RunConfig) (*entity.RunPlan, error) {
	environments, err := m.listEnvironments.List(ctx, root)
	if err != nil {
		return nil, err
	}
	local, err := m.findLocalVariables.Find(ctx, target.LocalID)
	if local, err = localVariablesOrDefaults(ctx, target.LocalID, local, err); err != nil {
		return nil, err
	}
	plan := entity.RunPlan{Target: target, Rows: rows, Config: config, Environments: environments, ActiveEnvironmentID: local.ActiveEnvironmentID}
	if active, found := entity.ActiveRunOf(m.manageActiveRuns.ListWorkspace(ctx, target.LocalID), target.ItemID); found {
		plan.ActiveRun = &active
	}
	return &plan, nil
}
