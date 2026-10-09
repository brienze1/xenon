package model

import (
	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/internal/domain/enums/authtype"
	"github.com/brienze1/xenon/pkg/yamlnode"
)

type Auth struct {
	Bearer *AuthBearer `yaml:"bearer" json:"bearer,omitempty"`
	Basic  *AuthBasic  `yaml:"basic" json:"basic,omitempty"`
	Type   string      `yaml:"type" json:"type"`
}

func NewAuth(auth entity.Auth) Auth {
	model := Auth{Type: auth.Type.String()}
	if auth.HasBearer() {
		bearer := NewAuthBearer(auth.Bearer)
		model.Bearer = &bearer
	}
	if auth.HasBasic() {
		basic := NewAuthBasic(auth.Basic)
		model.Basic = &basic
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
	return auth
}

func (a *Auth) Field() yamlnode.Field {
	fields := []yamlnode.Field{yamlnode.String("type", a.ToEntity().Type.String())}
	if a.Bearer != nil {
		fields = append(fields, yamlnode.Object("bearer", yamlnode.String("token", a.Bearer.Token)))
	}
	if a.Basic != nil {
		fields = append(fields, yamlnode.Object("basic",
			yamlnode.String("username", a.Basic.Username),
			yamlnode.String("password", a.Basic.Password),
		))
	}
	return yamlnode.Object("auth", fields...)
}
