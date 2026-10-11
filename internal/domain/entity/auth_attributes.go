package entity

import (
	"fmt"
	"slices"
	"strings"

	"github.com/brienze1/xenon/internal/domain/enums"
	"github.com/brienze1/xenon/internal/domain/enums/authtype"
	"github.com/brienze1/xenon/internal/domain/enums/clientauthentication"
	"github.com/brienze1/xenon/internal/domain/enums/digestalgorithm"
	"github.com/brienze1/xenon/internal/domain/enums/digestqop"
	"github.com/brienze1/xenon/internal/domain/enums/oauth2grant"
	"github.com/brienze1/xenon/internal/domain/enums/tokenplacement"
)

const (
	attributeAccessToken = "accessToken"
	attributeTokenType   = "tokenType"
)

type attributeSetter func(auth *Auth, value any)

var attributeSetters = map[enums.AuthType]map[string]attributeSetter{
	authtype.ApiKey: {
		"key":   func(a *Auth, v any) { a.ApiKey.Key = attributeText(v) },
		"value": func(a *Auth, v any) { a.ApiKey.Value = attributeText(v) },
		"in":    func(a *Auth, v any) { a.ApiKey.In = apiKeyLocation(attributeText(v)) },
	},
	authtype.OAuth2: {
		"grant_type":            func(a *Auth, v any) { a.OAuth2.GrantType = oauth2grant.GetOAuth2Grant(attributeText(v)) },
		"tokenName":             func(a *Auth, v any) { a.OAuth2.TokenName = attributeText(v) },
		"redirect_uri":          func(a *Auth, v any) { a.OAuth2.RedirectURI = attributeText(v) },
		"authUrl":               func(a *Auth, v any) { a.OAuth2.AuthURL = attributeText(v) },
		"accessTokenUrl":        func(a *Auth, v any) { a.OAuth2.AccessTokenURL = attributeText(v) },
		"refreshTokenUrl":       func(a *Auth, v any) { a.OAuth2.RefreshTokenURL = attributeText(v) },
		"clientId":              func(a *Auth, v any) { a.OAuth2.ClientID = attributeText(v) },
		"clientSecret":          func(a *Auth, v any) { a.OAuth2.ClientSecret = attributeText(v) },
		"username":              func(a *Auth, v any) { a.OAuth2.Username = attributeText(v) },
		"password":              func(a *Auth, v any) { a.OAuth2.Password = attributeText(v) },
		"scope":                 func(a *Auth, v any) { a.OAuth2.Scope = attributeText(v) },
		"state":                 func(a *Auth, v any) { a.OAuth2.State = attributeText(v) },
		"challengeAlgorithm":    func(a *Auth, v any) { a.OAuth2.ChallengeAlgorithm = attributeText(v) },
		"code_verifier":         func(a *Auth, v any) { a.OAuth2.CodeVerifier = attributeText(v) },
		"client_authentication": func(a *Auth, v any) { a.OAuth2.ClientAuthentication = clientAuthentication(attributeText(v)) },
		"addTokenTo":            func(a *Auth, v any) { a.OAuth2.AddTokenTo = tokenPlacement(attributeText(v)) },
		"headerPrefix":          func(a *Auth, v any) { a.OAuth2.HeaderPrefix = attributeText(v) },
		"useToken":              func(a *Auth, v any) { a.OAuth2.UseToken = attributeText(v) },
		"autoRefresh":           func(a *Auth, v any) { a.OAuth2.AutoRefresh = attributeBool(v) },
		"authRequestParams":     func(a *Auth, v any) { a.OAuth2.AuthRequestParams = attributeParams(v, OAuth2SendAsUrl) },
		"tokenRequestParams":    func(a *Auth, v any) { a.OAuth2.TokenRequestParams = attributeParams(v, OAuth2SendAsBody) },
		"refreshRequestParams":  func(a *Auth, v any) { a.OAuth2.RefreshRequestParams = attributeParams(v, OAuth2SendAsBody) },
		"audience":              func(a *Auth, v any) { a.OAuth2.Audience = attributeText(v) },
		"resource":              func(a *Auth, v any) { a.OAuth2.Resource = attributeText(v) },
	},
	authtype.AwsV4: {
		"accessKey":          func(a *Auth, v any) { a.AwsV4.AccessKey = attributeText(v) },
		"secretKey":          func(a *Auth, v any) { a.AwsV4.SecretKey = attributeText(v) },
		"sessionToken":       func(a *Auth, v any) { a.AwsV4.SessionToken = attributeText(v) },
		"region":             func(a *Auth, v any) { a.AwsV4.Region = attributeText(v) },
		"service":            func(a *Auth, v any) { a.AwsV4.Service = attributeText(v) },
		"addAuthDataToQuery": func(a *Auth, v any) { a.AwsV4.AddAuthDataToQuery = attributeBool(v) },
	},
	authtype.Digest: {
		"username":            func(a *Auth, v any) { a.Digest.Username = attributeText(v) },
		"password":            func(a *Auth, v any) { a.Digest.Password = attributeText(v) },
		"algorithm":           func(a *Auth, v any) { a.Digest.Algorithm = digestAlgorithm(attributeText(v)) },
		"qop":                 func(a *Auth, v any) { a.Digest.Qop = digestQop(attributeText(v)) },
		"realm":               func(a *Auth, v any) { a.Digest.Realm = attributeText(v) },
		"nonce":               func(a *Auth, v any) { a.Digest.Nonce = attributeText(v) },
		"opaque":              func(a *Auth, v any) { a.Digest.Opaque = attributeText(v) },
		"nonceCount":          func(a *Auth, v any) { a.Digest.NonceCount = attributeText(v) },
		"clientNonce":         func(a *Auth, v any) { a.Digest.ClientNonce = attributeText(v) },
		"disableRetryRequest": func(a *Auth, v any) { a.Digest.DisableRetryRequest = attributeBool(v) },
	},
}

