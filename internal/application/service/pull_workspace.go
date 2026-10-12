package service

import (
	"context"

	"github.com/brienze1/xenon/internal/application/adapter"
	"github.com/brienze1/xenon/internal/application/usecase"
	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/pkg/logger"
	"github.com/brienze1/xenon/pkg/timeutils"
)

type pullWorkspace struct {
	claimSyncOperation usecase.ClaimSyncOperation
	workspaces         linkedWorkspaces
	puller             syncPull
	status             syncStatusReader
}

func NewPullWorkspaceService(
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
) adapter.PullWorkspaceService {
	changes := newSyncChanges(readBranchHeads, readTreeVersions, identifySyncItems)
	return &pullWorkspace{
		claimSyncOperation: claimSyncOperation,
		workspaces:         newLinkedWorkspaces(findWorkspaceRegistry, findLocalSyncSettings, findGitAccounts, findGitCredential, findKnownHosts, scanHostKey, publishGitEvent),
		puller: newSyncPull(changes, publishGitEvent, fetchBranch, readTreeVersions, storeBlobs, convertSyncDocument, listCommitMessages, writeSyncCommit,
			moveBranch, applySyncedFiles, publishWorkspaceEvent),
		status: syncStatusReader{changes: changes, loadWorkspace: loadWorkspace, publishGitEvent: publishGitEvent},
	}
}

func (p *pullWorkspace) Pull(ctx context.Context, localId string) (*entity.PullResult, error) {
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
	session, err := p.puller.pull(ctx, *linked)
	if err != nil {
		return nil, err
	}
	status, err := p.status.read(ctx, *linked)
	if err != nil {
		return nil, err
	}
	pulled := status.SyncedAt(timeutils.Now())
	if session != nil {
		pulled = pulled.InConflict()
	}
	published, err := p.status.publish(ctx, localId, pulled)
	if err != nil {
		return nil, err
	}
	result := entity.PullResult{Status: *published, Session: session}

	logger.Info(ctx, "Finished", result.Status, map[string]any{"localId": localId, "conflicts": session != nil})
	return &result, nil
}
