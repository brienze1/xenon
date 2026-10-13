package entity

import (
	"slices"

	"github.com/brienze1/xenon/internal/domain/enums"
)

type VaultStatus struct {
	Mode                 enums.VaultMode
	LockedLocalIDs       []string
	BrokenLocalIDs       []string
	NeedsPassphraseSetup bool
}

func NewVaultStatus(settings VaultSettings, vaults []Vault) VaultStatus {
	status := VaultStatus{
		Mode:                 settings.Mode,
		NeedsPassphraseSetup: settings.NeedsPassphraseSetup(),
		LockedLocalIDs:       []string{},
		BrokenLocalIDs:       []string{},
	}
	for _, vault := range vaults {
		if vault.IsLocked() {
			status.LockedLocalIDs = append(status.LockedLocalIDs, vault.LocalID)
		}
		if vault.IsBroken() {
			status.BrokenLocalIDs = append(status.BrokenLocalIDs, vault.LocalID)
		}
	}
	return status
}

func (v VaultStatus) IsLocked(localId string) bool {
	return slices.Contains(v.LockedLocalIDs, localId)
}

func (v VaultStatus) IsBroken(localId string) bool {
	return slices.Contains(v.BrokenLocalIDs, localId)
}
