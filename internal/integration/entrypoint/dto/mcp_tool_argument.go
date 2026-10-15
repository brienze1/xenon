package dto

import (
	"slices"

	"github.com/brienze1/xenon/pkg/jsonschema"
)

type MCPToolArgument struct {
	property    jsonschema.Property
	usedBy      []string
	requiredFor []string
	always      bool
}

func NewMCPToolArgument(property jsonschema.Property, actions []string) MCPToolArgument {
	argument := MCPToolArgument{property: property, usedBy: actions}
	if len(property.UsedBy) > 0 {
		argument.usedBy = intersect(property.UsedBy, actions)
	}
	for _, action := range property.RequiredFor {
		if action == alwaysRequired {
			argument.always = true
			continue
		}
		if slices.Contains(actions, action) {
			argument.requiredFor = append(argument.requiredFor, action)
		}
	}
	return argument
}

func (a MCPToolArgument) mergedWith(other MCPToolArgument) MCPToolArgument {
	a.usedBy = append(a.usedBy, other.usedBy...)
	a.requiredFor = append(a.requiredFor, other.requiredFor...)
	a.always = a.always || other.always
	return a
}
