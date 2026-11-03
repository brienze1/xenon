package dto

import "github.com/brienze1/xenon/internal/domain/entity"

type VariableTable struct {
	Scope    string        `json:"scope"`
	Revision string        `json:"revision"`
	Rows     []VariableRow `json:"rows"`
	Warnings []string      `json:"warnings"`
}

func NewVariableTable(table entity.VariableTable) VariableTable {
	return VariableTable{Scope: table.Scope.String(), Revision: table.Revision, Rows: NewVariableRows(table.Entries()), Warnings: append([]string{}, table.Warnings...)}
}

func (v VariableTable) LogValue() any {
	return map[string]any{"scope": v.Scope, "revision": v.Revision, "rows": len(v.Rows), "warnings": v.Warnings}
}
