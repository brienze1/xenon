package service

import (
	"context"

	"github.com/brienze1/xenon/internal/application/usecase"
	"github.com/brienze1/xenon/internal/domain/entity"
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
	saveSyncSession    usecase.SaveSyncSession
	fetchBranch        usecase.FetchBranch
	listCommitMessages usecase.ListCommitMessages
	writeSyncCommit    usecase.WriteSyncCommit
	moveBranch         usecase.MoveBranch
	publishGitEvent    usecase.PublishGitEvent
	changes            syncChanges
	merger             syncMerger
	writer             syncWorktreeWriter
	renderer           syncConflictRenderer
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
	saveSyncSession usecase.SaveSyncSession,
	renderSyncItem usecase.RenderSyncItem,
) syncPull {
	return syncPull{
		saveSyncSession:    saveSyncSession,
		fetchBranch:        fetchBranch,
		listCommitMessages: listCommitMessages,
		writeSyncCommit:    writeSyncCommit,
		moveBranch:         moveBranch,
		publishGitEvent:    publishGitEvent,
		changes:            changes,
		merger:             syncMerger{readTreeVersions: readTreeVersions, storeBlobs: storeBlobs, convertSyncDocument: convertSyncDocument},
		writer:             syncWorktreeWriter{readTreeVersions: readTreeVersions, applySyncedFiles: applySyncedFiles, publishWorkspaceEvent: publishWorkspaceEvent},
		renderer:           syncConflictRenderer{readTreeVersions: readTreeVersions, renderSyncItem: renderSyncItem},
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
	if heads.HasNothingToPull() {
		return nil, p.saveSyncSession.DiscardSession(ctx, linked.LocalID)
	}
	if err := p.publishGitEvent.Progress(ctx, linked.LocalID, phaseMerge, 1, pullPhases); err != nil {
		return nil, err
	}
	snapshot, err := p.changes.snapshot(ctx, linked)
	if err != nil {
		return nil, err
	}
	target, err := p.target(ctx, linked, *heads)
	if err != nil {
		return nil, err
	}
	update := entity.NewWorktreeUpdate(snapshot.head, target.tree, snapshot.worktree)
	if len(target.conflicts) > 0 || update.HasConflicts() {
		uncommitted := update.ConflictPlans(snapshot.head, target.tree, snapshot.worktree)
		return p.session(ctx, linked, *heads, snapshot.worktree, target, uncommitted)
	}
	if err := p.publishGitEvent.Progress(ctx, linked.LocalID, phaseWrite, 2, pullPhases); err != nil {
		return nil, err
	}
	if err := p.writer.apply(ctx, linked, update); err != nil {
		return nil, err
	}
	if err := p.advance(ctx, linked, *heads, target.tree); err != nil {
		return nil, err
	}
	if err := p.saveSyncSession.DiscardSession(ctx, linked.LocalID); err != nil {
		return nil, err
	}
	return nil, p.publishGitEvent.Progress(ctx, linked.LocalID, phaseWrite, pullPhases, pullPhases)
}

func (p syncPull) target(ctx context.Context, linked entity.LinkedWorkspace, heads entity.BranchHeads) (*pullTarget, error) {
	if heads.IsFastForward() {
		tree, err := p.changes.tree(ctx, linked, heads.Remote)
		if err != nil {
			return nil, err
		}
		return &pullTarget{tree: tree, conflicts: []entity.PathPlan{}, fields: map[string][]entity.FieldConflict{}}, nil
	}
	return p.merger.merge(ctx, linked, heads, nil)
}

func (p syncPull) advance(ctx context.Context, linked entity.LinkedWorkspace, heads entity.BranchHeads, tree map[string]string) error {
	if heads.IsFastForward() {
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

func (p syncPull) session(ctx context.Context, linked entity.LinkedWorkspace, heads entity.BranchHeads, worktree entity.SyncedFiles, target *pullTarget, uncommitted []entity.PathPlan) (*entity.SyncSession, error) {
	plans := append(append([]entity.PathPlan{}, target.conflicts...), uncommitted...)
	present, blobs := worktree.LocatePlans(plans)
	refs, err := p.changes.identify(ctx, linked, present, blobs)
	if err != nil {
		return nil, err
	}
	found := make([]entity.ConflictItem, 0, len(plans))
	for _, plan := range target.conflicts {
		found = append(found, entity.NewConflictItem(refs.RefFor(plan.Path), plan, target.fields[plan.Path]))
	}
	for _, plan := range uncommitted {
		found = append(found, entity.NewUncommittedConflictItem(refs.RefFor(plan.Path), plan))
	}
	conflicts := make([]entity.ConflictItem, 0, len(found))
	for _, conflict := range found {
		rendered, err := p.renderer.render(ctx, linked, conflict, worktree)
		if err != nil {
			return nil, err
		}
		conflicts = append(conflicts, *rendered)
	}
	session := entity.NewSyncSession(idutils.NewUlid(), linked.LocalID, heads, heads.IsDiverged(), conflicts, timeutils.Now())
	if err := p.saveSyncSession.SaveSession(ctx, session); err != nil {
		return nil, err
	}
	return &session, nil
}
