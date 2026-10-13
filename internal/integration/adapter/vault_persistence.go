package adapter

import "github.com/brienze1/xenon/internal/application/usecase"

type VaultPersistence interface {
	usecase.FindVault
	usecase.SaveVault
	usecase.DeleteVault
	usecase.FindVaultSettings
	usecase.SaveVaultSettings
	usecase.SealVaultPassphrase
	usecase.VerifyVaultPassphrase
	usecase.UnlockVaults
	usecase.LockOAuthToken
	usecase.FindCookies
	usecase.SaveCookies
	usecase.SaveGitCredential
	usecase.FindGitCredential
	usecase.DeleteGitCredential
}
