package dto

import "github.com/brienze1/xenon/internal/domain/entity"

type VariableKeys struct {
	LocalId string   `json:"local_id" validate:"required,not_blank"`
	Scope   string   `json:"scope" validate:"variable_scope"`
	Keys    []string `json:"keys"`
}

func (v *VariableKeys) ScopeRef() entity.ScopeRef {
	return entity.ParseScopeRef(v.Scope)
}
