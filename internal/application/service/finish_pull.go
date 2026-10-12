package service

import (
	"context"
	"maps"

	"github.com/brienze1/xenon/internal/application/adapter"
	"github.com/brienze1/xenon/internal/application/usecase"
	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/pkg/logger"
	"github.com/brienze1/xenon/pkg/timeutils"
)

type finishPull struct {
	claimSyncOperation usecase.ClaimSyncOperation
	findSyncSession    usecase.FindSyncSession
	saveSyncSession    usecase.SaveSyncSession
	listCommitMessages usecase.ListCommitMessages
	writeSyncCommit    usecase.WriteSyncCommit
	moveBranch         usecase.MoveBranch
	workspaces         linkedWorkspaces
	resolver           syncResolver
	writer             syncWorktreeWriter
	status             syncStatusReader
}

func NewFinishPullService(
	findWorkspaceRegistry usecase.FindWorkspaceRegistry,
	findLocalSyncSettings usecase.FindLocalSyncSettings,
	findGitAccounts usecase.FindGitAccounts,
	findGitCredential usecase.FindGitCredential,
	findKnownHosts usecase.FindKnownHosts,
	scanHostKey usecase.ScanHostKey,
	publishGitEvent usecase.PublishGitEvent,
	claimSyncOperation usecase.ClaimSyncOperation,
	readBranchHeads usecase.ReadBranchHeads,
	readTreeVersions usecase.ReadTreeVersions,
	identifySyncItems usecase.IdentifySyncItems,
	storeBlobs usecase.StoreBlobs,
	convertSyncDocument usecase.ConvertSyncDocument,
	listCommitMessages usecase.ListCommitMessages,
	writeSyncCommit usecase.WriteSyncCommit,
	moveBranch usecase.MoveBranch,
	applySyncedFiles usecase.ApplySyncedFiles,
	publishWorkspaceEvent usecase.PublishWorkspaceEvent,
	loadWorkspace usecase.LoadWorkspace,
	findSyncSession usecase.FindSyncSession,
	saveSyncSession usecase.SaveSyncSession,
) adapter.FinishPullService {
	changes := newSyncChanges(readBranchHeads, readTreeVersions, identifySyncItems)
	merger := syncMerger{readTreeVersions: readTreeVersions, storeBlobs: storeBlobs, convertSyncDocument: convertSyncDocument}
	return &finishPull{
		claimSyncOperation: claimSyncOperation,
		findSyncSession:    findSyncSession,
		saveSyncSession:    saveSyncSession,
		listCommitMessages: listCommitMessages,
		writeSyncCommit:    writeSyncCommit,
		moveBranch:         moveBranch,
		workspaces:         newLinkedWorkspaces(findWorkspaceRegistry, findLocalSyncSettings, findGitAccounts, findGitCredential, findKnownHosts, scanHostKey, publishGitEvent),
		resolver:           syncResolver{readTreeVersions: readTreeVersions, storeBlobs: storeBlobs, convertSyncDocument: convertSyncDocument, changes: changes, merger: merger},
		writer:             syncWorktreeWriter{readTreeVersions: readTreeVersions, applySyncedFiles: applySyncedFiles, publishWorkspaceEvent: publishWorkspaceEvent},
		status:             newSyncStatusReader(changes, loadWorkspace, publishGitEvent, findSyncSession),
	}
}

func (f *finishPull) Finish(ctx context.Context, sessionId string) (*entity.SyncStatus, error) {
	logger.Info(ctx, "Started", map[string]any{"sessionId": sessionId})

	session, release, err := claimSession(ctx, f.findSyncSession, f.claimSyncOperation, sessionId)
	if err != nil {
		return nil, err
	}
	defer release()
	if !session.IsResolved() {
		return nil, usecase.RejectUnresolvedConflicts()
	}
	linked, err := f.workspaces.find(ctx, session.LocalID)
	if err != nil {
		return nil, err
	}
	account, err := f.workspaces.account(ctx, *linked)
	if err != nil {
		return nil, err
	}
	snapshot, err := f.resolver.changes.snapshot(ctx, *linked)
	if err != nil {
		return nil, err
	}
	if !session.StartedFrom(snapshot.heads.Local) {
		if err := f.saveSyncSession.DiscardSession(ctx, session.LocalID); err != nil {
			return nil, err
		}
		return nil, usecase.RejectStaleSyncSession()
	}
	resolved, err := f.resolver.resolve(ctx, *linked, *session, *snapshot)
	if err != nil {
		return nil, err
	}
	if err := f.writer.apply(ctx, *linked, resolved.update); err != nil {
		return nil, err
	}
	if err := f.advance(ctx, linked.WithAccess(*account, entity.GitAccess{}), *session, *resolved); err != nil {
		return nil, err
	}
	if err := f.saveSyncSession.DiscardSession(ctx, session.LocalID); err != nil {
		return nil, err
	}
	status, err := f.status.read(ctx, *linked)
	if err != nil {
		return nil, err
	}
	published, err := f.status.publish(ctx, session.LocalID, *status)
	if err != nil {
		return nil, err
	}

	logger.Info(ctx, "Finished", published, map[string]any{"sessionId": sessionId})
	return published, nil
}

func (f *finishPull) advance(ctx context.Context, author entity.LinkedWorkspace, session entity.SyncSession, resolved resolvedPull) error {
	if !session.Replay || maps.Equal(resolved.tree, resolved.remote) {
		return f.moveBranch.MoveBranch(ctx, author.ClonePath(), author.Branch(), session.RemoteHead)
	}
	messages, err := f.listCommitMessages.ListMessages(ctx, author.ClonePath(), session.Base, session.LocalHead)
	if err != nil {
		return err
	}
	commit := entity.NewSyncCommit(author, session.RemoteHead, usecase.SquashReplayMessage(messages), resolved.tree, timeutils.Now())
	_, err = f.writeSyncCommit.Write(ctx, author.ClonePath(), commit)
	return err
}
