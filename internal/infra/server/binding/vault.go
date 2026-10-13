package binding

import (
	"context"

	"github.com/brienze1/xenon/internal/integration/adapter"
	"github.com/brienze1/xenon/internal/integration/entrypoint/dto"
)

type VaultService struct {
	vault adapter.VaultController
}

func NewVaultService(vault adapter.VaultController) *VaultService {
	return &VaultService{vault: vault}
}

func (v *VaultService) Status(ctx context.Context) (*dto.VaultStatus, error) {
	return v.vault.Status(ctx)
}

func (v *VaultService) SetupPassphrase(ctx context.Context, request dto.SetupPassphrase) (*dto.VaultStatus, error) {
	return v.vault.SetupPassphrase(ctx, request)
}

func (v *VaultService) Unlock(ctx context.Context, request dto.UnlockVault) (*dto.VaultStatus, error) {
	return v.vault.Unlock(ctx, request)
}

func (v *VaultService) Reveal(ctx context.Context, request dto.SecretReference) (*dto.RevealedSecret, error) {
	return v.vault.Reveal(ctx, request)
}

func (v *VaultService) CopySecret(ctx context.Context, request dto.SecretReference) error {
	return v.vault.CopySecret(ctx, request)
}

func (v *VaultService) Reset(ctx context.Context, request dto.ResetVault) (*dto.VaultStatus, error) {
	return v.vault.Reset(ctx, request)
}
