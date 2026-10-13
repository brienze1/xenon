package service

import (
	"context"

	"github.com/brienze1/xenon/internal/application/adapter"
	"github.com/brienze1/xenon/internal/application/usecase"
	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/pkg/logger"
)

type commitChanges struct {
	claimSyncOperation usecase.ClaimSyncOperation
	committer          syncCommitter
	workspaces         linkedWorkspaces
	status             syncStatusReader
}

func NewCommitChangesService(
	findWorkspaceRegistry usecase.FindWorkspaceRegistry,
	findLocalSyncSettings usecase.FindLocalSyncSettings,
	findGitAccounts usecase.FindGitAccounts,
	claimSyncOperation usecase.ClaimSyncOperation,
	readBranchHeads usecase.ReadBranchHeads,
	readTreeVersions usecase.ReadTreeVersions,
	identifySyncItems usecase.IdentifySyncItems,
	storeBlobs usecase.StoreBlobs,
	writeSyncCommit usecase.WriteSyncCommit,
	loadWorkspace usecase.LoadWorkspace,
	publishGitEvent usecase.PublishGitEvent,
	findSyncSession usecase.FindSyncSession,
) adapter.CommitChangesService {
	return &commitChanges{
		claimSyncOperation: claimSyncOperation,
		committer:          syncCommitter{storeBlobs: storeBlobs, writeSyncCommit: writeSyncCommit},
		workspaces:         linkedWorkspaces{findWorkspaceRegistry: findWorkspaceRegistry, findLocalSyncSettings: findLocalSyncSettings, findGitAccounts: findGitAccounts},
		status:             newSyncStatusReader(newSyncChanges(readBranchHeads, readTreeVersions, identifySyncItems), loadWorkspace, publishGitEvent, findSyncSession),
	}
}

func (c *commitChanges) Commit(ctx context.Context, localId, message string, itemIds []string) (*entity.SyncStatus, error) {
	logger.Info(ctx, "Started", map[string]any{"localId": localId, "message": message, "itemIds": itemIds})

	release, err := c.claimSyncOperation.Claim(ctx, localId)
	if err != nil {
		return nil, err
	}
	defer release()
	linked, err := c.workspaces.find(ctx, localId)
	if err != nil {
		return nil, err
	}
	account, err := c.workspaces.account(ctx, *linked)
	if err != nil {
		return nil, err
	}
	snapshot, err := c.status.changes.snapshot(ctx, *linked)
	if err != nil {
		return nil, err
	}
	selected := snapshot.changes.Selected(itemIds)
	if selected.IsEmpty() {
		return nil, usecase.RejectEmptyCommit()
	}
	if err := c.committer.commit(ctx, linked.WithAccess(*account, entity.GitAccess{}), *snapshot, selected, message); err != nil {
		return nil, err
	}
	status, err := c.status.read(ctx, *linked)
	if err != nil {
		return nil, err
	}
	published, err := c.status.publish(ctx, localId, *status)
	if err != nil {
		return nil, err
	}

	logger.Info(ctx, "Finished", published, map[string]any{"localId": localId})
	return published, nil
}
