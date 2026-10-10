package dto

import "github.com/brienze1/xenon/internal/domain/entity"

type AddVariable struct {
	LocalId     string `json:"local_id" validate:"required,not_blank"`
	Scope       string `json:"scope" validate:"variable_scope"`
	Key         string `json:"key" validate:"required"`
	Value       string `json:"value"`
	Description string `json:"description"`
	Secret      bool   `json:"secret"`
}

func (a *AddVariable) ScopeRef() entity.ScopeRef {
	return entity.ParseScopeRef(a.Scope)
}

func (a *AddVariable) ToEntity() entity.VariableRow {
	row := entity.NewVariableRow(a.Key, "")
	row.Secret = a.Secret
	row.Description = a.Description
	return row
}

func (a AddVariable) LogValue() any {
	return map[string]any{"local_id": a.LocalId, "scope": a.Scope, "key": a.Key, "secret": a.Secret}
}
