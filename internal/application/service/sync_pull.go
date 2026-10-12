package service

import (
	"context"

	"github.com/brienze1/xenon/internal/application/usecase"
	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/internal/domain/enums"
	"github.com/brienze1/xenon/internal/domain/enums/divergence"
	"github.com/brienze1/xenon/pkg/idutils"
	"github.com/brienze1/xenon/pkg/timeutils"
)

const (
	phaseFetch = "fetch"
	phaseMerge = "merge"
	phaseWrite = "write"
	pullPhases = 3
)

type pullTarget struct {
	tree      map[string]string
	fields    map[string][]entity.FieldConflict
	conflicts []entity.PathPlan
}

type syncPull struct {
	fetchBranch        usecase.FetchBranch
	listCommitMessages usecase.ListCommitMessages
	writeSyncCommit    usecase.WriteSyncCommit
	moveBranch         usecase.MoveBranch
	publishGitEvent    usecase.PublishGitEvent
	changes            syncChanges
	merger             syncMerger
	writer             syncWorktreeWriter
}

func newSyncPull(
	changes syncChanges,
	publishGitEvent usecase.PublishGitEvent,
	fetchBranch usecase.FetchBranch,
	readTreeVersions usecase.ReadTreeVersions,
	storeBlobs usecase.StoreBlobs,
	convertSyncDocument usecase.ConvertSyncDocument,
	listCommitMessages usecase.ListCommitMessages,
	writeSyncCommit usecase.WriteSyncCommit,
	moveBranch usecase.MoveBranch,
	applySyncedFiles usecase.ApplySyncedFiles,
	publishWorkspaceEvent usecase.PublishWorkspaceEvent,
) syncPull {
	return syncPull{
		fetchBranch:        fetchBranch,
		listCommitMessages: listCommitMessages,
		writeSyncCommit:    writeSyncCommit,
		moveBranch:         moveBranch,
		publishGitEvent:    publishGitEvent,
		changes:            changes,
		merger:             syncMerger{readTreeVersions: readTreeVersions, storeBlobs: storeBlobs, convertSyncDocument: convertSyncDocument},
		writer:             syncWorktreeWriter{readTreeVersions: readTreeVersions, applySyncedFiles: applySyncedFiles, publishWorkspaceEvent: publishWorkspaceEvent},
	}
}

func (p syncPull) pull(ctx context.Context, linked entity.LinkedWorkspace) (*entity.SyncSession, error) {
	if err := p.publishGitEvent.Progress(ctx, linked.LocalID, phaseFetch, 0, pullPhases); err != nil {
		return nil, err
	}
	heads, err := p.fetchBranch.Fetch(ctx, linked.Access, linked.ClonePath(), linked.Branch())
	if err != nil {
		return nil, err
	}
	state := usecase.ClassifyDivergence(*heads)
	if state == divergence.UpToDate || state == divergence.Ahead {
		return nil, nil
	}
	if err := p.publishGitEvent.Progress(ctx, linked.LocalID, phaseMerge, 1, pullPhases); err != nil {
		return nil, err
	}
	snapshot, err := p.changes.snapshot(ctx, linked)
	if err != nil {
		return nil, err
	}
	target, err := p.target(ctx, linked, *heads, state)
	if err != nil {
		return nil, err
	}
	update := entity.NewWorktreeUpdate(snapshot.head, target.tree, snapshot.worktree)
	conflicts := append(target.conflicts, update.ConflictPlans(snapshot.head, target.tree, snapshot.worktree)...)
	if len(conflicts) > 0 {
		return p.session(ctx, linked, *heads, snapshot.worktree, conflicts, target.fields)
	}
	if err := p.publishGitEvent.Progress(ctx, linked.LocalID, phaseWrite, 2, pullPhases); err != nil {
		return nil, err
	}
	if err := p.writer.apply(ctx, linked, update); err != nil {
		return nil, err
	}
	if err := p.advance(ctx, linked, *heads, state, target.tree); err != nil {
		return nil, err
	}
	return nil, p.publishGitEvent.Progress(ctx, linked.LocalID, phaseWrite, pullPhases, pullPhases)
}

func (p syncPull) target(ctx context.Context, linked entity.LinkedWorkspace, heads entity.BranchHeads, state enums.Divergence) (*pullTarget, error) {
	if state == divergence.FastForward {
		tree, err := p.changes.tree(ctx, linked, heads.Remote)
		if err != nil {
			return nil, err
		}
		return &pullTarget{tree: tree, conflicts: []entity.PathPlan{}, fields: map[string][]entity.FieldConflict{}}, nil
	}
	return p.merger.merge(ctx, linked, heads)
}

func (p syncPull) advance(ctx context.Context, linked entity.LinkedWorkspace, heads entity.BranchHeads, state enums.Divergence, tree map[string]string) error {
	if state == divergence.FastForward {
		return p.moveBranch.MoveBranch(ctx, linked.ClonePath(), linked.Branch(), heads.Remote)
	}
	messages, err := p.listCommitMessages.ListMessages(ctx, linked.ClonePath(), heads.Base, heads.Local)
	if err != nil {
		return err
	}
	commit := entity.NewSyncCommit(linked, heads.Remote, usecase.SquashReplayMessage(messages), tree, timeutils.Now())
	_, err = p.writeSyncCommit.Write(ctx, linked.ClonePath(), commit)
	return err
}

func (p syncPull) session(ctx context.Context, linked entity.LinkedWorkspace, heads entity.BranchHeads, worktree entity.SyncedFiles, plans []entity.PathPlan, fields map[string][]entity.FieldConflict) (*entity.SyncSession, error) {
	present, blobs := make([]string, 0, len(plans)), map[string]string{}
	for _, plan := range plans {
		if _, found := worktree[plan.Path]; found {
			present = append(present, plan.Path)
			continue
		}
		blobs[plan.Path] = plan.AnyVersion()
	}
	refs, err := p.changes.identify(ctx, linked, present, blobs)
	if err != nil {
		return nil, err
	}
	conflicts := make([]entity.ConflictItem, 0, len(plans))
	for _, plan := range plans {
		conflicts = append(conflicts, entity.NewConflictItem(usecase.RefForPath(plan.Path, refs), plan, fields[plan.Path]))
	}
	session := entity.NewSyncSession(idutils.NewUlid(), heads, conflicts, timeutils.Now())
	return &session, nil
}
