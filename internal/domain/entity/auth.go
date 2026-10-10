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
	return a
}

func (a Auth) IsSupported() bool {
	switch a.Type {
	case authtype.Inherit, authtype.None, authtype.Bearer, authtype.Basic:
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

func (a Auth) LogValue() any {
	return map[string]any{"type": a.Type, "bearer": a.Bearer.LogValue(), "basic": a.Basic.LogValue()}
}
