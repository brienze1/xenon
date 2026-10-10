package service

import (
	"context"

	"github.com/brienze1/xenon/internal/application/adapter"
	"github.com/brienze1/xenon/internal/application/usecase"
	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/pkg/logger"
)

type inspectVariables struct {
	findWorkspaceRegistry usecase.FindWorkspaceRegistry
	listDynamicValueNames usecase.ListDynamicValueNames
	variables             requestVariables
}

func NewInspectVariablesService(
	findWorkspaceRegistry usecase.FindWorkspaceRegistry,
	findTreeItem usecase.FindTreeItem,
	loadWorkspace usecase.LoadWorkspace,
	findEnvironment usecase.FindEnvironment,
	findLocalVariables usecase.FindLocalVariables,
	findVault usecase.FindVault,
	listDynamicValueNames usecase.ListDynamicValueNames,
) adapter.InspectVariablesService {
	return &inspectVariables{
		findWorkspaceRegistry: findWorkspaceRegistry,
		listDynamicValueNames: listDynamicValueNames,
		variables: requestVariables{
			findTreeItem:       findTreeItem,
			loadWorkspace:      loadWorkspace,
			findEnvironment:    findEnvironment,
			findLocalVariables: findLocalVariables,
			findVault:          findVault,
		},
	}
}

func (i *inspectVariables) Inspect(ctx context.Context, localId string, request entity.Request) (*entity.VariableInspection, error) {
	logger.Info(ctx, "Started", map[string]any{"localId": localId, "request": request})

	workspace, err := i.findWorkspaceRegistry.FindWorkspace(ctx, localId)
	if err != nil {
		return nil, err
	}
	scopes, err := i.variables.load(ctx, localId, workspace.Path, request.ID)
	if err != nil {
		return nil, err
	}
	inspection := usecase.InspectRequestVariables(request, scopes.ancestors, scopes.context, i.listDynamicValueNames.Names(ctx))

	logger.Info(ctx, "Finished", map[string]any{"localId": localId, "request": request, "result": inspection})
	return &inspection, nil
}
