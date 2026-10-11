package usecase

import (
	"context"

	"github.com/brienze1/xenon/internal/domain/entity"
)

type SaveOAuthToken interface {
	SaveToken(ctx context.Context, localId string, token entity.OAuth2Token) error
}
