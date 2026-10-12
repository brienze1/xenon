package service

import (
	"context"

	"github.com/brienze1/xenon/internal/application/usecase"
	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/pkg/errs"
)

const phasePush = "push"

type syncPusher struct {
	pushBranch usecase.PushBranch
	puller     syncPull
}

func (s syncPusher) push(ctx context.Context, linked entity.LinkedWorkspace) (*entity.SyncSession, error) {
	err := s.pushOnce(ctx, linked)
	if !errs.IsConflictError(err) {
		return nil, err
	}
	session, err := s.puller.pull(ctx, linked)
	if err != nil || session != nil {
		return session, err
	}
	return nil, s.pushOnce(ctx, linked)
}

func (s syncPusher) pushOnce(ctx context.Context, linked entity.LinkedWorkspace) error {
	if err := s.puller.publishGitEvent.Progress(ctx, linked.LocalID, phasePush, 0, 1); err != nil {
		return err
	}
	heads, err := s.puller.changes.readBranchHeads.ReadHeads(ctx, linked.ClonePath(), linked.Branch())
	if err != nil || !heads.HasLocalCommits() {
		return err
	}
	return s.pushBranch.Push(ctx, linked.Access, linked.ClonePath(), linked.Branch(), heads.Remote)
}
