package webservice

import (
	"context"

	"github.com/brienze1/xenon/internal/domain/entity"
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

func (k *keychainWebservice) Probe(ctx context.Context) entity.KeychainProbe {
	logger.Debug(ctx, "Started")

	probe := entity.KeychainProbe{PassphraseForced: properties.Properties().Vault.ForcePassphrase}
	if !probe.PassphraseForced {
		probe.Available = k.isAvailable(ctx)
	}

	logger.Debug(ctx, "Finished", probe)
	return probe
}

func (k *keychainWebservice) isAvailable(ctx context.Context) bool {
	if err := k.keychain.Set(keychainProbeService, keychainProbeAccount, keychainProbeSecret); err != nil {
		logger.Warn(ctx, err, "The OS keychain cannot store entries")
		return false
	}
	secret, found, err := k.keychain.Get(keychainProbeService, keychainProbeAccount)
	if err != nil || !found || secret != keychainProbeSecret {
		logger.Warn(ctx, err, "The OS keychain cannot read entries back")
		return false
	}
	if err := k.keychain.Delete(keychainProbeService, keychainProbeAccount); err != nil {
		logger.Warn(ctx, err, "The OS keychain cannot delete entries")
		return false
	}
	return true
}
