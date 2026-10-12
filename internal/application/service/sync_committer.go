package service

import (
	"context"

	"github.com/brienze1/xenon/internal/application/usecase"
	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/pkg/timeutils"
)

type syncCommitter struct {
	storeBlobs      usecase.StoreBlobs
	writeSyncCommit usecase.WriteSyncCommit
}

func (s syncCommitter) commit(ctx context.Context, author entity.LinkedWorkspace, snapshot syncSnapshot, selected entity.ChangeSet, message string) error {
	if err := s.storeBlobs.StoreWorktreeBlobs(ctx, author.ClonePath(), author.Subfolder(), selected.PresentPaths(snapshot.worktree)); err != nil {
		return err
	}
	commit := entity.NewSyncCommit(author, snapshot.heads.Local, message, selected.TreeFrom(snapshot.head, snapshot.worktree), timeutils.Now())
	_, err := s.writeSyncCommit.Write(ctx, author.ClonePath(), commit)
	return err
}
