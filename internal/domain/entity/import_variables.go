package entity

type ImportVariables struct {
	Secrets     map[string]string
	Environment Environment
}

func NewImportVariables(environment Environment, secrets map[string]string) ImportVariables {
	return ImportVariables{Environment: environment, Secrets: copyValues(secrets)}
}

func (i ImportVariables) Rows() []VariableRow {
	return i.Environment.Variables
}

func (i ImportVariables) SecretOf(row VariableRow) (string, bool) {
	value, found := i.Secrets[row.Key]
	return value, row.Secret && found
}

func (i ImportVariables) LogValue() any {
	return map[string]any{"name": i.Environment.Name, "variables": len(i.Environment.Variables)}
}
