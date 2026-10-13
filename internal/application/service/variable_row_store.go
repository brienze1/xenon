package service

import (
	"context"

	"github.com/brienze1/xenon/internal/application/usecase"
	"github.com/brienze1/xenon/internal/domain/entity"
)

type variableRowStore struct {
	scopes            variableScopes
	saveEnvironment   usecase.SaveEnvironment
	saveTreeItem      usecase.SaveTreeItem
	saveWorkspaceFile usecase.SaveWorkspaceFile
}

func (v variableRowStore) save(ctx context.Context, root string, scope entity.ScopeRef, rows []entity.VariableRow) error {
	switch {
	case scope.IsEnvironment():
		environment, err := v.scopes.findEnvironment.Find(ctx, root, scope.ID)
		if err != nil {
			return err
		}
		_, err = v.saveEnvironment.Save(ctx, root, environment.WithVariables(rows))
		return err
	case scope.IsCollection():
		collection, err := v.scopes.collection(ctx, root, scope)
		if err != nil {
			return err
		}
		node, err := v.scopes.findTreeItem.FindNode(ctx, root, scope.ID)
		if err != nil {
			return err
		}
		_, err = v.saveTreeItem.SaveContainer(ctx, root, collection.WithVariables(rows).WithOrder(node.ChildOrder()))
		return err
	default:
		workspace, err := v.scopes.loadWorkspace.Load(ctx, root)
		if err != nil {
			return err
		}
		return v.saveWorkspaceFile.Save(ctx, root, workspace.WithVariables(rows))
	}
}
