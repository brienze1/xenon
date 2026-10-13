package service

import (
	"context"

	"github.com/brienze1/xenon/internal/application/usecase"
	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/pkg/idutils"
	"github.com/brienze1/xenon/pkg/logger"
	"github.com/brienze1/xenon/pkg/timeutils"
)

type importWorkspaceWriter struct {
	copyWorkspaceFolder   usecase.CopyWorkspaceFolder
	removeWorkspaceFolder usecase.RemoveWorkspaceFolder
	findWorkspaceRegistry usecase.FindWorkspaceRegistry
	saveWorkspaceRegistry usecase.SaveWorkspaceRegistry
	loadWorkspace         usecase.LoadWorkspace
	deleteVault           usecase.DeleteVault
	vaults                vaultWriter
}

func (w importWorkspaceWriter) importOne(ctx context.Context, target importTarget) entity.ImportOutcome {
	folder := target.planned.TargetFolder
	if err := w.copyWorkspaceFolder.CopyFolder(ctx, target.item.Workspace.Folder, folder); err != nil {
		return usecase.FailImportItem(target.item, "", err)
	}
	workspace, err := w.loadWorkspace.Load(ctx, folder)
	if err != nil {
		w.discard(ctx, folder, "")
		return usecase.FailImportItem(target.item, "", err)
	}
	registered := entity.NewRegisteredWorkspace(idutils.NewUlid(), folder, workspace.Name, timeutils.Now())
	if err := w.keepSecrets(ctx, registered.LocalID, target.item.Workspace.Secrets); err != nil {
		w.discard(ctx, folder, registered.LocalID)
		return usecase.FailImportItem(target.item, "", err)
	}
	if err := w.register(ctx, registered); err != nil {
		w.discard(ctx, folder, registered.LocalID)
		return usecase.FailImportItem(target.item, "", err)
	}
	return target.item.CompletedWorkspace(registered.LocalID, workspace.ID)
}

func (w importWorkspaceWriter) keepSecrets(ctx context.Context, localId string, secrets entity.SecretBundle) error {
	if secrets.IsEmpty() {
		return nil
	}
	_, err := w.vaults.update(ctx, localId, func(vault entity.Vault) entity.Vault {
		return vault.WithBundle(secrets)
	})
	return err
}

func (w importWorkspaceWriter) register(ctx context.Context, registered entity.RegisteredWorkspace) error {
	registry, err := w.findWorkspaceRegistry.Find(ctx)
	if err != nil {
		return err
	}
	registry.Register(registered)
	return w.saveWorkspaceRegistry.Save(ctx, *registry)
}

func (w importWorkspaceWriter) discard(ctx context.Context, folder, localId string) {
	if err := w.removeWorkspaceFolder.RemoveFolder(ctx, folder); err != nil {
		logger.Error(ctx, err, "The copied workspace folder could not be removed after a failed import", map[string]any{"folder": folder})
	}
	if localId == "" {
		return
	}
	if err := w.deleteVault.Delete(ctx, localId); err != nil {
		logger.Error(ctx, err, "The imported workspace's vault could not be removed after a failed import", map[string]any{"localId": localId})
	}
}
