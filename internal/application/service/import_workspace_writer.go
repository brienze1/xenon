package service

import (
	"context"

	"github.com/brienze1/xenon/internal/application/usecase"
	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/pkg/idutils"
	"github.com/brienze1/xenon/pkg/timeutils"
)

type importWorkspaceWriter struct {
	copyWorkspaceFolder   usecase.CopyWorkspaceFolder
	findWorkspaceRegistry usecase.FindWorkspaceRegistry
	saveWorkspaceRegistry usecase.SaveWorkspaceRegistry
	loadWorkspace         usecase.LoadWorkspace
	vaults                vaultWriter
}

func (w importWorkspaceWriter) importOne(ctx context.Context, target importTarget) entity.ImportOutcome {
	folder := target.planned.TargetFolder
	if err := w.copyWorkspaceFolder.CopyFolder(ctx, target.item.Workspace.Folder, folder); err != nil {
		return usecase.FailImportItem(target.item, "", err)
	}
	registered, workspace, err := w.register(ctx, folder)
	if err != nil {
		return usecase.FailImportItem(target.item, "", err)
	}
	if secrets := target.item.Workspace.Secrets; !secrets.IsEmpty() {
		_, err := w.vaults.update(ctx, registered.LocalID, func(vault entity.Vault) entity.Vault {
			return vault.WithBundle(secrets)
		})
		if err != nil {
			return usecase.FailImportItem(target.item, registered.LocalID, err)
		}
	}
	return target.item.CompletedWorkspace(registered.LocalID, workspace.ID)
}

func (w importWorkspaceWriter) register(ctx context.Context, folder string) (*entity.RegisteredWorkspace, *entity.Workspace, error) {
	registry, err := w.findWorkspaceRegistry.Find(ctx)
	if err != nil {
		return nil, nil, err
	}
	workspace, err := w.loadWorkspace.Load(ctx, folder)
	if err != nil {
		return nil, nil, err
	}
	registered := entity.NewRegisteredWorkspace(idutils.NewUlid(), folder, workspace.Name, timeutils.Now())
	registry.Register(registered)
	if err := w.saveWorkspaceRegistry.Save(ctx, *registry); err != nil {
		return nil, nil, err
	}
	return &registered, workspace, nil
}
