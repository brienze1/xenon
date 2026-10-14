package middleware

import (
	"github.com/brienze1/xenon/internal/integration/adapter"
	"github.com/brienze1/xenon/internal/integration/entrypoint/dto"
)

const (
	sessionIdHeader = "Mcp-Session-Id"
	deleteMethod    = "DELETE"
)

type mcpSessionClosing struct {
	call adapter.MCPCallController
}

// NewMCPSessionClosingMiddleware closes the MCP session named by Mcp-Session-Id when the client sends DELETE.
func NewMCPSessionClosingMiddleware(call adapter.MCPCallController) adapter.Middleware {
	return &mcpSessionClosing{call: call}
}

func (m *mcpSessionClosing) Handle(ctx adapter.MiddlewareContext) error {
	if sessionId, _ := ctx.Header(sessionIdHeader); ctx.Method() == deleteMethod && sessionId != "" {
		_ = m.call.CloseSession(ctx.Context(), dto.NewMCPSession(sessionId))
	}
	return ctx.Next()
}
