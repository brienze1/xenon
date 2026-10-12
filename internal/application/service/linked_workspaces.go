package service

import (
	"context"

	"github.com/brienze1/xenon/internal/application/usecase"
	"github.com/brienze1/xenon/internal/domain/entity"
)

type linkedWorkspaces struct {
	findWorkspaceRegistry usecase.FindWorkspaceRegistry
	findLocalSyncSettings usecase.FindLocalSyncSettings
	findGitAccounts       usecase.FindGitAccounts
	access                gitAccessResolver
}

func newLinkedWorkspaces(
	findWorkspaceRegistry usecase.FindWorkspaceRegistry,
	findLocalSyncSettings usecase.FindLocalSyncSettings,
	findGitAccounts usecase.FindGitAccounts,
	findGitCredential usecase.FindGitCredential,
	findKnownHosts usecase.FindKnownHosts,
	scanHostKey usecase.ScanHostKey,
	publishGitEvent usecase.PublishGitEvent,
) linkedWorkspaces {
	return linkedWorkspaces{
		findWorkspaceRegistry: findWorkspaceRegistry,
		findLocalSyncSettings: findLocalSyncSettings,
		findGitAccounts:       findGitAccounts,
		access:                gitAccessResolver{findGitCredential: findGitCredential, findKnownHosts: findKnownHosts, scanHostKey: scanHostKey, publishGitEvent: publishGitEvent},
	}
}

func (l linkedWorkspaces) settings(ctx context.Context, localId string) (*entity.RegisteredWorkspace, *entity.LocalSyncSettings, error) {
	registered, err := l.findWorkspaceRegistry.FindWorkspace(ctx, localId)
	if err != nil {
		return nil, nil, err
	}
	settings, err := l.findLocalSyncSettings.FindSyncSettings(ctx, localId)
	if err != nil {
		return nil, nil, err
	}
	return registered, settings, nil
}

func (l linkedWorkspaces) find(ctx context.Context, localId string) (*entity.LinkedWorkspace, error) {
	registered, settings, err := l.settings(ctx, localId)
	if err != nil {
		return nil, err
	}
	if !settings.IsLinked() {
		return nil, usecase.RejectNotLinked(localId)
	}
	linked := entity.NewLinkedWorkspace(localId, registered.Path, *settings)
	return &linked, nil
}

func (l linkedWorkspaces) account(ctx context.Context, linked entity.LinkedWorkspace) (*entity.GitAccount, error) {
	if !linked.Settings.HasAccount() {
		return nil, usecase.RejectMissingSyncAccount()
	}
	return l.findGitAccounts.Find(ctx, linked.Settings.AccountID)
}

func (l linkedWorkspaces) connect(ctx context.Context, linked entity.LinkedWorkspace) (*entity.LinkedWorkspace, error) {
	account, err := l.account(ctx, linked)
	if err != nil {
		return nil, err
	}
	access, err := l.access.resolve(ctx, *account, linked.Settings.Repository)
	if err != nil {
		return nil, err
	}
	connected := linked.WithAccess(*account, *access)
	return &connected, nil
}

func (l linkedWorkspaces) findConnected(ctx context.Context, localId string) (*entity.LinkedWorkspace, error) {
	linked, err := l.find(ctx, localId)
	if err != nil {
		return nil, err
	}
	return l.connect(ctx, *linked)
}
