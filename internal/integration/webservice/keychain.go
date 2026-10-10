package webservice

import (
	"context"

	"github.com/brienze1/xenon/internal/domain/enums"
	"github.com/brienze1/xenon/internal/domain/enums/vaultmode"
	"github.com/brienze1/xenon/internal/infra/properties"
	"github.com/brienze1/xenon/internal/integration/adapter"
	"github.com/brienze1/xenon/pkg/logger"
)

const (
	keychainProbeService = "dev.brienze.xenon.vault"
	keychainProbeAccount = "probe"
	keychainProbeSecret  = "xenon-keychain-probe"
)

type keychainWebservice struct {
	keychain adapter.Keychain
}

func NewKeychainWebservice(keychain adapter.Keychain) adapter.KeychainWebservice {
	return &keychainWebservice{keychain: keychain}
}

func (k *keychainWebservice) Probe(ctx context.Context) enums.VaultMode {
	logger.Debug(ctx, "Started")

	mode := k.probe(ctx)

	logger.Debug(ctx, "Finished", map[string]any{"mode": mode})
	return mode
}

func (k *keychainWebservice) probe(ctx context.Context) enums.VaultMode {
	if properties.Properties().Vault.ForcePassphrase {
		return vaultmode.Passphrase
	}
	if err := k.keychain.Set(keychainProbeService, keychainProbeAccount, keychainProbeSecret); err != nil {
		logger.Warn(ctx, err, "The OS keychain cannot store entries, using a passphrase")
		return vaultmode.Passphrase
	}
	secret, found, err := k.keychain.Get(keychainProbeService, keychainProbeAccount)
	if err != nil || !found || secret != keychainProbeSecret {
		logger.Warn(ctx, err, "The OS keychain cannot read entries back, using a passphrase")
		return vaultmode.Passphrase
	}
	if err := k.keychain.Delete(keychainProbeService, keychainProbeAccount); err != nil {
		logger.Warn(ctx, err, "The OS keychain cannot delete entries, using a passphrase")
		return vaultmode.Passphrase
	}
	return vaultmode.Keychain
}
