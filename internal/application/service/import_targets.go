package service

import (
	"context"

	"github.com/brienze1/xenon/internal/application/usecase"
	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/pkg/errs"
)

type importTargets struct {
	findWorkspaceRegistry usecase.FindWorkspaceRegistry
	loadTree              usecase.LoadTree
	findTreeItem          usecase.FindTreeItem
}

type importTarget struct {
	collection *entity.Collection
	workspace  entity.RegisteredWorkspace
	planned    entity.ImportPlanItem
	parentPath string
	item       entity.ImportItem
	siblings   entity.TreeNode
	tree       entity.Tree
	placement  entity.ImportPlacement
}

func (i importTargets) resolve(ctx context.Context, source entity.ImportSource, plan entity.ImportPlan) ([]importTarget, error) {
	targets := make([]importTarget, 0, len(plan.Items))
	for _, planned := range plan.Included() {
		target, err := i.target(ctx, source, planned)
		if err != nil {
			return nil, err
		}
		targets = append(targets, *target)
	}
	return targets, nil
}

func (i importTargets) target(ctx context.Context, source entity.ImportSource, planned entity.ImportPlanItem) (*importTarget, error) {
	item, found := source.Item(planned.Key)
	if !found {
		return nil, errs.NotFoundError("the import source has no item "+planned.Key, "IMP-16404")
	}
	workspace, err := i.findWorkspaceRegistry.FindWorkspace(ctx, planned.TargetLocalID)
	if err != nil {
		return nil, err
	}
	tree, err := i.loadTree.Load(ctx, workspace.Path)
	if err != nil {
		return nil, err
	}
	target := importTarget{workspace: *workspace, tree: *tree, item: item, planned: planned, siblings: tree.Root, placement: entity.NewTopLevelPlacement(item.Root)}
	if !planned.IsIntoParent() {
		return &target, nil
	}
	return i.intoParent(ctx, target)
}

func (i importTargets) intoParent(ctx context.Context, target importTarget) (*importTarget, error) {
	parentId := target.planned.TargetParentID
	parentNode, err := usecase.ResolveParent(target.tree, parentId)
	if err != nil {
		return nil, err
	}
	parent, err := i.findTreeItem.FindContainer(ctx, target.workspace.Path, parentId)
	if err != nil {
		return nil, err
	}
	ancestors, err := i.findTreeItem.FindAncestors(ctx, target.workspace.Path, parentId)
	if err != nil {
		return nil, err
	}
	collection, _ := entity.RootCollection(append(ancestors, *parent))
	target.siblings = *parentNode
	target.placement = target.item.Root.PlacedInto(*parent)
	target.collection = &collection
	target.parentPath = target.tree.PathOf(parentId)
	return &target, nil
}

func (t importTarget) clashes() []entity.Clash {
	clashes := usecase.FindImportClashes(t.item.Key, t.placement.Nodes, t.siblings.Children)
	if t.collection != nil {
		clashes = append(clashes, usecase.FindVariableClashes(t.item.Key, t.item.Root.Container.Variables, *t.collection)...)
	}
	return clashes
}
