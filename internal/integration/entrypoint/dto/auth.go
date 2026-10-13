package dto

import (
	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/internal/domain/enums/authtype"
)

type Auth struct {
	Unsupported *AuthUnsupported `json:"unsupported,omitempty"`
	ApiKey      *AuthApiKey      `json:"apikey,omitempty"`
	OAuth2      *AuthOAuth2      `json:"oauth2,omitempty"`
	AwsV4       *AuthAwsV4       `json:"awsv4,omitempty"`
	Digest      *AuthDigest      `json:"digest,omitempty"`
	Type        string           `json:"type" validate:"omitempty,oneof=inherit none bearer basic apikey oauth2 awsv4 digest unsupported"`
	Bearer      AuthBearer       `json:"bearer"`
	Basic       AuthBasic        `json:"basic"`
}

func NewAuth(auth entity.Auth) Auth {
	result := Auth{Type: auth.Type.String(), Bearer: NewAuthBearer(auth.Bearer), Basic: NewAuthBasic(auth.Basic), Unsupported: NewAuthUnsupported(auth.Unsupported)}
	if auth.HasApiKey() {
		result.ApiKey = NewAuthApiKey(auth.ApiKey)
	}
	if auth.HasOAuth2() {
		result.OAuth2 = NewAuthOAuth2(auth.OAuth2)
	}
	if auth.HasAwsV4() {
		result.AwsV4 = NewAuthAwsV4(auth.AwsV4)
	}
	if auth.HasDigest() {
		result.Digest = NewAuthDigest(auth.Digest)
	}
	return result
}

func (a *Auth) ToEntity() entity.Auth {
	auth := entity.DefaultAuth()
	if a.Type != "" {
		auth.Type = authtype.GetAuthType(a.Type)
	}
	auth.Bearer = a.Bearer.ToEntity()
	auth.Basic = a.Basic.ToEntity()
	auth.ApiKey = a.ApiKey.ToEntity()
	auth.OAuth2 = a.OAuth2.ToEntity()
	auth.AwsV4 = a.AwsV4.ToEntity()
	auth.Digest = a.Digest.ToEntity()
	return auth
}
