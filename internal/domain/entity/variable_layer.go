package entity

import "github.com/brienze1/xenon/internal/domain/enums"

type VariableLayer struct {
	Variables map[string]ContextVariable
	Scope     enums.VariableScope
}

// NewVariableLayer skips disabled rows; among duplicate keys the last enabled row wins.
func NewVariableLayer(scope enums.VariableScope, rows []VariableRow, current, secrets map[string]string, locked bool) VariableLayer {
	variables := make(map[string]ContextVariable, len(rows))
	for _, row := range rows {
		if row.IsActive() {
			variables[row.Key] = row.ContextVariable(scope, current, secrets).lockedIf(locked)
		}
	}
	return VariableLayer{Scope: scope, Variables: variables}
}

func (v VariableLayer) Lookup(name string) (ContextVariable, bool) {
	variable, found := v.Variables[name]
	return variable, found
}

func (v VariableLayer) secretValues() []string {
	values := make([]string, 0)
	for _, variable := range v.Variables {
		if variable.IsAvailableSecret() {
			values = append(values, variable.Value)
		}
	}
	return values
}
