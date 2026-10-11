package service

import (
	"context"

	"github.com/brienze1/xenon/internal/application/adapter"
	"github.com/brienze1/xenon/internal/application/usecase"
	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/internal/domain/enums/hostkeystatus"
	"github.com/brienze1/xenon/pkg/logger"
)

type verifyGitHost struct {
	findGitAccounts usecase.FindGitAccounts
	listRemote      usecase.ListRemote
	scanHostKey     usecase.ScanHostKey
	findKnownHosts  usecase.FindKnownHosts
	trustHostKey    usecase.TrustHostKey
	access          gitAccessResolver
}

func NewVerifyGitHostService(
	findGitAccounts usecase.FindGitAccounts,
	findGitCredential usecase.FindGitCredential,
	listRemote usecase.ListRemote,
	scanHostKey usecase.ScanHostKey,
	findKnownHosts usecase.FindKnownHosts,
	trustHostKey usecase.TrustHostKey,
	publishGitEvent usecase.PublishGitEvent,
) adapter.VerifyGitHostService {
	return &verifyGitHost{
		findGitAccounts: findGitAccounts,
		listRemote:      listRemote,
		scanHostKey:     scanHostKey,
		findKnownHosts:  findKnownHosts,
		trustHostKey:    trustHostKey,
		access: gitAccessResolver{
			findGitCredential: findGitCredential,
			findKnownHosts:    findKnownHosts,
			scanHostKey:       scanHostKey,
			publishGitEvent:   publishGitEvent,
		},
	}
}

func (v *verifyGitHost) TestConnection(ctx context.Context, accountId, repository string) (*entity.RemoteInfo, error) {
	logger.Info(ctx, "Started", map[string]any{"accountId": accountId, "repository": repository})

	account, err := v.findGitAccounts.Find(ctx, accountId)
	if err != nil {
		return nil, err
	}
	access, err := v.access.resolve(ctx, *account, repository)
	if err != nil {
		return nil, err
	}
	info, err := v.listRemote.List(ctx, *access)
	if err != nil {
		return nil, err
	}

	logger.Info(ctx, "Finished", info)
	return info, nil
}

func (v *verifyGitHost) ConfirmHostKey(ctx context.Context, repository, fingerprint string, accept bool) error {
	logger.Info(ctx, "Started", map[string]any{"repository": repository, "fingerprint": fingerprint, "accept": accept})

	if !accept {
		logger.Info(ctx, "Finished", map[string]any{"repository": repository, "accepted": false})
		return nil
	}
	presented, err := v.scanHostKey.Scan(ctx, repository)
	if err != nil {
		return err
	}
	if presented == nil {
		return usecase.RejectNotSshHost(repository)
	}
	known, err := v.findKnownHosts.FindKnownHosts(ctx)
	if err != nil {
		return err
	}
	switch {
	case known.StatusOf(*presented) == hostkeystatus.Changed:
		return usecase.RejectChangedHostKey(*presented)
	case known.StatusOf(*presented) == hostkeystatus.Trusted:
	case !presented.HasFingerprint(fingerprint):
		return usecase.RejectHostKeyMismatch(*presented)
	default:
		if err := v.trustHostKey.Trust(ctx, *presented); err != nil {
			return err
		}
	}

	logger.Info(ctx, "Finished", map[string]any{"repository": repository, "accepted": true})
	return nil
}
