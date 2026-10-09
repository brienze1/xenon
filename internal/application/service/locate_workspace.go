package service

import (
	"context"

	"github.com/brienze1/xenon/internal/application/adapter"
	"github.com/brienze1/xenon/internal/application/usecase"
	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/pkg/logger"
)

type locateWorkspace struct {
	findWorkspaceRegistry   usecase.FindWorkspaceRegistry
	saveWorkspaceRegistry   usecase.SaveWorkspaceRegistry
	validateWorkspaceFolder usecase.ValidateWorkspaceFolder
	loadWorkspace           usecase.LoadWorkspace
}

func NewLocateWorkspaceService(
	findWorkspaceRegistry usecase.FindWorkspaceRegistry,
	saveWorkspaceRegistry usecase.SaveWorkspaceRegistry,
	validateWorkspaceFolder usecase.ValidateWorkspaceFolder,
	loadWorkspace usecase.LoadWorkspace,
) adapter.LocateWorkspaceService {
	return &locateWorkspace{
		findWorkspaceRegistry:   findWorkspaceRegistry,
		saveWorkspaceRegistry:   saveWorkspaceRegistry,
		validateWorkspaceFolder: validateWorkspaceFolder,
		loadWorkspace:           loadWorkspace,
	}
}

func (l *locateWorkspace) Locate(ctx context.Context, localId, path string) (*entity.RegisteredWorkspace, error) {
	logger.Info(ctx, "Started", map[string]any{"localId": localId, "path": path})

	if _, err := l.findWorkspaceRegistry.FindWorkspace(ctx, localId); err != nil {
		return nil, err
	}
	registry, err := l.findWorkspaceRegistry.Find(ctx)
	if err != nil {
		return nil, err
	}
	if err := l.validateWorkspaceFolder.ValidateExisting(ctx, path, localId, *registry); err != nil {
		return nil, err
	}
	workspace, err := l.loadWorkspace.Load(ctx, path)
	if err != nil {
		return nil, err
	}
	located, _ := registry.Relocate(localId, path, workspace.Name)
	if err := l.saveWorkspaceRegistry.Save(ctx, *registry); err != nil {
		return nil, err
	}

	logger.Info(ctx, "Finished", located)
	return located, nil
}
