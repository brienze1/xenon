package dto

import (
	"reflect"

	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/internal/domain/enums/authtype"
)

type authAttribute struct {
	value     any
	fallback  any
	key       string
	xenonOnly bool
}

func (a authAttribute) isExported(types map[string]any) bool {
	if _, fromPostman := types[a.key]; fromPostman {
		return true
	}
	return !a.xenonOnly && !reflect.DeepEqual(a.value, a.fallback)
}

func typedAttributes(auth entity.Auth) []authAttribute {
	switch auth.Type {
	case authtype.Bearer:
		return []authAttribute{{key: tokenKey, value: auth.Bearer.Token, fallback: ""}}
	case authtype.Basic:
		return []authAttribute{{key: usernameKey, value: auth.Basic.Username, fallback: ""}, {key: passwordKey, value: auth.Basic.Password, fallback: ""}}
	case authtype.ApiKey:
		return []authAttribute{{key: "key", value: auth.ApiKey.Key, fallback: ""}, {key: "value", value: auth.ApiKey.Value, fallback: ""}, {key: "in", value: auth.ApiKey.In, fallback: entity.ApiKeyInHeader}}
	case authtype.AwsV4:
		return awsV4Attributes(auth.AwsV4)
	case authtype.Digest:
		return digestAttributes(auth.Digest)
	case authtype.OAuth2:
		return oauth2Attributes(auth.OAuth2)
	default:
		return []authAttribute{}
	}
}

func awsV4Attributes(aws entity.AwsV4Auth) []authAttribute {
	return []authAttribute{
		{key: "accessKey", value: aws.AccessKey, fallback: ""},
		{key: "secretKey", value: aws.SecretKey, fallback: ""},
		{key: "region", value: aws.Region, fallback: ""},
		{key: "service", value: aws.Service, fallback: ""},
		{key: "sessionToken", value: aws.SessionToken, fallback: ""},
		{key: "addAuthDataToQuery", value: aws.AddAuthDataToQuery, fallback: false},
	}
}

func digestAttributes(digest entity.DigestAuth) []authAttribute {
	defaults := entity.DefaultDigestAuth()
	return []authAttribute{
		{key: usernameKey, value: digest.Username, fallback: ""},
		{key: passwordKey, value: digest.Password, fallback: ""},
		{key: "realm", value: digest.Realm, fallback: ""},
		{key: "nonce", value: digest.Nonce, fallback: ""},
		{key: "algorithm", value: digest.Algorithm.String(), fallback: defaults.Algorithm.String()},
		{key: "qop", value: digest.Qop.String(), fallback: defaults.Qop.String()},
		{key: "nonceCount", value: digest.NonceCount, fallback: ""},
		{key: "clientNonce", value: digest.ClientNonce, fallback: ""},
		{key: "opaque", value: digest.Opaque, fallback: ""},
		{key: "disableRetryRequest", value: digest.DisableRetryRequest, fallback: false},
	}
}

func oauth2Attributes(oauth entity.OAuth2Config) []authAttribute {
	defaults := entity.DefaultOAuth2Config()
	return []authAttribute{
		{key: "grant_type", value: oauth.GrantType.String(), fallback: defaults.GrantType.String()},
		{key: "tokenName", value: oauth.TokenName, fallback: ""},
		{key: "authUrl", value: oauth.AuthURL, fallback: ""},
		{key: "accessTokenUrl", value: oauth.AccessTokenURL, fallback: ""},
		{key: "refreshTokenUrl", value: oauth.RefreshTokenURL, fallback: ""},
		{key: "clientId", value: oauth.ClientID, fallback: ""},
		{key: "clientSecret", value: oauth.ClientSecret, fallback: ""},
		{key: usernameKey, value: oauth.Username, fallback: ""},
		{key: passwordKey, value: oauth.Password, fallback: ""},
		{key: "scope", value: oauth.Scope, fallback: ""},
		{key: "state", value: oauth.State, fallback: ""},
		{key: "redirect_uri", value: oauth.RedirectURI, fallback: defaults.RedirectURI},
		{key: "challengeAlgorithm", value: oauth.ChallengeAlgorithm, fallback: defaults.ChallengeAlgorithm},
		{key: "code_verifier", value: oauth.CodeVerifier, fallback: ""},
		{key: "client_authentication", value: oauth.ClientAuthentication.String(), fallback: defaults.ClientAuthentication.String()},
		{key: "addTokenTo", value: oauth.AddTokenTo.String(), fallback: defaults.AddTokenTo.String()},
		{key: "headerPrefix", value: oauth.HeaderPrefix, fallback: defaults.HeaderPrefix},
		{key: "audience", value: oauth.Audience, fallback: ""},
		{key: "resource", value: oauth.Resource, fallback: ""},
		{key: "authRequestParams", value: oauth2Params(oauth.AuthRequestParams), fallback: []any{}},
		{key: "tokenRequestParams", value: oauth2Params(oauth.TokenRequestParams), fallback: []any{}},
		{key: "refreshRequestParams", value: oauth2Params(oauth.RefreshRequestParams), fallback: []any{}},
		{key: "useToken", value: oauth.UseToken, xenonOnly: true},
		{key: "autoRefresh", value: oauth.AutoRefresh, xenonOnly: true},
	}
}

func oauth2Params(params []entity.OAuth2Param) []any {
	rows := make([]any, 0, len(params))
	for _, param := range params {
		rows = append(rows, newObject().set("key", param.Key).set("value", param.Value).set("enabled", param.Enabled).set("send_as", param.SendAs))
	}
	return rows
}
