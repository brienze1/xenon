package dto

import (
	"bytes"
	"encoding/json"

	"github.com/brienze1/xenon/internal/domain/entity"
	mcpErrors "github.com/brienze1/xenon/internal/integration/entrypoint/errors"
)

type MCPToolCall struct {
	Tool      string          `json:"tool"`
	Action    string          `json:"action"`
	Session   string          `json:"session"`
	Client    string          `json:"client"`
	LocalId   string          `json:"local_id"`
	Arguments json.RawMessage `json:"arguments"`
}

func NewMCPToolCall(call MCPCall, tool, action, localId string) MCPToolCall {
	return MCPToolCall{Tool: tool, Action: action, Session: call.Session, Client: call.Client, LocalId: localId, Arguments: call.Args()}
}

// Decode reads the arguments into target; anything that does not fit is rejected as invalid arguments.
func (m MCPToolCall) Decode(target any) error {
	if err := json.NewDecoder(bytes.NewReader(m.Arguments)).Decode(target); err != nil {
		return mcpErrors.RejectInvalidArguments(m.Tool, err.Error())
	}
	return nil
}

func (m MCPToolCall) TabId() string {
	return entity.AISessionTabID(m.Session)
}

func (m MCPToolCall) LogValue() any {
	return map[string]any{"tool": m.Tool, "action": m.Action, "session": m.Session, "client": m.Client, "local_id": m.LocalId, "arguments_length": len(m.Arguments)}
}
