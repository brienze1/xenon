package service

import (
	"context"

	"github.com/brienze1/xenon/internal/application/adapter"
	"github.com/brienze1/xenon/internal/application/usecase"
	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/internal/domain/enums"
	"github.com/brienze1/xenon/pkg/logger"
)

type resolveConflicts struct {
	claimSyncOperation usecase.ClaimSyncOperation
	findSyncSession    usecase.FindSyncSession
	saveSyncSession    usecase.SaveSyncSession
	workspaces         linkedWorkspaces
	status             syncStatusReader
}

func NewResolveConflictsService(
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
	loadWorkspace usecase.LoadWorkspace,
	findSyncSession usecase.FindSyncSession,
	saveSyncSession usecase.SaveSyncSession,
) adapter.ResolveConflictsService {
	return &resolveConflicts{
		claimSyncOperation: claimSyncOperation,
		findSyncSession:    findSyncSession,
		saveSyncSession:    saveSyncSession,
		workspaces:         newLinkedWorkspaces(findWorkspaceRegistry, findLocalSyncSettings, findGitAccounts, findGitCredential, findKnownHosts, scanHostKey, publishGitEvent),
		status:             newSyncStatusReader(newSyncChanges(readBranchHeads, readTreeVersions, identifySyncItems), loadWorkspace, publishGitEvent, findSyncSession),
	}
}

func (r *resolveConflicts) Session(ctx context.Context, localId string) (*entity.SyncSession, error) {
	logger.Info(ctx, "Started", map[string]any{"localId": localId})

	if _, err := r.workspaces.find(ctx, localId); err != nil {
		return nil, err
	}
	session, err := r.findSyncSession.FindWorkspaceSession(ctx, localId)
	if err != nil {
		return nil, err
	}

	logger.Info(ctx, "Finished", map[string]any{"localId": localId, "open": session != nil})
	return session, nil
}

func (r *resolveConflicts) Resolve(ctx context.Context, sessionId, itemId string, resolution enums.ConflictResolution, fieldPath string) (*entity.ConflictItem, error) {
	logger.Info(ctx, "Started", map[string]any{"sessionId": sessionId, "itemId": itemId, "resolution": resolution, "fieldPath": fieldPath})

	session, err := r.findSyncSession.FindSession(ctx, sessionId)
	if err != nil {
		return nil, err
	}
	conflict, found := session.Conflict(itemId)
	if !found {
		return nil, usecase.RejectMissingConflict(itemId)
	}
	if !conflict.Accepts(resolution, fieldPath) {
		return nil, usecase.RejectConflictResolution(conflict.ItemName, resolution.String(), fieldPath)
	}
	resolved := conflict.Resolved(resolution, fieldPath)
	if err := r.saveSyncSession.UpdateSession(ctx, session.WithConflict(resolved)); err != nil {
		return nil, err
	}

	logger.Info(ctx, "Finished", map[string]any{"sessionId": sessionId, "itemId": itemId, "resolution": resolved.Resolution})
	return &resolved, nil
}

func (r *resolveConflicts) Cancel(ctx context.Context, sessionId string) (*entity.SyncStatus, error) {
	logger.Info(ctx, "Started", map[string]any{"sessionId": sessionId})

	session, release, err := claimSession(ctx, r.findSyncSession, r.claimSyncOperation, sessionId)
	if err != nil {
		return nil, err
	}
	defer release()
	linked, err := r.workspaces.find(ctx, session.LocalID)
	if err != nil {
		return nil, err
	}
	if err := r.saveSyncSession.DiscardSession(ctx, session.LocalID); err != nil {
		return nil, err
	}
	status, err := r.status.read(ctx, *linked)
	if err != nil {
		return nil, err
	}
	published, err := r.status.publish(ctx, session.LocalID, *status)
	if err != nil {
		return nil, err
	}

	logger.Info(ctx, "Finished", published, map[string]any{"sessionId": sessionId})
	return published, nil
}

func claimSession(ctx context.Context, findSyncSession usecase.FindSyncSession, claimSyncOperation usecase.ClaimSyncOperation, sessionId string) (*entity.SyncSession, func(), error) {
	session, err := findSyncSession.FindSession(ctx, sessionId)
	if err != nil {
		return nil, nil, err
	}
	release, err := claimSyncOperation.Claim(ctx, session.LocalID)
	if err != nil {
		return nil, nil, err
	}
	current, err := findSyncSession.FindSession(ctx, sessionId)
	if err != nil {
		release()
		return nil, nil, err
	}
	return current, release, nil
}
