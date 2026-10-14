package service

import (
	"context"
	"maps"

	"github.com/brienze1/xenon/internal/application/usecase"
	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/internal/domain/enums"
	"github.com/brienze1/xenon/pkg/logger"
)

type importVariablesWriter struct {
	saveEnvironment      usecase.SaveEnvironment
	deleteEnvironment    usecase.DeleteEnvironment
	saveWorkspaceFile    usecase.SaveWorkspaceFile
	loadWorkspace        usecase.LoadWorkspace
	values               scopeValues
	vaults               vaultWriter
	publishVariableEvent usecase.PublishVariableEvent
}

func (w importVariablesWriter) importOne(ctx context.Context, target importTarget, resolution enums.ClashResolution) entity.ImportOutcome {
	write, err := w.written(ctx, target, resolution)
	if err != nil {
		return usecase.FailImportItem(target.item, target.workspace.LocalID, err)
	}
	if write.IsEmpty() {
		return target.item.Skipped(target.workspace.LocalID)
	}
	if err := w.publishVariableEvent.VariablesChanged(ctx, target.workspace.LocalID, []entity.ScopeRef{write.Scope()}, ""); err != nil {
		logger.Error(ctx, err, "The imported variables could not be announced", map[string]any{"localId": target.workspace.LocalID})
	}
	return target.item.CompletedVariables(*write, target.workspace.LocalID)
}

func (w importVariablesWriter) written(ctx context.Context, target importTarget, resolution enums.ClashResolution) (*entity.ImportVariablesWrite, error) {
	if target.item.IsGlobals() {
		return w.globals(ctx, target, resolution)
	}
	write := usecase.PlanEnvironmentImport(*target.item.Variables, target.environments, resolution)
	if write.IsEmpty() {
		return &write, nil
	}
	return &write, w.environment(ctx, target.workspace, write)
}

func (w importVariablesWriter) environment(ctx context.Context, workspace entity.RegisteredWorkspace, write entity.ImportVariablesWrite) error {
	vault, err := w.storeSecrets(ctx, workspace.LocalID, write.Scope(), write.Environment.Variables, map[string]string{}, write.Secrets)
	if err != nil {
		return err
	}
	if write.ReplacesEnvironment() {
		if err := w.deleteEnvironment.Delete(ctx, workspace.Path, write.Replaced.ID); err != nil {
			w.vaults.restore(ctx, vault, write.Scope(), nil)
			return err
		}
	}
	if _, err := w.saveEnvironment.Save(ctx, workspace.Path, *write.Environment); err != nil {
		w.restoreReplaced(ctx, workspace.Path, write.Replaced)
		w.vaults.restore(ctx, vault, write.Scope(), nil)
		return err
	}
	if write.ReplacesEnvironment() {
		if _, err := w.values.drop(ctx, workspace.LocalID, *write.Replaced); err != nil {
			logger.Error(ctx, err, "The replaced environment's local and secret values could not be removed", map[string]any{"localId": workspace.LocalID})
		}
	}
	return nil
}

func (w importVariablesWriter) globals(ctx context.Context, target importTarget, resolution enums.ClashResolution) (*entity.ImportVariablesWrite, error) {
	scope := entity.NewGlobalsScopeRef()
	globals, err := w.loadWorkspace.Load(ctx, target.workspace.Path)
	if err != nil {
		return nil, err
	}
	vault, err := vaultFor(ctx, w.values.findVault, target.workspace.LocalID, globals.Variables, target.item.Variables.Rows())
	if err != nil {
		return nil, err
	}
	current := vault.SecretsFor(scope)
	write := usecase.PlanGlobalsImport(*target.item.Variables, *globals, current, resolution)
	written, err := w.storeSecrets(ctx, target.workspace.LocalID, scope, write.Globals.Variables, current, write.Secrets)
	if err != nil {
		return nil, err
	}
	if err := w.saveWorkspaceFile.Save(ctx, target.workspace.Path, *write.Globals); err != nil {
		w.vaults.restore(ctx, written, scope, current)
		return nil, err
	}
	return &write, nil
}

func (w importVariablesWriter) storeSecrets(ctx context.Context, localId string, scope entity.ScopeRef, rows []entity.VariableRow, current, secrets map[string]string) (entity.Vault, error) {
	if maps.Equal(current, secrets) {
		return entity.NewUnopenedVault(localId), nil
	}
	vault, err := vaultFor(ctx, w.values.findVault, localId, rows)
	if err != nil {
		return entity.Vault{}, err
	}
	return w.vaults.write(ctx, *vault, scope, secrets)
}

func (w importVariablesWriter) restoreReplaced(ctx context.Context, root string, replaced *entity.Environment) {
	if replaced == nil {
		return
	}
	if _, err := w.saveEnvironment.Save(ctx, root, *replaced); err != nil {
		logger.Error(ctx, err, "The replaced environment could not be restored after a failed import", map[string]any{"root": root})
	}
}
