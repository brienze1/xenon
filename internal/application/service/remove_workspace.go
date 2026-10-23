package service

import (
	"context"

	"github.com/brienze1/xenon/internal/application/adapter"
	"github.com/brienze1/xenon/internal/application/usecase"
	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/pkg/errs"
	"github.com/brienze1/xenon/pkg/logger"
)

type removeWorkspace struct {
	findWorkspaceRegistry usecase.FindWorkspaceRegistry
	saveWorkspaceRegistry usecase.SaveWorkspaceRegistry
	moveToTrash           usecase.MoveToTrash
	deleteUIState         usecase.DeleteUIState
	unwatchWorkspace      usecase.UnwatchWorkspace
	deleteVault           usecase.DeleteVault
	deleteLocalVariables  usecase.DeleteLocalVariables
	manageActiveRuns      usecase.ManageActiveRuns
}

func NewRemoveWorkspaceService(
	findWorkspaceRegistry usecase.FindWorkspaceRegistry,
	saveWorkspaceRegistry usecase.SaveWorkspaceRegistry,
	moveToTrash usecase.MoveToTrash,
	deleteUIState usecase.DeleteUIState,
	unwatchWorkspace usecase.UnwatchWorkspace,
	deleteVault usecase.DeleteVault,
	deleteLocalVariables usecase.DeleteLocalVariables,
	manageActiveRuns usecase.ManageActiveRuns,
) adapter.RemoveWorkspaceService {
	return &removeWorkspace{
		findWorkspaceRegistry: findWorkspaceRegistry,
		saveWorkspaceRegistry: saveWorkspaceRegistry,
		moveToTrash:           moveToTrash,
		deleteUIState:         deleteUIState,
		unwatchWorkspace:      unwatchWorkspace,
		deleteVault:           deleteVault,
		deleteLocalVariables:  deleteLocalVariables,
		manageActiveRuns:      manageActiveRuns,
	}
}

func (r *removeWorkspace) Remove(ctx context.Context, localId string, moveToTrash bool) (*entity.AppRegistry, error) {
	logger.Info(ctx, "Started", map[string]any{"localId": localId, "moveToTrash": moveToTrash})

	registered, err := r.findWorkspaceRegistry.FindWorkspace(ctx, localId)
	if err != nil {
		return nil, err
	}
	stopActiveRuns(ctx, r.manageActiveRuns, localId)
	if moveToTrash && !registered.IsMissing() {
		if err := r.moveToTrash.MoveToTrash(ctx, registered.Path); err != nil {
			return nil, err
		}
	}
	registry, err := r.findWorkspaceRegistry.Find(ctx)
	if err != nil {
		return nil, err
	}
	wasActive := registry.IsActive(localId)
	registry.Remove(localId)
	if err := r.saveWorkspaceRegistry.Save(ctx, *registry); err != nil {
		return nil, err
	}
	if err := r.deleteVault.Delete(ctx, localId); err != nil {
		if !errs.IsServiceUnavailableError(err) {
			return nil, err
		}
		logger.Warn(ctx, err, "The workspace vault could not be fully deleted", map[string]any{"localId": localId})
	}
	if err := r.deleteLocalVariables.Delete(ctx, localId); err != nil {
		return nil, err
	}
	if err := r.deleteUIState.Delete(ctx, localId); err != nil {
		return nil, err
	}
	if wasActive {
		if err := r.unwatchWorkspace.Unwatch(ctx); err != nil {
			return nil, err
		}
	}

	logger.Info(ctx, "Finished", map[string]any{"localId": localId, "moveToTrash": moveToTrash, "result": registry})
	return registry, nil
}
