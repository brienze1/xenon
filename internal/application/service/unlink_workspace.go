package service

import (
	"context"

	"github.com/brienze1/xenon/internal/application/adapter"
	"github.com/brienze1/xenon/internal/application/usecase"
	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/pkg/logger"
)

type unlinkWorkspace struct {
	claimSyncOperation     usecase.ClaimSyncOperation
	saveWorkspaceRegistry  usecase.SaveWorkspaceRegistry
	saveLocalSyncSettings  usecase.SaveLocalSyncSettings
	saveSyncSession        usecase.SaveSyncSession
	inspectWorkspaceFolder usecase.InspectWorkspaceFolder
	copyWorkspaceFolder    usecase.CopyWorkspaceFolder
	removeWorkspaceFolder  usecase.RemoveWorkspaceFolder
	loadWorkspace          usecase.LoadWorkspace
	watchWorkspace         usecase.WatchWorkspace
	workspaces             linkedWorkspaces
}

func NewUnlinkWorkspaceService(
	findWorkspaceRegistry usecase.FindWorkspaceRegistry,
	saveWorkspaceRegistry usecase.SaveWorkspaceRegistry,
	findLocalSyncSettings usecase.FindLocalSyncSettings,
	saveLocalSyncSettings usecase.SaveLocalSyncSettings,
	claimSyncOperation usecase.ClaimSyncOperation,
	saveSyncSession usecase.SaveSyncSession,
	inspectWorkspaceFolder usecase.InspectWorkspaceFolder,
	copyWorkspaceFolder usecase.CopyWorkspaceFolder,
	removeWorkspaceFolder usecase.RemoveWorkspaceFolder,
	loadWorkspace usecase.LoadWorkspace,
	watchWorkspace usecase.WatchWorkspace,
) adapter.UnlinkWorkspaceService {
	return &unlinkWorkspace{
		claimSyncOperation:     claimSyncOperation,
		saveWorkspaceRegistry:  saveWorkspaceRegistry,
		saveLocalSyncSettings:  saveLocalSyncSettings,
		saveSyncSession:        saveSyncSession,
		inspectWorkspaceFolder: inspectWorkspaceFolder,
		copyWorkspaceFolder:    copyWorkspaceFolder,
		removeWorkspaceFolder:  removeWorkspaceFolder,
		loadWorkspace:          loadWorkspace,
		watchWorkspace:         watchWorkspace,
		workspaces:             linkedWorkspaces{findWorkspaceRegistry: findWorkspaceRegistry, findLocalSyncSettings: findLocalSyncSettings},
	}
}

func (u *unlinkWorkspace) Unlink(ctx context.Context, localId, targetFolder string) (*entity.RegisteredWorkspace, error) {
	logger.Info(ctx, "Started", map[string]any{"localId": localId, "targetFolder": targetFolder})

	release, err := u.claimSyncOperation.Claim(ctx, localId)
	if err != nil {
		return nil, err
	}
	defer release()
	linked, err := u.workspaces.find(ctx, localId)
	if err != nil {
		return nil, err
	}
	if err := u.checkTarget(ctx, *linked, targetFolder); err != nil {
		return nil, err
	}
	if err := u.copyWorkspaceFolder.CopyFolder(ctx, linked.Root, targetFolder); err != nil {
		return nil, err
	}
	if err := u.saveLocalSyncSettings.SaveSyncSettings(ctx, localId, entity.LocalSyncSettings{}); err != nil {
		return nil, err
	}
	if err := u.saveSyncSession.DiscardSession(ctx, localId); err != nil {
		return nil, err
	}
	registered, err := u.repoint(ctx, localId, targetFolder)
	if err != nil {
		return nil, err
	}
	if err := u.removeWorkspaceFolder.RemoveFolder(ctx, linked.ClonePath()); err != nil {
		return nil, err
	}

	logger.Info(ctx, "Finished", map[string]any{"localId": localId, "targetFolder": targetFolder, "result": registered})
	return registered, nil
}

func (u *unlinkWorkspace) checkTarget(ctx context.Context, linked entity.LinkedWorkspace, targetFolder string) error {
	registry, err := u.workspaces.findWorkspaceRegistry.Find(ctx)
	if err != nil {
		return err
	}
	if registry.OverlapsAny(targetFolder) || linked.Settings.Encloses(targetFolder) {
		return usecase.RejectUnlinkTargetOverlap(targetFolder)
	}
	folder, err := u.inspectWorkspaceFolder.Inspect(ctx, targetFolder)
	if err != nil {
		return err
	}
	if !folder.IsEmptyOrAbsent() {
		return usecase.RejectUnlinkTargetInUse(targetFolder)
	}
	return nil
}

func (u *unlinkWorkspace) repoint(ctx context.Context, localId, targetFolder string) (*entity.RegisteredWorkspace, error) {
	workspace, err := u.loadWorkspace.Load(ctx, targetFolder)
	if err != nil {
		return nil, err
	}
	registry, err := u.workspaces.findWorkspaceRegistry.Find(ctx)
	if err != nil {
		return nil, err
	}
	registered, _ := registry.Relocate(localId, targetFolder, workspace.Name)
	if err := u.saveWorkspaceRegistry.Save(ctx, *registry); err != nil {
		return nil, err
	}
	if registry.IsActive(localId) {
		if err := u.watchWorkspace.Watch(ctx, localId, targetFolder); err != nil {
			return nil, err
		}
	}
	return registered, nil
}
