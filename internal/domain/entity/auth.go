package entity

import (
	"encoding/base64"

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

func (a Auth) Header() (string, bool) {
	switch a.Type {
	case authtype.Bearer:
		return "Bearer " + a.Bearer.Token, true
	case authtype.Basic:
		return "Basic " + base64.StdEncoding.EncodeToString([]byte(a.Basic.Username+":"+a.Basic.Password)), true
	default:
		return "", false
	}
}

func (a Auth) IsSupported() bool {
	switch a.Type {
	case authtype.Inherit, authtype.None, authtype.Bearer, authtype.Basic:
		return true
	default:
		return false
	}
}
