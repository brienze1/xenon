package dto

import (
	"bytes"
	"encoding/json"

	"github.com/brienze1/xenon/internal/domain/entity"
)

const mcpEmptyArguments = "{}"

type MCPCall struct {
	Session   string          `json:"session"`
	Client    string          `json:"client"`
	Arguments json.RawMessage `json:"arguments"`
}

func NewMCPCall(session, client string, arguments json.RawMessage) MCPCall {
	return MCPCall{Session: session, Client: entity.AIClientName(client), Arguments: arguments}
}

func (m MCPCall) Args() json.RawMessage {
	if len(bytes.TrimSpace(m.Arguments)) == 0 {
		return json.RawMessage(mcpEmptyArguments)
	}
	return m.Arguments
}

// Selector reads the string argument name; raw is its JSON text when it is present but not a string.
func (m MCPCall) Selector(name string) (value, raw string, isObject, isString bool) {
	var arguments map[string]json.RawMessage
	if err := json.Unmarshal(m.Args(), &arguments); err != nil {
		return "", "", false, false
	}
	selector, found := arguments[name]
	if !found {
		return "", "", true, true
	}
	if err := json.Unmarshal(selector, &value); err != nil {
		return "", string(selector), true, false
	}
	return value, string(selector), true, true
}

func (m MCPCall) LogValue() any {
	return map[string]any{"session": m.Session, "client": m.Client, "arguments_length": len(m.Arguments)}
}
