package dto

import "github.com/brienze1/xenon/internal/domain/entity"

type VariableTableReference struct {
	LocalId string `json:"local_id" validate:"required,not_blank"`
	Scope   string `json:"scope" validate:"variable_scope"`
}

func (v *VariableTableReference) ScopeRef() entity.ScopeRef {
	return entity.ParseScopeRef(v.Scope)
}
