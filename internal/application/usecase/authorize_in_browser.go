package usecase

import (
	"context"

	"github.com/brienze1/xenon/internal/domain/entity"
)

type AuthorizeInBrowser interface {
	Authorize(ctx context.Context, authorizationUrl, redirectUri string) (*entity.OAuth2Callback, error)
}
