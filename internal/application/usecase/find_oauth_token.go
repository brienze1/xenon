package usecase

import (
	"context"

	"github.com/brienze1/xenon/internal/domain/entity"
)

type FindOAuthToken interface {
	FindToken(ctx context.Context, localId, owner string) (*entity.OAuth2Token, error)
}
