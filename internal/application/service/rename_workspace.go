package service

import (
	"context"

	"github.com/brienze1/xenon/internal/application/adapter"
	"github.com/brienze1/xenon/internal/application/usecase"
	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/pkg/logger"
)

type renameWorkspace struct {
	findWorkspaceRegistry usecase.FindWorkspaceRegistry
	saveWorkspaceRegistry usecase.SaveWorkspaceRegistry
	loadWorkspace         usecase.LoadWorkspace
	saveWorkspaceFile     usecase.SaveWorkspaceFile
}

func NewRenameWorkspaceService(
	findWorkspaceRegistry usecase.FindWorkspaceRegistry,
	saveWorkspaceRegistry usecase.SaveWorkspaceRegistry,
	loadWorkspace usecase.LoadWorkspace,
	saveWorkspaceFile usecase.SaveWorkspaceFile,
) adapter.RenameWorkspaceService {
	return &renameWorkspace{
		findWorkspaceRegistry: findWorkspaceRegistry,
		saveWorkspaceRegistry: saveWorkspaceRegistry,
		loadWorkspace:         loadWorkspace,
		saveWorkspaceFile:     saveWorkspaceFile,
	}
}

func (r *renameWorkspace) Rename(ctx context.Context, localId, name string) (*entity.RegisteredWorkspace, error) {
	logger.Info(ctx, "Started", map[string]any{"localId": localId, "name": name})

	registered, err := r.findWorkspaceRegistry.FindWorkspace(ctx, localId)
	if err != nil {
		return nil, err
	}
	workspace, err := r.loadWorkspace.Load(ctx, registered.Path)
	if err != nil {
		return nil, err
	}
	renamed := workspace.Renamed(name)
	if err := r.saveWorkspaceFile.Save(ctx, registered.Path, renamed); err != nil {
		return nil, err
	}
	registry, err := r.findWorkspaceRegistry.Find(ctx)
	if err != nil {
		return nil, err
	}
	updated, _ := registry.Rename(localId, renamed.Name)
	if err := r.saveWorkspaceRegistry.Save(ctx, *registry); err != nil {
		return nil, err
	}

	logger.Info(ctx, "Finished", map[string]any{"localId": localId, "name": name, "result": updated})
	return updated, nil
}
