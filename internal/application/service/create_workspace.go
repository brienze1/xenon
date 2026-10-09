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

type createWorkspace struct {
	findWorkspaceRegistry   usecase.FindWorkspaceRegistry
	saveWorkspaceRegistry   usecase.SaveWorkspaceRegistry
	validateWorkspaceFolder usecase.ValidateWorkspaceFolder
	saveWorkspaceFile       usecase.SaveWorkspaceFile
}

func NewCreateWorkspaceService(
	findWorkspaceRegistry usecase.FindWorkspaceRegistry,
	saveWorkspaceRegistry usecase.SaveWorkspaceRegistry,
	validateWorkspaceFolder usecase.ValidateWorkspaceFolder,
	saveWorkspaceFile usecase.SaveWorkspaceFile,
) adapter.CreateWorkspaceService {
	return &createWorkspace{
		findWorkspaceRegistry:   findWorkspaceRegistry,
		saveWorkspaceRegistry:   saveWorkspaceRegistry,
		validateWorkspaceFolder: validateWorkspaceFolder,
		saveWorkspaceFile:       saveWorkspaceFile,
	}
}

func (c *createWorkspace) Create(ctx context.Context, name, path string) (*entity.RegisteredWorkspace, error) {
	logger.Info(ctx, "Started", map[string]any{"name": name, "path": path})

	registry, err := c.findWorkspaceRegistry.Find(ctx)
	if err != nil {
		return nil, err
	}
	if err := c.validateWorkspaceFolder.ValidateNew(ctx, path, *registry); err != nil {
		return nil, err
	}
	workspace := entity.NewWorkspace(idutils.NewUlid(), name)
	if err := c.saveWorkspaceFile.Create(ctx, path, workspace); err != nil {
		return nil, err
	}
	registered := entity.NewRegisteredWorkspace(idutils.NewUlid(), path, workspace.Name, timeutils.Now())
	registry.Register(registered)
	if err := c.saveWorkspaceRegistry.Save(ctx, *registry); err != nil {
		return nil, err
	}

	logger.Info(ctx, "Finished", map[string]any{"name": name, "path": path, "result": registered})
	return &registered, nil
}
