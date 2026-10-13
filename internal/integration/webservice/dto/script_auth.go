package dto

import (
	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/internal/domain/enums/authtype"
)

const (
	scriptAuthToken    = "token"
	scriptAuthUsername = "username"
	scriptAuthPassword = "password"
)

type ScriptAuth struct {
	Type   string           `json:"type"`
	Bearer []ScriptKeyValue `json:"bearer,omitempty"`
	Basic  []ScriptKeyValue `json:"basic,omitempty"`
}

func NewScriptAuth(auth entity.Auth) *ScriptAuth {
	switch auth.Type {
	case authtype.Bearer:
		return &ScriptAuth{Type: auth.Type.String(), Bearer: NewScriptKeyValues([]entity.KeyValue{entity.NewKeyValue(scriptAuthToken, auth.Bearer.Token)})}
	case authtype.Basic:
		return &ScriptAuth{Type: auth.Type.String(), Basic: NewScriptKeyValues([]entity.KeyValue{entity.NewKeyValue(scriptAuthUsername, auth.Basic.Username), entity.NewKeyValue(scriptAuthPassword, auth.Basic.Password)})}
	case authtype.None:
		return &ScriptAuth{Type: auth.Type.String()}
	default:
		return nil
	}
}

func (s *ScriptAuth) Equals(other *ScriptAuth) bool {
	if other == nil {
		return false
	}
	return s.Type == other.Type && sameKeyValues(s.Bearer, other.Bearer) && sameKeyValues(s.Basic, other.Basic)
}

func (s *ScriptAuth) ToEntity(written entity.Auth) entity.Auth {
	auth := written
	auth.Type = authtype.GetAuthType(s.Type)
	switch auth.Type {
	case authtype.Bearer:
		auth.Bearer = entity.AuthBearer{Token: scriptParameter(s.Bearer, scriptAuthToken)}
	case authtype.Basic:
		auth.Basic = entity.AuthBasic{Username: scriptParameter(s.Basic, scriptAuthUsername), Password: scriptParameter(s.Basic, scriptAuthPassword)}
	case authtype.None:
	default:
		return written
	}
	return auth
}
