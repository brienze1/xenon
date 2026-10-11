package service

import (
	"context"

	"github.com/brienze1/xenon/internal/application/adapter"
	"github.com/brienze1/xenon/internal/application/usecase"
	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/internal/domain/enums/itemkind"
	"github.com/brienze1/xenon/pkg/idutils"
	"github.com/brienze1/xenon/pkg/logger"
)

type manageTreeItem struct {
	findWorkspaceRegistry usecase.FindWorkspaceRegistry
	loadTree              usecase.LoadTree
	createTreeItem        usecase.CreateTreeItem
	renameTreeItem        usecase.RenameTreeItem
	duplicateTreeItem     usecase.DuplicateTreeItem
	deleteTreeItem        usecase.DeleteTreeItem
	moveTreeItem          usecase.MoveTreeItem
}

func NewManageTreeItemService(
	findWorkspaceRegistry usecase.FindWorkspaceRegistry,
	loadTree usecase.LoadTree,
	createTreeItem usecase.CreateTreeItem,
	renameTreeItem usecase.RenameTreeItem,
	duplicateTreeItem usecase.DuplicateTreeItem,
	deleteTreeItem usecase.DeleteTreeItem,
	moveTreeItem usecase.MoveTreeItem,
) adapter.ManageTreeItemService {
	return &manageTreeItem{
		findWorkspaceRegistry: findWorkspaceRegistry,
		loadTree:              loadTree,
		createTreeItem:        createTreeItem,
		renameTreeItem:        renameTreeItem,
		duplicateTreeItem:     duplicateTreeItem,
		deleteTreeItem:        deleteTreeItem,
		moveTreeItem:          moveTreeItem,
	}
}

func (m *manageTreeItem) CreateCollection(ctx context.Context, localId, name string) (*entity.TreeNode, error) {
	logger.Info(ctx, "Started", map[string]any{"localId": localId, "name": name})

	root, tree, err := m.load(ctx, localId)
	if err != nil {
		return nil, err
	}
	collection := entity.NewCollection(idutils.NewUlid(), name)
	fileName := usecase.DeriveFileName(collection.Name, collection.Kind, tree.Root.Entries)
	node, err := m.createTreeItem.CreateContainer(ctx, root, "", fileName, collection)
	if err != nil {
		return nil, err
	}

	logger.Info(ctx, "Finished", map[string]any{"localId": localId, "name": name, "result": *node})
	return node, nil
}

func (m *manageTreeItem) CreateFolder(ctx context.Context, localId, parentId, name string) (*entity.TreeNode, error) {
	logger.Info(ctx, "Started", map[string]any{"localId": localId, "parentId": parentId, "name": name})

	root, tree, err := m.load(ctx, localId)
	if err != nil {
		return nil, err
	}
	parent, err := usecase.ResolveParent(*tree, parentId)
	if err != nil {
		return nil, err
	}
	folder := entity.NewFolder(idutils.NewUlid(), name)
	fileName := usecase.DeriveFileName(folder.Name, folder.Kind, parent.Entries)
	node, err := m.createTreeItem.CreateContainer(ctx, root, parent.ID, fileName, folder)
	if err != nil {
		return nil, err
	}

	logger.Info(ctx, "Finished", map[string]any{"localId": localId, "parentId": parentId, "name": name, "result": *node})
	return node, nil
}

func (m *manageTreeItem) CreateRequest(ctx context.Context, localId, parentId, name string, request entity.Request) (*entity.TreeNode, error) {
	logger.Info(ctx, "Started", map[string]any{"localId": localId, "parentId": parentId, "name": name, "request": request})

	root, tree, err := m.load(ctx, localId)
	if err != nil {
		return nil, err
	}
	parent, err := usecase.ResolveParent(*tree, parentId)
	if err != nil {
		return nil, err
	}
	created := request.CopyAs(idutils.NewUlid(), name).WithFileRefsRelativeTo(root)
	fileName := usecase.DeriveFileName(created.Name, itemkind.Request, parent.Entries)
	saved, err := m.createTreeItem.CreateRequest(ctx, root, parent.ID, fileName, created)
	if err != nil {
		return nil, err
	}
	node := entity.NewRequestNode(*saved, fileName)

	logger.Info(ctx, "Finished", map[string]any{"localId": localId, "parentId": parentId, "name": name, "request": request, "result": node})
	return &node, nil
}

