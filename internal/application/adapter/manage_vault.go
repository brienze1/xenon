package adapter

import (
	"context"

	"github.com/brienze1/xenon/internal/domain/entity"
)

type ManageVaultService interface {
	Status(ctx context.Context) (*entity.VaultStatus, error)
	SetupPassphrase(ctx context.Context, setup entity.PassphraseSetup) (*entity.VaultStatus, error)
	Unlock(ctx context.Context, passphrase string) (*entity.VaultStatus, error)
	Reveal(ctx context.Context, localId string, scope entity.ScopeRef, key string) (string, error)
	CopySecret(ctx context.Context, localId string, scope entity.ScopeRef, key string) error
	Reset(ctx context.Context, localId string, confirm bool) (*entity.VaultStatus, error)
}
