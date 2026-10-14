package dto

import (
	"github.com/brienze1/xenon/internal/domain/entity"
)

const variableScopeKey = "_postman_variable_scope"

type Environment struct {
	Extra  map[string]any `yaml:",inline"`
	Name   string         `yaml:"name"`
	Values []Variable     `yaml:"values"`
}

func (e *Environment) ToPostmanV21(scope, fallbackName string, conversion *Conversion) map[string]any {
	document := copied(e.Extra)
	name := e.Name
	if name == "" {
		name = fallbackName
	}
	values := make([]any, 0, len(e.Values))
	for index := range e.Values {
		values = append(values, e.Values[index].toEnvironmentValue(conversion, entity.JoinImportPath(name, e.Values[index].Key)))
	}
	document["name"], document["values"], document[variableScopeKey] = name, values, scope
	return document
}
