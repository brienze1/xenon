package usecase

import (
	"context"

	"github.com/brienze1/xenon/internal/domain/entity"
)

type RefreshOAuthToken interface {
	Refresh(ctx context.Context, config entity.OAuth2Config, token entity.OAuth2Token) (*entity.OAuth2Token, error)
}
