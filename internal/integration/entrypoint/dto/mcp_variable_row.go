package dto

type MCPVariableRow struct {
	SecretValue *string `json:"secret_value"`
	Enabled     *bool   `json:"enabled"`
	Key         string  `json:"key"`
	Initial     string  `json:"initial"`
	Description string  `json:"description"`
	Secret      bool    `json:"secret"`
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
