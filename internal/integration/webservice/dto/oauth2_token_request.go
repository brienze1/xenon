package dto

import (
	"encoding/base64"
	"net/http"
	"net/url"
	"strings"

	"github.com/brienze1/xenon/internal/domain/entity"
)

const (
	grantAuthorizationCode = "authorization_code"
	grantClientCredentials = "client_credentials"
	grantPassword          = "password"
	grantRefreshToken      = "refresh_token"
	formContentType        = "application/x-www-form-urlencoded"
)

type OAuth2TokenRequest struct {
	Header http.Header
	Query  url.Values
	Form   url.Values
	Url    string
}

func NewOAuth2TokenRequest(grant entity.OAuth2GrantRequest) OAuth2TokenRequest {
	config := grant.Config
	request := newOAuth2TokenRequest(config, config.AccessTokenURL)
	switch {
	case config.UsesAuthorizationCode():
		request.Form.Set("grant_type", grantAuthorizationCode)
		request.Form.Set("code", grant.Code)
		request.Form.Set("redirect_uri", grant.RedirectURI)
		if config.UsesPkce() {
			request.Form.Set("code_verifier", grant.Verifier)
		}
	case config.IsPasswordGrant():
		request.Form.Set("grant_type", grantPassword)
		request.Form.Set("username", config.Username)
		request.Form.Set("password", config.Password)
		request.setScope(config.Scope)
	default:
		request.Form.Set("grant_type", grantClientCredentials)
		request.setScope(config.Scope)
	}
	request.setIfPresent("audience", config.Audience)
	request.setIfPresent("resource", config.Resource)
	return request.withParams(config.TokenRequestParams)
}

func NewOAuth2RefreshRequest(config entity.OAuth2Config, refreshToken string) OAuth2TokenRequest {
	request := newOAuth2TokenRequest(config, config.RefreshURL())
	request.Form.Set("grant_type", grantRefreshToken)
	request.Form.Set("refresh_token", refreshToken)
	return request.withParams(config.RefreshRequestParams)
}

func newOAuth2TokenRequest(config entity.OAuth2Config, target string) OAuth2TokenRequest {
	request := OAuth2TokenRequest{Url: target, Header: http.Header{}, Query: url.Values{}, Form: url.Values{}}
	request.Header.Set("Content-Type", formContentType)
	request.Header.Set("Accept", "application/json")
	if config.SendsClientInBody() {
		request.setIfPresent("client_id", config.ClientID)
		request.setIfPresent("client_secret", config.ClientSecret)
		return request
	}
	credentials := base64.StdEncoding.EncodeToString([]byte(config.ClientID + ":" + config.ClientSecret))
	request.Header.Set("Authorization", "Basic "+credentials)
	return request
}

func (o OAuth2TokenRequest) setScope(scope string) {
	o.setIfPresent("scope", scope)
}

func (o OAuth2TokenRequest) setIfPresent(key, value string) {
	if value != "" {
		o.Form.Set(key, value)
	}
}

func (o OAuth2TokenRequest) withParams(params []entity.OAuth2Param) OAuth2TokenRequest {
	for _, param := range params {
		switch {
		case !param.IsActive():
		case param.SendsInUrl():
			o.Query.Set(param.Key, param.Value)
		case param.SendsInHeader():
			o.Header.Set(param.Key, param.Value)
		default:
			o.Form.Set(param.Key, param.Value)
		}
	}
	return o
}

func (o OAuth2TokenRequest) Target() string {
	if len(o.Query) == 0 {
		return o.Url
	}
	separator := "?"
	if strings.Contains(o.Url, "?") {
		separator = "&"
	}
	return o.Url + separator + o.Query.Encode()
}

func (o OAuth2TokenRequest) Body() string {
	return o.Form.Encode()
}
