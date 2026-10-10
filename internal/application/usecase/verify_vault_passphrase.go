package usecase

import "context"

type VerifyVaultPassphrase interface {
	Verify(ctx context.Context, check, passphrase string) bool
}
