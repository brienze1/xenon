package entity

type VariableEdit struct {
	SecretValue  *string
	PreviousKey  string
	CurrentValue string
	Row          VariableRow
	SetCurrent   bool
}

func (v VariableEdit) SourceKey() string {
	if v.PreviousKey != "" {
		return v.PreviousKey
	}
	return v.Row.Key
}

func (v VariableEdit) LogValue() any {
	return map[string]any{"row": v.Row.LogValue(), "previousKey": v.PreviousKey, "setCurrent": v.SetCurrent, "setsSecret": v.SecretValue != nil}
}
