package dto

import (
	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/internal/domain/enums/authtype"
)

type Auth struct {
	Type   string     `json:"type" validate:"omitempty,oneof=inherit none bearer basic apikey oauth2 awsv4 digest"`
	Bearer AuthBearer `json:"bearer"`
	Basic  AuthBasic  `json:"basic"`
}

func NewAuth(auth entity.Auth) Auth {
	return Auth{Type: auth.Type.String(), Bearer: NewAuthBearer(auth.Bearer), Basic: NewAuthBasic(auth.Basic)}
}

func (a *Auth) ToEntity() entity.Auth {
	auth := entity.DefaultAuth()
	if a.Type != "" {
		auth.Type = authtype.GetAuthType(a.Type)
	}
	auth.Bearer = a.Bearer.ToEntity()
	auth.Basic = a.Basic.ToEntity()
	return auth
}
