package usecase

import "context"

type SealVaultPassphrase interface {
	Seal(ctx context.Context, passphrase string) (string, error)
}
