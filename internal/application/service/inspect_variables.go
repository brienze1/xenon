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
	generateDynamicValue  usecase.GenerateDynamicValue
	variables             requestVariables
}

func NewInspectVariablesService(
	findWorkspaceRegistry usecase.FindWorkspaceRegistry,
	findTreeItem usecase.FindTreeItem,
	loadWorkspace usecase.LoadWorkspace,
	findEnvironment usecase.FindEnvironment,
	findLocalVariables usecase.FindLocalVariables,
	findVault usecase.FindVault,
	generateDynamicValue usecase.GenerateDynamicValue,
) adapter.InspectVariablesService {
	return &inspectVariables{
		findWorkspaceRegistry: findWorkspaceRegistry,
		generateDynamicValue:  generateDynamicValue,
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

	scopes, err := i.load(ctx, localId, request.ID)
	if err != nil {
		return nil, err
	}
	inspection := usecase.InspectRequestVariables(request, scopes.ancestors, scopes.context, i.generateDynamicValue.Names(ctx))

	logger.Info(ctx, "Finished", map[string]any{"localId": localId, "request": request, "result": inspection})
	return &inspection, nil
}

func (i *inspectVariables) Suggest(ctx context.Context, localId, requestId, prefix string) ([]entity.VariableSuggestion, error) {
	logger.Info(ctx, "Started", map[string]any{"localId": localId, "requestId": requestId, "prefix": prefix})

	scopes, err := i.load(ctx, localId, requestId)
	if err != nil {
		return nil, err
	}
	suggestions := usecase.SuggestVariables(scopes.context, i.generateDynamicValue.Names(ctx), prefix)

	logger.Info(ctx, "Finished", map[string]any{"localId": localId, "requestId": requestId, "prefix": prefix, "result": len(suggestions)})
	return suggestions, nil
}

func (i *inspectVariables) load(ctx context.Context, localId, requestId string) (*requestScopes, error) {
	workspace, err := i.findWorkspaceRegistry.FindWorkspace(ctx, localId)
	if err != nil {
		return nil, err
	}
	return i.variables.load(ctx, localId, workspace.Path, requestId)
}
