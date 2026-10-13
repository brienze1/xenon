package dto

import "github.com/brienze1/xenon/internal/domain/entity"

type VaultStatus struct {
	Mode                 string   `json:"mode"`
	LockedLocalIds       []string `json:"locked_local_ids"`
	BrokenLocalIds       []string `json:"broken_local_ids"`
	NeedsPassphraseSetup bool     `json:"needs_passphrase_setup"`
}

func NewVaultStatus(status entity.VaultStatus) VaultStatus {
	return VaultStatus{
		Mode:                 status.Mode.String(),
		NeedsPassphraseSetup: status.NeedsPassphraseSetup,
		LockedLocalIds:       append([]string{}, status.LockedLocalIDs...),
		BrokenLocalIds:       append([]string{}, status.BrokenLocalIDs...),
	}
}
