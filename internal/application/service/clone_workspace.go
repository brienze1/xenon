package service

import (
	"context"

	"github.com/brienze1/xenon/internal/application/adapter"
	"github.com/brienze1/xenon/internal/application/usecase"
	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/pkg/idutils"
	"github.com/brienze1/xenon/pkg/logger"
	"github.com/brienze1/xenon/pkg/timeutils"
)

type cloneWorkspace struct {
	locateClone            usecase.LocateClone
	listRemote             usecase.ListRemote
	saveWorkspaceRegistry  usecase.SaveWorkspaceRegistry
	inspectWorkspaceFolder usecase.InspectWorkspaceFolder
	loadWorkspace          usecase.LoadWorkspace
	saveWorkspaceFile      usecase.SaveWorkspaceFile
	saveWorkspaceSync      usecase.SaveWorkspaceSync
	saveLocalSyncSettings  usecase.SaveLocalSyncSettings
	workspaces             linkedWorkspaces
	cloner                 syncCloner
}

func NewCloneWorkspaceService(
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
	listRemote usecase.ListRemote,
	cloneRepository usecase.CloneRepository,
	readBranchHeads usecase.ReadBranchHeads,
	readTreeVersions usecase.ReadTreeVersions,
	inspectWorkspaceFolder usecase.InspectWorkspaceFolder,
	removeWorkspaceFolder usecase.RemoveWorkspaceFolder,
	loadWorkspace usecase.LoadWorkspace,
	saveWorkspaceFile usecase.SaveWorkspaceFile,
	saveWorkspaceSync usecase.SaveWorkspaceSync,
) adapter.CloneWorkspaceService {
	return &cloneWorkspace{
		locateClone:            locateClone,
		listRemote:             listRemote,
		saveWorkspaceRegistry:  saveWorkspaceRegistry,
		inspectWorkspaceFolder: inspectWorkspaceFolder,
		loadWorkspace:          loadWorkspace,
		saveWorkspaceFile:      saveWorkspaceFile,
		saveWorkspaceSync:      saveWorkspaceSync,
		saveLocalSyncSettings:  saveLocalSyncSettings,
		workspaces:             newLinkedWorkspaces(findWorkspaceRegistry, findLocalSyncSettings, findGitAccounts, findGitCredential, findKnownHosts, scanHostKey, publishGitEvent),
		cloner: syncCloner{
			findWorkspaceRegistry: findWorkspaceRegistry, inspectWorkspaceFolder: inspectWorkspaceFolder, cloneRepository: cloneRepository,
			removeWorkspaceFolder: removeWorkspaceFolder, publishGitEvent: publishGitEvent,
			lfs: lfsDetector{readBranchHeads: readBranchHeads, readTreeVersions: readTreeVersions},
		},
	}
}

func (c *cloneWorkspace) Clone(ctx context.Context, shared entity.SyncSettings, accountId, clonePath, name string) (*entity.RegisteredWorkspace, error) {
	logger.Info(ctx, "Started", map[string]any{"repository": shared.Repository, "branch": shared.Branch, "subfolder": shared.Subfolder, "accountId": accountId, "clonePath": clonePath, "name": name})

	localId := idutils.NewUlid()
	if clonePath == "" {
		clonePath = c.locateClone.DefaultClonePath(ctx, localId)
	}
	if err := c.cloner.checkClonePath(ctx, clonePath); err != nil {
		return nil, err
	}
	linked, err := c.connect(ctx, localId, shared, accountId, clonePath)
	if err != nil {
		return nil, err
	}
	if err := c.cloner.clone(ctx, *linked); err != nil {
		return nil, err
	}
	root := linked.Settings.WorkspacePath()
	workspace, err := c.open(ctx, root, linked.Settings.Shared(), name)
	if err != nil {
		return nil, c.cloner.discard(ctx, clonePath, err)
	}
	if err := c.saveLocalSyncSettings.SaveSyncSettings(ctx, localId, linked.Settings); err != nil {
		return nil, err
	}
	registered, err := c.register(ctx, entity.NewRegisteredWorkspace(localId, root, workspace.Name, timeutils.Now()))
	if err != nil {
		return nil, err
	}

	logger.Info(ctx, "Finished", map[string]any{"localId": localId, "root": root, "result": registered})
	return registered, nil
}

func (c *cloneWorkspace) connect(ctx context.Context, localId string, shared entity.SyncSettings, accountId, clonePath string) (*entity.LinkedWorkspace, error) {
	connected, err := c.workspaces.connect(ctx, entity.NewLinkedWorkspace(localId, "", entity.NewLocalSyncSettings(shared, accountId, clonePath)))
	if err != nil || shared.HasBranch() {
		return connected, err
	}
	remote, err := c.listRemote.List(ctx, connected.Access)
	if err != nil {
		return nil, err
	}
	onDefaultBranch := connected.WithSettings(entity.NewLocalSyncSettings(shared.OnBranch(remote.CloneBranch()), accountId, clonePath))
	return &onDefaultBranch, nil
}

func (c *cloneWorkspace) open(ctx context.Context, root string, shared entity.SyncSettings, name string) (*entity.Workspace, error) {
	folder, err := c.inspectWorkspaceFolder.Inspect(ctx, root)
	if err != nil {
		return nil, err
	}
	if !folder.IsWorkspace() {
		created := entity.NewWorkspace(idutils.NewUlid(), shared.WorkspaceNameOr(name))
		if err := c.saveWorkspaceFile.Create(ctx, root, created); err != nil {
			return nil, err
		}
	}
	workspace, err := c.loadWorkspace.Load(ctx, root)
	if err != nil {
		return nil, err
	}
	if workspace.Sync.Differs(shared) {
		if err := c.saveWorkspaceSync.SaveSync(ctx, root, shared); err != nil {
			return nil, err
		}
	}
	return workspace, nil
}

func (c *cloneWorkspace) register(ctx context.Context, registered entity.RegisteredWorkspace) (*entity.RegisteredWorkspace, error) {
	registry, err := c.workspaces.findWorkspaceRegistry.Find(ctx)
	if err != nil {
		return nil, err
	}
	registry.Register(registered)
	if err := c.saveWorkspaceRegistry.Save(ctx, *registry); err != nil {
		return nil, err
	}
	return &registered, nil
}
