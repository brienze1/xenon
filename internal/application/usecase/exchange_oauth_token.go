package usecase

import (
	"context"

	"github.com/brienze1/xenon/internal/domain/entity"
)

type ExchangeOAuthToken interface {
	Exchange(ctx context.Context, grant entity.OAuth2GrantRequest) (*entity.OAuth2Token, error)
}
