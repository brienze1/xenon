package service

import (
	"context"

	"github.com/brienze1/xenon/internal/application/adapter"
	"github.com/brienze1/xenon/internal/application/usecase"
	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/pkg/logger"
)

type itemSettings struct {
	findWorkspaceRegistry usecase.FindWorkspaceRegistry
	findTreeItem          usecase.FindTreeItem
	saveTreeItem          usecase.SaveTreeItem
}

func NewItemSettingsService(
	findWorkspaceRegistry usecase.FindWorkspaceRegistry,
	findTreeItem usecase.FindTreeItem,
	saveTreeItem usecase.SaveTreeItem,
) adapter.ItemSettingsService {
	return &itemSettings{findWorkspaceRegistry: findWorkspaceRegistry, findTreeItem: findTreeItem, saveTreeItem: saveTreeItem}
}

func (i *itemSettings) Get(ctx context.Context, localId, id string) (*entity.Collection, error) {
	logger.Info(ctx, "Started", map[string]any{"localId": localId, "id": id})

	workspace, err := i.findWorkspaceRegistry.FindWorkspace(ctx, localId)
	if err != nil {
		return nil, err
	}
	container, err := i.findTreeItem.FindContainer(ctx, workspace.Path, id)
	if err != nil {
		return nil, err
	}

	logger.Info(ctx, "Finished", map[string]any{"localId": localId, "id": id, "kind": container.Kind})
	return container, nil
}

func (i *itemSettings) Save(ctx context.Context, localId string, settings entity.Collection) (*entity.Collection, error) {
	logger.Info(ctx, "Started", map[string]any{"localId": localId, "id": settings.ID})

	workspace, err := i.findWorkspaceRegistry.FindWorkspace(ctx, localId)
	if err != nil {
		return nil, err
	}
	container, err := i.findTreeItem.FindContainer(ctx, workspace.Path, settings.ID)
	if err != nil {
		return nil, err
	}
	saved, err := i.saveTreeItem.SaveContainer(ctx, workspace.Path, container.WithSettings(settings))
	if err != nil {
		return nil, err
	}

	logger.Info(ctx, "Finished", map[string]any{"localId": localId, "id": saved.ID})
	return saved, nil
}
