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

type signInGitHub struct {
	signInWithGitHub   usecase.SignInWithGitHub
	findGitHubUser     usecase.FindGitHubUser
	uploadSshPublicKey usecase.UploadSshPublicKey
	findGitAccounts    usecase.FindGitAccounts
	saveGitAccount     usecase.SaveGitAccount
	findGitCredential  usecase.FindGitCredential
	credentials        gitCredentialStore
}

func NewSignInGitHubService(
	signInWithGitHub usecase.SignInWithGitHub,
	findGitHubUser usecase.FindGitHubUser,
	uploadSshPublicKey usecase.UploadSshPublicKey,
	findGitAccounts usecase.FindGitAccounts,
	saveGitAccount usecase.SaveGitAccount,
	findGitCredential usecase.FindGitCredential,
	findVaultSettings usecase.FindVaultSettings,
	saveVaultSettings usecase.SaveVaultSettings,
	probeKeychain usecase.ProbeKeychain,
	publishVariableEvent usecase.PublishVariableEvent,
	saveGitCredential usecase.SaveGitCredential,
) adapter.SignInGitHubService {
	return &signInGitHub{
		signInWithGitHub:   signInWithGitHub,
		findGitHubUser:     findGitHubUser,
		uploadSshPublicKey: uploadSshPublicKey,
		findGitAccounts:    findGitAccounts,
		saveGitAccount:     saveGitAccount,
		findGitCredential:  findGitCredential,
		credentials: gitCredentialStore{
			findVaultSettings:    findVaultSettings,
			saveVaultSettings:    saveVaultSettings,
			probeKeychain:        probeKeychain,
			publishVariableEvent: publishVariableEvent,
			saveGitCredential:    saveGitCredential,
		},
	}
}

func (s *signInGitHub) Start(ctx context.Context) (*entity.GitHubDeviceCode, error) {
	logger.Info(ctx, "Started")

	if err := s.credentials.ensure(ctx); err != nil {
		return nil, err
	}
	code, err := s.signInWithGitHub.Start(ctx)
	if err != nil {
		return nil, err
	}

	logger.Info(ctx, "Finished", code)
	return code, nil
}

func (s *signInGitHub) Poll(ctx context.Context, sessionId string) (*entity.GitHubSignIn, *entity.GitAccount, error) {
	logger.Info(ctx, "Started", map[string]any{"sessionId": sessionId})

	signIn, err := s.signInWithGitHub.Poll(ctx, sessionId)
	if err != nil {
		return nil, nil, err
	}
	if !signIn.IsDone() {
		logger.Info(ctx, "Finished", signIn)
		return signIn, nil, nil
	}
	user, err := s.findGitHubUser.FindUser(ctx, signIn.Token)
	if err != nil {
		return nil, nil, err
	}
	account := entity.NewGitHubAccount(idutils.NewUlid(), *user, timeutils.Now())
	if err := s.credentials.save(ctx, account.ID, signIn.Token); err != nil {
		return nil, nil, err
	}
	if err := s.saveGitAccount.Save(ctx, account); err != nil {
		return nil, nil, err
	}

	logger.Info(ctx, "Finished", signIn, account)
	return signIn, &account, nil
}

func (s *signInGitHub) UploadPublicKey(ctx context.Context, sshAccountId, gitHubAccountId, title string) error {
	logger.Info(ctx, "Started", map[string]any{"sshAccountId": sshAccountId, "gitHubAccountId": gitHubAccountId, "title": title})

	sshAccount, err := s.findGitAccounts.Find(ctx, sshAccountId)
	if err != nil {
		return err
	}
	if !sshAccount.CanUploadPublicKey() {
		return usecase.RejectPublicKeyUpload("choose an SSH account with a key file or a generated key")
	}
	gitHubAccount, err := s.findGitAccounts.Find(ctx, gitHubAccountId)
	if err != nil {
		return err
	}
	if !gitHubAccount.CanCallGitHub() {
		return usecase.RejectPublicKeyUpload("choose a GitHub account to add the key to")
	}
	token, err := s.findGitCredential.FindCredential(ctx, gitHubAccount.ID)
	if err != nil {
		return err
	}
	if err := s.uploadSshPublicKey.UploadKey(ctx, token, title, sshAccount.PublicKey); err != nil {
		return err
	}

	logger.Info(ctx, "Finished", map[string]any{"sshAccountId": sshAccountId, "gitHubAccountId": gitHubAccountId})
	return nil
}
