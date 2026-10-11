package entity

type ExportScope struct {
	Ref  ScopeRef
	Rows []VariableRow
}

func NewExportScope(ref ScopeRef, rows []VariableRow) ExportScope {
	return ExportScope{Ref: ref, Rows: rows}
}

func (e ExportScope) HasSecrets() bool {
	return HasSecretRows(e.Rows)
}

func (e ExportScope) HasCurrentValues(local LocalVariables) bool {
	current := local.CurrentValues(e.Ref)
	for _, row := range e.Rows {
		if !row.Secret && current[row.Key] != "" {
			return true
		}
	}
	return false
}
