package entity

import "github.com/brienze1/xenon/internal/domain/enums"

type InspectedScope struct {
	Scope enums.VariableScope
	Ref   ScopeRef
}
