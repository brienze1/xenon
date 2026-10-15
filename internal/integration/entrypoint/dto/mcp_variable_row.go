package dto

type MCPVariableRow struct {
	SecretValue *string `json:"secret_value" description:"New value of a secret row; write-only."`
	Enabled     *bool   `json:"enabled" description:"Whether the row is enabled; default true."`
	Key         string  `json:"key" description:"Variable name."`
	Initial     string  `json:"initial" description:"Initial value, shared in the workspace files."`
	Description string  `json:"description" description:"Variable description."`
	Secret      bool    `json:"secret" description:"Whether the row is a secret."`
}

func (m *MCPVariableRow) MisplacesSecretValue() bool {
	return m.SecretValue != nil && !m.Secret
}

func (m *MCPVariableRow) Edit() VariableEdit {
	return NewVariableEdit(m.Key, m.InitialValue(), m.Description, m.Secret, m.Enabled, m.SecretValue)
}

// InitialValue is empty for a secret row; its value only travels as secret_value.
func (m *MCPVariableRow) InitialValue() string {
	if m.Secret {
		return ""
	}
	return m.Initial
}

func HasMisplacedSecretValue(rows []MCPVariableRow) bool {
	for index := range rows {
		if rows[index].MisplacesSecretValue() {
			return true
		}
	}
	return false
}

func NewVariableEdits(rows []MCPVariableRow) []VariableEdit {
	edits := make([]VariableEdit, 0, len(rows))
	for index := range rows {
		edits = append(edits, rows[index].Edit())
	}
	return edits
}
