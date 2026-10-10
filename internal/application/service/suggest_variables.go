package service

import (
	"context"

	"github.com/brienze1/xenon/internal/application/adapter"
	"github.com/brienze1/xenon/internal/application/usecase"
	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/pkg/logger"
)

type suggestVariables struct {
	findWorkspaceRegistry usecase.FindWorkspaceRegistry
	listDynamicValueNames usecase.ListDynamicValueNames
	variables             requestVariables
}

func NewSuggestVariablesService(
	findWorkspaceRegistry usecase.FindWorkspaceRegistry,
	findTreeItem usecase.FindTreeItem,
	loadWorkspace usecase.LoadWorkspace,
	findEnvironment usecase.FindEnvironment,
	findLocalVariables usecase.FindLocalVariables,
	findVault usecase.FindVault,
	listDynamicValueNames usecase.ListDynamicValueNames,
) adapter.SuggestVariablesService {
	return &suggestVariables{
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

func (s *suggestVariables) Suggest(ctx context.Context, localId, requestId, prefix string) ([]entity.VariableSuggestion, error) {
	logger.Info(ctx, "Started", map[string]any{"localId": localId, "requestId": requestId, "prefix": prefix})

	workspace, err := s.findWorkspaceRegistry.FindWorkspace(ctx, localId)
	if err != nil {
		return nil, err
	}
	scopes, err := s.variables.load(ctx, localId, workspace.Path, requestId)
	if err != nil {
		return nil, err
	}
	suggestions := usecase.SuggestVariables(scopes.context, s.listDynamicValueNames.Names(ctx), prefix)

	logger.Info(ctx, "Finished", map[string]any{"localId": localId, "requestId": requestId, "prefix": prefix, "result": suggestions})
	return suggestions, nil
}
