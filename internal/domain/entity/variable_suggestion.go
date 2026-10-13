package entity

import (
	"github.com/brienze1/xenon/internal/domain/enums"
	"github.com/brienze1/xenon/internal/domain/enums/variablescope"
)

type VariableSuggestion struct {
	Name        string
	Description string
	Scope       enums.VariableScope
	Secret      bool
	Dynamic     bool
}

func NewVariableSuggestion(variable ContextVariable) VariableSuggestion {
	return VariableSuggestion{Name: variable.Name, Description: variable.Description, Scope: variable.Scope, Secret: variable.Secret}
}

func NewDynamicVariableSuggestion(name string) VariableSuggestion {
	return VariableSuggestion{Name: name, Scope: variablescope.Dynamic, Dynamic: true}
}
