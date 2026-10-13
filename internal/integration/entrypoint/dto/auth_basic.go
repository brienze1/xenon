package dto

import "github.com/brienze1/xenon/internal/domain/entity"

type AuthBasic struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func NewAuthBasic(basic entity.AuthBasic) AuthBasic {
	return AuthBasic{Username: basic.Username, Password: basic.Password}
}

func (a *AuthBasic) ToEntity() entity.AuthBasic {
	return entity.AuthBasic{Username: a.Username, Password: a.Password}
}
