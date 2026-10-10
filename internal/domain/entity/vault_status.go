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

func (v VaultStatus) IsLocked(localId string) bool {
	return slices.Contains(v.LockedLocalIDs, localId)
}

func (v VaultStatus) IsBroken(localId string) bool {
	return slices.Contains(v.BrokenLocalIDs, localId)
}
