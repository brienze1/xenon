package service

import (
	"context"

	"github.com/brienze1/xenon/internal/application/usecase"
	"github.com/brienze1/xenon/internal/domain/entity"
)

type syncWorktreeWriter struct {
	readTreeVersions      usecase.ReadTreeVersions
	applySyncedFiles      usecase.ApplySyncedFiles
	publishWorkspaceEvent usecase.PublishWorkspaceEvent
}

func (s syncWorktreeWriter) apply(ctx context.Context, linked entity.LinkedWorkspace, update entity.WorktreeUpdate) error {
	if update.IsEmpty() {
		return nil
	}
	writes := make(map[string][]byte, len(update.Writes))
	for path, hash := range update.Writes {
		content, err := s.readTreeVersions.ReadBlob(ctx, linked.ClonePath(), hash)
		if err != nil {
			return err
		}
		writes[path] = content
	}
	changes, err := s.applySyncedFiles.ApplySyncedFiles(ctx, linked.Root, writes, update.Removals)
	if err != nil {
		return err
	}
	for _, change := range changes {
		if err := s.publishWorkspaceEvent.RequestChangedOnDisk(ctx, linked.LocalID, change); err != nil {
			return err
		}
	}
	return s.publishWorkspaceEvent.TreeChanged(ctx, linked.LocalID)
}
