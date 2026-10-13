package model

import (
	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/internal/domain/enums/vaultmode"
)

type VaultSettings struct {
	Mode            string `json:"mode"`
	PassphraseCheck string `json:"passphraseCheck,omitempty"`
}

func NewVaultSettings(settings entity.VaultSettings) VaultSettings {
	return VaultSettings{Mode: settings.Mode.String(), PassphraseCheck: settings.PassphraseCheck}
}

func (v *VaultSettings) ToEntity() entity.VaultSettings {
	return entity.NewVaultSettings(vaultmode.GetVaultMode(v.Mode), v.PassphraseCheck)
}
