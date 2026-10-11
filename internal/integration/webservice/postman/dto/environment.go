package dto

import (
	"encoding/json"

	"github.com/brienze1/xenon/internal/domain/entity"
)

const (
	globalsScope      = "globals"
	environmentScope  = "environment"
	globalsNameSuffix = " Globals"
)

var (
	environmentKeys = []string{"name", "values"}
	globalsKeys     = []string{"values"}
)

type Environment struct {
	Raw    map[string]any     `json:"-"`
	Name   string             `json:"name"`
	Scope  string             `json:"_postman_variable_scope"`
	Values []EnvironmentValue `json:"values"`
}

func (e *Environment) UnmarshalJSON(data []byte) error {
	type plain Environment
	if err := json.Unmarshal(data, (*plain)(e)); err != nil {
		return err
	}
	e.Raw = decodeRaw(data)
	return nil
}

func (e *Environment) ToEntity(newId func() string) entity.ImportVariables {
	environment := entity.NewEnvironment(newId(), e.Name)
	secrets := map[string]string{}
	for index := range e.Values {
		value := &e.Values[index]
		row := value.ToEntity()
		environment.Variables = append(environment.Variables, row)
		if secret := text(value.Value); value.IsSecret() && secret != "" {
			secrets[row.Key] = secret
		}
	}
	if passthrough := e.Passthrough(); len(passthrough) > 0 {
		environment.Postman = passthrough
	}
	return entity.NewImportVariables(environment, secrets)
}

func (e *Environment) Passthrough() map[string]any {
	if e.Scope == globalsScope {
		return without(e.Raw, globalsKeys...)
	}
	return without(e.Raw, environmentKeys...)
}

func NewEnvironment(environment entity.Environment, exportedAt, exportedUsing string) Environment {
	passthrough := map[string]any(environment.Postman)
	return newVariablesDocument(environment.ID, environment.Name, environmentScope, environment.Variables, passthrough, exportedAt, exportedUsing)
}

func NewGlobals(workspace entity.Workspace, exportedAt, exportedUsing string) Environment {
	passthrough := asMap(workspace.Postman[globalsScope])
	name := workspace.Name + globalsNameSuffix
	if stored, found := passthrough["name"]; found {
		name = text(stored)
	}
	return newVariablesDocument(workspace.ID, name, globalsScope, workspace.Variables, passthrough, exportedAt, exportedUsing)
}

func newVariablesDocument(id, name, scope string, rows []entity.VariableRow, passthrough map[string]any, exportedAt, exportedUsing string) Environment {
	values := make([]EnvironmentValue, 0, len(rows))
	for _, row := range rows {
		values = append(values, NewEnvironmentValue(row))
	}
	fields := without(passthrough, "name", "values", "_postman_variable_scope", "_postman_exported_at", "_postman_exported_using")
	if _, found := fields["id"]; !found {
		fields["id"] = id
	}
	fields["_postman_exported_at"] = exportedAt
	fields["_postman_exported_using"] = exportedUsing
	return Environment{Name: name, Scope: scope, Values: values, Raw: fields}
}

func (e Environment) MarshalJSON() ([]byte, error) {
	document := newObject().set("id", e.Raw["id"]).set("name", e.Name).set("values", e.Values).
		set("_postman_variable_scope", e.Scope).
		set("_postman_exported_at", e.Raw["_postman_exported_at"]).
		set("_postman_exported_using", e.Raw["_postman_exported_using"])
	return document.merged(e.Raw).MarshalJSON()
}
