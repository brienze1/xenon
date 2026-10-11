package webservice

import (
	"context"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/brienze1/xenon/internal/application/usecase"
	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/internal/integration/adapter"
	"github.com/brienze1/xenon/internal/integration/webservice/dto"
	"github.com/brienze1/xenon/pkg/errs"
	httpClient "github.com/brienze1/xenon/pkg/http"
	"github.com/brienze1/xenon/pkg/logger"
	"github.com/brienze1/xenon/pkg/timeutils"
)

const tokenEndpointTimeout = 30 * time.Second

type oauth2Webservice struct {
	newClient httpClient.ClientFactory
}

func NewOAuth2Webservice(newClient httpClient.ClientFactory) adapter.OAuth2Webservice {
	return &oauth2Webservice{newClient: newClient}
}

func (o *oauth2Webservice) Exchange(ctx context.Context, grant entity.OAuth2GrantRequest) (*entity.OAuth2Token, error) {
	logger.Debug(ctx, "Started", grant)

	token, err := o.request(ctx, dto.NewOAuth2TokenRequest(grant), grant.Config.Scope)
	if err != nil {
		return nil, err
	}

	logger.Debug(ctx, "Finished", token)
	return token, nil
}

func (o *oauth2Webservice) Refresh(ctx context.Context, config entity.OAuth2Config, token entity.OAuth2Token) (*entity.OAuth2Token, error) {
	logger.Debug(ctx, "Started", map[string]any{"config": config, "token": token})

	fresh, err := o.request(ctx, dto.NewOAuth2RefreshRequest(config, token.RefreshToken), config.Scope)
	if err != nil {
		return nil, err
	}

	logger.Debug(ctx, "Finished", fresh)
	return fresh, nil
}

func (o *oauth2Webservice) request(ctx context.Context, tokenRequest dto.OAuth2TokenRequest, scope string) (*entity.OAuth2Token, error) {
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, tokenRequest.Target(), strings.NewReader(tokenRequest.Body()))
	if err != nil {
		return nil, usecase.RejectInvalidOAuthConfig(err.Error())
	}
	request.Header = tokenRequest.Header
	client := o.newClient.Build(httpClient.Options{Timeout: tokenEndpointTimeout, FollowRedirects: true, VerifySsl: true})
	defer client.CloseIdleConnections()
	response, err := client.Do(ctx, request)
	if err != nil {
		return nil, errs.ServiceUnavailableError(err, "OAU-11503", "the OAuth 2.0 token endpoint could not be reached")
	}
	defer func() { _ = response.Body.Close() }()
	body, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, errs.ServiceUnavailableError(err, "OAU-12503", "the OAuth 2.0 token endpoint answer could not be read")
	}
	answer := dto.NewOAuth2TokenResponse(body)
	if answer.HasError() {
		return nil, usecase.RejectOAuthProviderError(answer.ErrorCode(response.StatusCode), answer.ErrorDescription)
	}
	token := answer.ToEntity(timeutils.Now(), scope)
	return &token, nil
}
