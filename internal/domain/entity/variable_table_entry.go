package entity

import "github.com/brienze1/xenon/internal/domain/enums"

type VariableTableEntry struct {
	SecretState        enums.SecretState
	CurrentValue       string
	Row                VariableRow
	HasCurrentOverride bool
	Duplicate          bool
}

func (v VariableTableEntry) HasSurroundingSpaces() bool {
	return v.Row.HasSurroundingSpaces()
}

func (v VariableTableEntry) LogValue() any {
	return map[string]any{"row": v.Row.LogValue(), "secretState": v.SecretState, "override": v.HasCurrentOverride, "duplicate": v.Duplicate}
}
