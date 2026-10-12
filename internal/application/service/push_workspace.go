package service

import (
	"context"

	"github.com/brienze1/xenon/internal/application/adapter"
	"github.com/brienze1/xenon/internal/application/usecase"
	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/pkg/logger"
	"github.com/brienze1/xenon/pkg/timeutils"
)

type pushWorkspace struct {
	claimSyncOperation usecase.ClaimSyncOperation
	workspaces         linkedWorkspaces
	pusher             syncPusher
	status             syncStatusReader
}

func NewPushWorkspaceService(
	findWorkspaceRegistry usecase.FindWorkspaceRegistry,
	findLocalSyncSettings usecase.FindLocalSyncSettings,
	findGitAccounts usecase.FindGitAccounts,
	findGitCredential usecase.FindGitCredential,
	findKnownHosts usecase.FindKnownHosts,
	scanHostKey usecase.ScanHostKey,
	publishGitEvent usecase.PublishGitEvent,
	claimSyncOperation usecase.ClaimSyncOperation,
	fetchBranch usecase.FetchBranch,
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
	pushBranch usecase.PushBranch,
	findSyncSession usecase.FindSyncSession,
	saveSyncSession usecase.SaveSyncSession,
	renderSyncItem usecase.RenderSyncItem,
) adapter.PushWorkspaceService {
	changes := newSyncChanges(readBranchHeads, readTreeVersions, identifySyncItems)
	return &pushWorkspace{
		claimSyncOperation: claimSyncOperation,
		workspaces:         newLinkedWorkspaces(findWorkspaceRegistry, findLocalSyncSettings, findGitAccounts, findGitCredential, findKnownHosts, scanHostKey, publishGitEvent),
		pusher: syncPusher{pushBranch: pushBranch, puller: newSyncPull(changes, publishGitEvent, fetchBranch, readTreeVersions, storeBlobs, convertSyncDocument,
			listCommitMessages, writeSyncCommit, moveBranch, applySyncedFiles, publishWorkspaceEvent, saveSyncSession, renderSyncItem)},
		status: newSyncStatusReader(changes, loadWorkspace, publishGitEvent, findSyncSession),
	}
}

func (p *pushWorkspace) Push(ctx context.Context, localId string) (*entity.SyncStatus, error) {
	logger.Info(ctx, "Started", map[string]any{"localId": localId})

	release, err := p.claimSyncOperation.Claim(ctx, localId)
	if err != nil {
		return nil, err
	}
	defer release()
	linked, err := p.workspaces.findConnected(ctx, localId)
	if err != nil {
		return nil, err
	}
	session, err := p.pusher.push(ctx, *linked)
	if err != nil {
		return nil, err
	}
	if session != nil {
		return p.publish(ctx, *linked, func(status entity.SyncStatus) entity.SyncStatus { return status.InConflict() })
	}
	published, err := p.publish(ctx, *linked, func(status entity.SyncStatus) entity.SyncStatus { return status.SyncedAt(timeutils.Now()) })
	if err != nil {
		return nil, err
	}

	logger.Info(ctx, "Finished", published, map[string]any{"localId": localId})
	return published, nil
}

func (p *pushWorkspace) publish(ctx context.Context, linked entity.LinkedWorkspace, decorate func(entity.SyncStatus) entity.SyncStatus) (*entity.SyncStatus, error) {
	status, err := p.status.read(ctx, linked)
	if err != nil {
		return nil, err
	}
	return p.status.publish(ctx, linked.LocalID, decorate(*status))
}
