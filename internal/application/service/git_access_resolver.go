package service

import (
	"context"

	"github.com/brienze1/xenon/internal/application/usecase"
	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/internal/domain/enums/hostkeystatus"
)

type gitAccessResolver struct {
	findGitCredential usecase.FindGitCredential
	findKnownHosts    usecase.FindKnownHosts
	scanHostKey       usecase.ScanHostKey
	publishGitEvent   usecase.PublishGitEvent
}

func (g gitAccessResolver) resolve(ctx context.Context, account entity.GitAccount, repository string) (*entity.GitAccess, error) {
	if !account.ServesRepository(repository) {
		return nil, usecase.RejectForeignRepository(repository, account.Host)
	}
	secret := ""
	if account.StoresSecret() {
		found, err := g.findGitCredential.FindCredential(ctx, account.ID)
		if err != nil {
			return nil, err
		}
		secret = found
	}
	access := entity.NewGitAccess(repository, account, secret)
	presented, err := g.scanHostKey.Scan(ctx, repository)
	if err != nil || presented == nil {
		return &access, err
	}
	known, err := g.findKnownHosts.FindKnownHosts(ctx)
	if err != nil {
		return nil, err
	}
	switch known.StatusOf(*presented) {
	case hostkeystatus.Trusted:
		trusted := access.WithTrustedHostKeys(known.KeysFor(presented.Host))
		return &trusted, nil
	case hostkeystatus.Changed:
		return nil, usecase.RejectChangedHostKey(*presented)
	default:
		if err := g.publishGitEvent.HostKeyUnknown(ctx, repository, *presented); err != nil {
			return nil, err
		}
		return nil, usecase.RejectUnknownHostKey(*presented)
	}
}
