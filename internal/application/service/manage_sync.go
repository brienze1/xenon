package service

import (
	"context"

	"github.com/brienze1/xenon/internal/application/adapter"
	"github.com/brienze1/xenon/internal/application/usecase"
	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/pkg/logger"
	"github.com/brienze1/xenon/pkg/timeutils"
)

type manageSync struct {
	claimSyncOperation    usecase.ClaimSyncOperation
	fetchBranch           usecase.FetchBranch
	saveLocalSyncSettings usecase.SaveLocalSyncSettings
	workspaces            linkedWorkspaces
	status                syncStatusReader
}

func NewManageSyncService(
	findWorkspaceRegistry usecase.FindWorkspaceRegistry,
	findLocalSyncSettings usecase.FindLocalSyncSettings,
	saveLocalSyncSettings usecase.SaveLocalSyncSettings,
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
	loadWorkspace usecase.LoadWorkspace,
	findSyncSession usecase.FindSyncSession,
) adapter.ManageSyncService {
	return &manageSync{
		claimSyncOperation:    claimSyncOperation,
		fetchBranch:           fetchBranch,
		saveLocalSyncSettings: saveLocalSyncSettings,
		workspaces:            newLinkedWorkspaces(findWorkspaceRegistry, findLocalSyncSettings, findGitAccounts, findGitCredential, findKnownHosts, scanHostKey, publishGitEvent),
		status:                newSyncStatusReader(newSyncChanges(readBranchHeads, readTreeVersions, identifySyncItems), loadWorkspace, publishGitEvent, findSyncSession),
	}
}

func (m *manageSync) Status(ctx context.Context, localId string) (*entity.SyncStatus, error) {
	logger.Info(ctx, "Started", map[string]any{"localId": localId})

	registered, settings, err := m.workspaces.settings(ctx, localId)
	if err != nil {
		return nil, err
	}
	if !settings.IsLinked() {
		status := entity.NotLinkedSyncStatus()
		logger.Info(ctx, "Finished", status, map[string]any{"localId": localId})
		return &status, nil
	}
	release, err := m.claimSyncOperation.Claim(ctx, localId)
	if err != nil {
		status := entity.SyncingSyncStatus(*settings)
		logger.Info(ctx, "Finished", status, map[string]any{"localId": localId})
		return &status, nil
	}
	defer release()
	linked := entity.NewLinkedWorkspace(localId, registered.Path, *settings)
	fetchErr := m.fetch(ctx, linked)
	status, err := m.status.read(ctx, linked)
	if err != nil {
		return nil, err
	}
	refreshed := status.SyncedAt(timeutils.Now())
	if fetchErr != nil {
		refreshed = status.WithError(entity.NewSyncError(fetchErr))
	}
	published, err := m.status.publish(ctx, localId, refreshed)
	if err != nil {
		return nil, err
	}

	logger.Info(ctx, "Finished", published, map[string]any{"localId": localId})
	return published, nil
}

func (m *manageSync) Changes(ctx context.Context, localId string) ([]entity.ChangeItem, error) {
	logger.Info(ctx, "Started", map[string]any{"localId": localId})

	linked, err := m.workspaces.find(ctx, localId)
	if err != nil {
		return nil, err
	}
	snapshot, err := m.status.changes.snapshot(ctx, *linked)
	if err != nil {
		return nil, err
	}

	logger.Info(ctx, "Finished", map[string]any{"localId": localId, "changes": snapshot.changes.Count()})
	return snapshot.changes.Items, nil
}

func (m *manageSync) SaveSettings(ctx context.Context, localId, accountId string, autoPull, autoPush bool) error {
	logger.Info(ctx, "Started", map[string]any{"localId": localId, "accountId": accountId, "autoPull": autoPull, "autoPush": autoPush})

	linked, err := m.workspaces.find(ctx, localId)
	if err != nil {
		return err
	}
	if accountId != "" {
		if _, err := m.workspaces.findGitAccounts.Find(ctx, accountId); err != nil {
			return err
		}
	}
	if err := m.saveLocalSyncSettings.SaveSyncSettings(ctx, localId, linked.Settings.WithOptions(accountId, autoPull, autoPush)); err != nil {
		return err
	}

	logger.Info(ctx, "Finished", map[string]any{"localId": localId, "accountId": accountId})
	return nil
}

func (m *manageSync) fetch(ctx context.Context, linked entity.LinkedWorkspace) error {
	connected, err := m.workspaces.connect(ctx, linked)
	if err != nil {
		return err
	}
	_, err = m.fetchBranch.Fetch(ctx, connected.Access, connected.ClonePath(), connected.Branch())
	return err
}
