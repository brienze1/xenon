package binding

import (
	"context"

	"github.com/brienze1/xenon/internal/integration/adapter"
	"github.com/brienze1/xenon/internal/integration/entrypoint/dto"
)

type OAuthService struct {
	oauth adapter.OAuthController
}

func NewOAuthService(oauth adapter.OAuthController) *OAuthService {
	return &OAuthService{oauth: oauth}
}

func (o *OAuthService) GetNewToken(ctx context.Context, request dto.OAuthTokenRequest) (*dto.TokenInfo, error) {
	return o.oauth.GetNewToken(ctx, request)
}

func (o *OAuthService) Refresh(ctx context.Context, request dto.OAuthTokenRequest) (*dto.TokenInfo, error) {
	return o.oauth.Refresh(ctx, request)
}

func (o *OAuthService) Get(ctx context.Context, request dto.OAuthTokenReference) (*dto.TokenInfo, error) {
	return o.oauth.Get(ctx, request)
}

func (o *OAuthService) Reveal(ctx context.Context, request dto.OAuthTokenPart) (*dto.RevealedSecret, error) {
	return o.oauth.Reveal(ctx, request)
}

func (o *OAuthService) Copy(ctx context.Context, request dto.OAuthTokenPart) error {
	return o.oauth.Copy(ctx, request)
}

func (o *OAuthService) Use(ctx context.Context, request dto.OAuthTokenUse) (*dto.TokenInfo, error) {
	return o.oauth.Use(ctx, request)
}

func (o *OAuthService) Clear(ctx context.Context, request dto.OAuthTokenReference) error {
	return o.oauth.Clear(ctx, request)
}
