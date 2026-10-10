package service

import (
	"context"

	"github.com/brienze1/xenon/internal/application/usecase"
	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/internal/domain/enums"
)

type importWriter struct {
	stageImportItem       usecase.StageImportItem
	commitImportItem      usecase.CommitImportItem
	rollbackImportItem    usecase.RollbackImportItem
	publishImportProgress usecase.PublishImportProgress
	publishWorkspaceEvent usecase.PublishWorkspaceEvent
}

func (w importWriter) write(ctx context.Context, sourceId string, plan entity.ImportPlan, targets []importTarget) entity.ImportResult {
	result := entity.NewImportResult()
	for index, target := range targets {
		_ = w.publishImportProgress.Progress(ctx, entity.ImportProgress{SourceID: sourceId, Done: index, Total: len(targets), CurrentItem: target.item.Name})
		result = result.WithOutcome(w.importOne(ctx, sourceId, target, plan.ResolutionFor(target.planned)))
	}
	_ = w.publishImportProgress.Progress(ctx, entity.ImportProgress{SourceID: sourceId, Done: len(targets), Total: len(targets)})
	for _, localId := range result.LocalIDs() {
		_ = w.publishWorkspaceEvent.TreeChanged(ctx, localId)
	}
	return result
}

func (w importWriter) importOne(ctx context.Context, sourceId string, target importTarget, resolution enums.ClashResolution) entity.ImportOutcome {
	nodes, replaced := usecase.ResolveImportClashes(target.placement.Nodes, target.siblings.Children, resolution)
	write := usecase.PlanImportWrite(target.placement.WithNodes(nodes), target.siblings, replaced, target.collection, target.item.Root.Container.Variables, resolution)
	if write.IsEmpty() {
		return target.item.Skipped(target.workspace.LocalID)
	}
	staged := write.Staged(sourceId, target.item.Key, target.workspace.Path, target.planned.TargetParentID)
	written, err := w.stageImportItem.Stage(ctx, staged)
	if err != nil {
		return usecase.FailedImport(target.item, target.workspace.LocalID, err)
	}
	if err := w.commitImportItem.Commit(ctx, *written); err != nil {
		_ = w.rollbackImportItem.Rollback(ctx, *written)
		return usecase.FailedImport(target.item, target.workspace.LocalID, err)
	}
	return target.item.Completed(write, entity.ImportLocation{LocalID: target.workspace.LocalID, ParentPath: target.parentPath, CollectionPath: target.collectionPath()})
}

func (t importTarget) collectionPath() string {
	if t.collection == nil {
		return ""
	}
	return t.tree.PathOf(t.collection.ID)
}
