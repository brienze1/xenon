package service

import (
	"context"

	"github.com/brienze1/xenon/internal/application/usecase"
	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/pkg/logger"
)

type vaultWriter struct {
	findVault            usecase.FindVault
	saveVault            usecase.SaveVault
	findVaultSettings    usecase.FindVaultSettings
	saveVaultSettings    usecase.SaveVaultSettings
	probeKeychain        usecase.ProbeKeychain
	publishVariableEvent usecase.PublishVariableEvent
}

func (v vaultWriter) write(ctx context.Context, vault entity.Vault, scope entity.ScopeRef, secrets map[string]string) (entity.Vault, error) {
	settings, err := v.findVaultSettings.FindSettings(ctx)
	if err != nil {
		return entity.Vault{}, err
	}
	if settings.IsUndecided() {
		if settings, err = v.chooseMode(ctx, *settings); err != nil {
			return entity.Vault{}, err
		}
	}
	current, err := v.findVault.Find(ctx, vault.LocalID)
	if err != nil {
		return entity.Vault{}, err
	}
	if err := usecase.CheckVaultWritable(*current, *settings); err != nil {
		return entity.Vault{}, err
	}
	updated := current.WithSecrets(scope, secrets).Unlocked()
	if err := v.saveVault.Save(ctx, updated); err != nil {
		return entity.Vault{}, err
	}
	return updated, nil
}

func (v vaultWriter) restore(ctx context.Context, written entity.Vault, scope entity.ScopeRef, previous map[string]string) {
	if !written.IsReadable() {
		return
	}
	if err := v.saveVault.Save(ctx, written.WithSecrets(scope, previous)); err != nil {
		logger.Error(ctx, err, "The vault could not be restored after a failed save", map[string]any{"localId": written.LocalID})
	}
}

func (v vaultWriter) chooseMode(ctx context.Context, settings entity.VaultSettings) (*entity.VaultSettings, error) {
	chosen := settings.WithMode(v.probeKeychain.Probe(ctx))
	if err := v.saveVaultSettings.SaveSettings(ctx, chosen); err != nil {
		return nil, err
	}
	if err := v.publishVariableEvent.VaultStatusChanged(ctx, entity.NewVaultStatus(chosen, nil)); err != nil {
		return nil, err
	}
	return &chosen, nil
}
