package adapter

import (
	"context"

	"github.com/brienze1/xenon/internal/integration/entrypoint/dto"
)

type OAuthController interface {
	GetNewToken(ctx context.Context, request dto.OAuthTokenRequest) (*dto.TokenInfo, error)
	Refresh(ctx context.Context, request dto.OAuthTokenRequest) (*dto.TokenInfo, error)
	Get(ctx context.Context, request dto.OAuthTokenReference) (*dto.TokenInfo, error)
	Reveal(ctx context.Context, request dto.OAuthTokenPart) (*dto.RevealedSecret, error)
	Copy(ctx context.Context, request dto.OAuthTokenPart) error
	Use(ctx context.Context, request dto.OAuthTokenUse) (*dto.TokenInfo, error)
	Clear(ctx context.Context, request dto.OAuthTokenReference) error
}
