package entity

import (
	"github.com/brienze1/xenon/internal/domain/enums"
	"github.com/brienze1/xenon/internal/domain/enums/vaultstate"
)

type VaultStorage struct {
	Secrets           map[string]map[string]string
	LocalID           string
	Mode              enums.VaultMode
	KeyFound          bool
	KeySealed         bool
	KeyUnreadable     bool
	KeyAvailable      bool
	ContentFound      bool
	ContentUnreadable bool
}

func (v VaultStorage) CanOpenContent() bool {
	return v.KeyFound && v.ContentFound && !v.KeySealed && !v.KeyUnreadable
}

func (v VaultStorage) State() enums.VaultState {
	switch {
	case v.KeySealed:
		return vaultstate.Locked
	case v.KeyUnreadable || v.KeyFound != v.ContentFound:
		return vaultstate.Broken
	case !v.KeyFound:
		return vaultstate.Empty
	case v.ContentUnreadable:
		return vaultstate.Broken
	default:
		return vaultstate.Unlocked
	}
}
