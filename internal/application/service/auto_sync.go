package service

import (
	"context"
	"time"

	"github.com/brienze1/xenon/internal/application/adapter"
	"github.com/brienze1/xenon/internal/application/usecase"
	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/pkg/logger"
	"github.com/brienze1/xenon/pkg/timeutils"
)

const (
	autoPullJob    = "auto-pull:"
	autoPushJob    = "auto-push:"
	autoPushDelay  = 3 * time.Second
	busyRetryDelay = 3 * time.Second
)

type autoSyncJob func(ctx context.Context, linked entity.LinkedWorkspace) error

type autoSync struct {
	claimSyncOperation usecase.ClaimSyncOperation
	scheduleSyncJob    usecase.ScheduleSyncJob
	workspaces         linkedWorkspaces
	committer          syncCommitter
	pusher             syncPusher
	status             syncStatusReader
}

func NewAutoSyncService(
	findWorkspaceRegistry usecase.FindWorkspaceRegistry,
	findLocalSyncSettings usecase.FindLocalSyncSettings,
	findGitAccounts usecase.FindGitAccounts,
	findGitCredential usecase.FindGitCredential,
	findKnownHosts usecase.FindKnownHosts,
	scanHostKey usecase.ScanHostKey,
	publishGitEvent usecase.PublishGitEvent,
	claimSyncOperation usecase.ClaimSyncOperation,
	scheduleSyncJob usecase.ScheduleSyncJob,
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
) adapter.AutoSyncService {
	changes := newSyncChanges(readBranchHeads, readTreeVersions, identifySyncItems)
	return &autoSync{
		claimSyncOperation: claimSyncOperation,
		scheduleSyncJob:    scheduleSyncJob,
		workspaces:         newLinkedWorkspaces(findWorkspaceRegistry, findLocalSyncSettings, findGitAccounts, findGitCredential, findKnownHosts, scanHostKey, publishGitEvent),
		committer:          syncCommitter{storeBlobs: storeBlobs, writeSyncCommit: writeSyncCommit},
		pusher: syncPusher{pushBranch: pushBranch, puller: newSyncPull(changes, publishGitEvent, fetchBranch, readTreeVersions, storeBlobs, convertSyncDocument,
			listCommitMessages, writeSyncCommit, moveBranch, applySyncedFiles, publishWorkspaceEvent, saveSyncSession, renderSyncItem)},
		status: newSyncStatusReader(changes, loadWorkspace, publishGitEvent, findSyncSession),
	}
}

func (a *autoSync) Opened(ctx context.Context, localId string) error {
	logger.Info(ctx, "Started", map[string]any{"localId": localId})

	_, settings, err := a.workspaces.settings(ctx, localId)
	if err != nil {
		return err
	}
	if settings.PullsWhenOpened() {
		a.schedule(ctx, autoPullJob, localId, 0, a.pull)
	}

	logger.Info(ctx, "Finished", map[string]any{"localId": localId, "autoPull": settings.PullsWhenOpened()})
	return nil
}

func (a *autoSync) Saved(ctx context.Context, localId string) error {
	logger.Info(ctx, "Started", map[string]any{"localId": localId})

	_, settings, err := a.workspaces.settings(ctx, localId)
	if err != nil {
		return err
	}
	if settings.PushesAfterSave() {
		a.schedule(ctx, autoPushJob, localId, autoPushDelay, a.push)
	}

	logger.Info(ctx, "Finished", map[string]any{"localId": localId, "autoPush": settings.PushesAfterSave()})
	return nil
}

func (a *autoSync) pull(ctx context.Context, linked entity.LinkedWorkspace) error {
	if !linked.Settings.PullsWhenOpened() {
		return nil
	}
	connected, err := a.workspaces.connect(ctx, linked)
	if err != nil {
		return err
	}
	_, err = a.pusher.puller.pull(ctx, *connected)
	return err
}

func (a *autoSync) push(ctx context.Context, linked entity.LinkedWorkspace) error {
	if !linked.Settings.PushesAfterSave() {
		return nil
	}
	connected, err := a.workspaces.connect(ctx, linked)
	if err != nil {
		return err
	}
	snapshot, err := a.status.changes.snapshot(ctx, *connected)
	if err != nil {
		return err
	}
	if !snapshot.changes.IsEmpty() {
		if err := a.committer.commit(ctx, *connected, *snapshot, snapshot.changes, snapshot.changes.UpdateMessage()); err != nil {
			return err
		}
	}
	_, err = a.pusher.push(ctx, *connected)
	return err
}

func (a *autoSync) schedule(ctx context.Context, job, localId string, delay time.Duration, run autoSyncJob) {
	a.scheduleSyncJob.Schedule(ctx, job+localId, delay, func(ctx context.Context) { a.run(ctx, job, localId, run) })
}

func (a *autoSync) run(ctx context.Context, job, localId string, run autoSyncJob) {
	logger.Info(ctx, "Started", map[string]any{"job": job, "localId": localId})

	release, err := a.claimSyncOperation.Claim(ctx, localId)
	if err != nil {
		logger.Info(ctx, "Finished", map[string]any{"job": job, "localId": localId, "retryIn": busyRetryDelay.String()})
		a.schedule(ctx, job, localId, busyRetryDelay, run)
		return
	}
	defer release()
	linked, err := a.workspaces.find(ctx, localId)
	if err != nil {
		logger.Warn(ctx, err, "The workspace is no longer linked", map[string]any{"job": job, "localId": localId})
		return
	}
	if err := a.report(ctx, *linked, run(ctx, *linked)); err != nil {
		logger.Error(ctx, err, "The sync status could not be reported", map[string]any{"job": job, "localId": localId})
		return
	}

	logger.Info(ctx, "Finished", map[string]any{"job": job, "localId": localId})
}

func (a *autoSync) report(ctx context.Context, linked entity.LinkedWorkspace, failure error) error {
	status, err := a.status.read(ctx, linked)
	if err != nil {
		return err
	}
	reported := status.SyncedAt(timeutils.Now())
	if failure != nil {
		reported = status.WithError(entity.NewSyncError(failure))
	}
	_, err = a.status.publish(ctx, linked.LocalID, reported)
	return err
}
