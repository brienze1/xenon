package service

import (
	"context"

	"github.com/brienze1/xenon/internal/application/adapter"
	"github.com/brienze1/xenon/internal/application/usecase"
	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/pkg/logger"
)

type reorderEnvironment struct {
	findWorkspaceRegistry usecase.FindWorkspaceRegistry
	findEnvironment       usecase.FindEnvironment
	findLocalVariables    usecase.FindLocalVariables
	orders                environmentOrders
	publishVariableEvent  usecase.PublishVariableEvent
}

func NewReorderEnvironmentService(
	findWorkspaceRegistry usecase.FindWorkspaceRegistry,
	listEnvironments usecase.ListEnvironments,
	findEnvironment usecase.FindEnvironment,
	loadWorkspace usecase.LoadWorkspace,
	saveWorkspaceFile usecase.SaveWorkspaceFile,
	findLocalVariables usecase.FindLocalVariables,
	publishVariableEvent usecase.PublishVariableEvent,
) adapter.ReorderEnvironmentService {
	return &reorderEnvironment{
		findWorkspaceRegistry: findWorkspaceRegistry,
		findEnvironment:       findEnvironment,
		findLocalVariables:    findLocalVariables,
		orders: environmentOrders{
			listEnvironments:  listEnvironments,
			loadWorkspace:     loadWorkspace,
			saveWorkspaceFile: saveWorkspaceFile,
		},
		publishVariableEvent: publishVariableEvent,
	}
}

func (r *reorderEnvironment) Reorder(ctx context.Context, localId, id string, index int) (*entity.EnvironmentList, error) {
	logger.Info(ctx, "Started", map[string]any{"localId": localId, "id": id, "index": index})

	workspace, err := r.findWorkspaceRegistry.FindWorkspace(ctx, localId)
	if err != nil {
		return nil, err
	}
	if _, err := r.findEnvironment.Find(ctx, workspace.Path, id); err != nil {
		return nil, err
	}
	list, err := r.orders.list(ctx, workspace.Path, "")
	if err != nil {
		return nil, err
	}
	order := list.OrderMoving(id, index)
	if !list.IsOrderedAs(order) {
		if err := r.orders.write(ctx, workspace.Path, order); err != nil {
			return nil, err
		}
	}
	if err := r.publishVariableEvent.VariablesChanged(ctx, localId, []entity.ScopeRef{entity.NewEnvironmentScopeRef(id)}, ""); err != nil {
		return nil, err
	}
	local, err := r.findLocalVariables.Find(ctx, localId)
	if local, err = localVariablesOrDefaults(ctx, localId, local, err); err != nil {
		return nil, err
	}
	reordered, err := r.orders.list(ctx, workspace.Path, local.ActiveEnvironmentID)
	if err != nil {
		return nil, err
	}

	logger.Info(ctx, "Finished", map[string]any{"localId": localId, "id": id, "index": index, "result": *reordered})
	return reordered, nil
}
