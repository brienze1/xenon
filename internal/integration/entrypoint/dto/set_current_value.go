package dto

import "github.com/brienze1/xenon/internal/domain/entity"

type SetCurrentValue struct {
	LocalId string `json:"local_id" validate:"required,not_blank"`
	Scope   string `json:"scope" validate:"variable_scope"`
	Key     string `json:"key" validate:"required"`
	Value   string `json:"value"`
}

func (s *SetCurrentValue) ScopeRef() entity.ScopeRef {
	return entity.ParseScopeRef(s.Scope)
}

func (s SetCurrentValue) LogValue() any {
	return map[string]any{"local_id": s.LocalId, "scope": s.Scope, "key": s.Key}
}
