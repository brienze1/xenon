package adapter

import "github.com/brienze1/xenon/internal/application/usecase"

type VaultPersistence interface {
	usecase.FindVault
	usecase.SaveVault
	usecase.DeleteVault
	usecase.FindVaultSettings
	usecase.SaveVaultSettings
	usecase.CheckVaultPassphrase
	usecase.UnlockVaults
}
