package service

import (
	"context"

	"github.com/brienze1/xenon/internal/application/adapter"
	"github.com/brienze1/xenon/internal/application/usecase"
	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/pkg/logger"
)

type removeWorkspace struct {
	findWorkspaceRegistry usecase.FindWorkspaceRegistry
	saveWorkspaceRegistry usecase.SaveWorkspaceRegistry
	moveToTrash           usecase.MoveToTrash
	deleteUIState         usecase.DeleteUIState
	watchWorkspace        usecase.WatchWorkspace
}

func NewRemoveWorkspaceService(
	findWorkspaceRegistry usecase.FindWorkspaceRegistry,
	saveWorkspaceRegistry usecase.SaveWorkspaceRegistry,
	moveToTrash usecase.MoveToTrash,
	deleteUIState usecase.DeleteUIState,
	watchWorkspace usecase.WatchWorkspace,
) adapter.RemoveWorkspaceService {
	return &removeWorkspace{
		findWorkspaceRegistry: findWorkspaceRegistry,
		saveWorkspaceRegistry: saveWorkspaceRegistry,
		moveToTrash:           moveToTrash,
		deleteUIState:         deleteUIState,
		watchWorkspace:        watchWorkspace,
	}
}

func (r *removeWorkspace) Remove(ctx context.Context, localId string, moveToTrash bool) (*entity.AppRegistry, error) {
	logger.Info(ctx, "Started", map[string]any{"localId": localId, "moveToTrash": moveToTrash})

	registered, err := r.findWorkspaceRegistry.FindWorkspace(ctx, localId)
	if err != nil {
		return nil, err
	}
	if moveToTrash && !registered.Missing {
		if err := r.moveToTrash.MoveToTrash(ctx, registered.Path); err != nil {
			return nil, err
		}
	}
	registry, err := r.findWorkspaceRegistry.Find(ctx)
	if err != nil {
		return nil, err
	}
	wasActive := registry.ActiveLocalID == localId
	registry.Remove(localId)
	if err := r.saveWorkspaceRegistry.Save(ctx, *registry); err != nil {
		return nil, err
	}
	if err := r.deleteUIState.Delete(ctx, localId); err != nil {
		return nil, err
	}
	if wasActive {
		if err := r.watchWorkspace.Unwatch(ctx); err != nil {
			return nil, err
		}
	}

	logger.Info(ctx, "Finished", registry)
	return registry, nil
}
