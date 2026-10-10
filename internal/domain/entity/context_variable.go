package entity

import "github.com/brienze1/xenon/internal/domain/enums"

type ContextVariable struct {
	Name    string
	Value   string
	Scope   enums.VariableScope
	Secret  bool
	Missing bool
}

func (c ContextVariable) IsResolvable() bool {
	return !c.Missing
}

func (c ContextVariable) LogValue() any {
	return map[string]any{"name": c.Name, "scope": c.Scope, "secret": c.Secret, "missing": c.Missing}
}
