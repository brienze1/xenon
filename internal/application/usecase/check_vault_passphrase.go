package usecase

import "context"

type CheckVaultPassphrase interface {
	Seal(ctx context.Context, passphrase string) (string, error)
	Verify(ctx context.Context, check, passphrase string) bool
}
