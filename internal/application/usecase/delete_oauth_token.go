package usecase

import "context"

type DeleteOAuthToken interface {
	DeleteToken(ctx context.Context, localId, owner string) error
}
