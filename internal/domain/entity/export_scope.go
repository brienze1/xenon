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

func (e ExportScope) CurrentValuesAsInitial(local LocalVariables) ([]VariableRow, bool) {
	current := local.CurrentValues(e.Ref)
	rows := make([]VariableRow, 0, len(e.Rows))
	changed := false
	for _, row := range e.Rows {
		applied := row.WithCurrentAsInitial(current)
		changed = changed || applied.InitialValue != row.InitialValue
		rows = append(rows, applied)
	}
	return rows, changed
}

func (e ExportScope) SecretValues(vault Vault) map[string]string {
	stored := vault.SecretsFor(e.Ref)
	values := map[string]string{}
	for _, row := range e.Rows {
		if value := stored[row.Key]; row.Secret && value != "" {
			values[row.Key] = value
		}
	}
	return values
}

func ExportScopesHaveSecrets(scopes []ExportScope) bool {
	for _, scope := range scopes {
		if scope.HasSecrets() {
			return true
		}
	}
	return false
}

func NewExportSecretBundle(scopes []ExportScope, vault Vault) SecretBundle {
	bundle := NewSecretBundle(nil)
	for _, scope := range scopes {
		bundle = bundle.With(scope.Ref, scope.SecretValues(vault))
	}
	return bundle
}
