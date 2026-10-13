package service

import (
	"context"

	"github.com/brienze1/xenon/internal/application/adapter"
	"github.com/brienze1/xenon/internal/application/usecase"
	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/pkg/logger"
)

type manageGitAccounts struct {
	findGitAccounts       usecase.FindGitAccounts
	saveGitAccount        usecase.SaveGitAccount
	deleteGitAccount      usecase.DeleteGitAccount
	deleteGitCredential   usecase.DeleteGitCredential
	deleteSshKey          usecase.DeleteSshKey
	findWorkspaceRegistry usecase.FindWorkspaceRegistry
	findLocalSyncSettings usecase.FindLocalSyncSettings
	saveLocalSyncSettings usecase.SaveLocalSyncSettings
}

func NewManageGitAccountsService(
	findGitAccounts usecase.FindGitAccounts,
	saveGitAccount usecase.SaveGitAccount,
	deleteGitAccount usecase.DeleteGitAccount,
	deleteGitCredential usecase.DeleteGitCredential,
	deleteSshKey usecase.DeleteSshKey,
	findWorkspaceRegistry usecase.FindWorkspaceRegistry,
	findLocalSyncSettings usecase.FindLocalSyncSettings,
	saveLocalSyncSettings usecase.SaveLocalSyncSettings,
) adapter.ManageGitAccountsService {
	return &manageGitAccounts{
		findGitAccounts:       findGitAccounts,
		saveGitAccount:        saveGitAccount,
		deleteGitAccount:      deleteGitAccount,
		deleteGitCredential:   deleteGitCredential,
		deleteSshKey:          deleteSshKey,
		findWorkspaceRegistry: findWorkspaceRegistry,
		findLocalSyncSettings: findLocalSyncSettings,
		saveLocalSyncSettings: saveLocalSyncSettings,
	}
}

func (m *manageGitAccounts) List(ctx context.Context) ([]entity.GitAccount, error) {
	logger.Info(ctx, "Started")

	accounts, err := m.findGitAccounts.FindAll(ctx)
	if err != nil {
		return nil, err
	}

	logger.Info(ctx, "Finished", map[string]any{"accounts": len(accounts)})
	return accounts, nil
}

func (m *manageGitAccounts) Update(ctx context.Context, id, displayName, commitName, commitEmail string) (*entity.GitAccount, error) {
	logger.Info(ctx, "Started", map[string]any{"id": id, "displayName": displayName, "commitName": commitName, "commitEmail": commitEmail})

	account, err := m.findGitAccounts.Find(ctx, id)
	if err != nil {
		return nil, err
	}
	updated := account.WithIdentity(displayName, commitName, commitEmail)
	if err := m.saveGitAccount.Save(ctx, updated); err != nil {
		return nil, err
	}

	logger.Info(ctx, "Finished", updated)
	return &updated, nil
}

func (m *manageGitAccounts) Remove(ctx context.Context, id string) error {
	logger.Info(ctx, "Started", map[string]any{"id": id})

	account, err := m.findGitAccounts.Find(ctx, id)
	if err != nil {
		return err
	}
	if err := m.deleteGitCredential.DeleteCredential(ctx, id); err != nil {
		return err
	}
	if account.IsGenerated() {
		if err := m.deleteSshKey.DeleteKey(ctx, account.KeyPath); err != nil {
			return err
		}
	}
	if err := m.deleteGitAccount.Delete(ctx, id); err != nil {
		return err
	}
	if err := m.unlinkWorkspaces(ctx, id); err != nil {
		return err
	}

	logger.Info(ctx, "Finished", map[string]any{"id": id})
	return nil
}

func (m *manageGitAccounts) unlinkWorkspaces(ctx context.Context, accountId string) error {
	registry, err := m.findWorkspaceRegistry.Find(ctx)
	if err != nil {
		return err
	}
	for _, localId := range registry.LocalIDs() {
		settings, err := m.findLocalSyncSettings.FindSyncSettings(ctx, localId)
		if err != nil {
			return err
		}
		if !settings.UsesAccount(accountId) {
			continue
		}
		if err := m.saveLocalSyncSettings.SaveSyncSettings(ctx, localId, settings.WithoutAccount()); err != nil {
			return err
		}
	}
	return nil
}
