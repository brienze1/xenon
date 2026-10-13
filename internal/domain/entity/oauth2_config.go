package entity

import (
	"net/url"
	"strings"

	"github.com/brienze1/xenon/internal/domain/enums"
	"github.com/brienze1/xenon/internal/domain/enums/clientauthentication"
	"github.com/brienze1/xenon/internal/domain/enums/oauth2grant"
	"github.com/brienze1/xenon/internal/domain/enums/tokenplacement"
)

const (
	DefaultOAuth2Callback = "http://127.0.0.1:53682/callback"
	DefaultHeaderPrefix   = "Bearer"
	UseAccessToken        = "access_token"
	UseIDToken            = "id_token"
	postmanCallbackHost   = "oauth.pstmn.io"
	schemeHttps           = "https"
	schemeHttp            = "http"
)

type OAuth2Config struct {
	GrantType            enums.OAuth2Grant
	ClientAuthentication enums.ClientAuthentication
	AddTokenTo           enums.TokenPlacement
	TokenName            string
	RedirectURI          string
	AuthURL              string
	AccessTokenURL       string
	RefreshTokenURL      string
	ClientID             string
	ClientSecret         string
	Username             string
	Password             string
	Scope                string
	State                string
	ChallengeAlgorithm   string
	CodeVerifier         string
	HeaderPrefix         string
	UseToken             string
	Audience             string
	Resource             string
	AuthRequestParams    []OAuth2Param
	TokenRequestParams   []OAuth2Param
	RefreshRequestParams []OAuth2Param
	AutoRefresh          bool
}

func DefaultOAuth2Config() OAuth2Config {
	return OAuth2Config{
		GrantType:            oauth2grant.AuthorizationCode,
		ClientAuthentication: clientauthentication.Header,
		AddTokenTo:           tokenplacement.Header,
		RedirectURI:          DefaultOAuth2Callback,
		ChallengeAlgorithm:   PkceS256,
		HeaderPrefix:         DefaultHeaderPrefix,
		UseToken:             UseAccessToken,
		AutoRefresh:          true,
		AuthRequestParams:    []OAuth2Param{},
		TokenRequestParams:   []OAuth2Param{},
		RefreshRequestParams: []OAuth2Param{},
	}
}

func (o OAuth2Config) IsBlank() bool {
	return o.AuthURL == "" && o.AccessTokenURL == "" && o.ClientID == "" && o.ClientSecret == "" && o.Scope == "" && o.TokenName == ""
}

func (o OAuth2Config) UsesBrowser() bool {
	return o.UsesAuthorizationCode() || o.IsImplicit()
}

func (o OAuth2Config) UsesAuthorizationCode() bool {
	return o.GrantType == oauth2grant.AuthorizationCode || o.UsesPkce()
}

func (o OAuth2Config) UsesPkce() bool {
	return o.GrantType == oauth2grant.AuthorizationCodeWithPkce
}

func (o OAuth2Config) IsImplicit() bool {
	return o.GrantType == oauth2grant.Implicit
}

func (o OAuth2Config) IsPasswordGrant() bool {
	return o.GrantType == oauth2grant.PasswordCredentials
}

func (o OAuth2Config) IsLegacyGrant() bool {
	return o.IsImplicit() || o.IsPasswordGrant()
}

func (o OAuth2Config) UsesPlainChallenge() bool {
	return o.ChallengeAlgorithm == PkcePlain
}

func (o OAuth2Config) SendsClientInBody() bool {
	return o.ClientAuthentication == clientauthentication.Body
}

func (o OAuth2Config) AddsTokenToQuery() bool {
	return o.AddTokenTo == tokenplacement.QueryParams
}

func (o OAuth2Config) RefreshURL() string {
	if o.RefreshTokenURL != "" {
		return o.RefreshTokenURL
	}
	return o.AccessTokenURL
}

func (o OAuth2Config) AuthorizationHeader(token string) string {
	if o.HeaderPrefix == "" {
		return token
	}
	return o.HeaderPrefix + " " + token
}

