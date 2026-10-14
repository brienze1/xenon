package dto

func NewSaveVariableTable(localId, scope string, rows []VariableEdit) SaveVariableTable {
	return SaveVariableTable{LocalId: localId, Scope: scope, Rows: rows}
}
