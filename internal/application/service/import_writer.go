package service

import (
	"context"
	"time"

	"github.com/brienze1/xenon/internal/application/usecase"
	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/internal/domain/enums"
)

const importHeartbeatInterval = 200 * time.Millisecond

type importWriter struct {
	stageImportItem       usecase.StageImportItem
	commitImportItem      usecase.CommitImportItem
	rollbackImportItem    usecase.RollbackImportItem
	publishImportProgress usecase.PublishImportProgress
	publishWorkspaceEvent usecase.PublishWorkspaceEvent
	variables             importVariablesWriter
}

func (w importWriter) write(ctx context.Context, sourceId string, plan entity.ImportPlan, targets []importTarget) entity.ImportResult {
	result := entity.NewImportResult()
	for index, target := range targets {
		progress := entity.ImportProgress{SourceID: sourceId, Done: index, Total: len(targets), CurrentItem: target.item.Name}
		_ = w.publishImportProgress.Progress(ctx, progress)
		stop := w.heartbeat(ctx, progress)
		result = result.WithOutcome(w.importOne(ctx, sourceId, target, plan.ResolutionFor(target.planned)))
		stop()
	}
	_ = w.publishImportProgress.Progress(ctx, entity.ImportProgress{SourceID: sourceId, Done: len(targets), Total: len(targets)})
	for _, localId := range result.LocalIDs() {
		_ = w.publishWorkspaceEvent.TreeChanged(ctx, localId)
	}
	return result
}

func (w importWriter) heartbeat(ctx context.Context, progress entity.ImportProgress) func() {
	ticker := time.NewTicker(importHeartbeatInterval)
	stopped := make(chan struct{})
	finished := make(chan struct{})
	go func() {
		defer close(finished)
		for {
			select {
			case <-stopped:
				return
			case <-ticker.C:
				_ = w.publishImportProgress.Progress(ctx, progress)
			}
		}
	}()
	return func() {
		ticker.Stop()
		close(stopped)
		<-finished
	}
}

func (w importWriter) importOne(ctx context.Context, sourceId string, target importTarget, resolution enums.ClashResolution) entity.ImportOutcome {
	if target.item.IsVariables() {
		return w.variables.importOne(ctx, target, resolution)
	}
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
