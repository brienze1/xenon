package dto

import "github.com/brienze1/xenon/internal/domain/entity"

type MCPEnvVariable struct {
	Scope       string `json:"scope"`
	Key         string `json:"key"`
	Value       string `json:"value"`
	Description string `json:"description"`
	Secret      bool   `json:"secret"`
}

func (m *MCPEnvVariable) AddVariable(localId string) AddVariable {
	return NewAddVariable(localId, m.Scope, m.Key, m.Value, m.Description, m.Secret)
}

func (m *MCPEnvVariable) SetCurrentValue(localId string) SetCurrentValue {
	return NewSetCurrentValue(localId, m.Scope, m.Key, m.Value)
}

func (m *MCPEnvVariable) VariableKeys(localId string) VariableKeys {
	return NewVariableKeys(localId, m.Scope, []string{m.Key})
}

func (m *MCPEnvVariable) Target() entity.AIActivityTarget {
	return entity.AIActivityTarget{Name: m.Key, Path: m.Scope}
}
