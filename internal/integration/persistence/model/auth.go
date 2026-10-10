package model

import (
	"go.yaml.in/yaml/v3"

	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/internal/domain/enums/authtype"
	"github.com/brienze1/xenon/pkg/yamlnode"
)

type Auth struct {
	Bearer      *AuthBearer      `yaml:"bearer" json:"bearer,omitempty"`
	Basic       *AuthBasic       `yaml:"basic" json:"basic,omitempty"`
	Unsupported *AuthUnsupported `yaml:"unsupported" json:"-"`
	Type        string           `yaml:"type" json:"type"`
	ApiKey      yaml.Node        `yaml:"apikey" json:"-"`
	OAuth2      yaml.Node        `yaml:"oauth2" json:"-"`
	AwsV4       yaml.Node        `yaml:"awsv4" json:"-"`
	Digest      yaml.Node        `yaml:"digest" json:"-"`
}

func NewAuth(auth entity.Auth) Auth {
	model := Auth{Type: auth.Type.String()}
	if auth.HasAttributes() {
		*model.attributeSlot(auth.Type.String()) = *authAttributesNode(auth.Attributes)
	}
	if auth.HasBearer() {
		bearer := NewAuthBearer(auth.Bearer)
		model.Bearer = &bearer
	}
	if auth.HasBasic() {
		basic := NewAuthBasic(auth.Basic)
		model.Basic = &basic
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
	if a.Unsupported != nil {
		auth.Unsupported = a.Unsupported.ToEntity()
	}
	if slot := a.attributeSlot(auth.Type.String()); slot != nil {
		auth.Attributes = authAttributesToEntity(slot)
	}
	return auth
}

func (a *Auth) attributeSlot(authType string) *yaml.Node {
	slots := map[string]*yaml.Node{
		authtype.ApiKey.String(): &a.ApiKey,
		authtype.OAuth2.String(): &a.OAuth2,
		authtype.AwsV4.String():  &a.AwsV4,
		authtype.Digest.String(): &a.Digest,
	}
	return slots[authType]
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
	if auth.HasAttributes() {
		fields = append(fields, yamlnode.Reserved(auth.Type.String(), authAttributesNode(auth.Attributes)))
	}
	if a.Unsupported != nil {
		fields = append(fields, yamlnode.Reserved("unsupported", a.Unsupported.Node()))
	}
	return yamlnode.Object("auth", fields...)
}
