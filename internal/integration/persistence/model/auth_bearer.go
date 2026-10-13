package model

import "github.com/brienze1/xenon/internal/domain/entity"

type AuthBearer struct {
	Token string `yaml:"token" json:"token"`
}

func NewAuthBearer(bearer entity.AuthBearer) AuthBearer {
	return AuthBearer{Token: bearer.Token}
}

func (a *AuthBearer) ToEntity() entity.AuthBearer {
	return entity.AuthBearer{Token: a.Token}
}
