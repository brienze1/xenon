package service

import (
	"context"

	"github.com/brienze1/xenon/internal/application/usecase"
	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/internal/domain/enums"
	"github.com/brienze1/xenon/internal/domain/enums/exportitemkind"
)

const globalsExportName = "globals"

type exportItems struct {
	loadTree           usecase.LoadTree
	findTreeItem       usecase.FindTreeItem
	findRequest        usecase.FindRequest
	listEnvironments   usecase.ListEnvironments
	findEnvironment    usecase.FindEnvironment
	loadWorkspace      usecase.LoadWorkspace
	findLocalVariables usecase.FindLocalVariables
}

func (e exportItems) item(ctx context.Context, root, id string) (entity.ExportNode, []entity.Collection, error) {
	found, err := e.findTreeItem.FindNode(ctx, root, id)
	if err != nil {
		return entity.ExportNode{}, nil, err
	}
	ancestors, err := e.findTreeItem.FindAncestors(ctx, root, id)
	if err != nil {
		return entity.ExportNode{}, nil, err
	}
	path := ""
	for _, ancestor := range ancestors {
		path = entity.JoinImportPath(path, ancestor.Name)
	}
	node, err := e.node(ctx, root, *found, entity.JoinImportPath(path, found.Name))
	return node, ancestors, err
}

func (e exportItems) collections(ctx context.Context, root string) ([]entity.ExportNode, error) {
	tree, err := e.loadTree.Load(ctx, root)
	if err != nil {
		return nil, err
	}
	nodes := make([]entity.ExportNode, 0, len(tree.Collections()))
	for _, collection := range tree.Collections() {
		if !collection.IsItem() {
			continue
		}
		node, err := e.node(ctx, root, collection, collection.Name)
		if err != nil {
			return nil, err
		}
		nodes = append(nodes, node)
	}
	return nodes, nil
}

func (e exportItems) node(ctx context.Context, root string, found entity.TreeNode, path string) (entity.ExportNode, error) {
	if found.IsRequest() {
		request, err := e.findRequest.Find(ctx, root, found.ID)
		if err != nil {
			return entity.ExportNode{}, err
		}
		return entity.NewRequestExportNode(*request, path), nil
	}
	container, err := e.findTreeItem.FindContainer(ctx, root, found.ID)
	if err != nil {
		return entity.ExportNode{}, err
	}
	children := make([]entity.ExportNode, 0, len(found.Children))
	for _, child := range found.Children {
		if !child.IsItem() || child.Ignored {
			continue
		}
		node, err := e.node(ctx, root, child, entity.JoinImportPath(path, child.Name))
		if err != nil {
			return entity.ExportNode{}, err
		}
		children = append(children, node)
	}
	return entity.NewContainerExportNode(*container, path, children), nil
}

func (e exportItems) local(ctx context.Context, localId string) (*entity.LocalVariables, error) {
	local, err := e.findLocalVariables.Find(ctx, localId)
	return localVariablesOrDefaults(ctx, localId, local, err)
}

func (e exportItems) scopes(ctx context.Context, root string, kind enums.ExportItemKind, id string) (string, []entity.ExportScope, error) {
	switch kind {
	case exportitemkind.Environment:
		environment, err := e.findEnvironment.Find(ctx, root, id)
		if err != nil {
			return "", nil, err
		}
		return environment.Name, []entity.ExportScope{entity.NewExportScope(environment.ScopeRef(), environment.Variables)}, nil
	case exportitemkind.Globals:
		workspace, err := e.loadWorkspace.Load(ctx, root)
		if err != nil {
			return "", nil, err
		}
		return globalsExportName, []entity.ExportScope{entity.NewExportScope(entity.NewGlobalsScopeRef(), workspace.Variables)}, nil
	case exportitemkind.Workspace:
		return e.workspaceScopes(ctx, root)
	default:
		return e.itemScopes(ctx, root, id)
	}
}

func (e exportItems) itemScopes(ctx context.Context, root, id string) (string, []entity.ExportScope, error) {
	found, err := e.findTreeItem.FindNode(ctx, root, id)
	if err != nil {
		return "", nil, err
	}
	ancestors, err := e.findTreeItem.FindAncestors(ctx, root, id)
	if err != nil {
		return "", nil, err
	}
	if found.IsCollection() {
		collection, err := e.findTreeItem.FindContainer(ctx, root, id)
		if err != nil {
			return "", nil, err
		}
		ancestors = []entity.Collection{*collection}
	}
	scopes := make([]entity.ExportScope, 0, 1)
	if collection, found := entity.RootCollection(ancestors); found {
		scopes = append(scopes, entity.NewExportScope(collection.ScopeRef(), collection.OwnVariables()))
	}
	return found.Name, scopes, nil
}

func (e exportItems) workspaceScopes(ctx context.Context, root string) (string, []entity.ExportScope, error) {
	workspace, err := e.loadWorkspace.Load(ctx, root)
	if err != nil {
		return "", nil, err
	}
	scopes := []entity.ExportScope{entity.NewExportScope(entity.NewGlobalsScopeRef(), workspace.Variables)}
	environments, err := e.listEnvironments.List(ctx, root)
	if err != nil {
		return "", nil, err
	}
	for _, environment := range environments {
		if !environment.IsMalformed() {
			scopes = append(scopes, entity.NewExportScope(environment.ScopeRef(), environment.Variables))
		}
	}
	collections, err := e.collections(ctx, root)
	if err != nil {
		return "", nil, err
	}
	for _, collection := range collections {
		scopes = append(scopes, entity.NewExportScope(collection.Container.ScopeRef(), collection.Container.OwnVariables()))
	}
	return workspace.Name, scopes, nil
}
