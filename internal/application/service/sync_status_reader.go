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
	return &status, nil
}

func (s syncStatusReader) publish(ctx context.Context, localId string, status entity.SyncStatus) (*entity.SyncStatus, error) {
	if err := s.publishGitEvent.StatusChanged(ctx, localId, status); err != nil {
		return nil, err
	}
	return &status, nil
}
