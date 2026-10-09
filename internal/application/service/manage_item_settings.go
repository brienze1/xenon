package service

import (
	"context"

	"github.com/brienze1/xenon/internal/application/adapter"
	"github.com/brienze1/xenon/internal/application/usecase"
	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/pkg/logger"
)

type manageItemSettings struct {
	findWorkspaceRegistry usecase.FindWorkspaceRegistry
	findTreeItem          usecase.FindTreeItem
	saveTreeItem          usecase.SaveTreeItem
}

func NewManageItemSettingsService(
	findWorkspaceRegistry usecase.FindWorkspaceRegistry,
	findTreeItem usecase.FindTreeItem,
	saveTreeItem usecase.SaveTreeItem,
) adapter.ManageItemSettingsService {
	return &manageItemSettings{findWorkspaceRegistry: findWorkspaceRegistry, findTreeItem: findTreeItem, saveTreeItem: saveTreeItem}
}

func (i *manageItemSettings) Get(ctx context.Context, localId, id string) (*entity.Collection, error) {
	logger.Info(ctx, "Started", map[string]any{"localId": localId, "id": id})

	workspace, err := i.findWorkspaceRegistry.FindWorkspace(ctx, localId)
	if err != nil {
		return nil, err
	}
	container, err := i.findTreeItem.FindContainer(ctx, workspace.Path, id)
	if err != nil {
		return nil, err
	}

	logger.Info(ctx, "Finished", map[string]any{"localId": localId, "id": id, "result": *container})
	return container, nil
}

func (i *manageItemSettings) Save(ctx context.Context, localId string, settings entity.Collection) (*entity.Collection, error) {
	logger.Info(ctx, "Started", map[string]any{"localId": localId, "settings": settings})

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

	logger.Info(ctx, "Finished", map[string]any{"localId": localId, "settings": settings, "result": *saved})
	return saved, nil
}