// NewAttributedAuth maps Postman-named attributes onto the typed auth; unknown attributes stay in Attributes.
func NewAttributedAuth(authType enums.AuthType, attributes []AuthAttribute) Auth {
	auth := Auth{Type: authType}
	switch authType {
	case authtype.ApiKey:
		auth.ApiKey = DefaultApiKeyAuth()
	case authtype.OAuth2:
		auth.OAuth2 = DefaultOAuth2Config()
	case authtype.Digest:
		auth.Digest = DefaultDigestAuth()
	}
	return auth.WithAttributes(attributes)
}

func (a Auth) WithAttributes(attributes []AuthAttribute) Auth {
	setters := attributeSetters[a.Type]
	a.Attributes = nil
	for _, attribute := range attributes {
		if setter, known := setters[attribute.Key]; known {
			setter(&a, attribute.Value)
			continue
		}
		a.Attributes = append(a.Attributes, attribute)
	}
	return a
}

// WithoutLiteralAccessToken takes a typed-in OAuth 2.0 access token out of the auth; variable references stay.
func (a Auth) WithoutLiteralAccessToken() (Auth, *OAuth2Token) {
	token := OAuth2Token{TokenType: a.attributeText(attributeTokenType), UseToken: a.OAuth2.UseToken}
	token.AccessToken = a.attributeText(attributeAccessToken)
	if !a.IsOAuth2() || token.AccessToken == "" || len(TokenizeReferences(attributeAccessToken, token.AccessToken)) > 0 {
		return a, nil
	}
	a.Attributes = slices.DeleteFunc(slices.Clone(a.Attributes), func(attribute AuthAttribute) bool {
		return attribute.Key == attributeAccessToken || attribute.Key == attributeTokenType
	})
	return a, &token
}

func (a Auth) attributeText(key string) string {
	for _, attribute := range a.Attributes {
		if attribute.Key == key {
			return attributeText(attribute.Value)
		}
	}
	return ""
}

func attributeText(value any) string {
	switch typed := value.(type) {
	case nil:
		return ""
	case string:
		return typed
	default:
		return fmt.Sprint(typed)
	}
}

func attributeBool(value any) bool {
	switch typed := value.(type) {
	case bool:
		return typed
	default:
		return strings.EqualFold(attributeText(value), "true")
	}
}

func attributeParams(value any, defaultSendAs string) []OAuth2Param {
	params := make([]OAuth2Param, 0)
	items, isList := value.([]any)
	if !isList {
		return params
	}
	for _, item := range items {
		entry, isMap := item.(map[string]any)
		if !isMap {
			continue
		}
		param := OAuth2Param{Key: attributeText(entry["key"]), Value: attributeText(entry["value"]), SendAs: attributeText(entry["send_as"]), Enabled: entry["enabled"] == nil || attributeBool(entry["enabled"])}
		if param.SendAs == "" {
			param.SendAs = defaultSendAs
		}
		params = append(params, param)
	}
	return params
}

func apiKeyLocation(value string) string {
	if strings.EqualFold(value, ApiKeyInQuery) {
		return ApiKeyInQuery
	}
	return ApiKeyInHeader
}

func clientAuthentication(value string) enums.ClientAuthentication {
	if strings.EqualFold(value, clientauthentication.Body.String()) {
		return clientauthentication.Body
	}
	return clientauthentication.Header
}

func tokenPlacement(value string) enums.TokenPlacement {
	if tokenplacement.GetTokenPlacement(value) == tokenplacement.QueryParams {
		return tokenplacement.QueryParams
	}
	return tokenplacement.Header
}

func digestAlgorithm(value string) enums.DigestAlgorithm {
	if algorithm := digestalgorithm.GetDigestAlgorithm(value); algorithm != digestalgorithm.Unknown {
		return algorithm
	}
	return digestalgorithm.MD5
}

func digestQop(value string) enums.DigestQop {
	if qop := digestqop.GetDigestQop(value); qop != digestqop.Unknown {
		return qop
	}
	return digestqop.None
}
