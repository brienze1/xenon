package entity

import (
	"slices"
	"strconv"
	"strings"

	"github.com/brienze1/xenon/internal/domain/enums/authtype"
	"github.com/brienze1/xenon/internal/domain/enums/oauth2grant"
)

const (
	SecurityTypeHttp   = "http"
	SecurityTypeApiKey = "apiKey"
	SecurityTypeOAuth2 = "oauth2"

	OAuth2FlowAuthorizationCode = "authorizationCode"
	OAuth2FlowClientCredentials = "clientCredentials"
	OAuth2FlowPassword          = "password"
	OAuth2FlowImplicit          = "implicit"

	awsV4Description = "AWS Signature Version 4"
)

// ApiSecurityScheme describes how a request authenticates: the type of its auth, never a credential.
type ApiSecurityScheme struct {
	Name             string
	Type             string
	Scheme           string
	In               string
	ParamName        string
	Flow             string
	AuthorizationURL string
	TokenURL         string
	Description      string
	Scopes           []string
}

// NewApiSecurityScheme describes the auth type; false when the auth sends nothing (none or inherit).
func NewApiSecurityScheme(auth Auth) (ApiSecurityScheme, bool) {
	switch auth.Type {
	case authtype.Bearer:
		return ApiSecurityScheme{Name: "bearerAuth", Type: SecurityTypeHttp, Scheme: "bearer"}, true
	case authtype.Basic:
		return ApiSecurityScheme{Name: "basicAuth", Type: SecurityTypeHttp, Scheme: "basic"}, true
	case authtype.Digest:
		return ApiSecurityScheme{Name: "digestAuth", Type: SecurityTypeHttp, Scheme: "digest"}, true
	case authtype.ApiKey:
		return ApiSecurityScheme{Name: "apiKeyAuth", Type: SecurityTypeApiKey, In: securityKeyLocation(auth.ApiKey), ParamName: auth.ApiKey.Key}, true
	case authtype.OAuth2:
		return newOAuth2Scheme(auth.OAuth2), true
	case authtype.AwsV4:
		return ApiSecurityScheme{Name: "awsV4Auth", Type: SecurityTypeApiKey, In: ApiKeyInHeader, ParamName: "Authorization", Description: awsV4Description}, true
	default:
		return ApiSecurityScheme{}, false
	}
}

func (a ApiSecurityScheme) Equal(other ApiSecurityScheme) bool {
	a.Name, other.Name = "", ""
	return a.Type == other.Type && a.Scheme == other.Scheme && a.In == other.In && a.ParamName == other.ParamName &&
		a.Flow == other.Flow && a.AuthorizationURL == other.AuthorizationURL && a.TokenURL == other.TokenURL &&
		a.Description == other.Description && slices.Equal(a.Scopes, other.Scopes)
}

func (a ApiSecurityScheme) named(taken []ApiSecurityScheme) ApiSecurityScheme {
	base := a.Name
	for suffix := 2; slices.ContainsFunc(taken, func(scheme ApiSecurityScheme) bool { return scheme.Name == a.Name }); suffix++ {
		a.Name = base + strconv.Itoa(suffix)
	}
	return a
}

func newOAuth2Scheme(config OAuth2Config) ApiSecurityScheme {
	scheme := ApiSecurityScheme{Name: "oauth2Auth", Type: SecurityTypeOAuth2, Scopes: strings.Fields(config.Scope)}
	switch config.GrantType {
	case oauth2grant.ClientCredentials:
		scheme.Flow, scheme.TokenURL = OAuth2FlowClientCredentials, config.AccessTokenURL
	case oauth2grant.PasswordCredentials:
		scheme.Flow, scheme.TokenURL = OAuth2FlowPassword, config.AccessTokenURL
	case oauth2grant.Implicit:
		scheme.Flow, scheme.AuthorizationURL = OAuth2FlowImplicit, config.AuthURL
	default:
		scheme.Flow, scheme.AuthorizationURL, scheme.TokenURL = OAuth2FlowAuthorizationCode, config.AuthURL, config.AccessTokenURL
	}
	return scheme
}

func securityKeyLocation(apiKey ApiKeyAuth) string {
	if apiKey.IsInQuery() {
		return ApiKeyInQuery
	}
	return ApiKeyInHeader
}
