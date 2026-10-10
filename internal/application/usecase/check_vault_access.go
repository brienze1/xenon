package usecase

import (
	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/pkg/errs"
)

func CheckVaultWritable(vault entity.Vault, settings entity.VaultSettings) error {
	if vault.IsWritable() {
		return nil
	}
	return vaultAccessError(vault, settings)
}

func CheckVaultReadable(vault entity.Vault, settings entity.VaultSettings) error {
	if vault.IsReadable() || vault.IsEmpty() {
		return nil
	}
	return vaultAccessError(vault, settings)
}

func vaultAccessError(vault entity.Vault, settings entity.VaultSettings) error {
	switch {
	case vault.IsBroken():
		return errs.ConflictError("the vault cannot be opened; reset it to store secrets again", "VLT-06409")
	case settings.NeedsPassphraseSetup():
		return errs.ConflictError("choose a vault passphrase before storing secrets", "VLT-05409")
	default:
		return errs.ConflictError("the vault is locked; unlock it with the passphrase", "VLT-04409")
	}
}
