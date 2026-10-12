package service

import (
	"context"

	"github.com/brienze1/xenon/internal/application/adapter"
	"github.com/brienze1/xenon/internal/application/usecase"
	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/internal/domain/enums"
	"github.com/brienze1/xenon/internal/domain/enums/linkmode"
	"github.com/brienze1/xenon/pkg/logger"
	"github.com/brienze1/xenon/pkg/timeutils"
)

const phaseClone = "clone"

type linkWorkspace struct {
	claimSyncOperation     usecase.ClaimSyncOperation
	locateClone            usecase.LocateClone
	saveWorkspaceRegistry  usecase.SaveWorkspaceRegistry
	inspectWorkspaceFolder usecase.InspectWorkspaceFolder
	cloneRepository        usecase.CloneRepository
	copyWorkspaceFolder    usecase.CopyWorkspaceFolder
	removeWorkspaceFolder  usecase.RemoveWorkspaceFolder
	clearWorkspaceFolder   usecase.ClearWorkspaceFolder
	saveWorkspaceSync      usecase.SaveWorkspaceSync
	saveLocalSyncSettings  usecase.SaveLocalSyncSettings
	watchWorkspace         usecase.WatchWorkspace
	workspaces             linkedWorkspaces
	lfs                    lfsDetector
	status                 syncStatusReader
}

func NewLinkWorkspaceService(
	findWorkspaceRegistry usecase.FindWorkspaceRegistry,
	saveWorkspaceRegistry usecase.SaveWorkspaceRegistry,
	findLocalSyncSettings usecase.FindLocalSyncSettings,
	saveLocalSyncSettings usecase.SaveLocalSyncSettings,
	locateClone usecase.LocateClone,
	findGitAccounts usecase.FindGitAccounts,
	findGitCredential usecase.FindGitCredential,
	findKnownHosts usecase.FindKnownHosts,
	scanHostKey usecase.ScanHostKey,
	publishGitEvent usecase.PublishGitEvent,
	claimSyncOperation usecase.ClaimSyncOperation,
	cloneRepository usecase.CloneRepository,
	readBranchHeads usecase.ReadBranchHeads,
	readTreeVersions usecase.ReadTreeVersions,
	identifySyncItems usecase.IdentifySyncItems,
	inspectWorkspaceFolder usecase.InspectWorkspaceFolder,
	copyWorkspaceFolder usecase.CopyWorkspaceFolder,
	removeWorkspaceFolder usecase.RemoveWorkspaceFolder,
	clearWorkspaceFolder usecase.ClearWorkspaceFolder,
	loadWorkspace usecase.LoadWorkspace,
	saveWorkspaceSync usecase.SaveWorkspaceSync,
	watchWorkspace usecase.WatchWorkspace,
) adapter.LinkWorkspaceService {
	return &linkWorkspace{
		claimSyncOperation:     claimSyncOperation,
		locateClone:            locateClone,
		saveWorkspaceRegistry:  saveWorkspaceRegistry,
		inspectWorkspaceFolder: inspectWorkspaceFolder,
		cloneRepository:        cloneRepository,
		copyWorkspaceFolder:    copyWorkspaceFolder,
		removeWorkspaceFolder:  removeWorkspaceFolder,
		clearWorkspaceFolder:   clearWorkspaceFolder,
		saveWorkspaceSync:      saveWorkspaceSync,
		saveLocalSyncSettings:  saveLocalSyncSettings,
		watchWorkspace:         watchWorkspace,
		workspaces:             newLinkedWorkspaces(findWorkspaceRegistry, findLocalSyncSettings, findGitAccounts, findGitCredential, findKnownHosts, scanHostKey, publishGitEvent),
		lfs:                    lfsDetector{readBranchHeads: readBranchHeads, readTreeVersions: readTreeVersions},
		status:                 syncStatusReader{changes: newSyncChanges(readBranchHeads, readTreeVersions, identifySyncItems), loadWorkspace: loadWorkspace, publishGitEvent: publishGitEvent},
	}
}

