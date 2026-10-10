package dto

import (
	"encoding/json"

	"github.com/brienze1/xenon/internal/domain/entity"
)

const (
	globalsScope       = "globals"
	secretVariableType = "secret"
)

var (
	environmentKeys      = []string{"name", "values"}
	globalsKeys          = []string{"values"}
	environmentValueKeys = []string{"key", "value", "enabled"}
)

type EnvironmentValue struct {
	Raw     map[string]any `json:"-"`
	Enabled *bool          `json:"enabled"`
	Key     any            `json:"key"`
	Value   any            `json:"value"`
	Type    string         `json:"type"`
}

func (e *EnvironmentValue) UnmarshalJSON(data []byte) error {
	type plain EnvironmentValue
	if err := json.Unmarshal(data, (*plain)(e)); err != nil {
		return err
	}
	e.Raw = decodeRaw(data)
	return nil
}

func (e *EnvironmentValue) IsSecret() bool {
	return e.Type == secretVariableType
}

func (e *EnvironmentValue) ToEntity() entity.VariableRow {
	row := entity.NewVariableRow(text(e.Key), text(e.Value))
	row.Enabled = e.Enabled == nil || *e.Enabled
	row.Secret = e.IsSecret()
	if passthrough := without(e.Raw, environmentValueKeys...); len(passthrough) > 0 {
		row.Postman = passthrough
	}
	return row.WithoutSecretValue()
}

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
