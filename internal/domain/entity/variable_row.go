package entity

import (
	"strings"

	"github.com/brienze1/xenon/internal/domain/enums"
)

const forbiddenVariableKeyCharacters = "{}\r\n"

type VariableRow struct {
	Key          string
	InitialValue string
	Description  string
	Secret       bool
	Enabled      bool
}

func NewVariableRow(key, initialValue string) VariableRow {
	return VariableRow{Key: key, InitialValue: initialValue, Enabled: true}
}

func (v VariableRow) HasValidKey() bool {
	return v.Key != "" && !strings.ContainsAny(v.Key, forbiddenVariableKeyCharacters)
}

func (v VariableRow) HasSurroundingSpaces() bool {
	return v.Key != strings.TrimSpace(v.Key)
}

func (v VariableRow) IsActive() bool {
	return v.Enabled && v.Key != ""
}

func (v VariableRow) HasStraySecretValue() bool {
	return v.Secret && v.InitialValue != ""
}

func (v VariableRow) WithoutSecretValue() VariableRow {
	if v.Secret {
		v.InitialValue = ""
	}
	return v
}

func (v VariableRow) ContextVariable(scope enums.VariableScope, current, secrets map[string]string) ContextVariable {
	if v.Secret {
		value, found := secrets[v.Key]
		return ContextVariable{Name: v.Key, Value: value, Scope: scope, Secret: true, Missing: !found}
	}
	value := v.InitialValue
	if override := current[v.Key]; override != "" {
		value = override
	}
	return ContextVariable{Name: v.Key, Value: value, Scope: scope}
}

func (v VariableRow) LogValue() any {
	return map[string]any{"key": v.Key, "secret": v.Secret, "enabled": v.Enabled}
}