func (l *linkWorkspace) Link(ctx context.Context, localId string, shared entity.SyncSettings, accountId, clonePath string, mode enums.LinkMode) (*entity.SyncStatus, error) {
	logger.Info(ctx, "Started", map[string]any{"localId": localId, "repository": shared.Repository, "branch": shared.Branch, "subfolder": shared.Subfolder, "accountId": accountId, "clonePath": clonePath, "mode": mode})

	release, err := l.claimSyncOperation.Claim(ctx, localId)
	if err != nil {
		return nil, err
	}
	defer release()
	registered, _, err := l.workspaces.settings(ctx, localId)
	if err != nil {
		return nil, err
	}
	if clonePath == "" {
		clonePath = l.locateClone.DefaultClonePath(ctx, localId)
	}
	if err := l.checkClonePath(ctx, clonePath); err != nil {
		return nil, err
	}
	linked := entity.NewLinkedWorkspace(localId, "", entity.NewLocalSyncSettings(shared, accountId, clonePath))
	connected, err := l.workspaces.connect(ctx, linked)
	if err != nil {
		return nil, err
	}
	if err := l.workspaces.access.publishGitEvent.Progress(ctx, localId, phaseClone, 0, 1); err != nil {
		return nil, err
	}
	if err := l.cloneRepository.Clone(ctx, connected.Access, clonePath, shared); err != nil {
		return nil, l.discard(ctx, clonePath, err)
	}
	root := connected.Settings.WorkspacePath()
	if err := l.place(ctx, *connected, registered.Path, root, mode); err != nil {
		return nil, l.discard(ctx, clonePath, err)
	}
	status, err := l.repoint(ctx, entity.NewLinkedWorkspace(localId, root, connected.Settings))
	if err != nil {
		return nil, err
	}

	logger.Info(ctx, "Finished", status, map[string]any{"localId": localId, "root": root})
	return status, nil
}

func (l *linkWorkspace) checkClonePath(ctx context.Context, clonePath string) error {
	registry, err := l.workspaces.findWorkspaceRegistry.Find(ctx)
	if err != nil {
		return err
	}
	if registry.OverlapsAny(clonePath) {
		return usecase.RejectClonePathOverlap(clonePath)
	}
	folder, err := l.inspectWorkspaceFolder.Inspect(ctx, clonePath)
	if err != nil {
		return err
	}
	if !folder.IsEmptyOrAbsent() {
		return usecase.RejectClonePathInUse(clonePath)
	}
	return nil
}

func (l *linkWorkspace) place(ctx context.Context, linked entity.LinkedWorkspace, source, root string, mode enums.LinkMode) error {
	usesLfs, err := l.lfs.detect(ctx, linked)
	if err != nil {
		return err
	}
	if usesLfs {
		return usecase.RejectLargeFileStorage(linked.Subfolder())
	}
	folder, err := l.inspectWorkspaceFolder.Inspect(ctx, root)
	if err != nil {
		return err
	}
	switch folder.LinkModeFor(mode) {
	case linkmode.AdoptRepo:
	case linkmode.ReplaceWithMine:
		if err := l.clearWorkspaceFolder.ClearFolder(ctx, root); err != nil {
			return err
		}
		return l.copyWorkspaceFolder.CopyInto(ctx, source, root)
	case linkmode.PlaceMine:
		return l.copyWorkspaceFolder.CopyInto(ctx, source, root)
	default:
		return usecase.RejectExistingWorkspace(linked.Subfolder())
	}
	return nil
}

func (l *linkWorkspace) repoint(ctx context.Context, linked entity.LinkedWorkspace) (*entity.SyncStatus, error) {
	workspace, err := l.status.loadWorkspace.Load(ctx, linked.Root)
	if err != nil {
		return nil, err
	}
	if workspace.Sync.Differs(linked.Settings.Shared()) {
		if err := l.saveWorkspaceSync.SaveSync(ctx, linked.Root, linked.Settings.Shared()); err != nil {
			return nil, err
		}
	}
	if err := l.saveLocalSyncSettings.SaveSyncSettings(ctx, linked.LocalID, linked.Settings); err != nil {
		return nil, err
	}
	registry, err := l.workspaces.findWorkspaceRegistry.Find(ctx)
	if err != nil {
		return nil, err
	}
	registry.Relocate(linked.LocalID, linked.Root, workspace.Name)
	if err := l.saveWorkspaceRegistry.Save(ctx, *registry); err != nil {
		return nil, err
	}
	if registry.IsActive(linked.LocalID) {
		if err := l.watchWorkspace.Watch(ctx, linked.LocalID, linked.Root); err != nil {
			return nil, err
		}
	}
	status, err := l.status.read(ctx, linked)
	if err != nil {
		return nil, err
	}
	return l.status.publish(ctx, linked.LocalID, status.SyncedAt(timeutils.Now()))
}

func (l *linkWorkspace) discard(ctx context.Context, clonePath string, cause error) error {
	if err := l.removeWorkspaceFolder.RemoveFolder(ctx, clonePath); err != nil {
		return err
	}
	return cause
}
