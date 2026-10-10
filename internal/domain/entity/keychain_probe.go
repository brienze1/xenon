package entity

import (
	"github.com/brienze1/xenon/internal/domain/enums"
	"github.com/brienze1/xenon/internal/domain/enums/vaultmode"
)

type KeychainProbe struct {
	Available        bool
	PassphraseForced bool
}

func (k KeychainProbe) VaultMode() enums.VaultMode {
	if k.Available && !k.PassphraseForced {
		return vaultmode.Keychain
	}
	return vaultmode.Passphrase
}
