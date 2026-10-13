package adapter

import (
	"context"

	"github.com/brienze1/xenon/internal/domain/entity"
)

type ManageOAuthService interface {
	GetNewToken(ctx context.Context, localId, owner string, draft *entity.OAuth2Config) (*entity.OAuth2Token, error)
	Refresh(ctx context.Context, localId, owner string, draft *entity.OAuth2Config) (*entity.OAuth2Token, error)
	Get(ctx context.Context, localId, owner string) (*entity.OAuth2Token, error)
	Reveal(ctx context.Context, localId, owner, which string) (string, error)
	Copy(ctx context.Context, localId, owner, which string) error
	Use(ctx context.Context, localId, owner, useToken string) (*entity.OAuth2Token, error)
	Clear(ctx context.Context, localId, owner string) error
}
