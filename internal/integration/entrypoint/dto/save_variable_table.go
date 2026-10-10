package dto

import "github.com/brienze1/xenon/internal/domain/entity"

type SaveVariableTable struct {
	LocalId       string         `json:"local_id" validate:"required,not_blank"`
	Scope         string         `json:"scope" validate:"variable_scope"`
	Rows          []VariableEdit `json:"rows"`
	ConfirmReveal bool           `json:"confirm_reveal"`
}

func (s *SaveVariableTable) ScopeRef() entity.ScopeRef {
	return entity.ParseScopeRef(s.Scope)
}

func (s *SaveVariableTable) ToEntity() []entity.VariableEdit {
	edits := make([]entity.VariableEdit, 0, len(s.Rows))
	for index := range s.Rows {
		edits = append(edits, s.Rows[index].ToEntity())
	}
	return edits
}

func (s SaveVariableTable) LogValue() any {
	rows := make([]any, 0, len(s.Rows))
	for _, row := range s.Rows {
		rows = append(rows, row.LogValue())
	}
	return map[string]any{"local_id": s.LocalId, "scope": s.Scope, "confirm_reveal": s.ConfirmReveal, "rows": rows}
}
