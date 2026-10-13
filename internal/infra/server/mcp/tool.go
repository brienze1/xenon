package mcp

import (
	"bytes"
	"context"
	"encoding/json"

	"github.com/brienze1/xenon/internal/integration/entrypoint/dto"
)

// Handler runs one tool call against the controllers; errors are reported to the client as tool errors.
type Handler func(ctx context.Context, call Call) (Result, error)

type Call struct {
	guard     func(ctx context.Context, text string) string
	Tool      string
	Action    string
	Session   string
	Client    string
	LocalId   string
	Arguments json.RawMessage
}

type Result struct {
	Payload  any
	Target   dto.AIActivityTarget
	Changes  []dto.AIChange
	Recorded bool
}

type tools struct {
	controllers Controllers
}

func newTools(controllers Controllers) *tools {
	return &tools{controllers: controllers}
}

// Decode reads the call arguments into target; anything that does not fit is MCP-07400.
func (c Call) Decode(target any) error {
	decoder := json.NewDecoder(bytes.NewReader(c.arguments()))
	if err := decoder.Decode(target); err != nil {
		return invalidArguments(c.Tool, err.Error())
	}
	return nil
}

// Masked runs the leak guard over a whole text before it is cut or sliced.
func (c Call) Masked(ctx context.Context, text string) string {
	return c.guard(ctx, text)
}

func (c Call) TabId() string {
	return mcpTabPrefix + c.Session
}

func (c Call) arguments() json.RawMessage {
	if len(bytes.TrimSpace(c.Arguments)) == 0 {
		return json.RawMessage(emptyArguments)
	}
	return c.Arguments
}

func (c Call) Unsupported() error {
	return unsupportedAction(c.Tool, c.Action)
}
