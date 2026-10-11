package usecase

import "context"

type LockOAuthToken interface {
	LockToken(ctx context.Context, localId, owner string) (unlock func())
}
