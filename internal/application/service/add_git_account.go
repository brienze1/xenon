package service

import (
	"context"

	"github.com/brienze1/xenon/internal/application/adapter"
	"github.com/brienze1/xenon/internal/application/usecase"
	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/internal/domain/enums/sshmode"
	"github.com/brienze1/xenon/pkg/idutils"
	"github.com/brienze1/xenon/pkg/logger"
	"github.com/brienze1/xenon/pkg/timeutils"
)

type addGitAccount struct {
	saveGitAccount usecase.SaveGitAccount
	verifyGitToken usecase.VerifyGitToken
	findGitHubUser usecase.FindGitHubUser
	checkSshAgent  usecase.CheckSshAgent
	readSshKey     usecase.ReadSshKey
	generateSshKey usecase.GenerateSshKey
	credentials    gitCredentialStore
}

func NewAddGitAccountService(
	saveGitAccount usecase.SaveGitAccount,
	verifyGitToken usecase.VerifyGitToken,
	findGitHubUser usecase.FindGitHubUser,
	checkSshAgent usecase.CheckSshAgent,
	readSshKey usecase.ReadSshKey,
	generateSshKey usecase.GenerateSshKey,
	findVaultSettings usecase.FindVaultSettings,
	saveVaultSettings usecase.SaveVaultSettings,
	probeKeychain usecase.ProbeKeychain,
	publishVariableEvent usecase.PublishVariableEvent,
	saveGitCredential usecase.SaveGitCredential,
) adapter.AddGitAccountService {
	return &addGitAccount{
		saveGitAccount: saveGitAccount,
		verifyGitToken: verifyGitToken,
		findGitHubUser: findGitHubUser,
		checkSshAgent:  checkSshAgent,
		readSshKey:     readSshKey,
		generateSshKey: generateSshKey,
		credentials: gitCredentialStore{
			findVaultSettings:    findVaultSettings,
			saveVaultSettings:    saveVaultSettings,
			probeKeychain:        probeKeychain,
			publishVariableEvent: publishVariableEvent,
			saveGitCredential:    saveGitCredential,
		},
	}
}

func (a *addGitAccount) AddToken(ctx context.Context, draft entity.GitAccount, token, repository string) (*entity.GitAccount, error) {
	logger.Info(ctx, "Started", draft, map[string]any{"repository": repository})

	account := entity.NewTokenAccount(idutils.NewUlid(), draft.Provider, draft.Host, draft.Username, timeutils.Now()).
		WithIdentity(draft.DisplayName, draft.CommitName, draft.CommitEmail)
	if repository != "" && !account.ServesRepository(repository) {
		return nil, usecase.RejectForeignRepository(repository, account.Host)
	}
	if repository == "" && !account.CanCallGitHub() {
		return nil, usecase.RejectUnverifiableToken()
	}
	if err := a.credentials.ensure(ctx); err != nil {
		return nil, err
	}
	verified, err := a.verify(ctx, account, token, repository)
	if err != nil {
		return nil, err
	}
	if err := a.store(ctx, verified, token); err != nil {
		return nil, err
	}

	logger.Info(ctx, "Finished", verified)
	return &verified, nil
}

func (a *addGitAccount) AddSsh(ctx context.Context, draft entity.GitAccount, keyPath, passphrase string) (*entity.GitAccount, string, error) {
	logger.Info(ctx, "Started", draft, map[string]any{"keyPath": keyPath, "hasPassphrase": passphrase != ""})

	id := idutils.NewUlid()
	key, err := a.key(ctx, id, draft, keyPath, passphrase)
	if err != nil {
		return nil, "", err
	}
	account := entity.NewSshAccount(id, draft.SshMode, draft.Host, *key, passphrase != "", timeutils.Now()).
		WithIdentity(draft.DisplayName, draft.CommitName, draft.CommitEmail)
	if account.StoresSecret() {
		if err := a.credentials.ensure(ctx); err != nil {
			return nil, "", err
		}
	}
	if err := a.store(ctx, account, passphrase); err != nil {
		return nil, "", err
	}

	logger.Info(ctx, "Finished", account)
	return &account, key.PublicKey, nil
}

func (a *addGitAccount) verify(ctx context.Context, account entity.GitAccount, token, repository string) (entity.GitAccount, error) {
	if repository != "" {
		_, err := a.verifyGitToken.Verify(ctx, entity.NewGitAccess(repository, account, token))
		return account, err
	}
	user, err := a.findGitHubUser.FindUser(ctx, token)
	if err != nil {
		return account, err
	}
	return account.WithGitHubUser(*user), nil
}

func (a *addGitAccount) key(ctx context.Context, id string, draft entity.GitAccount, keyPath, passphrase string) (*entity.SshKey, error) {
	switch draft.SshMode {
	case sshmode.Agent:
		return &entity.SshKey{}, a.checkSshAgent.CheckAgent(ctx)
	case sshmode.Generated:
		return a.generateSshKey.Generate(ctx, id)
	default:
		return a.readSshKey.Read(ctx, keyPath, passphrase)
	}
}

func (a *addGitAccount) store(ctx context.Context, account entity.GitAccount, secret string) error {
	if account.StoresSecret() {
		if err := a.credentials.save(ctx, account.ID, secret); err != nil {
			return err
		}
	}
	return a.saveGitAccount.Save(ctx, account)
}
