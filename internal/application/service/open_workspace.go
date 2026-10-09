package service

import (
	"context"

	"github.com/brienze1/xenon/internal/application/adapter"
	"github.com/brienze1/xenon/internal/application/usecase"
	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/pkg/idutils"
	"github.com/brienze1/xenon/pkg/logger"
	"github.com/brienze1/xenon/pkg/timeutils"
)

type openWorkspace struct {
	findWorkspaceRegistry   usecase.FindWorkspaceRegistry
	saveWorkspaceRegistry   usecase.SaveWorkspaceRegistry
	validateWorkspaceFolder usecase.ValidateWorkspaceFolder
	loadWorkspace           usecase.LoadWorkspace
}

func NewOpenWorkspaceService(
	findWorkspaceRegistry usecase.FindWorkspaceRegistry,
	saveWorkspaceRegistry usecase.SaveWorkspaceRegistry,
	validateWorkspaceFolder usecase.ValidateWorkspaceFolder,
	loadWorkspace usecase.LoadWorkspace,
) adapter.OpenWorkspaceService {
	return &openWorkspace{
		findWorkspaceRegistry:   findWorkspaceRegistry,
		saveWorkspaceRegistry:   saveWorkspaceRegistry,
		validateWorkspaceFolder: validateWorkspaceFolder,
		loadWorkspace:           loadWorkspace,
	}
}

func (o *openWorkspace) Open(ctx context.Context, path string) (*entity.RegisteredWorkspace, error) {
	logger.Info(ctx, "Started", map[string]any{"path": path})

	registry, err := o.findWorkspaceRegistry.Find(ctx)
	if err != nil {
		return nil, err
	}
	if err := o.validateWorkspaceFolder.ValidateExisting(ctx, path, "", *registry); err != nil {
		return nil, err
	}
	workspace, err := o.loadWorkspace.Load(ctx, path)
	if err != nil {
		return nil, err
	}
	registered := entity.NewRegisteredWorkspace(idutils.NewUlid(), path, workspace.Name, timeutils.Now())
	registry.Register(registered)
	if err := o.saveWorkspaceRegistry.Save(ctx, *registry); err != nil {
		return nil, err
	}

	logger.Info(ctx, "Finished", map[string]any{"path": path, "result": registered})
	return &registered, nil
}
