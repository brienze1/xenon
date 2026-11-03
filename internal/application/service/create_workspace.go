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
	findWorkspaceRegistry  usecase.FindWorkspaceRegistry
	saveWorkspaceRegistry  usecase.SaveWorkspaceRegistry
	inspectWorkspaceFolder usecase.InspectWorkspaceFolder
	saveWorkspaceFile      usecase.SaveWorkspaceFile
}

func NewCreateWorkspaceService(
	findWorkspaceRegistry usecase.FindWorkspaceRegistry,
	saveWorkspaceRegistry usecase.SaveWorkspaceRegistry,
	inspectWorkspaceFolder usecase.InspectWorkspaceFolder,
	saveWorkspaceFile usecase.SaveWorkspaceFile,
) adapter.CreateWorkspaceService {
	return &createWorkspace{
		findWorkspaceRegistry:  findWorkspaceRegistry,
		saveWorkspaceRegistry:  saveWorkspaceRegistry,
		inspectWorkspaceFolder: inspectWorkspaceFolder,
		saveWorkspaceFile:      saveWorkspaceFile,
	}
}

func (c *createWorkspace) Create(ctx context.Context, name, path string) (*entity.RegisteredWorkspace, error) {
	logger.Info(ctx, "Started", map[string]any{"name": name, "path": path})

	registered, err := c.create(ctx, name, path, (*entity.AppRegistry).Register)
	if err != nil {
		return nil, err
	}

	logger.Info(ctx, "Finished", map[string]any{"name": name, "path": path, "result": registered})
	return registered, nil
}

func (c *createWorkspace) CreateInBackground(ctx context.Context, name, path string) (*entity.RegisteredWorkspace, error) {
	logger.Info(ctx, "Started", map[string]any{"name": name, "path": path})

	registered, err := c.create(ctx, name, path, (*entity.AppRegistry).Add)
	if err != nil {
		return nil, err
	}

	logger.Info(ctx, "Finished", map[string]any{"name": name, "path": path, "result": registered})
	return registered, nil
}

func (c *createWorkspace) create(ctx context.Context, name, path string, register func(*entity.AppRegistry, entity.RegisteredWorkspace)) (*entity.RegisteredWorkspace, error) {
	registry, err := c.findWorkspaceRegistry.Find(ctx)
	if err != nil {
		return nil, err
	}
	folder, err := c.inspectWorkspaceFolder.Inspect(ctx, path)
	if err != nil {
		return nil, err
	}
	if err := usecase.ValidateNewWorkspaceFolder(*folder, *registry); err != nil {
		return nil, err
	}
	workspace := entity.NewWorkspace(idutils.NewUlid(), name)
	if err := c.saveWorkspaceFile.Create(ctx, path, workspace); err != nil {
		return nil, err
	}
	registered := entity.NewRegisteredWorkspace(idutils.NewUlid(), path, workspace.Name, timeutils.Now())
	register(registry, registered)
	if err := c.saveWorkspaceRegistry.Save(ctx, *registry); err != nil {
		return nil, err
	}
	return &registered, nil
}
