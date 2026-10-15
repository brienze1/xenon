package dto

import "reflect"

// MCPToolArguments registers the argument DTO that the given actions of a tool decode.
type MCPToolArguments struct {
	Shape   reflect.Type
	Actions []string
}

func NewMCPToolArguments[T any](actions ...string) MCPToolArguments {
	return MCPToolArguments{Actions: actions, Shape: reflect.TypeFor[T]()}
}
