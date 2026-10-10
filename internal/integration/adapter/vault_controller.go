package adapter

import (
	"context"

	"github.com/brienze1/xenon/internal/integration/entrypoint/dto"
)

type VaultController interface {
	Status(ctx context.Context) (*dto.VaultStatus, error)
	SetupPassphrase(ctx context.Context, request dto.SetupPassphrase) (*dto.VaultStatus, error)
	Unlock(ctx context.Context, request dto.UnlockVault) (*dto.VaultStatus, error)
	Reveal(ctx context.Context, request dto.SecretReference) (*dto.RevealedSecret, error)
	CopySecret(ctx context.Context, request dto.SecretReference) error
	Reset(ctx context.Context, request dto.ResetVault) (*dto.VaultStatus, error)
}