func (o OAuth2Config) UsesPostmanCallback() bool {
	parsed, err := url.Parse(o.RedirectURI)
	return err == nil && strings.EqualFold(parsed.Hostname(), postmanCallbackHost)
}

func (o OAuth2Config) CanObtainToken() bool {
	return o.InvalidReason() == ""
}

// InvalidReason explains why the configuration cannot obtain a token; empty when it can.
func (o OAuth2Config) InvalidReason() string {
	switch {
	case o.UsesBrowser() && strings.TrimSpace(o.AuthURL) == "":
		return "the auth URL is required for this grant type"
	case !o.IsImplicit() && strings.TrimSpace(o.AccessTokenURL) == "":
		return "the access token URL is required for this grant type"
	case o.UsesBrowser() && !isLoopbackCallback(o.RedirectURI):
		return "the callback URL must be a loopback address with a port, such as " + DefaultOAuth2Callback
	case o.UsesBrowser() && !isSafeBrowserUrl(o.AuthURL):
		return "the auth URL must use https (or http on a loopback address); Xenon does not open other URLs in the browser"
	default:
		return ""
	}
}

func (o OAuth2Config) MapText(transform func(string) string) OAuth2Config {
	for _, field := range []*string{&o.TokenName, &o.RedirectURI, &o.AuthURL, &o.AccessTokenURL, &o.RefreshTokenURL, &o.ClientID, &o.ClientSecret, &o.Username, &o.Password, &o.Scope, &o.State, &o.CodeVerifier, &o.HeaderPrefix, &o.Audience, &o.Resource} {
		*field = transform(*field)
	}
	o.AuthRequestParams = mapParams(o.AuthRequestParams, transform)
	o.TokenRequestParams = mapParams(o.TokenRequestParams, transform)
	o.RefreshRequestParams = mapParams(o.RefreshRequestParams, transform)
	return o
}

func (o OAuth2Config) TextFields(prefix string) []TextField {
	return []TextField{
		NewTextField(fieldName(prefix, "oauth2", "authUrl"), o.AuthURL),
		NewTextField(fieldName(prefix, "oauth2", "accessTokenUrl"), o.AccessTokenURL),
		NewTextField(fieldName(prefix, "oauth2", "refreshTokenUrl"), o.RefreshTokenURL),
		NewTextField(fieldName(prefix, "oauth2", "clientId"), o.ClientID),
		NewTextField(fieldName(prefix, "oauth2", "clientSecret"), o.ClientSecret),
		NewTextField(fieldName(prefix, "oauth2", "username"), o.Username),
		NewTextField(fieldName(prefix, "oauth2", "password"), o.Password),
		NewTextField(fieldName(prefix, "oauth2", "scope"), o.Scope),
	}
}

func (o OAuth2Config) LogValue() any {
	return map[string]any{"grantType": o.GrantType, "authUrl": o.AuthURL, "accessTokenUrl": o.AccessTokenURL, "clientId": o.ClientID, "scope": o.Scope, "redirectUri": o.RedirectURI, "hasClientSecret": o.ClientSecret != ""}
}

func mapParams(params []OAuth2Param, transform func(string) string) []OAuth2Param {
	if params == nil {
		return nil
	}
	mapped := make([]OAuth2Param, 0, len(params))
	for _, param := range params {
		mapped = append(mapped, param.MapText(transform))
	}
	return mapped
}

func isLoopbackCallback(raw string) bool {
	parsed, err := url.Parse(raw)
	return err == nil && strings.EqualFold(parsed.Scheme, schemeHttp) && isLoopbackHost(parsed.Hostname()) && parsed.Port() != ""
}

func isSafeBrowserUrl(raw string) bool {
	parsed, err := url.Parse(raw)
	if err != nil || parsed.Host == "" {
		return false
	}
	switch strings.ToLower(parsed.Scheme) {
	case schemeHttps:
		return true
	case schemeHttp:
		return isLoopbackHost(parsed.Hostname())
	default:
		return false
	}
}

func isLoopbackHost(host string) bool {
	return strings.EqualFold(host, "localhost") || host == "127.0.0.1" || host == "::1"
}
