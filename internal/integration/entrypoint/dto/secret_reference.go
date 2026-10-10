package dto

import "github.com/brienze1/xenon/internal/domain/entity"

type SecretReference struct {
	LocalId string `json:"local_id" validate:"required,not_blank"`
	Scope   string `json:"scope" validate:"variable_scope"`
	Key     string `json:"key" validate:"required"`
}

func (s *SecretReference) ScopeRef() entity.ScopeRef {
	return entity.ParseScopeRef(s.Scope)
}
