package usecase

import "context"

type UnlockVaults interface {
	Unlock(ctx context.Context, passphrase string)
}
