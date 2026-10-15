package service

import (
	"context"

	"github.com/brienze1/xenon/internal/application/usecase"
	"github.com/brienze1/xenon/internal/domain/entity"
)

type environmentOrders struct {
	listEnvironments  usecase.ListEnvironments
	loadWorkspace     usecase.LoadWorkspace
	saveWorkspaceFile usecase.SaveWorkspaceFile
}

func (e environmentOrders) list(ctx context.Context, root, activeId string) (*entity.EnvironmentList, error) {
	environments, err := e.listEnvironments.List(ctx, root)
	if err != nil {
		return nil, err
	}
	workspace, err := e.loadWorkspace.Load(ctx, root)
	if err != nil {
		return nil, err
	}
	list := entity.NewEnvironmentList(environments, workspace.EnvironmentOrder, activeId)
	return &list, nil
}

func (e environmentOrders) lists(ctx context.Context, root, id string) (bool, error) {
	workspace, err := e.loadWorkspace.Load(ctx, root)
	if err != nil {
		return false, err
	}
	return workspace.ListsEnvironment(id), nil
}

func (e environmentOrders) write(ctx context.Context, root string, order []string) error {
	workspace, err := e.loadWorkspace.Load(ctx, root)
	if err != nil {
		return err
	}
	return e.saveWorkspaceFile.Save(ctx, root, workspace.WithEnvironmentOrder(order))
}
