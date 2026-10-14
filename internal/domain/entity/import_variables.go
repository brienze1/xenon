package entity

import "github.com/brienze1/xenon/internal/domain/enums"

type ImportVariables struct {
	Secrets     map[string]string
	Warnings    []ImportWarning
	Environment Environment
}

func NewImportVariables(environment Environment, secrets map[string]string) ImportVariables {
	return ImportVariables{Environment: environment, Secrets: copyValues(secrets)}
}

func (i ImportVariables) WithWarnings(warnings []ImportWarning) ImportVariables {
	i.Warnings = warnings
	return i
}

func (i ImportVariables) ImportItem(key string, kind enums.ImportItemKind) ImportItem {
	return NewVariablesImportItem(key, kind, i).WithWarnings(i.Warnings)
}

func (i ImportVariables) Rows() []VariableRow {
	return i.Environment.Variables
}

func (i ImportVariables) SecretOf(row VariableRow) (string, bool) {
	value, found := i.Secrets[row.Key]
	return value, row.Secret && found
}

func (i ImportVariables) SecretValues() map[string]string {
	secrets := map[string]string{}
	for _, row := range i.Rows() {
		if value, found := i.SecretOf(row); found {
			secrets[row.Key] = value
		}
	}
	return secrets
}

func (i ImportVariables) LogValue() any {
	return map[string]any{"name": i.Environment.Name, "variables": len(i.Environment.Variables)}
}
