package service

import (
	"context"

	"github.com/brienze1/xenon/internal/application/adapter"
	"github.com/brienze1/xenon/internal/application/usecase"
	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/pkg/errs"
	"github.com/brienze1/xenon/pkg/logger"
)

const activeEnvironmentRemoved = "active_environment_removed"

type handleFileChange struct {
	findWorkspaceRegistry usecase.FindWorkspaceRegistry
	inspectFileChange     usecase.InspectFileChange
	unwatchWorkspace      usecase.UnwatchWorkspace
	publishWorkspaceEvent usecase.PublishWorkspaceEvent
	scopes                variableScopes
	findLocalVariables    usecase.FindLocalVariables
	saveLocalVariables    usecase.SaveLocalVariables
	findVault             usecase.FindVault
	saveVault             usecase.SaveVault
	publishVariableEvent  usecase.PublishVariableEvent
}

func NewHandleFileChangeService(
	findWorkspaceRegistry usecase.FindWorkspaceRegistry,
	inspectFileChange usecase.InspectFileChange,
	unwatchWorkspace usecase.UnwatchWorkspace,
	publishWorkspaceEvent usecase.PublishWorkspaceEvent,
	findEnvironment usecase.FindEnvironment,
	findTreeItem usecase.FindTreeItem,
	loadWorkspace usecase.LoadWorkspace,
	findLocalVariables usecase.FindLocalVariables,
	saveLocalVariables usecase.SaveLocalVariables,
	findVault usecase.FindVault,
	saveVault usecase.SaveVault,
	publishVariableEvent usecase.PublishVariableEvent,
) adapter.HandleFileChangeService {
	return &handleFileChange{
		findWorkspaceRegistry: findWorkspaceRegistry,
		inspectFileChange:     inspectFileChange,
		unwatchWorkspace:      unwatchWorkspace,
		publishWorkspaceEvent: publishWorkspaceEvent,
		scopes:                variableScopes{findEnvironment: findEnvironment, findTreeItem: findTreeItem, loadWorkspace: loadWorkspace},
		findLocalVariables:    findLocalVariables,
		saveLocalVariables:    saveLocalVariables,
		findVault:             findVault,
		saveVault:             saveVault,
		publishVariableEvent:  publishVariableEvent,
	}
}

func (h *handleFileChange) Handle(ctx context.Context, localId, path string) error {
	logger.Info(ctx, "Started", map[string]any{"localId": localId, "path": path})

	workspace, err := h.findWorkspaceRegistry.FindWorkspace(ctx, localId)
	if err != nil {
		return err
	}
	if workspace.IsMissing() {
		if err := h.unwatchWorkspace.Unwatch(ctx); err != nil {
			return err
		}
		if err := h.publishWorkspaceEvent.FolderMissing(ctx, localId); err != nil {
			return err
		}
		logger.Info(ctx, "Finished", map[string]any{"localId": localId, "path": path, "result": "folderMissing"})
		return nil
	}
	if !workspace.Watches(path) {
		logger.Info(ctx, "Finished", map[string]any{"localId": localId, "path": path, "result": "ignored"})
		return nil
	}
	change, err := h.inspectFileChange.Inspect(ctx, workspace.Path, path)
	if err != nil {
		return err
	}
	if change.IsOwnWrite() {
		logger.Info(ctx, "Finished", map[string]any{"localId": localId, "path": path, "result": "ownWrite"})
		return nil
	}
	for _, request := range change.Requests {
		if err := h.publishWorkspaceEvent.RequestChangedOnDisk(ctx, localId, request); err != nil {
			return err
		}
	}
	if err := h.publishWorkspaceEvent.TreeChanged(ctx, localId); err != nil {
		return err
	}
	if change.ChangesVariables() {
		if err := h.variablesChanged(ctx, localId, workspace.Path, *change); err != nil {
			return err
		}
	}

	logger.Info(ctx, "Finished", map[string]any{"localId": localId, "path": path, "result": change.Requests})
	return nil
}

func (h *handleFileChange) variablesChanged(ctx context.Context, localId, root string, change entity.FileChange) error {
	local, err := h.findLocalVariables.Find(ctx, localId)
	if local, err = localVariablesOrDefaults(ctx, localId, local, err); err != nil {
		return err
	}
	reason := ""
	if change.EnvironmentsChanged && local.HasActiveEnvironment() {
		if _, err := h.scopes.findEnvironment.Find(ctx, root, local.ActiveEnvironmentID); errs.IsNotFoundError(err) {
			cleared := local.WithActiveEnvironment("")
			local, reason = &cleared, activeEnvironmentRemoved
		}
	}
	for _, scope := range change.VariableScopes {
		if local, err = h.moveSecretOverrides(ctx, localId, root, scope, *local); err != nil {
			return err
		}
	}
	if err := h.saveLocalVariables.Save(ctx, localId, *local); err != nil {
		return err
	}
	return h.publishVariableEvent.VariablesChanged(ctx, localId, change.VariableScopes, reason)
}

func (h *handleFileChange) moveSecretOverrides(ctx context.Context, localId, root string, scope entity.ScopeRef, local entity.LocalVariables) (*entity.LocalVariables, error) {
	rows, err := h.scopes.rows(ctx, root, scope)
	if err != nil {
		logger.Warn(ctx, err, "The changed variables could not be read", map[string]any{"localId": localId, "scope": scope.String()})
		return &local, nil
	}
	vault, err := vaultFor(ctx, h.findVault, localId, rows)
	if err != nil {
		return nil, err
	}
	table := entity.NewVariableTable(scope, rows, local, *vault)
	moved := table.WithOverridesOfSecretRowsMoved(vault.IsReadable())
	if moved.SecretsDiffer(table) {
		if err := h.saveVault.Save(ctx, vault.WithSecrets(scope, moved.StoredSecrets())); err != nil {
			return nil, err
		}
	}
	updated := local.WithCurrentValues(scope, moved.StoredCurrent())
	return &updated, nil
}
