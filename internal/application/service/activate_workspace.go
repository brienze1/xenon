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
}

func NewActivateWorkspaceService(
	findWorkspaceRegistry usecase.FindWorkspaceRegistry,
	saveWorkspaceRegistry usecase.SaveWorkspaceRegistry,
	loadWorkspace usecase.LoadWorkspace,
) adapter.ActivateWorkspaceService {
	return &activateWorkspace{
		findWorkspaceRegistry: findWorkspaceRegistry,
		saveWorkspaceRegistry: saveWorkspaceRegistry,
		loadWorkspace:         loadWorkspace,
	}
}

func (a *activateWorkspace) Activate(ctx context.Context, localId string) (*entity.RegisteredWorkspace, error) {
	logger.Info(ctx, "Started", map[string]any{"localId": localId})

	registered, err := a.findWorkspaceRegistry.FindWorkspace(ctx, localId)
	if err != nil {
		return nil, err
	}
	workspace, err := a.loadWorkspace.Load(ctx, registered.Path)
	if err != nil {
		return nil, err
	}
	registry, err := a.findWorkspaceRegistry.Find(ctx)
	if err != nil {
		return nil, err
	}
	activated, _ := registry.Activate(localId, workspace.Name)
	if err := a.saveWorkspaceRegistry.Save(ctx, *registry); err != nil {
		return nil, err
	}

	logger.Info(ctx, "Finished", activated)
	return activated, nil
}
