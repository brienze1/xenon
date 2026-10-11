package service

import (
	"context"
	"time"

	"github.com/brienze1/xenon/internal/application/usecase"
	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/internal/domain/enums"
	"github.com/brienze1/xenon/pkg/logger"
)

const importHeartbeatInterval = 100 * time.Millisecond

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
		w.publishProgress(ctx, progress)
		stop := w.heartbeat(ctx, progress)
		result = result.WithOutcome(w.importOne(ctx, sourceId, target, plan.ResolutionFor(target.planned)))
		stop()
	}
	w.publishProgress(ctx, entity.ImportProgress{SourceID: sourceId, Done: len(targets), Total: len(targets)})
	for _, localId := range result.LocalIDs() {
		if err := w.publishWorkspaceEvent.TreeChanged(ctx, localId); err != nil {
			logger.Error(ctx, err, "The imported tree change could not be announced", map[string]any{"localId": localId})
		}
	}
	return result
}

func (w importWriter) publishProgress(ctx context.Context, progress entity.ImportProgress) {
	if err := w.publishImportProgress.Progress(ctx, progress); err != nil {
		logger.Error(ctx, err, "The import progress could not be announced", map[string]any{"progress": progress})
	}
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
				w.publishProgress(ctx, progress)
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
	named := usecase.NameImportFiles(nodes, target.siblings.EntriesReplacing(replaced))
	write := usecase.PlanImportWrite(target.placement.WithNodes(named), target.siblings, replaced, target.collection, target.item.Root.Container.Variables, resolution)
	if write.IsEmpty() {
		return target.item.Skipped(target.workspace.LocalID)
	}
	write, tokens := write.WithoutLiteralTokens()
	staged := write.Staged(sourceId, target.item.Key, target.workspace.Path, target.planned.TargetParentID)
	written, err := w.stageImportItem.Stage(ctx, staged)
	if err != nil {
		return usecase.FailImportItem(target.item, target.workspace.LocalID, err)
	}
	if err := w.commitImportItem.Commit(ctx, *written); err != nil {
		if rollbackErr := w.rollbackImportItem.Rollback(ctx, *written); rollbackErr != nil {
			logger.Error(ctx, rollbackErr, "The failed import item could not be rolled back", map[string]any{"staged": *written})
		}
		return usecase.FailImportItem(target.item, target.workspace.LocalID, err)
	}
	w.keepTokens(ctx, target.workspace.LocalID, tokens)
	return target.item.Completed(write, entity.ImportLocation{LocalID: target.workspace.LocalID, ParentPath: target.parentPath, CollectionPath: target.collectionPath()})
}

func (w importWriter) keepTokens(ctx context.Context, localId string, tokens []entity.OAuth2Token) {
	if len(tokens) == 0 {
		return
	}
	_, err := w.variables.vaults.update(ctx, localId, func(vault entity.Vault) entity.Vault {
		return vault.WithTokens(tokens)
	})
	if err != nil {
		logger.Warn(ctx, err, "The imported OAuth 2.0 tokens could not be kept in the vault", map[string]any{"localId": localId, "tokens": len(tokens)})
	}
}
