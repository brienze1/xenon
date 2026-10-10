package entity

import (
	"github.com/brienze1/xenon/internal/domain/enums"
	"github.com/brienze1/xenon/internal/domain/enums/authtype"
)

type Auth struct {
	Unsupported AuthUnsupported
	Type        enums.AuthType
	Bearer      AuthBearer
	Basic       AuthBasic
	Attributes  []AuthAttribute
}

func NewUnsupportedAuth(unsupported AuthUnsupported) Auth {
	return Auth{Type: authtype.Unsupported, Unsupported: unsupported}
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

func (a Auth) TextFields(prefix string) []TextField {
	switch a.Type {
	case authtype.Bearer:
		return []TextField{NewTextField(fieldName(prefix, "bearer", "token"), a.Bearer.Token)}
	case authtype.Basic:
		return []TextField{NewTextField(fieldName(prefix, "basic", "username"), a.Basic.Username), NewTextField(fieldName(prefix, "basic", "password"), a.Basic.Password)}
	default:
		return []TextField{}
	}
}

func (a Auth) IsUnsupported() bool {
	return a.Type == authtype.Unsupported
}

func (a Auth) HasAttributes() bool {
	return len(a.Attributes) > 0
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
	return map[string]any{"type": a.Type, "bearer": a.Bearer.LogValue(), "basic": a.Basic.LogValue(), "unsupported": a.Unsupported.LogValue()}
}
