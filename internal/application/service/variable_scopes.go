package service

import (
	"context"

	"github.com/brienze1/xenon/internal/application/usecase"
	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/pkg/errs"
)

type variableScopes struct {
	findEnvironment usecase.FindEnvironment
	findTreeItem    usecase.FindTreeItem
	loadWorkspace   usecase.LoadWorkspace
}

func (v variableScopes) rows(ctx context.Context, root string, scope entity.ScopeRef) ([]entity.VariableRow, error) {
	switch {
	case scope.IsEnvironment():
		environment, err := v.findEnvironment.Find(ctx, root, scope.ID)
		if err != nil {
			return nil, err
		}
		return environment.Variables, nil
	case scope.IsCollection():
		collection, err := v.collection(ctx, root, scope)
		if err != nil {
			return nil, err
		}
		return collection.Variables, nil
	default:
		workspace, err := v.loadWorkspace.Load(ctx, root)
		if err != nil {
			return nil, err
		}
		return workspace.Variables, nil
	}
}

func (v variableScopes) collection(ctx context.Context, root string, scope entity.ScopeRef) (*entity.Collection, error) {
	container, err := v.findTreeItem.FindContainer(ctx, root, scope.ID)
	if err != nil {
		return nil, err
	}
	if container.IsFolder() {
		return nil, errs.UnprocessableEntityError("folders have no variables; use the collection's", "VAR-06422")
	}
	return container, nil
}
