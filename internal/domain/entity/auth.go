package entity

import (
	"github.com/brienze1/xenon/internal/domain/enums"
	"github.com/brienze1/xenon/internal/domain/enums/authtype"
)

type Auth struct {
	Unsupported AuthUnsupported
	ApiKey      ApiKeyAuth
	Basic       AuthBasic
	Type        enums.AuthType
	Bearer      AuthBearer
	Digest      DigestAuth
	AwsV4       AwsV4Auth
	Attributes  []AuthAttribute
	OAuth2      OAuth2Config
}

func NewUnsupportedAuth(unsupported AuthUnsupported) Auth {
	return Auth{Type: authtype.Unsupported, Unsupported: unsupported}
}

func DefaultAuth() Auth {
	return Auth{Type: authtype.Inherit}
}

func (a Auth) IsInherited() bool {
	return a.Type == authtype.Inherit
}

func (a Auth) Header() (string, bool) {
	switch a.Type {
	case authtype.Bearer:
		return a.Bearer.Header(), true
	case authtype.Basic:
		return a.Basic.Header(), true
	default:
		return "", false
	}
}

func (a Auth) MapText(transform func(string) string) Auth {
	a.Bearer = a.Bearer.MapText(transform)
	a.Basic = a.Basic.MapText(transform)
	a.ApiKey = a.ApiKey.MapText(transform)
	a.OAuth2 = a.OAuth2.MapText(transform)
	a.AwsV4 = a.AwsV4.MapText(transform)
	a.Digest = a.Digest.MapText(transform)
	return a
}

func (a Auth) TextFields(prefix string) []TextField {
	switch a.Type {
	case authtype.Bearer:
		return []TextField{NewTextField(fieldName(prefix, "bearer", "token"), a.Bearer.Token)}
	case authtype.Basic:
		return []TextField{NewTextField(fieldName(prefix, "basic", "username"), a.Basic.Username), NewTextField(fieldName(prefix, "basic", "password"), a.Basic.Password)}
	case authtype.ApiKey:
		return a.ApiKey.TextFields(prefix)
	case authtype.OAuth2:
		return a.OAuth2.TextFields(prefix)
	case authtype.AwsV4:
		return a.AwsV4.TextFields(prefix)
	case authtype.Digest:
		return a.Digest.TextFields(prefix)
	default:
		return []TextField{}
	}
}

func (a Auth) IsUnsupported() bool {
	return a.Type == authtype.Unsupported
}

func (a Auth) HasAttributes() bool {
	return len(a.Attributes) > 0
}

func (a Auth) IsSupported() bool {
	switch a.Type {
	case authtype.Inherit, authtype.None, authtype.Bearer, authtype.Basic, authtype.ApiKey, authtype.OAuth2, authtype.AwsV4, authtype.Digest:
		return true
	default:
		return false
	}
}

func (a Auth) HasBearer() bool {
	return a.Type == authtype.Bearer || !a.Bearer.IsEmpty()
}

func (a Auth) HasBasic() bool {
	return a.Type == authtype.Basic || !a.Basic.IsEmpty()
}

func (a Auth) HasApiKey() bool {
	return a.IsApiKey() && a.ApiKey != (ApiKeyAuth{}) || !a.ApiKey.IsBlank()
}

func (a Auth) HasOAuth2() bool {
	return a.IsOAuth2() && a.OAuth2.GrantType != "" || !a.OAuth2.IsBlank()
}

func (a Auth) HasAwsV4() bool {
	return !a.AwsV4.IsBlank()
}

func (a Auth) HasDigest() bool {
	return a.IsDigest() && a.Digest.Algorithm != "" || !a.Digest.IsBlank()
}

func (a Auth) IsApiKey() bool {
	return a.Type == authtype.ApiKey
}

func (a Auth) IsOAuth2() bool {
	return a.Type == authtype.OAuth2
}

func (a Auth) IsAwsV4() bool {
	return a.Type == authtype.AwsV4
}

func (a Auth) IsDigest() bool {
	return a.Type == authtype.Digest
}

func (a Auth) LogValue() any {
	return map[string]any{"type": a.Type, "bearer": a.Bearer.LogValue(), "basic": a.Basic.LogValue(), "apikey": a.ApiKey.LogValue(), "oauth2": a.OAuth2.LogValue(), "awsv4": a.AwsV4.LogValue(), "digest": a.Digest.LogValue(), "unsupported": a.Unsupported.LogValue()}
}
