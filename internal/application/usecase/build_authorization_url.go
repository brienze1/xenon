package usecase

import (
	"net/url"
	"strings"

	"github.com/brienze1/xenon/internal/domain/entity"
)

const (
	responseTypeCode  = "code"
	responseTypeToken = "token"
)

func BuildAuthorizationUrl(config entity.OAuth2Config, state string, pkce entity.Pkce) string {
	query := url.Values{}
	query.Set("response_type", responseTypeCode)
	if config.IsImplicit() {
		query.Set("response_type", responseTypeToken)
	}
	query.Set("client_id", config.ClientID)
	query.Set("redirect_uri", config.RedirectURI)
	query.Set("state", state)
	setIfPresent(query, "scope", config.Scope)
	if config.UsesPkce() {
		query.Set("code_challenge", pkce.Challenge)
		query.Set("code_challenge_method", pkce.Method)
	}
	setIfPresent(query, "audience", config.Audience)
	setIfPresent(query, "resource", config.Resource)
	for _, param := range config.AuthRequestParams {
		if param.IsActive() {
			query.Set(param.Key, param.Value)
		}
	}
	separator := "?"
	if strings.Contains(config.AuthURL, "?") {
		separator = "&"
	}
	return config.AuthURL + separator + query.Encode()
}

func setIfPresent(query url.Values, key, value string) {
	if value != "" {
		query.Set(key, value)
	}
}
