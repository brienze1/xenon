package service

import (
	"github.com/brienze1/xenon/internal/application/usecase"
	"github.com/brienze1/xenon/internal/domain/entity"
)

type importTarget struct {
	collection   *entity.Collection
	globals      *entity.Workspace
	workspace    entity.RegisteredWorkspace
	planned      entity.ImportPlanItem
	parentPath   string
	item         entity.ImportItem
	environments []entity.Environment
	siblings     entity.TreeNode
	tree         entity.Tree
	placement    entity.ImportPlacement
}

func (t importTarget) clashes() []entity.Clash {
	if t.item.IsVariables() {
		return t.variableClashes()
	}
	clashes := usecase.FindImportClashes(t.item.Key, t.placement.Nodes, t.siblings.Children)
	if t.collection != nil {
		clashes = append(clashes, usecase.FindVariableClashes(t.item.Key, t.item.Root.Container.Variables, t.collection.Variables, t.collection.ID)...)
	}
	return clashes
}

func (t importTarget) variableClashes() []entity.Clash {
	if t.item.IsGlobals() {
		return usecase.FindVariableClashes(t.item.Key, t.item.Variables.Rows(), t.globals.Variables, t.globals.ID)
	}
	return usecase.FindEnvironmentClashes(t.item.Key, t.item.Name, t.environments)
}

func (t importTarget) collectionPath() string {
	if t.collection == nil {
		return ""
	}
	return t.tree.PathOf(t.collection.ID)
}
