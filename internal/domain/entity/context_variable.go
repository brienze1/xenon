package entity

import "github.com/brienze1/xenon/internal/domain/enums"

type ContextVariable struct {
	Name    string
	Value   string
	Scope   enums.VariableScope
	Secret  bool
	Missing bool
	Locked  bool
}

func (c ContextVariable) IsResolvable() bool {
	return !c.Missing && !c.Locked
}

func (c ContextVariable) IsUnavailableSecret() bool {
	return c.Secret && !c.IsResolvable()
}

func (c ContextVariable) IsAvailableSecret() bool {
	return c.Secret && c.IsResolvable() && c.Value != ""
}

func (c ContextVariable) lockedIf(locked bool) ContextVariable {
	if c.Secret && locked {
		c.Locked = true
		c.Missing = false
	}
	return c
}

func (c ContextVariable) LogValue() any {
	return map[string]any{"name": c.Name, "scope": c.Scope, "secret": c.Secret, "missing": c.Missing, "locked": c.Locked}
}
