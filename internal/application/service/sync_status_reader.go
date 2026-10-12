package service

import (
	"context"

	"github.com/brienze1/xenon/internal/application/usecase"
	"github.com/brienze1/xenon/internal/domain/entity"
)

type syncStatusReader struct {
	changes         syncChanges
	loadWorkspace   usecase.LoadWorkspace
	publishGitEvent usecase.PublishGitEvent
	findSyncSession usecase.FindSyncSession
}

func newSyncStatusReader(changes syncChanges, loadWorkspace usecase.LoadWorkspace, publishGitEvent usecase.PublishGitEvent, findSyncSession usecase.FindSyncSession) syncStatusReader {
	return syncStatusReader{changes: changes, loadWorkspace: loadWorkspace, publishGitEvent: publishGitEvent, findSyncSession: findSyncSession}
}

func (s syncStatusReader) read(ctx context.Context, linked entity.LinkedWorkspace) (*entity.SyncStatus, error) {
	snapshot, err := s.changes.snapshot(ctx, linked)
	if err != nil {
		return nil, err
	}
	workspace, err := s.loadWorkspace.Load(ctx, linked.Root)
	if err != nil {
		return nil, err
	}
	status := entity.NewSyncStatus(linked.Settings, snapshot.heads, snapshot.changes.Count()).WithSharedSettings(workspace.Sync)
	session, err := s.findSyncSession.FindWorkspaceSession(ctx, linked.LocalID)
	if err != nil {
		return nil, err
	}
	if session != nil {
		status = status.InConflict()
	}
	return &status, nil
}

func (s syncStatusReader) publish(ctx context.Context, localId string, status entity.SyncStatus) (*entity.SyncStatus, error) {
	if err := s.publishGitEvent.StatusChanged(ctx, localId, status); err != nil {
		return nil, err
	}
	return &status, nil
}
