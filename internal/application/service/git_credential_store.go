package service

import (
	"context"

	"github.com/brienze1/xenon/internal/application/usecase"
)

type gitCredentialStore struct {
	findVaultSettings    usecase.FindVaultSettings
	saveVaultSettings    usecase.SaveVaultSettings
	probeKeychain        usecase.ProbeKeychain
	publishVariableEvent usecase.PublishVariableEvent
	saveGitCredential    usecase.SaveGitCredential
}

func (g gitCredentialStore) ensure(ctx context.Context) error {
	settings, err := g.findVaultSettings.FindSettings(ctx)
	if err != nil {
		return err
	}
	if settings.IsUndecided() {
		writer := vaultWriter{saveVaultSettings: g.saveVaultSettings, probeKeychain: g.probeKeychain, publishVariableEvent: g.publishVariableEvent}
		if _, err := writer.chooseMode(ctx, *settings); err != nil {
			return err
		}
	}
	return g.saveGitCredential.CheckCredentialStore(ctx)
}

func (g gitCredentialStore) save(ctx context.Context, accountId, secret string) error {
	return g.saveGitCredential.SaveCredential(ctx, accountId, secret)
}
