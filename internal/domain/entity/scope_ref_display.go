package entity

import "github.com/brienze1/xenon/internal/domain/enums/variablescope"

const globalsDisplayName = "Globals"

func (s ScopeRef) IsGlobals() bool {
	return s.Scope == variablescope.Globals
}

// DisplayName is the name of a scope whose environment or collection name is unknown.
func (s ScopeRef) DisplayName() string {
	if s.IsGlobals() {
		return globalsDisplayName
	}
	return s.String()
}

func (s ScopeRef) Target(name string) AIActivityTarget {
	return AIActivityTarget{Name: name, Path: s.String()}
}
