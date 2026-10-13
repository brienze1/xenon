package mcp

import (
	"context"
	"strings"

	"github.com/brienze1/xenon/internal/domain/enums/itemkind"
	"github.com/brienze1/xenon/internal/integration/entrypoint/dto"
)

const itemPathSeparator = " / "

type workspaceItem struct {
	Id        string
	Kind      string
	Name      string
	Path      string
	ParentId  string
	RequestId string
	Children  int
}

func (w workspaceItem) isExample() bool {
	return w.Kind == itemkind.Example.String()
}

func (w workspaceItem) isRequest() bool {
	return w.Kind == itemkind.Request.String()
}

func (w workspaceItem) isContainer() bool {
	return w.Kind == itemkind.Collection.String() || w.Kind == itemkind.Folder.String()
}

func (w workspaceItem) target() dto.AIActivityTarget {
	return dto.AIActivityTarget{Id: w.Id, Name: w.Name, Path: w.Path}
}

// findWorkspaceItem answers MCP-08404 for an id that is not in the active workspace's tree.
func (t *tools) findWorkspaceItem(ctx context.Context, localId, id string) (workspaceItem, error) {
	tree, err := t.controllers.TreeSearch.GetTree(ctx, dto.GetTree{LocalId: localId})
	if err != nil {
		return workspaceItem{}, err
	}
	if item, found := searchWorkspaceItem(tree, nil, "", id); found {
		return item, nil
	}
	return workspaceItem{}, itemNotFound(id)
}

func searchWorkspaceItem(nodes []dto.TreeNode, ancestors []string, parentId, id string) (workspaceItem, bool) {
	for _, node := range nodes {
		names := append(append([]string{}, ancestors...), node.Name)
		if node.Id == id {
			return workspaceItem{Id: node.Id, Kind: node.Kind, Name: node.Name, Path: strings.Join(names, itemPathSeparator), ParentId: parentId, Children: len(node.Children) + len(node.Examples)}, true
		}
		for _, example := range node.Examples {
			if example.Id == id {
				path := strings.Join(append(names, example.Name), itemPathSeparator)
				return workspaceItem{Id: example.Id, Kind: example.Kind, Name: example.Name, Path: path, ParentId: node.Id, RequestId: node.Id}, true
			}
		}
		if item, found := searchWorkspaceItem(node.Children, names, node.Id, id); found {
			return item, true
		}
	}
	return workspaceItem{}, false
}
