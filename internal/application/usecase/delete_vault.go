package usecase

import "context"

type DeleteVault interface {
	Delete(ctx context.Context, localId string) error
}
