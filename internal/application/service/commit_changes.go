package service

import (
	"context"

	"github.com/brienze1/xenon/internal/application/adapter"
	"github.com/brienze1/xenon/internal/application/usecase"
	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/pkg/logger"
	"github.com/brienze1/xenon/pkg/timeutils"
)

type commitChanges struct {
	claimSyncOperation usecase.ClaimSyncOperation
	storeBlobs         usecase.StoreBlobs
	writeSyncCommit    usecase.WriteSyncCommit
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
) adapter.CommitChangesService {
	return &commitChanges{
		claimSyncOperation: claimSyncOperation,
		storeBlobs:         storeBlobs,
		writeSyncCommit:    writeSyncCommit,
		workspaces:         linkedWorkspaces{findWorkspaceRegistry: findWorkspaceRegistry, findLocalSyncSettings: findLocalSyncSettings, findGitAccounts: findGitAccounts},
		status:             syncStatusReader{changes: newSyncChanges(readBranchHeads, readTreeVersions, identifySyncItems), loadWorkspace: loadWorkspace, publishGitEvent: publishGitEvent},
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
	if err := c.storeBlobs.StoreWorktreeBlobs(ctx, linked.ClonePath(), linked.Subfolder(), selected.PresentPaths(snapshot.worktree)); err != nil {
		return nil, err
	}
	author := linked.WithAccess(*account, entity.GitAccess{})
	commit := entity.NewSyncCommit(author, snapshot.heads.Local, message, selected.TreeFrom(snapshot.head, snapshot.worktree), timeutils.Now())
	if _, err := c.writeSyncCommit.Write(ctx, linked.ClonePath(), commit); err != nil {
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
