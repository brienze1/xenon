package service

import (
	"context"

	"github.com/brienze1/xenon/internal/application/adapter"
	"github.com/brienze1/xenon/internal/application/usecase"
	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/pkg/logger"
)

type activateWorkspace struct {
	findWorkspaceRegistry usecase.FindWorkspaceRegistry
	saveWorkspaceRegistry usecase.SaveWorkspaceRegistry
	loadWorkspace         usecase.LoadWorkspace
	loadTree              usecase.LoadTree
	findUIState           usecase.FindUIState
	watchWorkspace        usecase.WatchWorkspace
	publishWorkspaceEvent usecase.PublishWorkspaceEvent
	findActiveRunControls usecase.FindActiveRunControls
}

func NewActivateWorkspaceService(
	findWorkspaceRegistry usecase.FindWorkspaceRegistry,
	saveWorkspaceRegistry usecase.SaveWorkspaceRegistry,
	loadWorkspace usecase.LoadWorkspace,
	loadTree usecase.LoadTree,
	findUIState usecase.FindUIState,
	watchWorkspace usecase.WatchWorkspace,
	publishWorkspaceEvent usecase.PublishWorkspaceEvent,
	findActiveRunControls usecase.FindActiveRunControls,
) adapter.ActivateWorkspaceService {
	return &activateWorkspace{
		findWorkspaceRegistry: findWorkspaceRegistry,
		saveWorkspaceRegistry: saveWorkspaceRegistry,
		loadWorkspace:         loadWorkspace,
		loadTree:              loadTree,
		findUIState:           findUIState,
		watchWorkspace:        watchWorkspace,
		publishWorkspaceEvent: publishWorkspaceEvent,
		findActiveRunControls: findActiveRunControls,
	}
}

func (a *activateWorkspace) Activate(ctx context.Context, localId string) (*entity.WorkspaceSnapshot, error) {
	logger.Info(ctx, "Started", map[string]any{"localId": localId})

	registered, err := a.findWorkspaceRegistry.FindWorkspace(ctx, localId)
	if err != nil {
		return nil, err
	}
	if registered.IsMissing() {
		if err := a.publishWorkspaceEvent.FolderMissing(ctx, localId); err != nil {
			return nil, err
		}
	}
	workspace, err := a.loadWorkspace.Load(ctx, registered.Path)
	if err != nil {
		return nil, err
	}
	registry, err := a.findWorkspaceRegistry.Find(ctx)
	if err != nil {
		return nil, err
	}
	if !registry.IsActive(localId) {
		stopActiveRuns(ctx, a.findActiveRunControls, registry.ActiveLocalID)
	}
	activated, _ := registry.Activate(localId, workspace.Name)
	if err := a.saveWorkspaceRegistry.Save(ctx, *registry); err != nil {
		return nil, err
	}
	tree, err := a.loadTree.Load(ctx, activated.Path)
	if err != nil {
		return nil, err
	}
	state, err := a.findUIState.Find(ctx, localId)
	if state, err = uiStateOrDefaults(ctx, localId, state, err); err != nil {
		return nil, err
	}
	if err := a.watchWorkspace.Watch(ctx, localId, activated.Path); err != nil {
		return nil, err
	}
	snapshot := entity.WorkspaceSnapshot{Workspace: *activated, Tree: *tree, UIState: *state}

	logger.Info(ctx, "Finished", map[string]any{"localId": localId, "result": snapshot})
	return &snapshot, nil
}
