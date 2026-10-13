package model

import (
	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/internal/domain/enums/authtype"
	"github.com/brienze1/xenon/pkg/yamlnode"
)

type Auth struct {
	Bearer      *AuthBearer      `yaml:"bearer" json:"bearer,omitempty"`
	Basic       *AuthBasic       `yaml:"basic" json:"basic,omitempty"`
	ApiKey      *AuthApiKey      `yaml:"apikey" json:"apikey,omitempty"`
	OAuth2      *AuthOAuth2      `yaml:"oauth2" json:"oauth2,omitempty"`
	AwsV4       *AuthAwsV4       `yaml:"awsv4" json:"awsv4,omitempty"`
	Digest      *AuthDigest      `yaml:"digest" json:"digest,omitempty"`
	Unsupported *AuthUnsupported `yaml:"unsupported" json:"-"`
	Type        string           `yaml:"type" json:"type"`
	Extra       []yamlnode.Field `yaml:"-" json:"-"`
}

func NewAuth(auth entity.Auth) Auth {
	model := Auth{Type: auth.Type.String(), Extra: authExtraFields(auth.Attributes)}
	if auth.HasBearer() {
		bearer := NewAuthBearer(auth.Bearer)
		model.Bearer = &bearer
	}
	if auth.HasBasic() {
		basic := NewAuthBasic(auth.Basic)
		model.Basic = &basic
	}
	if auth.HasApiKey() {
		apiKey := NewAuthApiKey(auth.ApiKey)
		model.ApiKey = &apiKey
	}
	if auth.HasOAuth2() {
		oauth2 := NewAuthOAuth2(auth.OAuth2)
		model.OAuth2 = &oauth2
	}
	if auth.HasAwsV4() {
		awsV4 := NewAuthAwsV4(auth.AwsV4)
		model.AwsV4 = &awsV4
	}
	if auth.HasDigest() {
		digest := NewAuthDigest(auth.Digest)
		model.Digest = &digest
	}
	if !auth.Unsupported.IsEmpty() {
		unsupported := NewAuthUnsupported(auth.Unsupported)
		model.Unsupported = &unsupported
	}
	return model
}

func (a *Auth) ToEntity() entity.Auth {
	auth := entity.DefaultAuth()
	if a.Type != "" {
		auth.Type = authtype.GetAuthType(a.Type)
	}
	if a.Bearer != nil {
		auth.Bearer = a.Bearer.ToEntity()
	}
	if a.Basic != nil {
		auth.Basic = a.Basic.ToEntity()
	}
	if a.ApiKey != nil {
		auth.ApiKey = a.ApiKey.ToEntity()
	}
	if a.OAuth2 != nil {
		auth.OAuth2 = a.OAuth2.ToEntity()
	}
	if a.AwsV4 != nil {
		auth.AwsV4 = a.AwsV4.ToEntity()
	}
	if a.Digest != nil {
		auth.Digest = a.Digest.ToEntity()
	}
	if a.Unsupported != nil {
		auth.Unsupported = a.Unsupported.ToEntity()
	}
	if len(a.Extra) > 0 {
		auth.Attributes = authAttributesOf(a.Extra)
	}
	return auth
}

func (a *Auth) Field() yamlnode.Field {
	auth := a.ToEntity()
	fields := []yamlnode.Field{yamlnode.String("type", auth.Type.String())}
	if a.Bearer != nil {
		fields = append(fields, yamlnode.Object("bearer", yamlnode.String("token", a.Bearer.Token)))
	}
	if a.Basic != nil {
		fields = append(fields, yamlnode.Object("basic",
			yamlnode.String("username", a.Basic.Username),
			yamlnode.String("password", a.Basic.Password),
		))
	}
	if a.ApiKey != nil {
		fields = append(fields, a.subObject(authtype.ApiKey.String(), a.ApiKey.fields()))
	}
	if a.OAuth2 != nil {
		fields = append(fields, a.subObject(authtype.OAuth2.String(), a.OAuth2.fields()))
	}
	if a.AwsV4 != nil {
		fields = append(fields, a.subObject(authtype.AwsV4.String(), a.AwsV4.fields()))
	}
	if a.Digest != nil {
		fields = append(fields, a.subObject(authtype.Digest.String(), a.Digest.fields()))
	}
	if a.Unsupported != nil {
		fields = append(fields, yamlnode.Reserved("unsupported", a.Unsupported.Node()))
	}
	return yamlnode.Object("auth", fields...)
}

func (a *Auth) subObject(key string, fields []yamlnode.Field) yamlnode.Field {
	if key == a.Type {
		fields = append(fields, a.Extra...)
	}
	return yamlnode.Object(key, fields...)
}
