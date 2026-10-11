package service

import (
	"context"

	"github.com/brienze1/xenon/internal/application/usecase"
	"github.com/brienze1/xenon/internal/domain/entity"
)

type importTargets struct {
	findWorkspaceRegistry  usecase.FindWorkspaceRegistry
	loadTree               usecase.LoadTree
	findTreeItem           usecase.FindTreeItem
	listEnvironments       usecase.ListEnvironments
	loadWorkspace          usecase.LoadWorkspace
	inspectWorkspaceFolder usecase.InspectWorkspaceFolder
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
	resolved, err := usecase.ResolveImportItem(source, planned.Key)
	if err != nil {
		return nil, err
	}
	if resolved.IsRequest() && !planned.IsIntoParent() {
		return nil, usecase.RejectTopLevelRequest()
	}
	item := resolved.Named(planned.Name)
	if item.IsWorkspace() {
		return i.workspaceTarget(ctx, item, planned)
	}
	workspace, err := i.findWorkspaceRegistry.FindWorkspace(ctx, planned.TargetLocalID)
	if err != nil {
		return nil, err
	}
	if item.IsVariables() {
		return i.variablesTarget(ctx, *workspace, item, planned)
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

func (i importTargets) workspaceTarget(ctx context.Context, item entity.ImportItem, planned entity.ImportPlanItem) (*importTarget, error) {
	if planned.TargetFolder == "" {
		return nil, usecase.RejectMissingWorkspaceFolder()
	}
	registry, err := i.findWorkspaceRegistry.Find(ctx)
	if err != nil {
		return nil, err
	}
	folder, err := i.inspectWorkspaceFolder.Inspect(ctx, planned.TargetFolder)
	if err != nil {
		return nil, err
	}
	if err := usecase.ValidateNewWorkspaceFolder(*folder, *registry); err != nil {
		return nil, err
	}
	return &importTarget{item: item, planned: planned}, nil
}

func (i importTargets) variablesTarget(ctx context.Context, workspace entity.RegisteredWorkspace, item entity.ImportItem, planned entity.ImportPlanItem) (*importTarget, error) {
	environments, err := i.listEnvironments.List(ctx, workspace.Path)
	if err != nil {
		return nil, err
	}
	globals, err := i.loadWorkspace.Load(ctx, workspace.Path)
	if err != nil {
		return nil, err
	}
	return &importTarget{workspace: workspace, item: item, planned: planned, environments: environments, globals: globals}, nil
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
