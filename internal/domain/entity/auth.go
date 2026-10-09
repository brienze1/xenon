package entity

import (
	"github.com/brienze1/xenon/internal/domain/enums"
	"github.com/brienze1/xenon/internal/domain/enums/authtype"
)

type Auth struct {
	Type   enums.AuthType
	Bearer AuthBearer
	Basic  AuthBasic
}

func DefaultAuth() Auth {
	return Auth{Type: authtype.Inherit}
}

func (a Auth) IsInherited() bool {
	return a.Type == authtype.Inherit
}

func (a Auth) IsSupported() bool {
	switch a.Type {
	case authtype.Inherit, authtype.None, authtype.Bearer, authtype.Basic:
		return true
	default:
		return false
	}
}
