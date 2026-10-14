package steps

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

const (
	defaultMcpClientName = "bdd-client"
	mcpCallTimeout       = 30 * time.Second
)

func (t *testUtils) mcpEndpoint() string {
	return "http://127.0.0.1:" + t.vars["mcp_port"] + "/mcp"
}

func (t *testUtils) session(ctx context.Context) (*mcp.ClientSession, error) {
	if session, open := t.mcpSessions[t.mcpClientName]; open {
		return session, nil
	}
	client := mcp.NewClient(&mcp.Implementation{Name: t.mcpClientName, Version: "1.0.0"}, nil)
	session, err := client.Connect(ctx, &mcp.StreamableClientTransport{Endpoint: t.mcpEndpoint(), MaxRetries: -1}, nil)
	if err != nil {
		return nil, fmt.Errorf("connecting an MCP client to %s: %w", t.mcpEndpoint(), err)
	}
	t.mcpSessions[t.mcpClientName] = session
	return session, nil
}

func toolText(result *mcp.CallToolResult) string {
	var builder strings.Builder
	for _, content := range result.Content {
		if text, isText := content.(*mcp.TextContent); isText {
			builder.WriteString(text.Text)
		}
	}
	return builder.String()
}

func wrapToolError(text string) []byte {
	var payload any
	if err := json.Unmarshal([]byte(text), &payload); err != nil {
		payload = map[string]any{"message": text}
	}
	wrapped, _ := json.Marshal(map[string]any{"error": payload})
	return wrapped
}
