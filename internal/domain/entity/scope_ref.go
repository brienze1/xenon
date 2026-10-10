package entity

import (
	"strings"

	"github.com/brienze1/xenon/internal/domain/enums"
	"github.com/brienze1/xenon/internal/domain/enums/variablescope"
)

const (
	environmentScopePrefix = "env:"
	collectionScopePrefix  = "collection:"
	globalsScopeName       = "globals"
)

type ScopeRef struct {
	Scope enums.VariableScope
	ID    string
}

func NewEnvironmentScopeRef(id string) ScopeRef {
	return ScopeRef{Scope: variablescope.Environment, ID: id}
}

func NewCollectionScopeRef(id string) ScopeRef {
	return ScopeRef{Scope: variablescope.Collection, ID: id}
}

func NewGlobalsScopeRef() ScopeRef {
	return ScopeRef{Scope: variablescope.Globals}
}

func ParseScopeRef(raw string) ScopeRef {
	if id, found := strings.CutPrefix(raw, environmentScopePrefix); found && id != "" {
		return NewEnvironmentScopeRef(id)
	}
	if id, found := strings.CutPrefix(raw, collectionScopePrefix); found && id != "" {
		return NewCollectionScopeRef(id)
	}
	if raw == globalsScopeName {
		return NewGlobalsScopeRef()
	}
	return ScopeRef{Scope: variablescope.Unknown, ID: raw}
}

func (s ScopeRef) IsValid() bool {
	return s.Scope != variablescope.Unknown
}

func (s ScopeRef) String() string {
	switch s.Scope {
	case variablescope.Environment:
		return environmentScopePrefix + s.ID
	case variablescope.Collection:
		return collectionScopePrefix + s.ID
	case variablescope.Globals:
		return globalsScopeName
	default:
		return s.ID
	}
}
