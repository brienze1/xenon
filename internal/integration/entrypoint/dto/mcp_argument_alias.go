package dto

import (
	"slices"
)

type MCPArgumentAlias struct {
	prefix   string
	argument string
	actions  []string
}

func NewMCPArgumentAlias(prefix, argument string, actions []string) MCPArgumentAlias {
	return MCPArgumentAlias{prefix: prefix, argument: argument, actions: actions}
}

func (a MCPArgumentAlias) appliesTo(action string) bool {
	return len(a.actions) == 0 || slices.Contains(a.actions, action)
}
