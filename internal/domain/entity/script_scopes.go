package entity

import (
	"github.com/brienze1/xenon/internal/domain/enums"
	"github.com/brienze1/xenon/internal/domain/enums/variablescope"
)

type ScriptScopes struct {
	Environment ScriptScope
	Collection  ScriptScope
	Globals     ScriptScope
	Local       ScriptScope
}

func (s ScriptScopes) WithLocal(local ScriptScope) ScriptScopes {
	s.Local = local
	return s
}

func (s ScriptScopes) Scope(kind enums.VariableScope) ScriptScope {
	switch kind {
	case variablescope.Environment:
		return s.Environment
	case variablescope.Collection:
		return s.Collection
	case variablescope.Globals:
		return s.Globals
	default:
		return s.Local
	}
}
