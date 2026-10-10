package entity

import (
	"strings"

	"github.com/brienze1/xenon/internal/domain/enums"
	"github.com/brienze1/xenon/internal/domain/enums/clashresolution"
)

const forbiddenVariableKeyCharacters = "{}\r\n"

type VariableRow struct {
	Postman      SourceFields
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
		return ContextVariable{Name: v.Key, Value: value, Description: v.Description, Scope: scope, Secret: true, Missing: !found}
	}
	value := v.InitialValue
	if override := current[v.Key]; override != "" {
		value = override
	}
	return ContextVariable{Name: v.Key, Value: value, Description: v.Description, Scope: scope}
}

func (v VariableRow) LogValue() any {
	return map[string]any{"key": v.Key, "secret": v.Secret, "enabled": v.Enabled}
}

func HasSecretRows(rows []VariableRow) bool {
	for _, row := range rows {
		if row.Secret {
			return true
		}
	}
	return false
}

func AnyHasSecretRows(rows ...[]VariableRow) bool {
	for _, scopeRows := range rows {
		if HasSecretRows(scopeRows) {
			return true
		}
	}
	return false
}

func UsesSecretRows(environment *Environment, ancestors []Collection, globals []VariableRow) bool {
	rows := [][]VariableRow{globals}
	if environment != nil {
		rows = append(rows, environment.Variables)
	}
	if collection, found := RootCollection(ancestors); found {
		rows = append(rows, collection.OwnVariables())
	}
	return AnyHasSecretRows(rows...)
}

func VariableRowIndex(rows []VariableRow, key string) (int, bool) {
	for index, row := range rows {
		if row.Key == key {
			return index, true
		}
	}
	return -1, false
}

func MergeImportedVariables(existing, imported []VariableRow, resolution enums.ClashResolution) ([]VariableRow, []VariableRow) {
	merged := append([]VariableRow{}, existing...)
	kept := make([]VariableRow, 0)
	for _, row := range imported {
		index, clashes := VariableRowIndex(merged, row.Key)
		switch {
		case !clashes:
			merged = append(merged, row)
		case resolution == clashresolution.Replace:
			merged[index] = row
		case resolution == clashresolution.KeepBoth:
			kept = append(kept, row)
		}
	}
	return merged, kept
}
