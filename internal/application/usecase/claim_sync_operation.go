package usecase

import "context"

type ClaimSyncOperation interface {
	Claim(ctx context.Context, localId string) (release func(), err error)
}
