package entity

import (
	"github.com/brienze1/xenon/internal/domain/enums/authtype"
	"github.com/brienze1/xenon/internal/domain/enums/importorigin"
)

func (a Auth) CredentialFields() []TextField {
	switch a.Type {
	case authtype.Bearer:
		return []TextField{NewTextField("bearer.token", a.Bearer.Token)}
	case authtype.Basic:
		return []TextField{NewTextField("basic.password", a.Basic.Password)}
	case authtype.ApiKey:
		return []TextField{NewTextField("apikey.value", a.ApiKey.Value)}
	case authtype.OAuth2:
		return []TextField{NewTextField("oauth2.clientSecret", a.OAuth2.ClientSecret), NewTextField("oauth2.password", a.OAuth2.Password)}
	case authtype.AwsV4:
		return []TextField{NewTextField("awsv4.secretKey", a.AwsV4.SecretKey), NewTextField("awsv4.sessionToken", a.AwsV4.SessionToken)}
	case authtype.Digest:
		return []TextField{NewTextField("digest.password", a.Digest.Password)}
	default:
		return []TextField{}
	}
}

func (a Auth) IsBrunoOnly() bool {
	return a.IsUnsupported() && a.Unsupported.Origin == importorigin.Bruno
}

func (a Auth) HasXenonOnlyOAuth2Fields() bool {
	defaults := DefaultOAuth2Config()
	return a.IsOAuth2() && (a.OAuth2.UseToken != defaults.UseToken || a.OAuth2.AutoRefresh != defaults.AutoRefresh)
}