func (m *manageTreeItem) Rename(ctx context.Context, localId, id, name string) (*entity.TreeNode, error) {
	logger.Info(ctx, "Started", map[string]any{"localId": localId, "id": id, "name": name})

	root, tree, err := m.load(ctx, localId)
	if err != nil {
		return nil, err
	}
	item, err := usecase.ResolveItem(*tree, id)
	if err != nil {
		return nil, err
	}
	fileName := usecase.DeriveFileName(name, item.Kind, tree.SiblingEntries(id))
	node, err := m.renameTreeItem.Rename(ctx, root, id, entity.NormalizeItemName(name), fileName)
	if err != nil {
		return nil, err
	}

	logger.Info(ctx, "Finished", map[string]any{"localId": localId, "id": id, "name": name, "result": *node})
	return node, nil
}

func (m *manageTreeItem) Duplicate(ctx context.Context, localId, id string) (*entity.TreeNode, error) {
	logger.Info(ctx, "Started", map[string]any{"localId": localId, "id": id})

	root, tree, err := m.load(ctx, localId)
	if err != nil {
		return nil, err
	}
	item, err := usecase.ResolveItem(*tree, id)
	if err != nil {
		return nil, err
	}
	parent, _ := tree.Parent(id)
	fileName := usecase.DeriveFileName(item.CopyName(), item.Kind, parent.Entries)
	node, err := m.duplicateTreeItem.Duplicate(ctx, root, id, item.CopyAs(item.CopyName(), fileName, idutils.NewUlid))
	if err != nil {
		return nil, err
	}

	logger.Info(ctx, "Finished", map[string]any{"localId": localId, "id": id, "result": *node})
	return node, nil
}

func (m *manageTreeItem) Delete(ctx context.Context, localId, id string) ([]string, error) {
	logger.Info(ctx, "Started", map[string]any{"localId": localId, "id": id})

	root, tree, err := m.load(ctx, localId)
	if err != nil {
		return nil, err
	}
	item, err := usecase.ResolveItem(*tree, id)
	if err != nil {
		return nil, err
	}
	if err := m.deleteTreeItem.Delete(ctx, root, id); err != nil {
		return nil, err
	}
	deleted := item.RequestIDs()

	logger.Info(ctx, "Finished", map[string]any{"localId": localId, "id": id, "result": deleted})
	return deleted, nil
}

func (m *manageTreeItem) Move(ctx context.Context, localId, id, parentId string, index int) (*entity.Tree, error) {
	logger.Info(ctx, "Started", map[string]any{"localId": localId, "id": id, "parentId": parentId, "index": index})

	workspace, err := m.findWorkspaceRegistry.FindWorkspace(ctx, localId)
	if err != nil {
		return nil, err
	}
	root := workspace.Path
	tree, err := m.loadTree.LoadIndexed(ctx, root)
	if err != nil {
		return nil, err
	}
	move, err := usecase.ResolveTreeMove(*tree, id, parentId, index)
	if err != nil {
		return nil, err
	}
	moved, err := m.moveTreeItem.Move(ctx, root, *move)
	if err != nil {
		return nil, err
	}

	logger.Info(ctx, "Finished", map[string]any{"localId": localId, "id": id, "parentId": parentId, "index": index, "collections": len(moved.Collections())})
	return moved, nil
}

func (m *manageTreeItem) load(ctx context.Context, localId string) (string, *entity.Tree, error) {
	workspace, err := m.findWorkspaceRegistry.FindWorkspace(ctx, localId)
	if err != nil {
		return "", nil, err
	}
	tree, err := m.loadTree.Load(ctx, workspace.Path)
	if err != nil {
		return "", nil, err
	}
	return workspace.Path, tree, nil
}
