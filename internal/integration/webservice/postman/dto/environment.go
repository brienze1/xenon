package dto

import (
	"encoding/json"

	"github.com/brienze1/xenon/internal/domain/entity"
)

const globalsScope = "globals"

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
