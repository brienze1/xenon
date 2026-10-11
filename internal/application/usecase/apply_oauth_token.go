package usecase

import "github.com/brienze1/xenon/internal/domain/entity"

const accessTokenParam = "access_token"

func ApplyOAuthToken(request entity.EffectiveRequest, config entity.OAuth2Config, token *entity.OAuth2Token) entity.EffectiveRequest {
	if token == nil || token.IsMac() || token.Value() == "" {
		return request
	}
	if config.AddsTokenToQuery() {
		return request.WithQueryParam(accessTokenParam, token.Value())
	}
	return request.WithHeader(entity.HeaderAuthorization, config.AuthorizationHeader(token.Value()))
}
