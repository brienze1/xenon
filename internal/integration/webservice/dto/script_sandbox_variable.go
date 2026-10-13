package dto

import "github.com/brienze1/xenon/internal/domain/entity"

type ScriptSandboxVariable struct {
	Value    any    `json:"value"`
	Key      string `json:"key"`
	Type     string `json:"type,omitempty"`
	Secret   bool   `json:"secret,omitempty"`
	Disabled bool   `json:"disabled,omitempty"`
}

func NewScriptSandboxVariable(variable entity.ScriptVariable) ScriptSandboxVariable {
	return ScriptSandboxVariable{Key: variable.Key, Value: variable.Value, Type: variable.Type, Secret: variable.Secret, Disabled: variable.Disabled}
}

func (s ScriptSandboxVariable) ToEntity() entity.ScriptVariable {
	return entity.ScriptVariable{Key: s.Key, Value: s.Value, Type: s.Type, Secret: s.Secret, Disabled: s.Disabled}
}
