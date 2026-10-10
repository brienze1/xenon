package entity

import (
	"github.com/brienze1/xenon/internal/domain/enums"
	"github.com/brienze1/xenon/internal/domain/enums/vaultmode"
)

type VaultSettings struct {
	Mode            enums.VaultMode
	PassphraseCheck string
}

func NewVaultSettings(mode enums.VaultMode, passphraseCheck string) VaultSettings {
	if mode == vaultmode.Unknown || mode == "" {
		mode = vaultmode.None
	}
	return VaultSettings{Mode: mode, PassphraseCheck: passphraseCheck}
}

func (v VaultSettings) IsUndecided() bool {
	return v.Mode == vaultmode.None
}

func (v VaultSettings) UsesPassphrase() bool {
	return v.Mode == vaultmode.Passphrase
}

func (v VaultSettings) HasPassphrase() bool {
	return v.PassphraseCheck != ""
}

func (v VaultSettings) NeedsPassphraseSetup() bool {
	return v.UsesPassphrase() && !v.HasPassphrase()
}

func (v VaultSettings) WithMode(mode enums.VaultMode) VaultSettings {
	v.Mode = mode
	return v
}

func (v VaultSettings) WithPassphraseCheck(check string) VaultSettings {
	v.Mode = vaultmode.Passphrase
	v.PassphraseCheck = check
	return v
}

func (v VaultSettings) LogValue() any {
	return map[string]any{"mode": v.Mode, "hasPassphrase": v.HasPassphrase()}
}
