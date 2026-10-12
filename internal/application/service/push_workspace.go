package service

import (
	"context"

	"github.com/brienze1/xenon/internal/application/adapter"
	"github.com/brienze1/xenon/internal/application/usecase"
	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/pkg/errs"
	"github.com/brienze1/xenon/pkg/logger"
	"github.com/brienze1/xenon/pkg/timeutils"
)

const phasePush = "push"

type pushWorkspace struct {
	pushBranch         usecase.PushBranch
	claimSyncOperation usecase.ClaimSyncOperation
	workspaces         linkedWorkspaces
	puller             syncPull
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
) adapter.PushWorkspaceService {
	changes := newSyncChanges(readBranchHeads, readTreeVersions, identifySyncItems)
	return &pushWorkspace{
		pushBranch:         pushBranch,
		claimSyncOperation: claimSyncOperation,
		workspaces:         newLinkedWorkspaces(findWorkspaceRegistry, findLocalSyncSettings, findGitAccounts, findGitCredential, findKnownHosts, scanHostKey, publishGitEvent),
		puller: newSyncPull(changes, publishGitEvent, fetchBranch, readTreeVersions, storeBlobs, convertSyncDocument, listCommitMessages, writeSyncCommit,
			moveBranch, applySyncedFiles, publishWorkspaceEvent),
		status: syncStatusReader{changes: changes, loadWorkspace: loadWorkspace, publishGitEvent: publishGitEvent},
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
	err = p.push(ctx, *linked)
	if errs.IsConflictError(err) {
		session, pullErr := p.puller.pull(ctx, *linked)
		switch {
		case pullErr != nil:
			return nil, pullErr
		case session != nil:
			return p.publish(ctx, *linked, func(status entity.SyncStatus) entity.SyncStatus { return status.InConflict() })
		}
		err = p.push(ctx, *linked)
	}
	if err != nil {
		return nil, err
	}
	published, err := p.publish(ctx, *linked, func(status entity.SyncStatus) entity.SyncStatus { return status.SyncedAt(timeutils.Now()) })
	if err != nil {
		return nil, err
	}

	logger.Info(ctx, "Finished", published, map[string]any{"localId": localId})
	return published, nil
}

func (p *pushWorkspace) push(ctx context.Context, linked entity.LinkedWorkspace) error {
	if err := p.puller.publishGitEvent.Progress(ctx, linked.LocalID, phasePush, 0, 1); err != nil {
		return err
	}
	heads, err := p.puller.changes.readBranchHeads.ReadHeads(ctx, linked.ClonePath(), linked.Branch())
	if err != nil || !heads.HasLocalCommits() {
		return err
	}
	return p.pushBranch.Push(ctx, linked.Access, linked.ClonePath(), linked.Branch(), heads.Remote)
}

func (p *pushWorkspace) publish(ctx context.Context, linked entity.LinkedWorkspace, decorate func(entity.SyncStatus) entity.SyncStatus) (*entity.SyncStatus, error) {
	status, err := p.status.read(ctx, linked)
	if err != nil {
		return nil, err
	}
	return p.status.publish(ctx, linked.LocalID, decorate(*status))
}
